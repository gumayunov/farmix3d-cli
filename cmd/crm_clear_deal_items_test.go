package cmd

import (
	"testing"
	"farmix-cli/internal/bitrix"
)

func TestCrmClearDealItemsValidation(t *testing.T) {
	tests := []struct {
		name        string
		dealID      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid deal ID",
			dealID:      "123",
			expectError: false,
		},
		{
			name:        "valid zero deal ID",
			dealID:      "0",
			expectError: false,
		},
		{
			name:        "empty deal ID should fail",
			dealID:      "",
			expectError: true,
			errorMsg:    "deal ID cannot be empty",
		},
		{
			name:        "non-numeric deal ID should fail",
			dealID:      "abc123",
			expectError: true,
			errorMsg:    "deal ID must be a number",
		},
		{
			name:        "negative deal ID should be valid (strconv.Atoi accepts it)",
			dealID:      "-123",
			expectError: false,
		},
		{
			name:        "decimal deal ID should fail",
			dealID:      "123.45",
			expectError: true,
			errorMsg:    "deal ID must be a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bitrix.ValidateDealID(tt.dealID)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for dealID '%s', but got none", tt.dealID)
					return
				}
				
				if tt.errorMsg != "" && !containsSubstring(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for dealID '%s', but got: %v", tt.dealID, err)
				}
			}
		})
	}
}

func TestCrmClearDealItemsCommandStructure(t *testing.T) {
	// Test that the command is properly structured
	if crmClearDealItemsCmd.Use != "crm-clear-deal-items" {
		t.Errorf("Expected command use to be 'crm-clear-deal-items', got '%s'", crmClearDealItemsCmd.Use)
	}
	
	if crmClearDealItemsCmd.Short == "" {
		t.Error("Expected command to have a short description")
	}
	
	if crmClearDealItemsCmd.Long == "" {
		t.Error("Expected command to have a long description")
	}
	
	// Check that required flags are set
	dealIdFlag := crmClearDealItemsCmd.Flags().Lookup("deal-id")
	if dealIdFlag == nil {
		t.Error("Expected 'deal-id' flag to exist")
	}
	
	dryRunFlag := crmClearDealItemsCmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected 'dry-run' flag to exist")
	}
}

// Helper function to check if string contains substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && (len(substr) == 0 || findSubstring(str, substr) >= 0)
}

func findSubstring(str, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(str)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if str[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}