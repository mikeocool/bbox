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
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0},
			expectError: false,
		},
		{
			name:        "Zero-size bbox",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 1.0, MaxY: 4.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: MaxX (%f) must be greater than MinX (%f)", 1.0, 1.0),
		},
		{
			name:        "Negative-width bbox",
			bbox:        Bbox{MinX: 3.0, MinY: 2.0, MaxX: 1.0, MaxY: 4.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: MaxX (%f) must be greater than MinX (%f)", 1.0, 3.0),
		},
		{
			name:        "Zero-height bbox",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 2.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: MaxY (%f) must be greater than MinY (%f)", 2.0, 2.0),
		},
		{
			name:        "Negative-height bbox",
			bbox:        Bbox{MinX: 1.0, MinY: 4.0, MaxX: 3.0, MaxY: 2.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: MaxY (%f) must be greater than MinY (%f)", 2.0, 4.0),
		},
		{
			name:        "Invalid width and height",
			bbox:        Bbox{MinX: 3.0, MinY: 4.0, MaxX: 1.0, MaxY: 2.0},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: MaxX (%f) must be greater than MinX (%f)", 1.0, 3.0),
		},
		{
			name:        "Zero value bbox",
			bbox:        Bbox{},
			expectError: true,
			errorMsg:    fmt.Sprintf("invalid bbox: MaxX (%f) must be greater than MinX (%f)", 0.0, 0.0),
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
		bbox := Bbox{MinX: -1e10, MinY: -1e10, MaxX: 1e10, MaxY: 1e10}
		if err := bbox.Validate(); err != nil {
			t.Errorf("Unexpected error with very large values: %v", err)
		}
	})

	// Test with very small differences
	t.Run("Very small differences", func(t *testing.T) {
		bbox := Bbox{MinX: 0.0, MinY: 0.0, MaxX: 0.0000001, MaxY: 0.0000001}
		if err := bbox.Validate(); err != nil {
			t.Errorf("Unexpected error with very small differences: %v", err)
		}
	})

	// Test with floating point precision issues
	t.Run("Floating point precision", func(t *testing.T) {
		// These values should be different enough to avoid floating point comparison issues
		bbox := Bbox{MinX: 1.0, MinY: 1.0, MaxX: 1.0 + 1e-10, MaxY: 1.0 + 1e-10}
		if err := bbox.Validate(); err != nil {
			t.Errorf("Unexpected error with floating point precision: %v", err)
		}
	})
}