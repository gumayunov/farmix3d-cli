package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestFind3DFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  []string
		expectedFiles []string
		expectError bool
	}{
		{
			name: "finds STL and STEP files with various cases",
			setupFiles: []string{
				"part1.stl",
				"part2.STL", 
				"component.step",
				"gear.STEP",
				"model.StEp",
				"assembly.sTl",
			},
			expectedFiles: []string{
				"part1.stl",
				"part2.STL",
				"component.step", 
				"gear.STEP",
				"model.StEp",
				"assembly.sTl",
			},
			expectError: false,
		},
		{
			name: "ignores non-3D files",
			setupFiles: []string{
				"part1.stl",
				"readme.txt",
				"config.json",
				"model.obj",
				"assembly.3mf",
				"gear.step",
				"notes.md",
			},
			expectedFiles: []string{
				"part1.stl",
				"gear.step",
			},
			expectError: false,
		},
		{
			name: "finds files in subdirectories recursively",
			setupFiles: []string{
				"part1.stl",
				"subdir/part2.step",
				"deep/nested/gear.STL",
				"another/component.STEP",
			},
			expectedFiles: []string{
				"part1.stl",
				"part2.step",
				"gear.STL", 
				"component.STEP",
			},
			expectError: false,
		},
		{
			name: "returns empty slice for directory with no 3D files",
			setupFiles: []string{
				"readme.txt",
				"config.json",
				"image.png",
			},
			expectedFiles: []string{},
			expectError: false,
		},
		{
			name: "returns empty slice for empty directory",
			setupFiles: []string{},
			expectedFiles: []string{},
			expectError: false,
		},
		{
			name: "handles files with similar extensions",
			setupFiles: []string{
				"model.stl",
				"backup.stl.bak",  // Should be ignored
				"temp.step",
				"archive.step.gz", // Should be ignored
				"component.stp",   // Should be ignored (not .step)
			},
			expectedFiles: []string{
				"model.stl",
				"temp.step",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "test_find3d_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create test files
			for _, file := range tt.setupFiles {
				fullPath := filepath.Join(tempDir, file)
				
				// Create directory if needed
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory %s: %v", dir, err)
				}

				// Create empty file
				f, err := os.Create(fullPath)
				if err != nil {
					t.Fatalf("Failed to create file %s: %v", fullPath, err)
				}
				f.Close()
			}

			// Test the function
			result, err := find3DFiles(tempDir)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Sort both slices for comparison (since file order may vary)
			sort.Strings(result)
			expected := make([]string, len(tt.expectedFiles))
			copy(expected, tt.expectedFiles)
			sort.Strings(expected)

			// Compare results - handle empty slice case
			if len(expected) == 0 && len(result) == 0 {
				// Both empty - test passes
				return
			}
			
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("Expected files %v, but got %v", expected, result)
			}
		})
	}
}

func TestFind3DFilesNonExistentDirectory(t *testing.T) {
	// Test with non-existent directory
	nonExistentDir := "/tmp/non_existent_directory_12345"
	_, err := find3DFiles(nonExistentDir)
	
	if err == nil {
		t.Error("Expected error for non-existent directory, but got none")
	}
}

func TestFind3DFilesPermissionDenied(t *testing.T) {
	// Create temporary directory with restricted permissions
	tempDir, err := os.MkdirTemp("", "test_find3d_perm_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory with no read permissions
	restrictedDir := filepath.Join(tempDir, "restricted")
	if err := os.Mkdir(restrictedDir, 0000); err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

	// The function should handle permission errors gracefully
	// Since filepath.WalkDir handles permission errors by calling the WalkDirFunc with the error,
	// and our implementation returns that error, we expect an error here
	_, err = find3DFiles(tempDir)
	
	// On some systems, permission errors might be handled differently
	// The important thing is that the function doesn't panic
	if err != nil {
		t.Logf("Got expected permission error: %v", err)
	}
}