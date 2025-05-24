package core

import (
	"strings"
	"testing"
)

func TestTemplatedFormat(t *testing.T) {
	// Test with zero value Bbox
	t.Run("Zero value bbox", func(t *testing.T) {
		result, err := TemplatedFormat("{{.Left}} {{.Bottom}} {{.Right}} {{.Top}}", Bbox{})
		if err != nil {
			t.Errorf("Unexpected error with zero value bbox: %v", err)
		}
		expected := "0 0 0 0"
		if strings.TrimSpace(result) != expected {
			t.Errorf("Expected %q but got %q", expected, result)
		}
	})

	tests := []struct {
		name        string
		template    string
		bbox        Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Basic format",
			template:    "{{.Left}} {{.Bottom}} {{.Right}} {{.Top}}",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "1 2 3 4",
			expectError: false,
		},
		{
			name:        "JSON format",
			template:    "{\"min_x\":{{.Left}},\"min_y\":{{.Bottom}},\"max_x\":{{.Right}},\"max_y\":{{.Top}}}",
			bbox:        Bbox{Left: 10.5, Bottom: 20.5, Right: 30.5, Top: 40.5},
			expected:    "{\"min_x\":10.5,\"min_y\":20.5,\"max_x\":30.5,\"max_y\":40.5}",
			expectError: false,
		},
		{
			name:        "With missing function",
			template:    "Width: {{.Right}} - {{.Left}} = {{sub .Right .Left}}, Height: {{.Top}} - {{.Bottom}} = {{sub .Top .Bottom}}",
			bbox:        Bbox{Left: 10, Bottom: 20, Right: 30, Top: 50},
			expected:    "",
			expectError: true, // Will error because the "sub" function is not defined
		},
		{
			name:        "Template execution error",
			template:    "{{if .NonExistentMethod.Call}}This will fail at execution time{{end}}",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "",
			expectError: true, // Will error during execution, not parsing
		},
		{
			name:        "Invalid template syntax",
			template:    "{{if .Left}}Only one closing bracket",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Malformed template",
			template:    "{{.Left} {{.Bottom}}",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "",
			expectError: true,
		},
		{
			name:        "With special characters",
			template:    "<bbox min=\"{{.Left}},{{.Bottom}}\" max=\"{{.Right}},{{.Top}}\"/>",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "<bbox min=\"1,2\" max=\"3,4\"/>",
			expectError: false,
		},
		{
			name:        "Mixed text and fields",
			template:    "Min: ({{.Left}}, {{.Bottom}}), Max: ({{.Right}}, {{.Top}})",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "Min: (1, 2), Max: (3, 4)",
			expectError: false,
		},
		{
			name:        "With formatting",
			template:    "{{printf \"%.2f\" .Left}} {{printf \"%.2f\" .Bottom}} {{printf \"%.2f\" .Right}} {{printf \"%.2f\" .Top}}",
			bbox:        Bbox{Left: 1.123, Bottom: 2.456, Right: 3.789, Top: 4.012},
			expected:    "1.12 2.46 3.79 4.01",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TemplatedFormat(tc.template, tc.bbox)

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
		result, err := TemplatedFormat("", Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0})
		if err != nil {
			t.Errorf("Unexpected error with empty template: %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty result for empty template, got %q", result)
		}
	})
}

func TestWktFormat(t *testing.T) {
	tests := []struct {
		name        string
		bbox        Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Zero value bbox",
			bbox:        Bbox{},
			expected:    "POLYGON((0 0, 0 0, 0 0, 0 0, 0 0))",
			expectError: false,
		},
		{
			name:        "Integer coordinates",
			bbox:        Bbox{Left: 1, Bottom: 2, Right: 3, Top: 4},
			expected:    "POLYGON((1 2, 3 2, 3 4, 1 4, 1 2))",
			expectError: false,
		},
		{
			name:        "Decimal coordinates",
			bbox:        Bbox{Left: 10.5, Bottom: 20.25, Right: 30.75, Top: 40.125},
			expected:    "POLYGON((10.5 20.25, 30.75 20.25, 30.75 40.125, 10.5 40.125, 10.5 20.25))",
			expectError: false,
		},
		{
			name:        "Negative coordinates",
			bbox:        Bbox{Left: -10, Bottom: -20, Right: -5, Top: -15},
			expected:    "POLYGON((-10 -20, -5 -20, -5 -15, -10 -15, -10 -20))",
			expectError: false,
		},
		{
			name:        "Mixed sign coordinates",
			bbox:        Bbox{Left: -10.5, Bottom: 20.25, Right: -5.75, Top: 15.125},
			expected:    "POLYGON((-10.5 20.25, -5.75 20.25, -5.75 15.125, -10.5 15.125, -10.5 20.25))",
			expectError: false,
		},
		{
			name:        "Large coordinates",
			bbox:        Bbox{Left: 1000000.123, Bottom: 2000000.456, Right: 3000000.789, Top: 4000000.012},
			expected:    "POLYGON((1.000000123e+06 2.000000456e+06, 3.000000789e+06 2.000000456e+06, 3.000000789e+06 4.000000012e+06, 1.000000123e+06 4.000000012e+06, 1.000000123e+06 2.000000456e+06))",
			expectError: false,
		},
		{
			name:        "Very small coordinates",
			bbox:        Bbox{Left: 0.0001, Bottom: 0.0002, Right: 0.0003, Top: 0.0004},
			expected:    "POLYGON((0.0001 0.0002, 0.0003 0.0002, 0.0003 0.0004, 0.0001 0.0004, 0.0001 0.0002))",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := WktFormat(tc.bbox)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestWktFormatStructure(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}
	result, err := WktFormat(bbox)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Run("Starts with POLYGON", func(t *testing.T) {
		if !strings.HasPrefix(result, "POLYGON") {
			t.Error("WKT result should start with 'POLYGON'")
		}
	})

	t.Run("Has proper parentheses structure", func(t *testing.T) {
		if !strings.HasPrefix(result, "POLYGON((") {
			t.Error("WKT result should start with 'POLYGON(('")
		}
		if !strings.HasSuffix(result, "))") {
			t.Error("WKT result should end with '))'")
		}
	})

	t.Run("Polygon is closed", func(t *testing.T) {
		// The first and last coordinate pairs should be the same
		if !strings.Contains(result, "1 2, 3 2, 3 4, 1 4, 1 2") {
			t.Error("WKT polygon should be closed (first and last coordinates should be the same)")
		}
	})

	t.Run("Has correct number of coordinate pairs", func(t *testing.T) {
		// Count the number of coordinate pairs by counting ", "
		coordSeparators := strings.Count(result, ", ")
		// Should have 4 separators for 5 coordinate pairs (4 corners + 1 closure)
		expected := 4
		if coordSeparators != expected {
			t.Errorf("Expected %d coordinate separators, got %d", expected, coordSeparators)
		}
	})

	t.Run("Coordinates are space-separated", func(t *testing.T) {
		// Check that coordinates within each pair are separated by spaces, not commas
		coords := strings.TrimPrefix(strings.TrimSuffix(result, "))"), "POLYGON((")
		pairs := strings.Split(coords, ", ")
		for i, pair := range pairs {
			parts := strings.Split(pair, " ")
			if len(parts) != 2 {
				t.Errorf("Coordinate pair %d should have exactly 2 space-separated values, got %d: %q", i, len(parts), pair)
			}
		}
	})
}

func TestFormat(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}

	tests := []struct {
		name        string
		formatType  string
		expected    string
		expectError bool
	}{
		{
			name:        "Comma format",
			formatType:  FormatComma,
			expected:    "1,2,3,4",
			expectError: false,
		},
		{
			name:        "Space format",
			formatType:  FormatSpace,
			expected:    "1 2 3 4",
			expectError: false,
		},
		{
			name:        "GeoJSON format",
			formatType:  FormatGeoJSON,
			expected:    `{"type":"Polygon","coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}`,
			expectError: false,
		},
		{
			name:        "WKT format",
			formatType:  FormatWKT,
			expected:    "POLYGON((1 2, 3 2, 3 4, 1 4, 1 2))",
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
			result, err := Format(bbox, tc.formatType)

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

func TestGetFormatter(t *testing.T) {
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
			name:       "GeoJSON formatter",
			formatType: FormatGeoJSON,
			expectNil:  false,
		},
		{
			name:       "WKT formatter",
			formatType: FormatWKT,
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
			formatter := GetFormatter(tc.formatType)

			if tc.expectNil && formatter != nil {
				t.Errorf("Expected nil formatter but got one")
			}
			if !tc.expectNil && formatter == nil {
				t.Errorf("Expected formatter but got nil")
			}

			// Test that the formatter works if it's not nil
			if !tc.expectNil && formatter != nil {
				bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}
				result, err := formatter(bbox)
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

func TestCommaFormat(t *testing.T) {
	tests := []struct {
		name        string
		bbox        Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Zero value bbox",
			bbox:        Bbox{},
			expected:    "0,0,0,0",
			expectError: false,
		},
		{
			name:        "Integer coordinates",
			bbox:        Bbox{Left: 1, Bottom: 2, Right: 3, Top: 4},
			expected:    "1,2,3,4",
			expectError: false,
		},
		{
			name:        "Decimal coordinates",
			bbox:        Bbox{Left: 10.5, Bottom: 20.25, Right: 30.75, Top: 40.125},
			expected:    "10.5,20.25,30.75,40.125",
			expectError: false,
		},
		{
			name:        "Negative coordinates",
			bbox:        Bbox{Left: -10, Bottom: -20, Right: -5, Top: -15},
			expected:    "-10,-20,-5,-15",
			expectError: false,
		},
		{
			name:        "Mixed sign coordinates",
			bbox:        Bbox{Left: -10.5, Bottom: 20.25, Right: -5.75, Top: 15.125},
			expected:    "-10.5,20.25,-5.75,15.125",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CommaFormat(tc.bbox)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestGeojsonFormat(t *testing.T) {
	tests := []struct {
		name     string
		bbox     Bbox
		expected string
	}{
		{
			name: "Basic rectangle",
			bbox: Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected: `{"type":"Polygon","coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}`,
		},
		{
			name: "Zero value bbox",
			bbox: Bbox{},
			expected: `{"type":"Polygon","coordinates":[[[0,0],[0,0],[0,0],[0,0],[0,0]]]}`,
		},
		{
			name: "Negative coordinates",
			bbox: Bbox{Left: -10.0, Bottom: -20.0, Right: -5.0, Top: -15.0},
			expected: `{"type":"Polygon","coordinates":[[[-10,-20],[-5,-20],[-5,-15],[-10,-15],[-10,-20]]]}`,
		},
		{
			name: "Mixed positive/negative coordinates",
			bbox: Bbox{Left: -1.5, Bottom: -2.5, Right: 1.5, Top: 2.5},
			expected: `{"type":"Polygon","coordinates":[[[-1.5,-2.5],[1.5,-2.5],[1.5,2.5],[-1.5,2.5],[-1.5,-2.5]]]}`,
		},
		{
			name: "Decimal coordinates",
			bbox: Bbox{Left: 10.25, Bottom: 20.75, Right: 30.125, Top: 40.875},
			expected: `{"type":"Polygon","coordinates":[[[10.25,20.75],[30.125,20.75],[30.125,40.875],[10.25,40.875],[10.25,20.75]]]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GeojsonFormat(tc.bbox)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestGeojsonFormatStructure(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}
	result, err := GeojsonFormat(bbox)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Run("Contains required GeoJSON fields", func(t *testing.T) {
		if !strings.Contains(result, `"type":"Polygon"`) {
			t.Error("Result should contain type field with Polygon value")
		}
		if !strings.Contains(result, `"coordinates"`) {
			t.Error("Result should contain coordinates field")
		}
	})

	t.Run("Coordinates are properly nested", func(t *testing.T) {
		// Should have three levels of brackets: [[[...]]]
		// One for the coordinates array, one for the polygon rings, one for the actual coordinates
		if !strings.Contains(result, `[[[`) {
			t.Error("Coordinates should be properly nested with three opening brackets")
		}
		if !strings.Contains(result, `]]]`) {
			t.Error("Coordinates should be properly nested with three closing brackets")
		}
	})

	t.Run("Polygon is closed", func(t *testing.T) {
		// The first and last coordinate pairs should be the same
		if !strings.Contains(result, `[1,2],[3,2],[3,4],[1,4],[1,2]`) {
			t.Error("Polygon should be closed (first and last coordinates should be the same)")
		}
	})

	t.Run("Has correct number of coordinate pairs", func(t *testing.T) {
		// Count the number of coordinate pairs by counting "],["
		coordSeparators := strings.Count(result, "],[")
		// Should have 4 separators for 5 coordinate pairs (4 corners + 1 closure)
		expected := 4
		if coordSeparators != expected {
			t.Errorf("Expected %d coordinate separators, got %d", expected, coordSeparators)
		}
	})
}