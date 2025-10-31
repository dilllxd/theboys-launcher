package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"theboyslauncher/process_registry"
)

func TestProcessRegistry(t *testing.T) {
	fmt.Println("Testing Process Registry Functionality")
	fmt.Println("=====================================")

	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "theboyslauncher_test")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Test 1: Create a new process registry
	fmt.Println("\n1. Creating new process registry...")
	registry := process_registry.NewProcessRegistry(tempDir)
	if registry == nil {
		t.Error("Could not create process registry")
		return
	}
	fmt.Println("✅ SUCCESS: Process registry created")

	// Test 2: Create a process record
	fmt.Println("\n2. Creating process record...")
	record := &process_registry.PersistentProcessRecord{
		ProcessID:      1234,
		Executable:     "minecraft.exe",
		Arguments:      "--demo",
		WorkingDir:     "C:\\Games\\Minecraft",
		StartTime:      time.Now(),
		LastSeen:       time.Now(),
		Status:         process_registry.ProcessStatus_Running,
		ModpackName:    "Test Modpack",
		ModpackVersion: "1.0.0",
	}

	err := registry.AddOrUpdateRecord(record)
	if err != nil {
		t.Errorf("Could not add process record: %v", err)
		return
	}
	fmt.Println("✅ SUCCESS: Process record added")

	// Test 3: Save the registry
	fmt.Println("\n3. Saving process registry...")
	err = registry.Save()
	if err != nil {
		t.Errorf("Could not save registry: %v", err)
		return
	}
	fmt.Println("✅ SUCCESS: Registry saved")

	// Test 4: Load the registry
	fmt.Println("\n4. Loading process registry...")
	newRegistry := process_registry.NewProcessRegistry(tempDir)
	err = newRegistry.Load()
	if err != nil {
		t.Errorf("Could not load registry: %v", err)
		return
	}
	fmt.Println("✅ SUCCESS: Registry loaded")

	// Test 5: Verify the loaded record
	fmt.Println("\n5. Verifying loaded record...")
	records := newRegistry.GetAllRecords()
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
		return
	}

	loadedRecord := records[0]
	if loadedRecord.ProcessID != 1234 || loadedRecord.Executable != "minecraft.exe" {
		t.Error("Record data mismatch")
		return
	}
	fmt.Println("✅ SUCCESS: Record data verified")

	// Test 6: Clean up stale records
	fmt.Println("\n6. Testing cleanup of stale records...")
	err = newRegistry.CleanupStaleRecords()
	if err != nil {
		t.Errorf("Could not cleanup stale records: %v", err)
		return
	}
	fmt.Println("✅ SUCCESS: Cleanup completed")

	// Test 7: Get record by process ID
	fmt.Println("\n7. Testing get record by process ID...")
	foundRecord := newRegistry.GetRecordByProcessID(1234)
	if foundRecord == nil {
		t.Error("Could not find record by process ID")
		return
	}
	fmt.Println("✅ SUCCESS: Record found by process ID")

	fmt.Println("\n🎉 All tests passed! Process registry functionality is working correctly.")
}
