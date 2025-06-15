package input

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestInputParams_GetBbox(t *testing.T) {
	tests := []struct {
		name        string
		params      InputParams
		expectError bool
		errorMsg    string
		expectBbox  *core.Bbox
	}{
		// RawBuilder tests
		{
			name: "RawBuilder - invalid",
			params: InputParams{
				Raw: []byte("some raw data"),
			},
			expectError: true,
			errorMsg:    "could not parse value: some",
		},
		{
			name: "RawBuilder - with unexpected field",
			params: InputParams{
				Raw:   []byte("some raw data"),
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
			expectBbox:  &core.Bbox{}, // PlaceBuilder returns empty Bbox
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
				Left:   floatPtr(1.0),
			},
			expectError: true,
			errorMsg:    "Unexpected argument: Left with place",
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
			expectBbox: &core.Bbox{
				Left:   8.0,  // 10.0 - 4/2
				Bottom: 17.0, // 20.0 - 6/2
				Right:  12.0, // 10.0 + 4/2
				Top:    23.0, // 20.0 + 6/2
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

		// File Builder tests
		{
			name: "FileBuilder - blank value in slice",
			params: InputParams{
				File: []string{""},
			},
			expectError: true,
			errorMsg:    "File: no valid file paths provided",
		},
		{
			name: "FileBuilder - whitespace value in slice",
			params: InputParams{
				File: []string{"   "},
			},
			expectError: true,
			errorMsg:    "File: no valid file paths provided",
		},

		// BoundsBuilder tests
		{
			name: "BoundsBuilder - Left and Right",
			params: InputParams{
				Left:   floatPtr(1.0),
				Right:  floatPtr(5.0),
				Bottom: floatPtr(2.0),
				Top:    floatPtr(8.0),
			},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Right:  5.0,
				Bottom: 2.0,
				Top:    8.0,
			},
		},
		{
			name: "BoundsBuilder - Left and Width",
			params: InputParams{
				Left:   floatPtr(1.0),
				Width:  "4",
				Bottom: floatPtr(2.0),
				Height: "6",
			},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Right:  5.0, // 1.0 + 4
				Bottom: 2.0,
				Top:    8.0, // 2.0 + 6
			},
		},
		{
			name: "BoundsBuilder - Right and Width",
			params: InputParams{
				Right:  floatPtr(5.0),
				Width:  "4",
				Top:    floatPtr(8.0),
				Height: "6",
			},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0, // 5.0 - 4
				Right:  5.0,
				Bottom: 2.0, // 8.0 - 6
				Top:    8.0,
			},
		},
		{
			name: "BoundsBuilder - only Left (invalid)",
			params: InputParams{
				Left: floatPtr(1.0),
			},
			expectError: true,
			errorMsg:    "min specified without max or length",
		},
		{
			name: "BoundsBuilder - only Right (invalid)",
			params: InputParams{
				Right: floatPtr(5.0),
			},
			expectError: true,
			errorMsg:    "max specified without min or legnth",
		},
		{
			name: "BoundsBuilder - Left, Right and Width (invalid)",
			params: InputParams{
				Left:   floatPtr(1.0),
				Right:  floatPtr(5.0),
				Width:  "4",
				Bottom: floatPtr(2.0),
				Top:    floatPtr(8.0),
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

		// buffer
		{
			name: "BoundsBuilder - Left, Right and Width (invalid)",
			params: InputParams{
				Buffer: 2.0,
			},
			expectError: true,
			errorMsg:    "Cannot specify buffer without a bounding box",
		},
		{
			name: "Buffered bounds",
			params: InputParams{
				Buffer: 2.0,
				Left:   floatPtr(1.0),
				Right:  floatPtr(5.0),
				Bottom: floatPtr(2.0),
				Top:    floatPtr(8.0),
			},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   -1.0,
				Right:  7.0,
				Bottom: 0.0,
				Top:    10.0,
			},
		},
		{
			name: "Invalid buffer",
			params: InputParams{
				Buffer: -2.0,
				Left:   floatPtr(1.0),
				Right:  floatPtr(2.0),
				Bottom: floatPtr(1.0),
				Top:    floatPtr(2.0),
			},
			expectError: true,
			errorMsg:    fmt.Sprintf("cannot shrink box with width %f by %f", 1.0, -2.0),
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
				if bbox.Left != tc.expectBbox.Left {
					t.Errorf("Expected Left %f but got %f", tc.expectBbox.Left, bbox.Left)
				}
				if bbox.Bottom != tc.expectBbox.Bottom {
					t.Errorf("Expected Bottom %f but got %f", tc.expectBbox.Bottom, bbox.Bottom)
				}
				if bbox.Right != tc.expectBbox.Right {
					t.Errorf("Expected Right %f but got %f", tc.expectBbox.Right, bbox.Right)
				}
				if bbox.Top != tc.expectBbox.Top {
					t.Errorf("Expected Top %f but got %f", tc.expectBbox.Top, bbox.Top)
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
			name: "Has Left",
			params: InputParams{
				Left: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has Bottom",
			params: InputParams{
				Bottom: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has Right",
			params: InputParams{
				Right: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has Top",
			params: InputParams{
				Top: floatPtr(1.0),
			},
			expected: true,
		},
		{
			name: "Has all coordinates",
			params: InputParams{
				Left:   floatPtr(1.0),
				Bottom: floatPtr(2.0),
				Right:  floatPtr(3.0),
				Top:    floatPtr(4.0),
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
				Left:   floatPtr(1.0),
				Bottom: floatPtr(2.0),
				Right:  floatPtr(3.0),
				Top:    floatPtr(4.0),
				Center: []float64{5.0, 6.0},
				Width:  "100",
				Height: "200",
				Raw:    []byte("raw data"),
				Place:  "New York",
			},
			expected: []string{"Left", "Bottom", "Right", "Top", "Center", "Width", "Height", "Raw", "Place"},
		},
		{
			name: "Mixed field types",
			params: InputParams{
				Left:   floatPtr(0.0),  // zero value pointer should still count as set
				Center: []float64{},    // empty slice should be considered empty
				Width:  "",             // empty string should be considered empty
				Raw:    []byte("data"), // non-empty string
			},
			expected: []string{"Left", "Raw"},
		},
		{
			name: "Nil vs zero values",
			params: InputParams{
				Left:   nil,           // nil pointer
				Bottom: floatPtr(0.0), // zero value but not nil
				Center: nil,           // nil slice
				Width:  "",            // empty string
			},
			expected: []string{"Bottom"},
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
				Left:   floatPtr(1.0),
				Right:  floatPtr(5.0),
				Bottom: floatPtr(2.0),
				Top:    floatPtr(8.0),
			},
			expectError: false, // Should use BoundsBuilder instead
		},
		{
			name: "Zero value coordinates",
			params: InputParams{
				Left:   floatPtr(0.0), // zero value but not nil
				Right:  floatPtr(1.0),
				Bottom: floatPtr(0.0),
				Top:    floatPtr(1.0),
			},
			expectError: false,
		},
		{
			name: "Single coordinate with zero value",
			params: InputParams{
				Left: floatPtr(0.0), // only one coordinate set
			},
			expectError: true,
			errorMsg:    "min specified without max or length",
		},
		{
			name: "String fields with whitespace",
			params: InputParams{
				Raw: []byte(" "), // whitespace should be considered non-empty
			},
			expectError: true,
			errorMsg:    "invalid input",
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

func TestFileBuilder(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		expectError bool
		errorMsg    string
		expectBbox  *core.Bbox
	}{
		{
			name:        "Single valid file",
			files:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson")},
			expectError: false,
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99755413385825, Right: -91.14794444117372, Top: 48.01355378301334},
		},
		{
			name:        "Multiple valid files - union",
			files:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson"), getTestDataPath(t, "../integration_tests/data/subset_b.geojson")},
			expectError: false,
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99067253859491, Right: -90.92072645384923, Top: 48.07394149630552},
		},
		{
			name:        "Empty file",
			files:       []string{getTestDataPath(t, "../integration_tests/data/empty.geojson")},
			expectError: true, // Assuming empty file causes an error
			errorMsg:    "no features found",
		},
		{
			name:        "Non-existent file",
			files:       []string{getTestDataPath(t, "non_existent_file.geojson")},
			expectError: true,
		},
		{
			name:        "Mixed valid and empty files",
			files:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson"), getTestDataPath(t, "../integration_tests/data/empty.geojson")},
			expectError: false,
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99755413385825, Right: -91.14794444117372, Top: 48.01355378301334},
		},
		{
			name:        "Empty string in file list",
			files:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson"), "", getTestDataPath(t, "../integration_tests/data/subset_b.geojson")},
			expectError: false, // Empty strings should be skipped
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99067253859491, Right: -90.92072645384923, Top: 48.07394149630552},
		},
		{
			name:        "Empty file list",
			files:       []string{},
			expectError: true,
			errorMsg:    "no usable builder for the provided parameters",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := InputParams{
				File: tc.files,
			}

			bbox, err := params.GetBbox()

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

			// If not expecting an error, verify the bbox structure is valid
			if !tc.expectError {
				// Basic sanity checks - left should be <= right, bottom should be <= top
				if bbox.Left > bbox.Right {
					t.Errorf("Invalid bbox: Left (%f) > Right (%f)", bbox.Left, bbox.Right)
				}
				if bbox.Bottom > bbox.Top {
					t.Errorf("Invalid bbox: Bottom (%f) > Top (%f)", bbox.Bottom, bbox.Top)
				}

				// If we have expected bbox values, compare them
				if tc.expectBbox != nil {
					if bbox.Left != tc.expectBbox.Left {
						t.Errorf("Expected Left %f but got %f", tc.expectBbox.Left, bbox.Left)
					}
					if bbox.Bottom != tc.expectBbox.Bottom {
						t.Errorf("Expected Bottom %f but got %f", tc.expectBbox.Bottom, bbox.Bottom)
					}
					if bbox.Right != tc.expectBbox.Right {
						t.Errorf("Expected Right %f but got %f", tc.expectBbox.Right, bbox.Right)
					}
					if bbox.Top != tc.expectBbox.Top {
						t.Errorf("Expected Top %f but got %f", tc.expectBbox.Top, bbox.Top)
					}
				}
			}
		})
	}
}

func TestFileBuilder_IsUsable(t *testing.T) {
	tests := []struct {
		name     string
		params   InputParams
		expected bool
	}{
		{
			name:     "No files",
			params:   InputParams{},
			expected: false,
		},
		{
			name: "Empty file slice",
			params: InputParams{
				File: []string{},
			},
			expected: false,
		},
		{
			name: "Single file",
			params: InputParams{
				File: []string{"test.geojson"},
			},
			expected: true,
		},
		{
			name: "Multiple files",
			params: InputParams{
				File: []string{"test1.geojson", "test2.geojson"},
			},
			expected: true,
		},
		{
			name: "File slice with empty string",
			params: InputParams{
				File: []string{""},
			},
			expected: true, // Non-empty slice, even if contains empty string
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FileBuilder.IsUsable(&tc.params)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}

func getTestDataPath(t *testing.T, filename string) string {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}
	testDir := filepath.Dir(testFile)
	return filepath.Join(testDir, filename)
}

// Helper function to create float64 pointers
func floatPtr(f float64) *float64 {
	return &f
}
