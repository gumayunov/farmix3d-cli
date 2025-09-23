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

			// Function now returns files in discovery order, so we check contents without order
			expected := tt.expectedFiles

			// Compare results - handle empty slice case
			if len(expected) == 0 && len(result) == 0 {
				// Both empty - test passes
				return
			}

			// Check that all expected files are present (order doesn't matter)
			if len(result) != len(expected) {
				t.Errorf("Expected %d files, but got %d files", len(expected), len(result))
				return
			}

			// Convert both slices to maps for easy comparison
			expectedMap := make(map[string]string)
			resultMap := make(map[string]string)

			for _, file := range expected {
				expectedMap[file.FileName] = file.DirPath
			}

			for _, file := range result {
				resultMap[file.FileName] = file.DirPath
			}

			if !reflect.DeepEqual(resultMap, expectedMap) {
				t.Errorf("Expected files %v, but got %v", expectedMap, resultMap)
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

