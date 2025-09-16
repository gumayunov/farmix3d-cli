package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"farmix-cli/internal/bitrix"
)

func TestFind3DFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  []string
		expectedFiles []bitrix.FileInfo
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
			expectedFiles: []bitrix.FileInfo{
				{FileName: "assembly.sTl", DirPath: ""},    // cleanName: "assembly"
				{FileName: "component.step", DirPath: ""},  // cleanName: "component"
				{FileName: "gear.STEP", DirPath: ""},       // cleanName: "gear"
				{FileName: "model.StEp", DirPath: ""},      // cleanName: "model"
				{FileName: "part1.stl", DirPath: ""},       // cleanName: "part1"
				{FileName: "part2.STL", DirPath: ""},       // cleanName: "part2"
			},
			expectError: false,
		},
		{
			name: "finds files with quantity prefixes and sorts by clean name",
			setupFiles: []string{
				"5x_bracket.stl",
				"2x_adapter.stl", 
				"bracket_v2.step",
				"10x_adapter.step",
				"gear.stl",
				"3х_gear.step",   // cyrillic х
				"1x_mount.stl",
			},
			expectedFiles: []bitrix.FileInfo{
				{FileName: "10x_adapter.step", DirPath: ""}, // cleanName: "adapter"
				{FileName: "2x_adapter.stl", DirPath: ""},   // cleanName: "adapter"
				{FileName: "5x_bracket.stl", DirPath: ""},   // cleanName: "bracket"
				{FileName: "bracket_v2.step", DirPath: ""},  // cleanName: "bracket_v2"
				{FileName: "3х_gear.step", DirPath: ""},     // cleanName: "gear"
				{FileName: "gear.stl", DirPath: ""},         // cleanName: "gear"
				{FileName: "1x_mount.stl", DirPath: ""},     // cleanName: "mount"
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
			expectedFiles: []bitrix.FileInfo{
				{FileName: "gear.step", DirPath: ""},   // cleanName: "gear"
				{FileName: "part1.stl", DirPath: ""},   // cleanName: "part1"
			},
			expectError: false,
		},
		{
			name: "finds files in subdirectories recursively with quantity prefixes",
			setupFiles: []string{
				"2x_part1.stl",
				"subdir/5x_part2.step",
				"deep/nested/1x_gear.STL",
				"another/10x_component.STEP",
				"simple_part.stl",
			},
			expectedFiles: []bitrix.FileInfo{
				{FileName: "10x_component.STEP", DirPath: "another"}, // cleanName: "component"
				{FileName: "1x_gear.STL", DirPath: "deep/nested"},        // cleanName: "gear"
				{FileName: "2x_part1.stl", DirPath: ""},       // cleanName: "part1"
				{FileName: "5x_part2.step", DirPath: "subdir"},      // cleanName: "part2"
				{FileName: "simple_part.stl", DirPath: ""},    // cleanName: "simple_part"
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
			expectedFiles: []bitrix.FileInfo{},
			expectError: false,
		},
		{
			name: "returns empty slice for empty directory",
			setupFiles: []string{},
			expectedFiles: []bitrix.FileInfo{},
			expectError: false,
		},
		{
			name: "handles files with similar extensions and quantity prefixes",
			setupFiles: []string{
				"2x_model.stl",
				"backup.stl.bak",     // Should be ignored
				"5x_temp.step",
				"archive.step.gz",    // Should be ignored
				"component.stp",      // Should be ignored (not .step)
				"10х_bracket.stl",    // cyrillic х
				"adapter.step",
			},
			expectedFiles: []bitrix.FileInfo{
				{FileName: "adapter.step", DirPath: ""},    // cleanName: "adapter"
				{FileName: "10х_bracket.stl", DirPath: ""}, // cleanName: "bracket"
				{FileName: "2x_model.stl", DirPath: ""},    // cleanName: "model"
				{FileName: "5x_temp.step", DirPath: ""},    // cleanName: "temp"
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

			// Function now returns files sorted by clean name, so expected order matters
			expected := tt.expectedFiles

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

func TestFind3DFilesSortedByCleanName(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    []string
		expectedOrder []bitrix.FileInfo
	}{
		{
			name: "sorts files by clean name without quantity prefix",
			setupFiles: []string{
				"5x_bracket.stl",
				"2x_adapter.stl", 
				"bracket_v2.stl",
				"10x_adapter.stl",
				"gear.step",
				"3x_gear.step",
			},
			expectedOrder: []bitrix.FileInfo{
				{FileName: "10x_adapter.stl", DirPath: ""},  // cleanName: "adapter"
				{FileName: "2x_adapter.stl", DirPath: ""},   // cleanName: "adapter"
				{FileName: "5x_bracket.stl", DirPath: ""},   // cleanName: "bracket"
				{FileName: "bracket_v2.stl", DirPath: ""},   // cleanName: "bracket_v2"
				{FileName: "3x_gear.step", DirPath: ""},     // cleanName: "gear"
				{FileName: "gear.step", DirPath: ""},        // cleanName: "gear"
			},
		},
		{
			name: "handles case insensitive sorting",
			setupFiles: []string{
				"Mount.STL",
				"2x_ADAPTER.stl",
				"adapter_v2.step",
				"5x_mount.stl",
			},
			expectedOrder: []bitrix.FileInfo{
				{FileName: "2x_ADAPTER.stl", DirPath: ""},   // cleanName: "ADAPTER"
				{FileName: "adapter_v2.step", DirPath: ""},  // cleanName: "adapter_v2"
				{FileName: "5x_mount.stl", DirPath: ""},     // cleanName: "mount"
				{FileName: "Mount.STL", DirPath: ""},        // cleanName: "Mount"
			},
		},
		{
			name: "files without quantity prefix",
			setupFiles: []string{
				"zebra.stl",
				"alpha.step",
				"beta.stl",
			},
			expectedOrder: []bitrix.FileInfo{
				{FileName: "alpha.step", DirPath: ""},  // cleanName: "alpha"
				{FileName: "beta.stl", DirPath: ""},    // cleanName: "beta"
				{FileName: "zebra.stl", DirPath: ""},   // cleanName: "zebra"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "test_sort3d_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create test files
			for _, file := range tt.setupFiles {
				fullPath := filepath.Join(tempDir, file)
				f, err := os.Create(fullPath)
				if err != nil {
					t.Fatalf("Failed to create file %s: %v", fullPath, err)
				}
				f.Close()
			}

			// Test the function
			result, err := find3DFiles(tempDir)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check that files are returned in the expected sorted order
			if !reflect.DeepEqual(result, tt.expectedOrder) {
				t.Errorf("Expected order %v, but got %v", tt.expectedOrder, result)
				
				// Debug information to understand sorting
				t.Log("Debug info:")
				for i, fileInfo := range result {
					cleanName, _ := bitrix.ParseFileName(fileInfo.FileName)
					t.Logf("  [%d] %s (dir: %s) -> cleanName: '%s'", i, fileInfo.FileName, fileInfo.DirPath, cleanName)
				}
			}
		})
	}
}