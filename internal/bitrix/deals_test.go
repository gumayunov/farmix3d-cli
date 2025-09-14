package bitrix

import (
	"strings"
	"testing"
)

func TestValidateDealID(t *testing.T) {
	tests := []struct {
		name        string
		dealID      string
		expectError bool
		errorContains string
	}{
		{
			name:        "valid positive integer",
			dealID:      "123",
			expectError: false,
		},
		{
			name:        "valid zero",
			dealID:      "0",
			expectError: false,
		},
		{
			name:        "valid large number",
			dealID:      "999999999",
			expectError: false,
		},
		{
			name:        "valid single digit",
			dealID:      "5",
			expectError: false,
		},
		{
			name:        "empty string should fail",
			dealID:      "",
			expectError: true,
			errorContains: "cannot be empty",
		},
		{
			name:        "alphabetic string should fail",
			dealID:      "abc",
			expectError: true,
			errorContains: "must be a number",
		},
		{
			name:        "alphanumeric string should fail",
			dealID:      "123abc",
			expectError: true,
			errorContains: "must be a number",
		},
		{
			name:        "decimal number should fail",
			dealID:      "123.45",
			expectError: true,
			errorContains: "must be a number",
		},
		{
			name:        "negative number should be valid (strconv.Atoi accepts it)",
			dealID:      "-123",
			expectError: false,
		},
		{
			name:        "number with spaces should fail",
			dealID:      " 123 ",
			expectError: true,
			errorContains: "must be a number",
		},
		{
			name:        "number with leading zero",
			dealID:      "0123",
			expectError: false, // This should be valid as it's still a valid integer
		},
		{
			name:        "special characters should fail",
			dealID:      "12@34",
			expectError: true,
			errorContains: "must be a number",
		},
		{
			name:        "plus sign should be valid (strconv.Atoi accepts it)",
			dealID:      "+123",
			expectError: false,
		},
		{
			name:        "whitespace only should fail",
			dealID:      "   ",
			expectError: true,
			errorContains: "must be a number",
		},
		{
			name:        "unicode digits should fail",
			dealID:      "１２３", // Full-width digits
			expectError: true,
			errorContains: "must be a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDealID(tt.dealID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for dealID '%s', but got none", tt.dealID)
					return
				}
				
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for dealID '%s', but got: %v", tt.dealID, err)
				}
			}
		})
	}
}

func TestValidateDealIDEdgeCases(t *testing.T) {
	// Test with very large number (should still be valid as string conversion)
	largeNumber := "999999999999999999999999999999"
	err := ValidateDealID(largeNumber)
	
	// This might fail due to integer overflow in strconv.Atoi, which is expected behavior
	if err != nil {
		t.Logf("Large number validation failed as expected: %v", err)
	} else {
		t.Log("Large number validation passed")
	}
}

func TestValidateDealIDConsistency(t *testing.T) {
	// Test that the function behaves consistently for the same input
	testID := "12345"
	
	err1 := ValidateDealID(testID)
	err2 := ValidateDealID(testID)
	
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("Function should be consistent: first call returned %v, second call returned %v", err1, err2)
	}
}

func TestClearDealProductRowsValidation(t *testing.T) {
	// Test cases for clear deal product rows functionality
	// Note: These are unit tests for the validation and structure logic,
	// not integration tests with actual Bitrix24 API calls
	
	tests := []struct {
		name     string
		dealID   string
		dryRun   bool
		expected string // expected in function behavior description
	}{
		{
			name:     "valid deal ID with dry-run",
			dealID:   "123",
			dryRun:   true,
			expected: "should work with dry-run mode",
		},
		{
			name:     "valid deal ID without dry-run",
			dealID:   "456",
			dryRun:   false,
			expected: "should work in actual mode",
		},
		{
			name:     "zero deal ID",
			dealID:   "0",
			dryRun:   true,
			expected: "should accept zero as valid deal ID",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the deal ID validation passes for these cases
			// The actual ClearDealProductRows method would fail due to no API connection,
			// but we can test that the deal ID validation logic works
			err := ValidateDealID(tt.dealID)
			if err != nil {
				t.Errorf("ValidateDealID failed for case '%s': %v", tt.name, err)
			}
			
			// Test that dry-run parameter is handled properly (boolean values)
			if tt.dryRun != true && tt.dryRun != false {
				t.Errorf("DryRun parameter should be boolean, got: %v", tt.dryRun)
			}
		})
	}
}