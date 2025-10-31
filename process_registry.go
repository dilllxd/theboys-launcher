package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ProcessStatusCacheEntry represents a cached process status entry
type ProcessStatusCacheEntry struct {
	IsRunning   bool
	CachedAt    time.Time
	Error       error // nil if no error occurred
}

// ProcessStatusCache caches process status with TTL to reduce external command executions
type ProcessStatusCache struct {
	entries map[int]*ProcessStatusCacheEntry // key: PID
	mutex   sync.RWMutex
	ttl     time.Duration
}

// NewProcessStatusCache creates a new process status cache with the specified TTL
func NewProcessStatusCache(ttl time.Duration) *ProcessStatusCache {
	return &ProcessStatusCache{
		entries: make(map[int]*ProcessStatusCacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached process status, or executes the check if not cached or expired
func (cache *ProcessStatusCache) Get(pid int) (bool, error) {
	cache.mutex.RLock()
	entry, exists := cache.entries[pid]
	cache.mutex.RUnlock()

	// If entry exists and is not expired, return cached value
	if exists && time.Since(entry.CachedAt) < cache.ttl {
		return entry.IsRunning, entry.Error
	}

	// Cache miss or expired, check process status
	isRunning, err := isProcessRunning(pid)

	// Update cache
	cache.mutex.Lock()
	cache.entries[pid] = &ProcessStatusCacheEntry{
		IsRunning: isRunning,
		CachedAt:  time.Now(),
		Error:     err,
	}
	cache.mutex.Unlock()

	return isRunning, err
}

// Invalidate removes a specific PID from the cache
func (cache *ProcessStatusCache) Invalidate(pid int) {
	cache.mutex.Lock()
	delete(cache.entries, pid)
	cache.mutex.Unlock()
}

// Clear removes all entries from the cache
func (cache *ProcessStatusCache) Clear() {
	cache.mutex.Lock()
	cache.entries = make(map[int]*ProcessStatusCacheEntry)
	cache.mutex.Unlock()
}

// Cleanup removes expired entries from the cache
func (cache *ProcessStatusCache) Cleanup() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	now := time.Now()
	for pid, entry := range cache.entries {
		if now.Sub(entry.CachedAt) > cache.ttl {
			delete(cache.entries, pid)
		}
	}
}

// ProcessStatus represents the current status of a tracked process
type ProcessStatus int

const (
	ProcessStatusUnknown ProcessStatus = iota
	ProcessStatusStarting
	ProcessStatusRunning
	ProcessStatusStopping
	ProcessStatusStopped
	ProcessStatusCrashed
	ProcessStatusOrphaned
)

func (ps ProcessStatus) String() string {
	switch ps {
	case ProcessStatusStarting:
		return "starting"
	case ProcessStatusRunning:
		return "running"
	case ProcessStatusStopping:
		return "stopping"
	case ProcessStatusStopped:
		return "stopped"
	case ProcessStatusCrashed:
		return "crashed"
	case ProcessStatusOrphaned:
		return "orphaned"
	default:
		return "unknown"
	}
}

// ProcessStatusFromJSON is used for JSON unmarshaling
func ProcessStatusFromJSON(s string) ProcessStatus {
	switch s {
	case "starting":
		return ProcessStatusStarting
	case "running":
		return ProcessStatusRunning
	case "stopping":
		return ProcessStatusStopping
	case "stopped":
		return ProcessStatusStopped
	case "crashed":
		return ProcessStatusCrashed
	case "orphaned":
		return ProcessStatusOrphaned
	default:
		return ProcessStatusUnknown
	}
}

// MarshalJSON implements custom JSON marshaling for ProcessStatus
func (ps ProcessStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.String())
}

// UnmarshalJSON implements custom JSON unmarshaling for ProcessStatus
func (ps *ProcessStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ps = ProcessStatusFromJSON(s)
	return nil
}

// PersistentProcessRecord contains information about a process that needs to persist across launcher sessions
type PersistentProcessRecord struct {
	ID               string        `json:"id"`
	ModpackID        string        `json:"modpack_id"`
	ModpackName      string        `json:"modpack_name"`
	PID              int           `json:"pid"`
	Executable       string        `json:"executable"`
	WorkingDir       string        `json:"working_dir"`
	StartTime        time.Time     `json:"start_time"`
	LastSeen         time.Time     `json:"last_seen"`
	Status           ProcessStatus `json:"status"`
	JavaVersion      string        `json:"java_version,omitempty"`
	MinecraftVersion string        `json:"minecraft_version,omitempty"`
	InstanceName     string        `json:"instance_name,omitempty"`
	LauncherPath     string        `json:"launcher_path,omitempty"`
}

// IsExpired checks if a process record is older than the specified duration
func (r *PersistentProcessRecord) IsExpired(duration time.Duration) bool {
	return time.Since(r.LastSeen) > duration
}

// ProcessRegistry manages persistent process records
type ProcessRegistry struct {
	records      map[string]*PersistentProcessRecord
	mutex        sync.RWMutex
	registryPath string
	statusCache  *ProcessStatusCache
}

// NewProcessRegistry creates a new process registry
func NewProcessRegistry(rootDir string) (*ProcessRegistry, error) {
	registryPath := getRegistryPath(rootDir)

	registry := &ProcessRegistry{
		records:      make(map[string]*PersistentProcessRecord),
		registryPath: registryPath,
		// The process status cache uses a 2-second TTL. This value is chosen as a balance between
		// cache freshness (timely detection of process status changes) and minimizing system call
		// overhead (frequent process status checks can be expensive). For most modpack/game
		// processes, status changes (start/stop) are infrequent compared to this interval, so
		// 2 seconds provides responsive updates without excessive polling. Adjust if needed
		// based on observed performance or process lifecycle patterns.
		statusCache:  NewProcessStatusCache(2 * time.Second), // 2-second TTL
	}

	// Load existing records
	if err := registry.Load(); err != nil {
		logf("Warning: Failed to load process registry: %v", err)
		// Continue with empty registry
	}

	return registry, nil
}

// getRegistryPath returns the platform-specific path for the process registry
func getRegistryPath(rootDir string) string {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\TheBoysLauncher\processes.json
		appData := os.Getenv("APPDATA")
		if appData == "" {
			// Fallback to root directory
			configDir = filepath.Join(rootDir, "config")
		} else {
			configDir = filepath.Join(appData, "TheBoysLauncher")
		}
	case "darwin":
		// macOS: ~/Library/Application Support/TheBoysLauncher/processes.json
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to root directory
			configDir = filepath.Join(rootDir, "config")
		} else {
			configDir = filepath.Join(home, "Library", "Application Support", "TheBoysLauncher")
		}
	default: // linux and others
		// Linux: ~/.config/TheBoysLauncher/processes.json
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to root directory
			configDir = filepath.Join(rootDir, "config")
		} else {
			configDir = filepath.Join(home, ".config", "TheBoysLauncher")
		}
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logf("Warning: Failed to create config directory: %v", err)
		// Fallback to root directory
		return filepath.Join(rootDir, "config", "processes.json")
	}

	return filepath.Join(configDir, "processes.json")
}

// Load loads the process registry from disk
func (pr *ProcessRegistry) Load() error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if _, err := os.Stat(pr.registryPath); os.IsNotExist(err) {
		// Registry doesn't exist yet, start with empty
		pr.records = make(map[string]*PersistentProcessRecord)
		return nil
	}

	data, err := os.ReadFile(pr.registryPath)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	var records map[string]*PersistentProcessRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to parse registry file: %w", err)
	}

	pr.records = records
	logf("Loaded %d process records from registry", len(pr.records))
	return nil
}

// Save saves the process registry to disk using atomic writes
func (pr *ProcessRegistry) Save() error {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	// Create temporary file for atomic write
	tempPath := pr.registryPath + ".tmp"

	data, err := json.MarshalIndent(pr.records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Write to temporary file first
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary registry file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, pr.registryPath); err != nil {
		// Clean up temp file if rename fails
		os.Remove(tempPath)
		return fmt.Errorf("failed to atomically save registry: %w", err)
	}

	return nil
}

// AddRecord adds a new process record to the registry
func (pr *ProcessRegistry) AddRecord(record *PersistentProcessRecord) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	pr.records[record.ID] = record
	return pr.Save()
}

// UpdateRecord updates an existing process record
func (pr *ProcessRegistry) UpdateRecord(id string, updateFunc func(*PersistentProcessRecord)) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	record, exists := pr.records[id]
	if !exists {
		return fmt.Errorf("process record with ID %s not found", id)
	}

	updateFunc(record)
	return pr.Save()
}

// RemoveRecord removes a process record from the registry
func (pr *ProcessRegistry) RemoveRecord(id string) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	record, exists := pr.records[id]
	if !exists {
		return fmt.Errorf("process record with ID %s not found", id)
	}

	// Invalidate cache entry for this PID
	pr.statusCache.Invalidate(record.PID)

	delete(pr.records, id)
	return pr.Save()
}

// GetRecord retrieves a process record by ID
func (pr *ProcessRegistry) GetRecord(id string) (*PersistentProcessRecord, error) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	record, exists := pr.records[id]
	if !exists {
		return nil, fmt.Errorf("process record with ID %s not found", id)
	}

	// Return a copy to prevent concurrent modification
	recordCopy := *record
	return &recordCopy, nil
}

// GetAllRecords returns all process records
func (pr *ProcessRegistry) GetAllRecords() []*PersistentProcessRecord {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	records := make([]*PersistentProcessRecord, 0, len(pr.records))
	for _, record := range pr.records {
		// Return copies to prevent concurrent modification
		recordCopy := *record
		records = append(records, &recordCopy)
	}

	return records
}

// GetRecordsByModpackID returns all process records for a specific modpack
func (pr *ProcessRegistry) GetRecordsByModpackID(modpackID string) []*PersistentProcessRecord {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	var records []*PersistentProcessRecord
	for _, record := range pr.records {
		if record.ModpackID == modpackID {
			// Return copies to prevent concurrent modification
			recordCopy := *record
			records = append(records, &recordCopy)
		}
	}

	return records
}

// CleanupExpiredRecords removes records older than the specified duration
func (pr *ProcessRegistry) CleanupExpiredRecords(maxAge time.Duration) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	var toRemove []string
	now := time.Now()

	for id, record := range pr.records {
		if now.Sub(record.LastSeen) > maxAge {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		delete(pr.records, id)
		logf("Removed expired process record: %s", id)
	}

	if len(toRemove) > 0 {
		return pr.Save()
	}

	return nil
}

// ValidateProcesses validates all registered processes and updates their status
func (pr *ProcessRegistry) ValidateProcesses() error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	var toRemove []string
	now := time.Now()

	// Cleanup expired cache entries periodically
	pr.statusCache.Cleanup()

	for id, record := range pr.records {
		// Check if process is still running using the cache
		isRunning, err := pr.statusCache.Get(record.PID)
		if err != nil {
			logf("Warning: Failed to check process %d: %v", record.PID, err)
			// Assume process is not running if we can't check
			isRunning = false
		}

		if isRunning {
			// Process is still running, update last seen time
			record.LastSeen = now
			if record.Status != ProcessStatusRunning {
				record.Status = ProcessStatusRunning
				logf("Process %d (modpack: %s) is running", record.PID, record.ModpackName)
			}
		} else {
			// Process is not running, update status
			switch record.Status {
			case ProcessStatusStarting, ProcessStatusRunning:
				// Process was running but now it's not - it likely stopped or crashed
				if now.Sub(record.StartTime) < time.Minute {
					// If it ran for less than a minute, consider it crashed
					record.Status = ProcessStatusCrashed
				} else {
					record.Status = ProcessStatusStopped
				}
				logf("Process %d (modpack: %s) is %s", record.PID, record.ModpackName, record.Status)
			}

			// Remove old stopped/crashed records (older than 24 hours)
			if now.Sub(record.LastSeen) > 24*time.Hour {
				toRemove = append(toRemove, id)
			}
		}
	}

	// Remove old records
	for _, id := range toRemove {
		delete(pr.records, id)
		logf("Removed old process record: %s", id)
	}

	return pr.Save()
}

// GetRunningProcesses returns all currently running processes
func (pr *ProcessRegistry) GetRunningProcesses() []*PersistentProcessRecord {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	var running []*PersistentProcessRecord
	for _, record := range pr.records {
		if record.Status == ProcessStatusRunning {
			// Return copies to prevent concurrent modification
			recordCopy := *record
			running = append(running, &recordCopy)
		}
	}

	return running
}

// GetProcessByPID finds a process record by PID
func (pr *ProcessRegistry) GetProcessByPID(pid int) (*PersistentProcessRecord, error) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	for _, record := range pr.records {
		if record.PID == pid {
			// Return a copy to prevent concurrent modification
			recordCopy := *record
			return &recordCopy, nil
		}
	}

	return nil, fmt.Errorf("process record with PID %d not found", pid)
}

// UpdateProcessStatus updates the status of a process record
func (pr *ProcessRegistry) UpdateProcessStatus(id string, status ProcessStatus) error {
	return pr.UpdateRecord(id, func(record *PersistentProcessRecord) {
		record.Status = status
		record.LastSeen = time.Now()
	})
}

// UpdateProcessLastSeen updates the last seen time of a process record
func (pr *ProcessRegistry) UpdateProcessLastSeen(id string) error {
	return pr.UpdateRecord(id, func(record *PersistentProcessRecord) {
		record.LastSeen = time.Now()
	})
}

// ClearProcessStatusCache clears the process status cache
func (pr *ProcessRegistry) ClearProcessStatusCache() {
	pr.statusCache.Clear()
}

// GetProcessStatusCacheStats returns statistics about the process status cache
func (pr *ProcessRegistry) GetProcessStatusCacheStats() (entryCount int, ttl time.Duration) {
	pr.statusCache.mutex.RLock()
	defer pr.statusCache.mutex.RUnlock()
	
	entryCount = len(pr.statusCache.entries)
	ttl = pr.statusCache.ttl
	
	return entryCount, ttl
}

// Global registry instance
var globalRegistry *ProcessRegistry
var registryOnce sync.Once

// GetGlobalProcessRegistry returns the global process registry instance
func GetGlobalProcessRegistry(rootDir string) (*ProcessRegistry, error) {
	var err error
	registryOnce.Do(func() {
		globalRegistry, err = NewProcessRegistry(rootDir)
	})
	return globalRegistry, err
}
