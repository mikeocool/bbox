package core

import (
	"fmt"
	"testing"
)

func TestBboxValidate(t *testing.T) {
	tests := []struct {
		name        string
		bbox        Bbox
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid bbox",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expectError: false,
		},
		{
			name:        "Zero-size bbox",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 1.0, Top: 4.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: Right (%f) must be greater than Left (%f)", 1.0, 1.0),
		},
		{
			name:        "Negative-width bbox",
			bbox:        Bbox{Left: 3.0, Bottom: 2.0, Right: 1.0, Top: 4.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: Right (%f) must be greater than Left (%f)", 1.0, 3.0),
		},
		{
			name:        "Zero-height bbox",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 2.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: Top (%f) must be greater than Bottom (%f)", 2.0, 2.0),
		},
		{
			name:        "Negative-height bbox",
			bbox:        Bbox{Left: 1.0, Bottom: 4.0, Right: 3.0, Top: 2.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: Top (%f) must be greater than Bottom (%f)", 2.0, 4.0),
		},
		{
			name:        "Invalid width and height",
			bbox:        Bbox{Left: 3.0, Bottom: 4.0, Right: 1.0, Top: 2.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: Right (%f) must be greater than Left (%f)", 1.0, 3.0),
		},
		{
			name:        "Zero value bbox",
			bbox:        Bbox{},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: Right (%f) must be greater than Left (%f)", 0.0, 0.0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.bbox.Validate()

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If expecting an error, verify the error message
			if tc.expectError && err != nil {
				if err.Error() != tc.errorMsg {
					t.Errorf("Expected error message %q but got %q", tc.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestBboxValidateEdgeCases(t *testing.T) {
	// Test with very large values
	t.Run("Very large values", func(t *testing.T) {
		bbox := Bbox{Left: -1e10, Bottom: -1e10, Right: 1e10, Top: 1e10}
		if err := bbox.Validate(); err != nil {
			t.Errorf("Unexpected error with very large values: %v", err)
		}
	})

	// Test with very small differences
	t.Run("Very small differences", func(t *testing.T) {
		bbox := Bbox{Left: 0.0, Bottom: 0.0, Right: 0.0000001, Top: 0.0000001}
		if err := bbox.Validate(); err != nil {
			t.Errorf("Unexpected error with very small differences: %v", err)
		}
	})

	// Test with floating point precision issues
	t.Run("Floating point precision", func(t *testing.T) {
		// These values should be different enough to avoid floating point comparison issues
		bbox := Bbox{Left: 1.0, Bottom: 1.0, Right: 1.0 + 1e-10, Top: 1.0 + 1e-10}
		if err := bbox.Validate(); err != nil {
			t.Errorf("Unexpected error with floating point precision: %v", err)
		}
	})
}