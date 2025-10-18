// Package modpack provides modpack management functionality
package modpack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"theboys-launcher/internal/config"
	"theboys-launcher/internal/logging"
)

// Repository represents a modpack repository
type Repository struct {
	config    *config.Config
	logger    *logging.Logger
	cachePath string
}

// NewRepository creates a new modpack repository
func NewRepository(cfg *config.Config, logger *logging.Logger) *Repository {
	cachePath := filepath.Join(cfg.CacheDir, "modpacks")

	// Ensure cache directory exists
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		logger.Error("Failed to create modpack cache directory: %v", err)
	}

	return &Repository{
		config:    cfg,
		logger:    logger,
		cachePath: cachePath,
	}
}

// ModpackSource represents a source for modpacks
type ModpackSource interface {
	GetModpacks() ([]*Modpack, error)
	GetModpack(id string) (*Modpack, error)
	SearchModpacks(query string, limit int) ([]*Modpack, error)
	GetIconURL(modpack *Modpack) string
}

// CurseForgeSource represents CurseForge as a modpack source
type CurseForgeSource struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewCurseForgeSource creates a new CurseForge source
func NewCurseForgeSource(apiKey string) *CurseForgeSource {
	return &CurseForgeSource{
		apiKey:  apiKey,
		baseURL: "https://api.curseforge.com/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetModpacks retrieves modpacks from CurseForge
func (cfs *CurseForgeSource) GetModpacks() ([]*Modpack, error) {
	// This is a placeholder implementation
	// In a real implementation, you would use the CurseForge API
	// For now, return some sample modpacks
	return []*Modpack{
		{
			ID:          "curseforge-all-the-mods-9",
			Name:        "All the Mods 9",
			Description: "A curated collection of the best mods",
			Author:      "All the Mods Team",
			Version:     "0.2.23",
			Type:        ModpackTypeCurseForge,
			Status:      StatusNotInstalled,
			MinecraftVersion: MinecraftVersion{
				ID:          "1.20.1",
				ReleaseType: "release",
				Date:        "2023-06-12",
			},
			ModLoader: ModLoader{
				Type:    "forge",
				Version: "47.2.0",
			},
			IconURL:              "https://media.forgecdn.net/avatars/447/200/637777833848931823.png",
			RequiredMemory:       6144,
			TotalSize:            2500000000, // 2.5GB
			DateCreated:          time.Now().Add(-30 * 24 * time.Hour),
			DateModified:         time.Now().Add(-1 * 24 * time.Hour),
			Tags:                 []string{"technology", "magic", "adventure", "kitchen-sink"},
			Features:             []string{"Quests", "Custom Recipes", "Performance Optimizations"},
			AutoUpdate:           true,
		},
		{
			ID:          "curseforge-skyfactory-5",
			Name:        "SkyFactory 5",
			Description: "The ultimate skyblock experience",
			Author:      "Bacon_Donut",
			Version:     "5.0.2",
			Type:        ModpackTypeCurseForge,
			Status:      StatusNotInstalled,
			MinecraftVersion: MinecraftVersion{
				ID:          "1.19.2",
				ReleaseType: "release",
				Date:        "2022-08-05",
			},
			ModLoader: ModLoader{
				Type:    "forge",
				Version: "43.2.0",
			},
			IconURL:              "https://media.forgecdn.net/avatars/393/819/637735447864384833.png",
			RequiredMemory:       4096,
			TotalSize:            1500000000, // 1.5GB
			DateCreated:          time.Now().Add(-90 * 24 * time.Hour),
			DateModified:         time.Now().Add(-7 * 24 * time.Hour),
			Tags:                 []string{"skyblock", "tech", "magic", "automation"},
			Features:             []string{"Skyblock", "Automation", "Custom World Gen"},
			AutoUpdate:           false,
		},
	}, nil
}

// GetModpack retrieves a specific modpack from CurseForge
func (cfs *CurseForgeSource) GetModpack(id string) (*Modpack, error) {
	modpacks, err := cfs.GetModpacks()
	if err != nil {
		return nil, err
	}

	for _, modpack := range modpacks {
		if modpack.ID == id {
			return modpack, nil
		}
	}

	return nil, fmt.Errorf("modpack not found: %s", id)
}

// SearchModpacks searches for modpacks on CurseForge
func (cfs *CurseForgeSource) SearchModpacks(query string, limit int) ([]*Modpack, error) {
	modpacks, err := cfs.GetModpacks()
	if err != nil {
		return nil, err
	}

	var results []*Modpack
	count := 0

	for _, modpack := range modpacks {
		if count >= limit {
			break
		}

		// Simple search implementation (case-insensitive)
		if containsIgnoreCase(modpack.Name, query) ||
		   containsIgnoreCase(modpack.Description, query) ||
		   containsIgnoreCase(modpack.Author, query) {
			results = append(results, modpack)
			count++
		}
	}

	return results, nil
}

// GetIconURL returns the icon URL for a modpack
func (cfs *CurseForgeSource) GetIconURL(modpack *Modpack) string {
	return modpack.IconURL
}

// LocalSource represents local modpacks
type LocalSource struct {
	modpacksDir string
	logger      *logging.Logger
}

// NewLocalSource creates a new local source
func NewLocalSource(modpacksDir string, logger *logging.Logger) *LocalSource {
	return &LocalSource{
		modpacksDir: modpacksDir,
		logger:      logger,
	}
}

// GetModpacks retrieves local modpacks
func (ls *LocalSource) GetModpacks() ([]*Modpack, error) {
	var modpacks []*Modpack

	// Scan modpacks directory for modpack instances
	entries, err := os.ReadDir(ls.modpacksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return modpacks, nil
		}
		return nil, fmt.Errorf("failed to read modpacks directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modpack, err := ls.loadModpackFromDirectory(filepath.Join(ls.modpacksDir, entry.Name()))
		if err != nil {
			ls.logger.Error("Failed to load modpack from directory %s: %v", entry.Name(), err)
			continue
		}

		if modpack != nil {
			modpacks = append(modpacks, modpack)
		}
	}

	return modpacks, nil
}

// GetModpack retrieves a specific local modpack
func (ls *LocalSource) GetModpack(id string) (*Modpack, error) {
	modpacks, err := ls.GetModpacks()
	if err != nil {
		return nil, err
	}

	for _, modpack := range modpacks {
		if modpack.ID == id {
			return modpack, nil
		}
	}

	return nil, fmt.Errorf("local modpack not found: %s", id)
}

// SearchModpacks searches local modpacks
func (ls *LocalSource) SearchModpacks(query string, limit int) ([]*Modpack, error) {
	modpacks, err := ls.GetModpacks()
	if err != nil {
		return nil, err
	}

	var results []*Modpack
	count := 0

	for _, modpack := range modpacks {
		if count >= limit {
			break
		}

		if containsIgnoreCase(modpack.Name, query) ||
		   containsIgnoreCase(modpack.Description, query) ||
		   containsIgnoreCase(modpack.Author, query) {
			results = append(results, modpack)
			count++
		}
	}

	return results, nil
}

// GetIconURL returns the icon path for a local modpack
func (ls *LocalSource) GetIconURL(modpack *Modpack) string {
	// Check for local icon file
	iconPath := filepath.Join(modpack.InstallPath, "icon.png")
	if _, err := os.Stat(iconPath); err == nil {
		return "file://" + iconPath
	}

	return ""
}

// loadModpackFromDirectory loads a modpack from a directory
func (ls *LocalSource) loadModpackFromDirectory(dirPath string) (*Modpack, error) {
	// Look for modpack.json file
	configPath := filepath.Join(dirPath, "modpack.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try to infer modpack from directory structure
		return ls.inferModpackFromDirectory(dirPath)
	}

	// Load modpack configuration
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read modpack.json: %w", err)
	}

	var modpack Modpack
	if err := json.Unmarshal(data, &modpack); err != nil {
		return nil, fmt.Errorf("failed to parse modpack.json: %w", err)
	}

	// Update status and paths
	modpack.Status = StatusInstalled
	modpack.InstallPath = dirPath
	modpack.Type = ModpackTypeLocal

	// Get directory info
	info, err := os.Stat(dirPath)
	if err == nil {
		modpack.DateModified = info.ModTime()
	}

	return &modpack, nil
}

// inferModpackFromDirectory tries to infer modpack information from directory structure
func (ls *LocalSource) inferModpackFromDirectory(dirPath string) (*Modpack, error) {
	// Check for common indicators of a modpack
	hasMods := false
	hasConfig := false
	hasJar := false

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if name == "mods" || name == "resourcepacks" || name == "shaderpacks" {
				hasMods = true
			} else if name == "config" {
				hasConfig = true
			}
		} else {
			if filepath.Ext(name) == ".jar" {
				hasJar = true
			}
		}
	}

	// Only consider it a modpack if it has some indicators
	if !hasMods && !hasConfig && !hasJar {
		return nil, nil
	}

	dirName := filepath.Base(dirPath)

	return &Modpack{
		ID:          fmt.Sprintf("local-%s", sanitizeID(dirName)),
		Name:        dirName,
		Description: fmt.Sprintf("Local modpack found in %s", dirName),
		Author:      "Local",
		Version:     "Unknown",
		Type:        ModpackTypeLocal,
		Status:      StatusInstalled,
		InstallPath: dirPath,
		MinecraftVersion: MinecraftVersion{
			ID:          "Unknown",
			ReleaseType: "release",
		},
		ModLoader: ModLoader{
			Type:    "unknown",
			Version: "unknown",
		},
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}, nil
}

// GetCachedModpacks retrieves modpacks from cache
func (r *Repository) GetCachedModpacks() ([]*Modpack, error) {
	cacheFile := filepath.Join(r.cachePath, "modpacks.json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Modpack{}, nil
		}
		return nil, err
	}

	var cached struct {
		Modpacks []*Modpack `json:"modpacks"`
		Updated  time.Time  `json:"updated"`
	}

	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	// Check if cache is stale (older than 24 hours)
	if time.Since(cached.Updated) > 24*time.Hour {
		r.logger.Info("Modpack cache is stale, refreshing")
		return nil, fmt.Errorf("cache expired")
	}

	return cached.Modpacks, nil
}

// CacheModpacks saves modpacks to cache
func (r *Repository) CacheModpacks(modpacks []*Modpack) error {
	cacheFile := filepath.Join(r.cachePath, "modpacks.json")

	cacheData := struct {
		Modpacks []*Modpack `json:"modpacks"`
		Updated  time.Time  `json:"updated"`
	}{
		Modpacks: modpacks,
		Updated:  time.Now(),
	}

	data, err := json.MarshalIndent(cacheData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// GetAllModpacks retrieves all modpacks from all sources
func (r *Repository) GetAllModpacks() ([]*Modpack, error) {
	var allModpacks []*Modpack

	// Get local modpacks
	localSource := NewLocalSource(filepath.Join(r.config.LauncherDir, "instances"), r.logger)
	localModpacks, err := localSource.GetModpacks()
	if err != nil {
		r.logger.Error("Failed to get local modpacks: %v", err)
	} else {
		allModpacks = append(allModpacks, localModpacks...)
	}

	// Get remote modpacks (try cache first)
	cachedModpacks, err := r.GetCachedModpacks()
	if err != nil {
		r.logger.Info("Refreshing modpack cache")

		// Get from CurseForge (no API key required for basic access)
		cfSource := NewCurseForgeSource("")
		remoteModpacks, err := cfSource.GetModpacks()
		if err != nil {
			r.logger.Error("Failed to get remote modpacks: %v", err)
		} else {
			allModpacks = append(allModpacks, remoteModpacks...)

			// Cache the results
			if err := r.CacheModpacks(remoteModpacks); err != nil {
				r.logger.Error("Failed to cache modpacks: %v", err)
			}
		}
	} else {
		allModpacks = append(allModpacks, cachedModpacks...)
	}

	return allModpacks, nil
}

// Helper functions
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func sanitizeID(s string) string {
	// Simple ID sanitization
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove invalid characters
	valid := regexp.MustCompile(`^[a-z0-9-]+$`)
	for !valid.MatchString(s) {
		s = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(s, "")
		if s == "" {
			return "unknown"
		}
	}

	return s
}