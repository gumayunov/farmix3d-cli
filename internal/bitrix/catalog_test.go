package bitrix

import (
	"reflect"
	"testing"
)

func TestParseFileName(t *testing.T) {
	tests := []struct {
		name             string
		fileName         string
		expectedCleanName string
		expectedQuantity float64
	}{
		{
			name:             "simple stl file without quantity",
			fileName:         "bracket.stl",
			expectedCleanName: "bracket",
			expectedQuantity: 1.0,
		},
		{
			name:             "simple step file without quantity",
			fileName:         "gear.step",
			expectedCleanName: "gear",
			expectedQuantity: 1.0,
		},
		{
			name:             "file with 2x prefix (latin x)",
			fileName:         "2x_part.stl",
			expectedCleanName: "part",
			expectedQuantity: 2.0,
		},
		{
			name:             "file with 3х prefix (cyrillic х)",
			fileName:         "3х_mount.step",
			expectedCleanName: "mount",
			expectedQuantity: 3.0,
		},
		{
			name:             "file with 10x prefix",
			fileName:         "10x_bracket.STL",
			expectedCleanName: "bracket",
			expectedQuantity: 10.0,
		},
		{
			name:             "complex filename with underscore",
			fileName:         "5x_complex_part_v2.step",
			expectedCleanName: "complex_part_v2",
			expectedQuantity: 5.0,
		},
		{
			name:             "filename with numbers but no quantity prefix",
			fileName:         "part123.stl",
			expectedCleanName: "part123",
			expectedQuantity: 1.0,
		},
		{
			name:             "filename with x but no quantity prefix",
			fileName:         "matrix_x_part.stl",
			expectedCleanName: "matrix_x_part",
			expectedQuantity: 1.0,
		},
		{
			name:             "quantity zero should default to 1.0",
			fileName:         "0x_invalid.stl",
			expectedCleanName: "0x_invalid", // No match, treated as regular name
			expectedQuantity: 1.0,
		},
		{
			name:             "uppercase extension",
			fileName:         "4x_TEST.STEP",
			expectedCleanName: "TEST",
			expectedQuantity: 4.0,
		},
		{
			name:             "file with space separator (1x SMA)",
			fileName:         "1x SMA Hear.stl",
			expectedCleanName: "SMA Hear",
			expectedQuantity: 1.0,
		},
		{
			name:             "file with space separator (2x ACC)",
			fileName:         "2x Acc Bracket.stl",
			expectedCleanName: "Acc Bracket",
			expectedQuantity: 2.0,
		},
		{
			name:             "file with cyrillic x and space",
			fileName:         "3х Tank Mount.step",
			expectedCleanName: "Tank Mount",
			expectedQuantity: 3.0,
		},
		{
			name:             "file with space separator and complex name",
			fileName:         "5x Tank Mount Att Radar.stl",
			expectedCleanName: "Tank Mount Att Radar",
			expectedQuantity: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanName, quantity := ParseFileName(tt.fileName)
			
			if cleanName != tt.expectedCleanName {
				t.Errorf("ParseFileName(%s) cleanName = %s, expected %s", tt.fileName, cleanName, tt.expectedCleanName)
			}
			
			if quantity != tt.expectedQuantity {
				t.Errorf("ParseFileName(%s) quantity = %.1f, expected %.1f", tt.fileName, quantity, tt.expectedQuantity)
			}
		})
	}
}

func TestFormatProductName(t *testing.T) {
	tests := []struct {
		name          string
		cleanName     string
		quantity      float64
		expectedName  string
	}{
		{
			name:         "simple name with quantity 1",
			cleanName:    "bracket",
			quantity:     1.0,
			expectedName: "Изделие \"bracket\"",
		},
		{
			name:         "simple name with quantity 4",
			cleanName:    "bracket",
			quantity:     4.0,
			expectedName: "Изделие \"bracket Q4\"",
		},
		{
			name:         "complex name with underscores and quantity 2",
			cleanName:    "complex_part_v2",
			quantity:     2.0,
			expectedName: "Изделие \"complex_part_v2 Q2\"",
		},
		{
			name:         "name with numbers and quantity 1",
			cleanName:    "gear123",
			quantity:     1.0,
			expectedName: "Изделие \"gear123\"",
		},
		{
			name:         "name with numbers and quantity 10",
			cleanName:    "gear123",
			quantity:     10.0,
			expectedName: "Изделие \"gear123 Q10\"",
		},
		{
			name:         "empty name with quantity 1",
			cleanName:    "",
			quantity:     1.0,
			expectedName: "Изделие \"\"",
		},
		{
			name:         "empty name with quantity 3",
			cleanName:    "",
			quantity:     3.0,
			expectedName: "Изделие \" Q3\"",
		},
		{
			name:         "name with special characters and quantity 1",
			cleanName:    "part-model_final",
			quantity:     1.0,
			expectedName: "Изделие \"part-model_final\"",
		},
		{
			name:         "name with special characters and quantity 5",
			cleanName:    "part-model_final",
			quantity:     5.0,
			expectedName: "Изделие \"part-model_final Q5\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatProductName(tt.cleanName, tt.quantity)
			
			if result != tt.expectedName {
				t.Errorf("FormatProductName(%s, %.1f) = %s, expected %s", tt.cleanName, tt.quantity, result, tt.expectedName)
			}
		})
	}
}

func TestCreateDealProductRows(t *testing.T) {
	tests := []struct {
		name        string
		products    []ProductInfo
		expected    []DealProductRow
	}{
		{
			name:     "empty product list",
			products: []ProductInfo{},
			expected: nil, // Go returns nil slice for empty input
		},
		{
			name: "single product with default quantity",
			products: []ProductInfo{
				{ID: "123", Quantity: 1.0},
			},
			expected: []DealProductRow{
				{
					ProductID: ProductIDString("123"),
					Quantity:  1.0,
					Price:     0.0,
				},
			},
		},
		{
			name: "multiple products with different quantities",
			products: []ProductInfo{
				{ID: "123", Quantity: 2.0},
				{ID: "456", Quantity: 5.0},
				{ID: "789", Quantity: 1.0},
			},
			expected: []DealProductRow{
				{
					ProductID: ProductIDString("123"),
					Quantity:  2.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("456"),
					Quantity:  5.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("789"),
					Quantity:  1.0,
					Price:     0.0,
				},
			},
		},
		{
			name: "products with various quantities",
			products: []ProductInfo{
				{ID: "1", Quantity: 10.0},
				{ID: "0", Quantity: 3.0},
				{ID: "999999", Quantity: 1.0},
				{ID: "42", Quantity: 7.0},
			},
			expected: []DealProductRow{
				{
					ProductID: ProductIDString("1"),
					Quantity:  10.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("0"),
					Quantity:  3.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("999999"),
					Quantity:  1.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("42"),
					Quantity:  7.0,
					Price:     0.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateDealProductRows(tt.products)

			// Handle nil vs empty slice comparison
			if tt.expected == nil && result == nil {
				// Both nil - test passes
			} else if len(tt.expected) == 0 && len(result) == 0 {
				// Both empty - test passes
			} else if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CreateDealProductRows() = %v, expected %v", result, tt.expected)
			}

			// Additional checks for non-empty results
			if len(tt.products) > 0 {
				// Check that all prices are 0.0
				for i, row := range result {
					if row.Price != 0.0 {
						t.Errorf("Row %d: expected price 0.0, got %f", i, row.Price)
					}
				}

				// Check that ProductIDs and quantities match input
				for i, row := range result {
					expectedID := ProductIDString(tt.products[i].ID)
					if row.ProductID != expectedID {
						t.Errorf("Row %d: expected ProductID %s, got %s", i, expectedID, row.ProductID)
					}
					
					if row.Quantity != tt.products[i].Quantity {
						t.Errorf("Row %d: expected quantity %.1f, got %.1f", i, tt.products[i].Quantity, row.Quantity)
					}
				}

				// Check result length matches input length
				if len(result) != len(tt.products) {
					t.Errorf("Expected %d rows, got %d", len(tt.products), len(result))
				}
			}
		})
	}
}

func TestCreateDealProductRowsNilInput(t *testing.T) {
	// Test with nil slice (should behave same as empty slice)
	result := CreateDealProductRows(nil)
	
	// For nil input, Go's range returns nothing, so we get nil slice
	if result != nil {
		t.Errorf("Expected nil slice for nil input, got %v", result)
	}
}

func TestCreateDealProductRowsConsistency(t *testing.T) {
	// Test that the function produces consistent results for the same input
	products := []ProductInfo{
		{ID: "123", Quantity: 2.0},
		{ID: "456", Quantity: 5.0},
		{ID: "789", Quantity: 1.0},
	}
	
	result1 := CreateDealProductRows(products)
	result2 := CreateDealProductRows(products)
	
	if !reflect.DeepEqual(result1, result2) {
		t.Error("Function should produce consistent results for the same input")
	}
}

func TestCreateDealProductRowsImmutability(t *testing.T) {
	// Test that modifying the input slice doesn't affect already returned results
	products := []ProductInfo{
		{ID: "123", Quantity: 2.0},
		{ID: "456", Quantity: 3.0},
	}
	result := CreateDealProductRows(products)
	
	// Modify the input slice
	products[0].ID = "999"
	products[0].Quantity = 10.0
	
	// Check that the result wasn't affected
	if result[0].ProductID != ProductIDString("123") {
		t.Error("Function result should not be affected by modifications to input slice")
	}
	if result[0].Quantity != 2.0 {
		t.Error("Function result quantity should not be affected by modifications to input slice")
	}
}

func TestCreateDealProductRowsProductIDStringConversion(t *testing.T) {
	// Test that ProductIDString conversion works correctly
	products := []ProductInfo{
		{ID: "123", Quantity: 1.0},
		{ID: "0", Quantity: 2.0},
		{ID: "999999", Quantity: 3.0},
	}
	result := CreateDealProductRows(products)

	for i, row := range result {
		// Test String() method
		if row.ProductID.String() != products[i].ID {
			t.Errorf("Row %d: ProductID.String() = %s, expected %s", i, row.ProductID.String(), products[i].ID)
		}

		// Test that ProductIDString can be compared to string
		if string(row.ProductID) != products[i].ID {
			t.Errorf("Row %d: string(ProductID) = %s, expected %s", i, string(row.ProductID), products[i].ID)
		}
	}
}

func TestFormatDirPrefix(t *testing.T) {
	tests := []struct {
		name     string
		dirPath  string
		expected string
	}{
		{
			name:     "empty directory path",
			dirPath:  "",
			expected: "",
		},
		{
			name:     "single directory",
			dirPath:  "parts",
			expected: "parts ",
		},
		{
			name:     "nested directories with slash separator",
			dirPath:  "arms/mechanisms",
			expected: "arms.mechanisms ",
		},
		{
			name:     "deep nested directories",
			dirPath:  "project/parts/brackets/v2",
			expected: "project.parts.brackets.v2 ",
		},
		{
			name:     "directory with special characters",
			dirPath:  "project_v1/test-parts",
			expected: "project_v1.test-parts ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDirPrefix(tt.dirPath)

			if result != tt.expected {
				t.Errorf("formatDirPrefix(%s) = %s, expected %s", tt.dirPath, result, tt.expected)
			}
		})
	}
}

func TestEnsureCompaniesFolderLogic(t *testing.T) {
	// Test that EnsureCompaniesFolder would work correctly
	// This is a conceptual test since we can't test actual Bitrix24 API calls

	tests := []struct {
		name         string
		folderName   string
		expectError  bool
	}{
		{
			name:        "companies folder name constant is correct",
			folderName:  COMPANIES_FOLDER_NAME,
			expectError: false,
		},
		{
			name:        "companies folder name is not empty",
			folderName:  COMPANIES_FOLDER_NAME,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the constant is properly defined
			if COMPANIES_FOLDER_NAME == "" {
				t.Error("COMPANIES_FOLDER_NAME should not be empty")
			}

			if COMPANIES_FOLDER_NAME != "Компании" {
				t.Errorf("COMPANIES_FOLDER_NAME = %s, expected 'Компании'", COMPANIES_FOLDER_NAME)
			}
		})
	}
}

func TestFormatProductNameWithDir(t *testing.T) {
	tests := []struct {
		name          string
		cleanName     string
		dirPath       string
		quantity      float64
		expectedName  string
	}{
		{
			name:         "simple name with empty directory and quantity 1",
			cleanName:    "bracket",
			dirPath:      "",
			quantity:     1.0,
			expectedName: "Изделие \"bracket\"",
		},
		{
			name:         "simple name with directory and quantity 1",
			cleanName:    "bracket",
			dirPath:      "parts",
			quantity:     1.0,
			expectedName: "Изделие \"parts bracket\"",
		},
		{
			name:         "simple name with nested directory and quantity 1",
			cleanName:    "gear",
			dirPath:      "arms/mechanisms",
			quantity:     1.0,
			expectedName: "Изделие \"arms.mechanisms gear\"",
		},
		{
			name:         "simple name with empty directory and quantity 4",
			cleanName:    "bracket",
			dirPath:      "",
			quantity:     4.0,
			expectedName: "Изделие \"bracket Q4\"",
		},
		{
			name:         "simple name with directory and quantity 2",
			cleanName:    "mount",
			dirPath:      "parts",
			quantity:     2.0,
			expectedName: "Изделие \"parts mount Q2\"",
		},
		{
			name:         "complex name with nested directory and quantity 10",
			cleanName:    "complex_part_v2",
			dirPath:      "project/brackets",
			quantity:     10.0,
			expectedName: "Изделие \"project.brackets complex_part_v2 Q10\"",
		},
		{
			name:         "name with numbers and deep nested directory",
			dirPath:      "arms/mechanisms/v3",
			cleanName:    "gear123",
			quantity:     5.0,
			expectedName: "Изделие \"arms.mechanisms.v3 gear123 Q5\"",
		},
		{
			name:         "empty name with directory and quantity 3",
			cleanName:    "",
			dirPath:      "test",
			quantity:     3.0,
			expectedName: "Изделие \"test  Q3\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatProductNameWithDir(tt.cleanName, tt.dirPath, tt.quantity)

			if result != tt.expectedName {
				t.Errorf("FormatProductNameWithDir(%s, %s, %.1f) = %s, expected %s", tt.cleanName, tt.dirPath, tt.quantity, result, tt.expectedName)
			}
		})
	}
}