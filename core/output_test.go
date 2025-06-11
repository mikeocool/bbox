package core

import (
	"strings"
	"testing"
)

func TestTemplatedFormat(t *testing.T) {
	// Test with zero value Bbox
	t.Run("Zero value bbox", func(t *testing.T) {
		result, err := TemplatedFormat(OutputSettings{FormatDetails: "{{.Left}} {{.Bottom}} {{.Right}} {{.Top}}"}, Bbox{})
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
			result, err := TemplatedFormat(OutputSettings{FormatDetails: tc.template}, tc.bbox)

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
		result, err := TemplatedFormat(OutputSettings{FormatDetails: ""}, Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0})
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
			result, err := WktFormat(OutputSettings{}, tc.bbox)

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

func TestUrlFormat(t *testing.T) {
	tests := []struct {
		name        string
		urlType     string
		bbox        Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "osm - Zero value bbox",
			urlType:     "openstreetmap.com",
			bbox:        Bbox{},
			expected:    "https://www.openstreetmap.org/?box=yes&minlon=0&minlat=0&maxlon=0&maxlat=0",
			expectError: false,
		},
		{
			name:        "osm - Real world example (London)",
			urlType:     "openstreetmap.com",
			bbox:        Bbox{Left: -0.489, Bottom: 51.28, Right: 0.236, Top: 51.686},
			expected:    "https://www.openstreetmap.org/?box=yes&minlon=-0.489&minlat=51.28&maxlon=0.236&maxlat=51.686",
			expectError: false,
		},
		{
			name:        "osm - Case insenstive",
			urlType:     "OpenStreetMap.com",
			bbox:        Bbox{Left: -0.489, Bottom: 51.28, Right: 0.236, Top: 51.686},
			expected:    "https://www.openstreetmap.org/?box=yes&minlon=-0.489&minlat=51.28&maxlon=0.236&maxlat=51.686",
			expectError: false,
		},
		{
			name:        "osm - Large coordinates",
			urlType:     "openstreetmap.com",
			bbox:        Bbox{Left: -180, Bottom: -90, Right: 180, Top: 90},
			expected:    "https://www.openstreetmap.org/?box=yes&minlon=-180&minlat=-90&maxlon=180&maxlat=90",
			expectError: false,
		},
		{
			name:        "osm alias - Basic test",
			urlType:     "osm",
			bbox:        Bbox{Left: -0.489, Bottom: 51.28, Right: 0.236, Top: 51.686},
			expected:    "https://www.openstreetmap.org/?box=yes&minlon=-0.489&minlat=51.28&maxlon=0.236&maxlat=51.686",
			expectError: false,
		},
		{
			name:        "geojson.io - Zero value bbox",
			urlType:     "geojson.io",
			bbox:        Bbox{},
			expected:    "https://geojson.io/#data=data:application/json,%7B%22type%22%3A%22Polygon%22%2C%22coordinates%22%3A%5B%5B%5B0%2C0%5D%2C%5B0%2C0%5D%2C%5B0%2C0%5D%2C%5B0%2C0%5D%2C%5B0%2C0%5D%5D%5D%7D",
			expectError: false,
		},
		{
			name:        "geojson.io - Basic rectangle",
			urlType:     "geojson.io",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "https://geojson.io/#data=data:application/json,%7B%22type%22%3A%22Polygon%22%2C%22coordinates%22%3A%5B%5B%5B1%2C2%5D%2C%5B3%2C2%5D%2C%5B3%2C4%5D%2C%5B1%2C4%5D%2C%5B1%2C2%5D%5D%5D%7D",
			expectError: false,
		},
		{
			name:        "geojson.io - Real world example (New York City)",
			urlType:     "geojson.io",
			bbox:        Bbox{Left: -74.25909, Bottom: 40.477399, Right: -73.700181, Top: 40.916178},
			expected:    "https://geojson.io/#data=data:application/json,%7B%22type%22%3A%22Polygon%22%2C%22coordinates%22%3A%5B%5B%5B-74.25909%2C40.477399%5D%2C%5B-73.700181%2C40.477399%5D%2C%5B-73.700181%2C40.916178%5D%2C%5B-74.25909%2C40.916178%5D%2C%5B-74.25909%2C40.477399%5D%5D%5D%7D",
			expectError: false,
		},
		{
			name:        "geojson.io - Global extent (world bounds)",
			urlType:     "geojson.io",
			bbox:        Bbox{Left: -180, Bottom: -90, Right: 180, Top: 90},
			expected:    "https://geojson.io/#data=data:application/json,%7B%22type%22%3A%22Polygon%22%2C%22coordinates%22%3A%5B%5B%5B-180%2C-90%5D%2C%5B180%2C-90%5D%2C%5B180%2C90%5D%2C%5B-180%2C90%5D%2C%5B-180%2C-90%5D%5D%5D%7D",
			expectError: false,
		},
		{
			name:        "geojson.io - High precision decimal coordinates",
			urlType:     "geojson.io",
			bbox:        Bbox{Left: 10.123456789, Bottom: 20.987654321, Right: 30.111111111, Top: 40.999999999},
			expected:    "https://geojson.io/#data=data:application/json,%7B%22type%22%3A%22Polygon%22%2C%22coordinates%22%3A%5B%5B%5B10.123456789%2C20.987654321%5D%2C%5B30.111111111%2C20.987654321%5D%2C%5B30.111111111%2C40.999999999%5D%2C%5B10.123456789%2C40.999999999%5D%2C%5B10.123456789%2C20.987654321%5D%5D%5D%7D",
			expectError: false,
		},
		// Error cases
		{
			name:        "Error - Empty urlType",
			urlType:     "",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Error - Unknown urlType",
			urlType:     "unknown-provider",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Error - Invalid urlType",
			urlType:     "invalid.provider.com",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := UrlFormat(OutputSettings{FormatDetails: tc.urlType}, tc.bbox)

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
	result, err := WktFormat(OutputSettings{}, bbox)
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
			name:        "template format",
			formatType:  "go-template={{.Top}} {{.Right}} {{.Bottom}} {{.Left}}",
			expected:    "4 3 2 1",
			expectError: false,
		},
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
			formatType:  FormatGeoJson,
			expected:    `{"type":"Polygon","coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}`,
			expectError: false,
		},
		{
			name:        "WKT format",
			formatType:  FormatWkt,
			expected:    "POLYGON((1 2, 3 2, 3 4, 1 4, 1 2))",
			expectError: false,
		},
		{
			name:        "URL format",
			formatType:  "url=openstreetmap.org",
			expected:    "https://www.openstreetmap.org/?box=yes&minlon=1&minlat=2&maxlon=3&maxlat=4",
			expectError: false,
		},
		{
			name:        "URL missing type",
			formatType:  "url",
			expected:    "",
			expectError: true,
		},
		{
			name:        "URL missing type after =",
			formatType:  "url=",
			expected:    "",
			expectError: true,
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
			formatType, formatDetails := ParseFormat(tc.formatType)
			settings := OutputSettings{FormatType: formatType, FormatDetails: formatDetails}
			result, err := FormatBbox(bbox, settings)

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

func TestGetBboxFormatter(t *testing.T) {
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
			formatter := GetBboxFormatter(tc.formatType)

			if tc.expectNil && formatter != nil {
				t.Errorf("Expected nil formatter but got one")
			}
			if !tc.expectNil && formatter == nil {
				t.Errorf("Expected formatter but got nil")
			}

			// Test that the formatter works if it's not nil
			if !tc.expectNil && formatter != nil {
				bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}
				result, err := formatter(OutputSettings{}, bbox)
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
			result, err := CommaFormat(OutputSettings{}, tc.bbox)

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
			name:     "Basic rectangle",
			bbox:     Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected: `{"type":"Polygon","coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}`,
		},
		{
			name:     "Zero value bbox",
			bbox:     Bbox{},
			expected: `{"type":"Polygon","coordinates":[[[0,0],[0,0],[0,0],[0,0],[0,0]]]}`,
		},
		{
			name:     "Negative coordinates",
			bbox:     Bbox{Left: -10.0, Bottom: -20.0, Right: -5.0, Top: -15.0},
			expected: `{"type":"Polygon","coordinates":[[[-10,-20],[-5,-20],[-5,-15],[-10,-15],[-10,-20]]]}`,
		},
		{
			name:     "Mixed positive/negative coordinates",
			bbox:     Bbox{Left: -1.5, Bottom: -2.5, Right: 1.5, Top: 2.5},
			expected: `{"type":"Polygon","coordinates":[[[-1.5,-2.5],[1.5,-2.5],[1.5,2.5],[-1.5,2.5],[-1.5,-2.5]]]}`,
		},
		{
			name:     "Decimal coordinates",
			bbox:     Bbox{Left: 10.25, Bottom: 20.75, Right: 30.125, Top: 40.875},
			expected: `{"type":"Polygon","coordinates":[[[10.25,20.75],[30.125,20.75],[30.125,40.875],[10.25,40.875],[10.25,20.75]]]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GeojsonFormat(OutputSettings{}, tc.bbox)
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
	result, err := GeojsonFormat(OutputSettings{}, bbox)
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

func TestGeojsonFormatTypes(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}

	tests := []struct {
		name          string
		geojsonType   string
		expected      string
		expectError   bool
		shouldContain []string
	}{
		{
			name:        "coordinates type",
			geojsonType: "coordinates",
			expected:    `[[[1,2],[3,2],[3,4],[1,4],[1,2]]]`,
			expectError: false,
		},
		{
			name:        "geometry type explicit",
			geojsonType: "geometry",
			expected:    `{"type":"Polygon","coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}`,
			expectError: false,
		},
		{
			name:        "empty type defaults to geometry",
			geojsonType: "",
			expected:    `{"type":"Polygon","coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}`,
			expectError: false,
		},
		{
			name:        "feature type",
			geojsonType: "feature",
			expectError: false,
			shouldContain: []string{
				`"type":"Feature"`,
				`"geometry":{"type":"Polygon"`,
				`"coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]`,
			},
		},
		{
			name:        "featurecollection type",
			geojsonType: "featurecollection",
			expectError: false,
			shouldContain: []string{
				`"type":"FeatureCollection"`,
				`"features":[`,
				`"type":"Feature"`,
				`"geometry":{"type":"Polygon"`,
				`"coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]`,
			},
		},
		{
			name:        "unknown type defaults to featurecollection",
			geojsonType: "unknown",
			expectError: false,
			shouldContain: []string{
				`"type":"FeatureCollection"`,
				`"features":[`,
				`"type":"Feature"`,
			},
		},
		{
			name:        "case insensitive - COORDINATES",
			geojsonType: "COORDINATES",
			expected:    `[[[1,2],[3,2],[3,4],[1,4],[1,2]]]`,
			expectError: false,
		},
		{
			name:        "case insensitive - GEOMETRY",
			geojsonType: "GEOMETRY",
			expected:    `{"type":"Polygon","coordinates":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}`,
			expectError: false,
		},
		{
			name:        "case insensitive - FEATURE",
			geojsonType: "FEATURE",
			expectError: false,
			shouldContain: []string{
				`"type":"Feature"`,
				`"geometry":{"type":"Polygon"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			settings := OutputSettings{
				GeojsonType: tc.geojsonType,
			}
			result, err := GeojsonFormat(settings, bbox)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.expected != "" {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}

			if len(tc.shouldContain) > 0 {
				for _, substr := range tc.shouldContain {
					if !strings.Contains(result, substr) {
						t.Errorf("Result should contain %q, but got %q", substr, result)
					}
				}
			}
		})
	}
}

func TestGeojsonFormatWithIndent(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}

	tests := []struct {
		name          string
		indent        int
		geojsonType   string
		shouldContain []string
	}{
		{
			name:        "geometry with 2 space indent",
			indent:      2,
			geojsonType: "geometry",
			shouldContain: []string{
				"{\n  \"type\": \"Polygon\"",
				"\n  \"coordinates\": [",
			},
		},
		{
			name:        "feature with 4 space indent",
			indent:      4,
			geojsonType: "feature",
			shouldContain: []string{
				"{\n    \"type\": \"Feature\"",
				"\n    \"geometry\": {",
				"\n        \"type\": \"Polygon\"",
			},
		},
		{
			name:        "no indent",
			indent:      0,
			geojsonType: "geometry",
			shouldContain: []string{
				`{"type":"Polygon"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			settings := OutputSettings{
				GeojsonType:   tc.geojsonType,
				GeojsonIndent: tc.indent,
			}
			result, err := GeojsonFormat(settings, bbox)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			for _, substr := range tc.shouldContain {
				if !strings.Contains(result, substr) {
					t.Errorf("Result should contain %q, but got %q", substr, result)
				}
			}
		})
	}
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

func TestSpaceFormatCollection(t *testing.T) {
	tests := []struct {
		name        string
		boxes       []Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Empty collection",
			boxes:       []Bbox{},
			expected:    "",
			expectError: false,
		},
		{
			name:        "Single bbox - zero values",
			boxes:       []Bbox{{Left: 0, Bottom: 0, Right: 0, Top: 0}},
			expected:    "0 0 0 0",
			expectError: false,
		},
		{
			name:        "Single bbox - integer coordinates",
			boxes:       []Bbox{{Left: 1, Bottom: 2, Right: 3, Top: 4}},
			expected:    "1 2 3 4",
			expectError: false,
		},
		{
			name:        "Single bbox - decimal coordinates",
			boxes:       []Bbox{{Left: 10.5, Bottom: 20.25, Right: 30.75, Top: 40.125}},
			expected:    "10.5 20.25 30.75 40.125",
			expectError: false,
		},
		{
			name:        "Single bbox - negative coordinates",
			boxes:       []Bbox{{Left: -10, Bottom: -20, Right: -5, Top: -15}},
			expected:    "-10 -20 -5 -15",
			expectError: false,
		},
		{
			name: "Multiple bboxes",
			boxes: []Bbox{
				{Left: 1, Bottom: 2, Right: 3, Top: 4},
				{Left: 5, Bottom: 6, Right: 7, Top: 8},
				{Left: 9, Bottom: 10, Right: 11, Top: 12},
			},
			expected:    "1 2 3 4\n5 6 7 8\n9 10 11 12",
			expectError: false,
		},
		{
			name: "Multiple bboxes with mixed coordinate types",
			boxes: []Bbox{
				{Left: 0, Bottom: 0, Right: 0, Top: 0},
				{Left: -10.5, Bottom: 20.25, Right: -5.75, Top: 15.125},
				{Left: 100, Bottom: 200, Right: 300, Top: 400},
			},
			expected:    "0 0 0 0\n-10.5 20.25 -5.75 15.125\n100 200 300 400",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SpaceFormatCollection(OutputSettings{}, tc.boxes)

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
