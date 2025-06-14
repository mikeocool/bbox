package output

import (
	"strings"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestTemplatedFormatPoint(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		point       [2]float64
		expected    string
		expectError bool
	}{
		{
			name:        "Basic format",
			template:    "{{.X}} {{.Y}}",
			point:       [2]float64{1, 2},
			expected:    "1 2",
			expectError: false,
		},
		{
			name:        "JSON format",
			template:    "{\"x\":{{.X}},\"y\":{{.Y}}}",
			point:       [2]float64{10.5, 20.5},
			expected:    "{\"x\":10.5,\"y\":20.5}",
			expectError: false,
		},
		{
			name:        "With missing function",
			template:    "{{.X}} {{sub .Y .X}}",
			point:       [2]float64{10.5, 20.5},
			expected:    "",
			expectError: true, // Will error because the "sub" function is not defined
		},
		{
			name:        "Template execution error",
			template:    "{{if .NonExistentMethod.Call}}This will fail at execution time{{end}}",
			point:       [2]float64{10.5, 20.5},
			expected:    "",
			expectError: true, // Will error during execution, not parsing
		},
		{
			name:        "Invalid template syntax",
			template:    "{{if .Left}}Only one closing bracket",
			point:       [2]float64{10.5, 20.5},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Malformed template",
			template:    "{{.Left} {{.Bottom}}",
			point:       [2]float64{10.5, 20.5},
			expected:    "",
			expectError: true,
		},
		{
			name:        "With formatting",
			template:    "{{printf \"%.2f\" .X}} {{printf \"%.2f\" .Y}}",
			point:       [2]float64{1.123, 2.456},
			expected:    "1.12 2.46",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TemplatedFormatPoint(OutputSettings{FormatDetails: tc.template}, tc.point)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError {
				if strings.TrimSpace(result) != strings.TrimSpace(tc.expected) {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}
		})
	}

	// Add a test for empty template
	t.Run("Empty template", func(t *testing.T) {
		result, err := TemplatedFormat(OutputSettings{FormatDetails: ""}, core.Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0})
		if err != nil {
			t.Errorf("Unexpected error with empty template: %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty result for empty template, got %q", result)
		}
	})
}

func TestCommaFormatPoint(t *testing.T) {
	tests := []struct {
		name     string
		point    [2]float64
		expected string
	}{
		{
			name:     "Zero coordinates",
			point:    [2]float64{0.0, 0.0},
			expected: "0,0",
		},
		{
			name:     "Positive integers",
			point:    [2]float64{1.0, 2.0},
			expected: "1,2",
		},
		{
			name:     "Decimal coordinates",
			point:    [2]float64{10.5, 20.25},
			expected: "10.5,20.25",
		},
		{
			name:     "Negative coordinates",
			point:    [2]float64{-10.0, -20.0},
			expected: "-10,-20",
		},
		{
			name:     "Mixed sign coordinates",
			point:    [2]float64{-10.5, 20.25},
			expected: "-10.5,20.25",
		},
		{
			name:     "Large coordinates",
			point:    [2]float64{1000000.123, 2000000.456},
			expected: "1.000000123e+06,2.000000456e+06",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CommaFormatPoint(OutputSettings{}, tc.point)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestSpaceFormatPoint(t *testing.T) {
	tests := []struct {
		name     string
		point    [2]float64
		expected string
	}{
		{
			name:     "Zero coordinates",
			point:    [2]float64{0.0, 0.0},
			expected: "0 0",
		},
		{
			name:     "Positive integers",
			point:    [2]float64{1.0, 2.0},
			expected: "1 2",
		},
		{
			name:     "Decimal coordinates",
			point:    [2]float64{10.5, 20.25},
			expected: "10.5 20.25",
		},
		{
			name:     "Negative coordinates",
			point:    [2]float64{-10.0, -20.0},
			expected: "-10 -20",
		},
		{
			name:     "Mixed sign coordinates",
			point:    [2]float64{-10.5, 20.25},
			expected: "-10.5 20.25",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SpaceFormatPoint(OutputSettings{}, tc.point)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestTabFormatPoint(t *testing.T) {
	tests := []struct {
		name     string
		point    [2]float64
		expected string
	}{
		{
			name:     "Zero coordinates",
			point:    [2]float64{0.0, 0.0},
			expected: "0\t0",
		},
		{
			name:     "Positive integers",
			point:    [2]float64{1.0, 2.0},
			expected: "1\t2",
		},
		{
			name:     "Decimal coordinates",
			point:    [2]float64{10.5, 20.25},
			expected: "10.5\t20.25",
		},
		{
			name:     "Negative coordinates",
			point:    [2]float64{-10.0, -20.0},
			expected: "-10\t-20",
		},
		{
			name:     "Mixed sign coordinates",
			point:    [2]float64{-10.5, 20.25},
			expected: "-10.5\t20.25",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TabFormatPoint(OutputSettings{}, tc.point)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestWktFormatPoint(t *testing.T) {
	tests := []struct {
		name     string
		point    [2]float64
		expected string
	}{
		{
			name:     "Zero coordinates",
			point:    [2]float64{0.0, 0.0},
			expected: "POINT (0 0)",
		},
		{
			name:     "Positive integers",
			point:    [2]float64{1.0, 2.0},
			expected: "POINT (1 2)",
		},
		{
			name:     "Decimal coordinates",
			point:    [2]float64{10.5, 20.25},
			expected: "POINT (10.5 20.25)",
		},
		{
			name:     "Negative coordinates",
			point:    [2]float64{-10.0, -20.0},
			expected: "POINT (-10 -20)",
		},
		{
			name:     "Mixed sign coordinates",
			point:    [2]float64{-10.5, 20.25},
			expected: "POINT (-10.5 20.25)",
		},
		{
			name:     "Large coordinates",
			point:    [2]float64{1000000.123, 2000000.456},
			expected: "POINT (1.000000123e+06 2.000000456e+06)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := WktFormatPoint(OutputSettings{}, tc.point)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestGeojsonFormatPoint(t *testing.T) {
	tests := []struct {
		name     string
		point    [2]float64
		expected string
	}{
		{
			name:     "Zero coordinates",
			point:    [2]float64{0.0, 0.0},
			expected: `{"type":"Point","coordinates":[0,0]}`,
		},
		{
			name:     "Positive integers",
			point:    [2]float64{1.0, 2.0},
			expected: `{"type":"Point","coordinates":[1,2]}`,
		},
		{
			name:     "Decimal coordinates",
			point:    [2]float64{10.5, 20.25},
			expected: `{"type":"Point","coordinates":[10.5,20.25]}`,
		},
		{
			name:     "Negative coordinates",
			point:    [2]float64{-10.0, -20.0},
			expected: `{"type":"Point","coordinates":[-10,-20]}`,
		},
		{
			name:     "Mixed sign coordinates",
			point:    [2]float64{-10.5, 20.25},
			expected: `{"type":"Point","coordinates":[-10.5,20.25]}`,
		},
		{
			name:     "Large coordinates",
			point:    [2]float64{1000000.123, 2000000.456},
			expected: `{"type":"Point","coordinates":[1000000.123,2000000.456]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GeojsonFormatPoint(OutputSettings{}, tc.point)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestGeojsonFormatPointStructure(t *testing.T) {
	point := [2]float64{1.0, 2.0}
	result, err := GeojsonFormatPoint(OutputSettings{}, point)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Run("Contains required GeoJSON fields", func(t *testing.T) {
		if !strings.Contains(result, `"type":"Point"`) {
			t.Error("Result should contain type field with Point value")
		}
		if !strings.Contains(result, `"coordinates"`) {
			t.Error("Result should contain coordinates field")
		}
	})

	t.Run("Coordinates are properly formatted", func(t *testing.T) {
		if !strings.Contains(result, `"coordinates":[1,2]`) {
			t.Error("Coordinates should be formatted as an array [x,y]")
		}
	})
}

func TestGetPointFormatter(t *testing.T) {
	tests := []struct {
		name       string
		formatType string
		expectNil  bool
	}{
		{
			name:       "Comma formatter",
			formatType: FormatComma,
			expectNil:  false,
		},
		{
			name:       "Space formatter",
			formatType: FormatSpace,
			expectNil:  false,
		},
		{
			name:       "Tab formatter",
			formatType: FormatTab,
			expectNil:  false,
		},
		{
			name:       "GeoJSON formatter",
			formatType: FormatGeoJson,
			expectNil:  false,
		},
		{
			name:       "WKT formatter",
			formatType: FormatWkt,
			expectNil:  false,
		},
		{
			name:       "Invalid formatter",
			formatType: "invalid",
			expectNil:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatter := GetPointFormatter(tc.formatType)

			if tc.expectNil && formatter != nil {
				t.Errorf("Expected nil formatter but got one")
			}
			if !tc.expectNil && formatter == nil {
				t.Errorf("Expected formatter but got nil")
			}

			// Test that the formatter works if it's not nil
			if !tc.expectNil && formatter != nil {
				point := [2]float64{1.0, 2.0}
				result, err := formatter(OutputSettings{}, point)
				if err != nil {
					t.Errorf("Formatter returned error: %v", err)
				}
				if result == "" {
					t.Errorf("Formatter returned empty result")
				}
			}
		})
	}
}

func TestFormatPoint(t *testing.T) {
	point := [2]float64{1.0, 2.0}

	tests := []struct {
		name        string
		formatType  string
		expected    string
		expectError bool
	}{
		{
			name:        "Comma format",
			formatType:  FormatComma,
			expected:    "1,2",
			expectError: false,
		},
		{
			name:        "Space format",
			formatType:  FormatSpace,
			expected:    "1 2",
			expectError: false,
		},
		{
			name:        "Tab format",
			formatType:  FormatTab,
			expected:    "1\t2",
			expectError: false,
		},
		{
			name:        "GeoJSON format",
			formatType:  FormatGeoJson,
			expected:    `{"type":"Point","coordinates":[1,2]}`,
			expectError: false,
		},
		{
			name:        "WKT format",
			formatType:  FormatWkt,
			expected:    "POINT (1 2)",
			expectError: false,
		},
		{
			name:        "Invalid format",
			formatType:  "invalid",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FormatPoint(point, OutputSettings{FormatType: tc.formatType})

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectError && result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}
