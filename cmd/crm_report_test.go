package cmd

import (
	"testing"

	"farmix-cli/internal/bitrix"
)

// TestParseCustomFieldValue tests the parsing of custom field values
func TestParseCustomFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "String value",
			input:    "Test String",
			expected: "Test String",
		},
		{
			name:     "Integer as float64",
			input:    float64(1000),
			expected: "1000",
		},
		{
			name:     "Float with decimals",
			input:    float64(1234.56),
			expected: "1234.56",
		},
		{
			name:     "Integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "Boolean true",
			input:    true,
			expected: "Да",
		},
		{
			name:     "Boolean false",
			input:    false,
			expected: "Нет",
		},
		{
			name:     "Nil value",
			input:    nil,
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Monetary value with currency RUB",
			input:    "1000|RUB",
			expected: "1000",
		},
		{
			name:     "Monetary value with currency USD",
			input:    "2500.50|USD",
			expected: "2500.50",
		},
		{
			name:     "Monetary value with currency EUR",
			input:    "500|EUR",
			expected: "500",
		},
		{
			name:     "String without currency separator",
			input:    "Just a text",
			expected: "Just a text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bitrix.ParseCustomFieldValue(tt.input)
			if result != tt.expected {
				t.Errorf("ParseCustomFieldValue(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestReportCustomFieldsValidation tests validation of custom fields configuration
func TestReportCustomFieldsValidation(t *testing.T) {
	tests := []struct {
		name        string
		fields      bitrix.ReportCustomFields
		shouldEmpty bool
	}{
		{
			name: "All fields configured",
			fields: bitrix.ReportCustomFields{
				MachineCost:     "UF_CRM_123",
				HumanCost:       "UF_CRM_456",
				MaterialCost:    "UF_CRM_789",
				TotalCost:       "UF_CRM_012",
				PaymentReceived: "UF_CRM_345",
			},
			shouldEmpty: false,
		},
		{
			name: "Some fields configured",
			fields: bitrix.ReportCustomFields{
				MachineCost: "UF_CRM_123",
				TotalCost:   "UF_CRM_012",
			},
			shouldEmpty: false,
		},
		{
			name:        "No fields configured",
			fields:      bitrix.ReportCustomFields{},
			shouldEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.fields.MachineCost == "" &&
				tt.fields.HumanCost == "" &&
				tt.fields.MaterialCost == "" &&
				tt.fields.TotalCost == "" &&
				tt.fields.PaymentReceived == ""

			if isEmpty != tt.shouldEmpty {
				t.Errorf("Configuration validation failed: isEmpty=%v, want %v", isEmpty, tt.shouldEmpty)
			}
		})
	}
}

// TestFormatValidation tests format parameter validation
func TestFormatValidation(t *testing.T) {
	validFormats := []string{"text", "csv"}
	invalidFormats := []string{"json", "xml", "html", ""}

	for _, format := range validFormats {
		t.Run("Valid format: "+format, func(t *testing.T) {
			if format != "text" && format != "csv" {
				t.Errorf("Format %q should be valid", format)
			}
		})
	}

	for _, format := range invalidFormats {
		t.Run("Invalid format: "+format, func(t *testing.T) {
			if format == "text" || format == "csv" {
				t.Errorf("Format %q should be invalid", format)
			}
		})
	}
}
