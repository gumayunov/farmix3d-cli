package bitrix

import (
	"reflect"
	"testing"
)

func TestCreateDealProductRows(t *testing.T) {
	tests := []struct {
		name        string
		productIDs  []string
		expected    []DealProductRow
	}{
		{
			name:       "empty product IDs",
			productIDs: []string{},
			expected:   nil, // Go returns nil slice for empty input
		},
		{
			name:       "single product ID",
			productIDs: []string{"123"},
			expected: []DealProductRow{
				{
					ProductID: ProductIDString("123"),
					Quantity:  1.0,
					Price:     0.0,
				},
			},
		},
		{
			name:       "multiple product IDs",
			productIDs: []string{"123", "456", "789"},
			expected: []DealProductRow{
				{
					ProductID: ProductIDString("123"),
					Quantity:  1.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("456"),
					Quantity:  1.0,
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
			name:       "product IDs with different formats",
			productIDs: []string{"1", "0", "999999", "42"},
			expected: []DealProductRow{
				{
					ProductID: ProductIDString("1"),
					Quantity:  1.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("0"),
					Quantity:  1.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("999999"),
					Quantity:  1.0,
					Price:     0.0,
				},
				{
					ProductID: ProductIDString("42"),
					Quantity:  1.0,
					Price:     0.0,
				},
			},
		},
		{
			name:       "large number of products",
			productIDs: make([]string, 100),
			expected:   make([]DealProductRow, 100),
		},
	}

	// Setup the large test case
	for i := 0; i < 100; i++ {
		tests[len(tests)-1].productIDs[i] = string(rune('0' + (i % 10))) // "0", "1", ..., "9", "0", "1", ...
		tests[len(tests)-1].expected[i] = DealProductRow{
			ProductID: ProductIDString(string(rune('0' + (i % 10)))),
			Quantity:  1.0,
			Price:     0.0,
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateDealProductRows(tt.productIDs)

			// Handle nil vs empty slice comparison
			if tt.expected == nil && result == nil {
				// Both nil - test passes
			} else if len(tt.expected) == 0 && len(result) == 0 {
				// Both empty - test passes
			} else if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CreateDealProductRows() = %v, expected %v", result, tt.expected)
			}

			// Additional checks for non-empty results
			if len(tt.productIDs) > 0 {
				// Check that all quantities are 1.0
				for i, row := range result {
					if row.Quantity != 1.0 {
						t.Errorf("Row %d: expected quantity 1.0, got %f", i, row.Quantity)
					}
				}

				// Check that all prices are 0.0
				for i, row := range result {
					if row.Price != 0.0 {
						t.Errorf("Row %d: expected price 0.0, got %f", i, row.Price)
					}
				}

				// Check that ProductIDs match input
				for i, row := range result {
					expectedID := ProductIDString(tt.productIDs[i])
					if row.ProductID != expectedID {
						t.Errorf("Row %d: expected ProductID %s, got %s", i, expectedID, row.ProductID)
					}
				}

				// Check result length matches input length
				if len(result) != len(tt.productIDs) {
					t.Errorf("Expected %d rows, got %d", len(tt.productIDs), len(result))
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
	productIDs := []string{"123", "456", "789"}
	
	result1 := CreateDealProductRows(productIDs)
	result2 := CreateDealProductRows(productIDs)
	
	if !reflect.DeepEqual(result1, result2) {
		t.Error("Function should produce consistent results for the same input")
	}
}

func TestCreateDealProductRowsImmutability(t *testing.T) {
	// Test that modifying the input slice doesn't affect already returned results
	productIDs := []string{"123", "456"}
	result := CreateDealProductRows(productIDs)
	
	// Modify the input slice
	productIDs[0] = "999"
	
	// Check that the result wasn't affected
	if result[0].ProductID != ProductIDString("123") {
		t.Error("Function result should not be affected by modifications to input slice")
	}
}

func TestCreateDealProductRowsProductIDStringConversion(t *testing.T) {
	// Test that ProductIDString conversion works correctly
	productIDs := []string{"123", "0", "999999"}
	result := CreateDealProductRows(productIDs)
	
	for i, row := range result {
		// Test String() method
		if row.ProductID.String() != productIDs[i] {
			t.Errorf("Row %d: ProductID.String() = %s, expected %s", i, row.ProductID.String(), productIDs[i])
		}
		
		// Test that ProductIDString can be compared to string
		if string(row.ProductID) != productIDs[i] {
			t.Errorf("Row %d: string(ProductID) = %s, expected %s", i, string(row.ProductID), productIDs[i])
		}
	}
}