package cmd

import (
	"testing"

	"farmix-cli/internal/bitrix"
)

func TestValidateAddStoreParameters(t *testing.T) {
	tests := []struct {
		name    string
		dealID  string
		wantErr bool
	}{
		{
			name:    "valid deal ID",
			dealID:  "123",
			wantErr: false,
		},
		{
			name:    "empty deal ID",
			dealID:  "",
			wantErr: true,
		},
		{
			name:    "non-numeric deal ID",
			dealID:  "abc",
			wantErr: true,
		},
		{
			name:    "deal ID with spaces",
			dealID:  "123 456",
			wantErr: true,
		},
		{
			name:    "negative deal ID",
			dealID:  "-123",
			wantErr: false, // strconv.Atoi accepts negative numbers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bitrix.ValidateDealID(tt.dealID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDealID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddStoreDefaults(t *testing.T) {
	// Test default values
	if addStoreStoreID != "1" {
		t.Errorf("Expected default store ID to be '1', got '%s'", addStoreStoreID)
	}

	if addStoreCurrency != "RUB" {
		t.Errorf("Expected default currency to be 'RUB', got '%s'", addStoreCurrency)
	}

	if addStoreDryRun != false {
		t.Errorf("Expected default dry-run to be false, got %v", addStoreDryRun)
	}
}

func TestStoreIDConfigLogic(t *testing.T) {
	tests := []struct {
		name             string
		flagValue        string
		configValue      string
		expectedResult   string
		shouldUseConfig  bool
	}{
		{
			name:            "flag not specified, config has value",
			flagValue:       "1", // default value
			configValue:     "5",
			expectedResult:  "5",
			shouldUseConfig: true,
		},
		{
			name:            "flag specified explicitly",
			flagValue:       "3",
			configValue:     "5",
			expectedResult:  "3",
			shouldUseConfig: false,
		},
		{
			name:            "flag not specified, config empty",
			flagValue:       "1", // default value
			configValue:     "",
			expectedResult:  "1",
			shouldUseConfig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from runCRMAddStore
			currentStoreID := tt.flagValue
			if currentStoreID == "1" { // Check if it's the default value
				if tt.configValue != "" {
					currentStoreID = tt.configValue
				}
			}

			if currentStoreID != tt.expectedResult {
				t.Errorf("Expected store ID '%s', got '%s'", tt.expectedResult, currentStoreID)
			}
		})
	}
}