package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestProcessStatusCache tests the process status cache functionality
func TestProcessStatusCache(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "process_cache_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a process registry with a short TTL for testing
	registry, err := NewProcessRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create process registry: %v", err)
	}

	// Test cache stats
	entryCount, ttl := registry.GetProcessStatusCacheStats()
	if entryCount != 0 {
		t.Errorf("Expected 0 cache entries, got %d", entryCount)
	}
	if ttl != 2*time.Second {
		t.Errorf("Expected TTL of 2 seconds, got %v", ttl)
	}

	// Create a test process record
	testPID := 99999 // Use a PID that likely doesn't exist
	record := &PersistentProcessRecord{
		ID:          "test-process-1",
		ModpackID:   "test-modpack",
		ModpackName: "Test Modpack",
		PID:         testPID,
		Executable:  "/path/to/test",
		WorkingDir:  "/path/to/test",
		StartTime:   time.Now(),
		LastSeen:    time.Now(),
		Status:      ProcessStatusStarting,
	}

	// Add the record to the registry
	err = registry.AddRecord(record)
	if err != nil {
		t.Fatalf("Failed to add process record: %v", err)
	}

	// Validate processes (this should cache the process status)
	err = registry.ValidateProcesses()
	if err != nil {
		t.Fatalf("Failed to validate processes: %v", err)
	}

	// Check cache stats again
	entryCount, _ = registry.GetProcessStatusCacheStats()
	if entryCount != 1 {
		t.Errorf("Expected 1 cache entry after validation, got %d", entryCount)
	}

	// Validate processes again (should use cache)
	err = registry.ValidateProcesses()
	if err != nil {
		t.Fatalf("Failed to validate processes with cache: %v", err)
	}

	// Clear the cache
	registry.ClearProcessStatusCache()
	entryCount, _ = registry.GetProcessStatusCacheStats()
	if entryCount != 0 {
		t.Errorf("Expected 0 cache entries after clear, got %d", entryCount)
	}

	// Remove the record (should also invalidate cache)
	err = registry.AddRecord(record)
	if err != nil {
		t.Fatalf("Failed to re-add process record: %v", err)
	}

	// Validate to populate cache
	err = registry.ValidateProcesses()
	if err != nil {
		t.Fatalf("Failed to validate processes before removal: %v", err)
	}

	// Remove the record
	err = registry.RemoveRecord("test-process-1")
	if err != nil {
		t.Fatalf("Failed to remove process record: %v", err)
	}

	// Check cache stats after removal
	entryCount, _ = registry.GetProcessStatusCacheStats()
	if entryCount != 0 {
		t.Errorf("Expected 0 cache entries after record removal, got %d", entryCount)
	}
}

// TestProcessStatusCacheTTL tests the TTL functionality of the cache
func TestProcessStatusCacheTTL(t *testing.T) {
	// Create a cache with a very short TTL for testing
	cache := NewProcessStatusCache(100 * time.Millisecond)

	// Test cache miss
	_, err := cache.Get(12345)
	if err != nil {
		t.Errorf("Expected no error on cache miss, got: %v", err)
	}

	// Test cache hit
	isRunning, err := cache.Get(12345)
	if err != nil {
		t.Errorf("Expected no error on cache hit, got: %v", err)
	}
	// The value should be the same as before (cached)
	isRunning2, err := cache.Get(12345)
	if err != nil {
		t.Errorf("Expected no error on second cache hit, got: %v", err)
	}
	if isRunning != isRunning2 {
		t.Errorf("Expected same running status from cache, got %v and %v", isRunning, isRunning2)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Test cache expiration - verify that Get still works after TTL
	_, err = cache.Get(12345)
	if err != nil {
		t.Errorf("Expected no error on cache refresh, got: %v", err)
	}

	// Test invalidation
	cache.Invalidate(12345)
	entryCount := len(cache.entries)
	if entryCount != 0 {
		t.Errorf("Expected 0 entries after invalidation, got %d", entryCount)
	}

	// Test cleanup
	cache.Get(12345) // Add entry
	cache.Get(67890) // Add another entry
	time.Sleep(150 * time.Millisecond)
	cache.Cleanup()
	entryCount = len(cache.entries)
	if entryCount != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", entryCount)
	}
}

// TestProcessStatusCacheConcurrency tests thread safety of the cache
func TestProcessStatusCacheConcurrency(t *testing.T) {
	cache := NewProcessStatusCache(1 * time.Second)

	// Start multiple goroutines accessing the cache
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				pid := id*100 + j
				cache.Get(pid)
				if j%10 == 0 {
					cache.Invalidate(pid)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we reach here without panics or deadlocks, the test passes
	t.Log("Concurrency test passed")
}
