package core

import (
	"testing"
)

func TestInputParams_GetBbox(t *testing.T) {
	tests := []struct {
		name        string
		params      InputParams
		expectError bool
		errorMsg    string
		expectBbox  *Bbox
	}{
		// RawBuilder tests
		{
			name: "RawBuilder - valid",
			params: InputParams{
				Raw: "some raw data",
			},
			expectError: false,
			expectBbox:  &Bbox{}, // RawBuilder returns empty Bbox
		},
		{
			name: "RawBuilder - with unexpected field",
			params: InputParams{
				Raw:   "some raw data",
				Place: "unexpected",
			},
			expectError: true,
			errorMsg:    "Unexpected argument: Place with ",
		},

		// PlaceBuilder tests
		{
			name: "PlaceBuilder - valid",
			params: InputParams{
				Place:  "New York",
				Width:  "100",
				Height: "200",
			},
			expectError: false,
			expectBbox:  &Bbox{}, // PlaceBuilder returns empty Bbox
		},
		{
			name: "PlaceBuilder - missing width",
			params: InputParams{
				Place:  "New York",
				Height: "200",
			},
			expectError: true,
			errorMsg:    "width: width required",
		},
		{
			name: "PlaceBuilder - missing height",
			params: InputParams{
				Place: "New York",
				Width: "100",
			},
			expectError: true,
			errorMsg:    "height: height required",
		},
		{
			name: "PlaceBuilder - with unexpected field",
			params: InputParams{
				Place:  "New York",
				Width:  "100",
				Height: "200",
				MinX:   floatPtr(1.0),
			},
			expectError: true,
			errorMsg:    "Unexpected argument: MinX with place",
		},

		// CenterBuilder tests
		{
			name: "CenterBuilder - valid",
			params: InputParams{
				Center: []float64{10.0, 20.0},
				Width:  "4",
				Height: "6",
			},
			expectError: false,
			expectBbox: &Bbox{
				MinX: 8.0,  // 10.0 - 4/2
				MinY: 17.0, // 20.0 - 6/2
				MaxX: 12.0, // 10.0 + 4/2
				MaxY: 23.0, // 20.0 + 6/2
			},
		},
		{
			name: "CenterBuilder - invalid center coordinates",
			params: InputParams{
				Center: []float64{10.0},
				Width:  "4",
				Height: "6",
			},
			expectError: true,
			errorMsg:    "center: invalid center coordinates",
		},
		{
			name: "CenterBuilder - missing width",
			params: InputParams{
				Center: []float64{10.0, 20.0},
				Height: "6",
			},
			expectError: true,
			errorMsg:    "width: width required",
		},
		{
			name: "CenterBuilder - missing height",
			params: InputParams{
				Center: []float64{10.0, 20.0},
				Width:  "4",
			},
			expectError: true,
			errorMsg:    "height: height required",
		},
		{
			name: "CenterBuilder - invalid width format",
			params: InputParams{
				Center: []float64{10.0, 20.0},
				Width:  "invalid",
				Height: "6",
			},
			expectError: true,
		},
		{
			name: "CenterBuilder - invalid height format",
			params: InputParams{
				Center: []float64{10.0, 20.0},
				Width:  "4",
				Height: "invalid",
			},
			expectError: true,
		},

		// BoundsBuilder tests
		{
			name: "BoundsBuilder - MinX and MaxX",
			params: InputParams{
				MinX: floatPtr(1.0),
				MaxX: floatPtr(5.0),
				MinY: floatPtr(2.0),
				MaxY: floatPtr(8.0),
			},
			expectError: false,
			expectBbox: &Bbox{
				MinX: 1.0,
				MaxX: 5.0,
				MinY: 2.0,
				MaxY: 8.0,
			},
		},
		{
			name: "BoundsBuilder - MinX and Width",
			params: InputParams{
				MinX:   floatPtr(1.0),
				Width:  "4",
				MinY:   floatPtr(2.0),
				Height: "6",
			},
			expectError: false,
			expectBbox: &Bbox{
				MinX: 1.0,
				MaxX: 5.0, // 1.0 + 4
				MinY: 2.0,
				MaxY: 8.0, // 2.0 + 6
			},
		},
		{
			name: "BoundsBuilder - MaxX and Width",
			params: InputParams{
				MaxX:   floatPtr(5.0),
				Width:  "4",
				MaxY:   floatPtr(8.0),
				Height: "6",
			},
			expectError: false,
			expectBbox: &Bbox{
				MinX: 1.0, // 5.0 - 4
				MaxX: 5.0,
				MinY: 2.0, // 8.0 - 6
				MaxY: 8.0,
			},
		},
		{
			name: "BoundsBuilder - only MinX (invalid)",
			params: InputParams{
				MinX: floatPtr(1.0),
			},
			expectError: true,
			errorMsg:    "min specified without max or length",
		},
		{
			name: "BoundsBuilder - only MaxX (invalid)",
			params: InputParams{
				MaxX: floatPtr(5.0),
			},
			expectError: true,
			errorMsg:    "max specified without min or legnth",
		},
		{
			name: "BoundsBuilder - MinX, MaxX and Width (invalid)",
			params: InputParams{
				MinX:  floatPtr(1.0),
				MaxX:  floatPtr(5.0),
				Width: "4",
				MinY:  floatPtr(2.0),
				MaxY:  floatPtr(8.0),
			},
			expectError: true,
			errorMsg:    "must specify only two of: min, max, and length",
		},

		// No usable builder
		{
			name:        "No usable builder",
			params:      InputParams{},
			expectError: true,
			errorMsg:    "no usable builder for the provided parameters",
		},


	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bbox, err := tc.params.GetBbox()

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// If expecting an error, verify the error message
			if tc.expectError && err != nil {
				if tc.errorMsg != "" && err.Error() != tc.errorMsg {
					t.Errorf("Expected error message %q but got %q", tc.errorMsg, err.Error())
				}
				return
			}

			// If not expecting an error, verify the bbox
			if tc.expectBbox != nil {
				if bbox.MinX != tc.expectBbox.MinX {
					t.Errorf("Expected MinX %f but got %f", tc.expectBbox.MinX, bbox.MinX)
				}
				if bbox.MinY != tc.expectBbox.MinY {
					t.Errorf("Expected MinY %f but got %f", tc.expectBbox.MinY, bbox.MinY)
				}
				if bbox.MaxX != tc.expectBbox.MaxX {
					t.Errorf("Expected MaxX %f but got %f", tc.expectBbox.MaxX, bbox.MaxX)
				}
				if bbox.MaxY != tc.expectBbox.MaxY {
					t.Errorf("Expected MaxY %f but got %f", tc.expectBbox.MaxY, bbox.MaxY)
				}
			}
		})
	}
}

func TestInputParams_HasAnyCoordinates(t *testing.T) {
	tests := []struct {
		name     string
		params   InputParams
		expected bool
	}{
		{
			name:     "No coordinates",
			params:   InputParams{},
			expected: false,
		},
		{
			name: "Has MinX",
			params: InputParams{
				MinX: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has MinY",
			params: InputParams{
				MinY: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has MaxX",
			params: InputParams{
				MaxX: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has MaxY",
			params: InputParams{
				MaxY: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has all coordinates",
			params: InputParams{
				MinX: floatPtr(1.0),
				MinY: floatPtr(2.0),
				MaxX: floatPtr(3.0),
				MaxY: floatPtr(4.0),
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.params.HasAnyCoordinates()
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}

func TestInputParams_HasWidth(t *testing.T) {
	tests := []struct {
		name     string
		params   InputParams
		expected bool
	}{
		{
			name:     "No width",
			params:   InputParams{},
			expected: false,
		},
		{
			name: "Empty width",
			params: InputParams{
				Width: "",
			},
			expected: false,
		},
		{
			name: "Has width",
			params: InputParams{
				Width: "100",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.params.HasWidth()
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}

func TestInputParams_HasHeight(t *testing.T) {
	tests := []struct {
		name     string
		params   InputParams
		expected bool
	}{
		{
			name:     "No height",
			params:   InputParams{},
			expected: false,
		},
		{
			name: "Empty height",
			params: InputParams{
				Height: "",
			},
			expected: false,
		},
		{
			name: "Has height",
			params: InputParams{
				Height: "200",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.params.HasHeight()
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}

func TestInputParams_getSetFields(t *testing.T) {
	tests := []struct {
		name     string
		params   InputParams
		expected []string
	}{
		{
			name:     "Empty params",
			params:   InputParams{},
			expected: []string{},
		},
		{
			name: "All fields set",
			params: InputParams{
				MinX:   floatPtr(1.0),
				MinY:   floatPtr(2.0),
				MaxX:   floatPtr(3.0),
				MaxY:   floatPtr(4.0),
				Center: []float64{5.0, 6.0},
				Width:  "100",
				Height: "200",
				Raw:    "raw data",
				Place:  "New York",
			},
			expected: []string{"MinX", "MinY", "MaxX", "MaxY", "Center", "Width", "Height", "Raw", "Place"},
		},
		{
			name: "Mixed field types",
			params: InputParams{
				MinX:   floatPtr(0.0), // zero value pointer should still count as set
				Center: []float64{},   // empty slice should be considered empty
				Width:  "",            // empty string should be considered empty
				Raw:    "data",        // non-empty string
			},
			expected: []string{"MinX", "Raw"},
		},
		{
			name: "Nil vs zero values",
			params: InputParams{
				MinX:   nil,           // nil pointer
				MinY:   floatPtr(0.0), // zero value but not nil
				Center: nil,           // nil slice
				Width:  "",            // empty string
			},
			expected: []string{"MinY"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.params.getSetFields()

			// Convert to maps for easier comparison since order might vary
			resultMap := make(map[string]bool)
			for _, field := range result {
				resultMap[field] = true
			}
			expectedMap := make(map[string]bool)
			for _, field := range tc.expected {
				expectedMap[field] = true
			}

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d fields but got %d: %v", len(tc.expected), len(result), result)
				return
			}

			for _, expectedField := range tc.expected {
				if !resultMap[expectedField] {
					t.Errorf("Expected field %s to be in result %v", expectedField, result)
				}
			}

			for _, resultField := range result {
				if !expectedMap[resultField] {
					t.Errorf("Unexpected field %s in result %v", resultField, result)
				}
			}
		})
	}
}

func TestInputParams_EdgeCases(t *testing.T) {
	// Test edge cases that involve the field validation logic
	tests := []struct {
		name        string
		params      InputParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "Empty center slice",
			params: InputParams{
				Center: []float64{}, // empty slice, should not trigger CenterBuilder
				MinX:   floatPtr(1.0),
				MaxX:   floatPtr(5.0),
				MinY:   floatPtr(2.0),
				MaxY:   floatPtr(8.0),
			},
			expectError: false, // Should use BoundsBuilder instead
		},
		{
			name: "Zero value coordinates",
			params: InputParams{
				MinX: floatPtr(0.0), // zero value but not nil
				MaxX: floatPtr(1.0),
				MinY: floatPtr(0.0),
				MaxY: floatPtr(1.0),
			},
			expectError: false,
		},
		{
			name: "Single coordinate with zero value",
			params: InputParams{
				MinX: floatPtr(0.0), // only one coordinate set
			},
			expectError: true,
			errorMsg:    "min specified without max or length",
		},
		{
			name: "String fields with whitespace",
			params: InputParams{
				Raw: " ", // whitespace should be considered non-empty
			},
			expectError: false, // PlaceBuilder should be used
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.params.GetBbox()

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.expectError && tc.errorMsg != "" && err != nil {
				if err.Error() != tc.errorMsg {
					t.Errorf("Expected error message %q but got %q", tc.errorMsg, err.Error())
				}
			}
		})
	}
}

// Helper function to create float64 pointers
func floatPtr(f float64) *float64 {
	return &f
}
