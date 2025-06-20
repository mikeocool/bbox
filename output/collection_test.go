package output

import (
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestSpaceFormatCollection(t *testing.T) {
	tests := []struct {
		name        string
		boxes       []core.Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Empty collection",
			boxes:       []core.Bbox{},
			expected:    "",
			expectError: false,
		},
		{
			name:        "Single bbox - zero values",
			boxes:       []core.Bbox{{Left: 0, Bottom: 0, Right: 0, Top: 0}},
			expected:    "0 0 0 0",
			expectError: false,
		},
		{
			name:        "Single bbox - integer coordinates",
			boxes:       []core.Bbox{{Left: 1, Bottom: 2, Right: 3, Top: 4}},
			expected:    "1 2 3 4",
			expectError: false,
		},
		{
			name:        "Single bbox - decimal coordinates",
			boxes:       []core.Bbox{{Left: 10.5, Bottom: 20.25, Right: 30.75, Top: 40.125}},
			expected:    "10.5 20.25 30.75 40.125",
			expectError: false,
		},
		{
			name:        "Single bbox - negative coordinates",
			boxes:       []core.Bbox{{Left: -10, Bottom: -20, Right: -5, Top: -15}},
			expected:    "-10 -20 -5 -15",
			expectError: false,
		},
		{
			name: "Multiple bboxes",
			boxes: []core.Bbox{
				{Left: 1, Bottom: 2, Right: 3, Top: 4},
				{Left: 5, Bottom: 6, Right: 7, Top: 8},
				{Left: 9, Bottom: 10, Right: 11, Top: 12},
			},
			expected:    "1 2 3 4\n5 6 7 8\n9 10 11 12",
			expectError: false,
		},
		{
			name: "Multiple bboxes with mixed coordinate types",
			boxes: []core.Bbox{
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

func TestWktFormatCollection(t *testing.T) {
	tests := []struct {
		name     string
		boxes    []core.Bbox
		settings OutputSettings
		expected string
	}{
		{
			name:     "empty collection",
			boxes:    []core.Bbox{},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION()",
		},
		{
			name: "single bbox",
			boxes: []core.Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION(POLYGON((1 2, 3 2, 3 4, 1 4, 1 2)))",
		},
		{
			name: "two bboxes",
			boxes: []core.Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
				{Left: 5.0, Bottom: 6.0, Right: 7.0, Top: 8.0},
			},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION(POLYGON((1 2, 3 2, 3 4, 1 4, 1 2)),\nPOLYGON((5 6, 7 6, 7 8, 5 8, 5 6)))",
		},
		{
			name: "three bboxes with decimal coordinates",
			boxes: []core.Bbox{
				{Left: 1.5, Bottom: 2.5, Right: 3.5, Top: 4.5},
				{Left: 10.25, Bottom: 20.75, Right: 30.125, Top: 40.875},
				{Left: -1.0, Bottom: -2.0, Right: -0.5, Top: -1.5},
			},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION(POLYGON((1.5 2.5, 3.5 2.5, 3.5 4.5, 1.5 4.5, 1.5 2.5)),\nPOLYGON((10.25 20.75, 30.125 20.75, 30.125 40.875, 10.25 40.875, 10.25 20.75)),\nPOLYGON((-1 -2, -0.5 -2, -0.5 -1.5, -1 -1.5, -1 -2)))",
		},
		{
			name: "bbox with zero coordinates",
			boxes: []core.Bbox{
				{Left: 0.0, Bottom: 0.0, Right: 1.0, Top: 1.0},
			},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION(POLYGON((0 0, 1 0, 1 1, 0 1, 0 0)))",
		},
		{
			name: "bbox with negative coordinates",
			boxes: []core.Bbox{
				{Left: -10.0, Bottom: -20.0, Right: -5.0, Top: -15.0},
			},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION(POLYGON((-10 -20, -5 -20, -5 -15, -10 -15, -10 -20)))",
		},
		{
			name: "bbox with very large coordinates",
			boxes: []core.Bbox{
				{Left: 1000000.0, Bottom: 2000000.0, Right: 3000000.0, Top: 4000000.0},
			},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION(POLYGON((1e+06 2e+06, 3e+06 2e+06, 3e+06 4e+06, 1e+06 4e+06, 1e+06 2e+06)))",
		},
		{
			name: "bbox with very small coordinates",
			boxes: []core.Bbox{
				{Left: 0.000001, Bottom: 0.000002, Right: 0.000003, Top: 0.000004},
			},
			settings: OutputSettings{},
			expected: "GEOMETRYCOLLECTION(POLYGON((1e-06 2e-06, 3e-06 2e-06, 3e-06 4e-06, 1e-06 4e-06, 1e-06 2e-06)))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WktFormatCollection(tt.settings, tt.boxes)
			if err != nil {
				t.Errorf("WktFormatCollection() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("WktFormatCollection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWktFormatCollectionSettings(t *testing.T) {
	// Test that OutputSettings parameter doesn't affect the output
	// (since WktFormatCollection doesn't use any settings currently)
	boxes := []core.Bbox{
		{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
	}

	settings1 := OutputSettings{
		FormatType:    "wkt",
		FormatDetails: "some details",
		GeojsonIndent: 2,
		GeojsonType:   "feature",
	}

	settings2 := OutputSettings{}

	result1, err1 := WktFormatCollection(settings1, boxes)
	result2, err2 := WktFormatCollection(settings2, boxes)

	if err1 != nil || err2 != nil {
		t.Errorf("WktFormatCollection() unexpected errors: %v, %v", err1, err2)
	}

	if result1 != result2 {
		t.Errorf("WktFormatCollection() results should be identical regardless of settings: %v != %v", result1, result2)
	}
}

func TestTemplatedFormatCollection(t *testing.T) {
	tests := []struct {
		name        string
		settings    OutputSettings
		boxes       []core.Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Empty collection with simple template",
			settings:    OutputSettings{FormatDetails: "{{ range . }}{{ .Left }},{{ .Bottom }}{{ end }}"},
			boxes:       []core.Bbox{},
			expected:    "",
			expectError: false,
		},
		{
			name:        "Single bbox - access fields",
			settings:    OutputSettings{FormatDetails: "{{ range . }}{{ .Left }},{{ .Bottom }},{{ .Right }},{{ .Top }}{{ end }}"},
			boxes:       []core.Bbox{{Left: 1, Bottom: 2, Right: 3, Top: 4}},
			expected:    "1,2,3,4",
			expectError: false,
		},
		{
			name:        "Single bbox - formatted output",
			settings:    OutputSettings{FormatDetails: "{{ range . }}Box: [{{ printf \"%.2f\" .Left }}, {{ printf \"%.2f\" .Bottom }}, {{ printf \"%.2f\" .Right }}, {{ printf \"%.2f\" .Top }}]{{ end }}"},
			boxes:       []core.Bbox{{Left: 1.123, Bottom: 2.456, Right: 3.789, Top: 4.012}},
			expected:    "Box: [1.12, 2.46, 3.79, 4.01]",
			expectError: false,
		},
		{
			name:     "Multiple bboxes - iterate with range",
			settings: OutputSettings{FormatDetails: "Boxes:\n{{ range . }}  - ({{ .Left }},{{ .Bottom }},{{ .Right }},{{ .Top }})\n{{ end }}"},
			boxes: []core.Bbox{
				{Left: 1, Bottom: 2, Right: 3, Top: 4},
				{Left: 5, Bottom: 6, Right: 7, Top: 8},
			},
			expected:    "Boxes:\n  - (1,2,3,4)\n  - (5,6,7,8)\n",
			expectError: false,
		},
		{
			name:     "Multiple bboxes - formatted JSON-like output",
			settings: OutputSettings{FormatDetails: "[\n{{ range $i, $box := . }}  {{ if $i }},{{ end }}  {\n    \"left\": {{ .Left }},\n    \"bottom\": {{ .Bottom }},\n    \"right\": {{ .Right }},\n    \"top\": {{ .Top }}\n  }\n{{ end }}\n]"},
			boxes: []core.Bbox{
				{Left: 10, Bottom: 20, Right: 30, Top: 40},
				{Left: 50, Bottom: 60, Right: 70, Top: 80},
			},
			expected:    "[\n    {\n    \"left\": 10,\n    \"bottom\": 20,\n    \"right\": 30,\n    \"top\": 40\n  }\n  ,  {\n    \"left\": 50,\n    \"bottom\": 60,\n    \"right\": 70,\n    \"top\": 80\n  }\n\n]",
			expectError: false,
		},
		{
			name:        "Invalid template syntax",
			settings:    OutputSettings{FormatDetails: "{{ .Left "},
			boxes:       []core.Bbox{{Left: 1, Bottom: 2, Right: 3, Top: 4}},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid template field reference",
			settings:    OutputSettings{FormatDetails: "{{ range . }}{{ .InvalidField }}{{ end }}"},
			boxes:       []core.Bbox{{Left: 1, Bottom: 2, Right: 3, Top: 4}},
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TemplatedFormatCollection(tc.settings, tc.boxes)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError && err == nil {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestGeojsonlFormatCollection(t *testing.T) {
	tests := []struct {
		name        string
		settings    OutputSettings
		boxes       []core.Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Empty collection",
			settings:    OutputSettings{GeojsonIndent: 0},
			boxes:       []core.Bbox{},
			expected:    "",
			expectError: false,
		},
		{
			name:        "Single bbox",
			settings:    OutputSettings{GeojsonIndent: 0},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}},
			expected:    "{\"type\":\"Polygon\",\"coordinates\":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}",
			expectError: false,
		},
		{
			name:     "Multiple bboxes with decimal coordinates",
			settings: OutputSettings{GeojsonIndent: 0},
			boxes: []core.Bbox{
				{Left: 1.5, Bottom: 2.5, Right: 3.5, Top: 4.5},
				{Left: 10.25, Bottom: 20.75, Right: 30.125, Top: 40.875},
				{Left: -1.0, Bottom: -2.0, Right: -0.5, Top: -1.5},
			},
			expected:    "{\"type\":\"Polygon\",\"coordinates\":[[[1.5,2.5],[3.5,2.5],[3.5,4.5],[1.5,4.5],[1.5,2.5]]]}\n{\"type\":\"Polygon\",\"coordinates\":[[[10.25,20.75],[30.125,20.75],[30.125,40.875],[10.25,40.875],[10.25,20.75]]]}\n{\"type\":\"Polygon\",\"coordinates\":[[[-1,-2],[-0.5,-2],[-0.5,-1.5],[-1,-1.5],[-1,-2]]]}",
			expectError: false,
		},
		{
			name:        "Error when indentation is requested",
			settings:    OutputSettings{GeojsonIndent: 2},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Multiple bboxes - coordinates type",
			settings:    OutputSettings{GeojsonIndent: 0, GeojsonType: "coordinates"},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}, {Left: 5.0, Bottom: 6.0, Right: 7.0, Top: 8.0}},
			expected:    "[[[1,2],[3,2],[3,4],[1,4],[1,2]]]\n[[[5,6],[7,6],[7,8],[5,8],[5,6]]]",
			expectError: false,
		},
		{
			name:        "Multiple bboxes - feature type",
			settings:    OutputSettings{GeojsonIndent: 0, GeojsonType: "feature"},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}, {Left: 5.0, Bottom: 6.0, Right: 7.0, Top: 8.0}},
			expected:    "{\"type\":\"Feature\",\"geometry\":{\"type\":\"Polygon\",\"coordinates\":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}}\n{\"type\":\"Feature\",\"geometry\":{\"type\":\"Polygon\",\"coordinates\":[[[5,6],[7,6],[7,8],[5,8],[5,6]]]}}",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GeojsonlFormatCollection(tc.settings, tc.boxes)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError && err == nil {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}

			// Check specific error message for indentation error
			if tc.expectError && tc.settings.GeojsonIndent != 0 && err != nil {
				expectedErrMsg := "GeoJSONL format does not support indentation"
				if err.Error() != expectedErrMsg {
					t.Errorf("Expected error message %q but got %q", expectedErrMsg, err.Error())
				}
			}
		})
	}
}

func TestFormatCollectionGeojsonl(t *testing.T) {
	tests := []struct {
		name        string
		settings    OutputSettings
		boxes       []core.Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Single bbox via FormatCollection",
			settings:    OutputSettings{FormatType: FormatGeoJsonl, GeojsonIndent: 0},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}},
			expected:    "{\"type\":\"Polygon\",\"coordinates\":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}",
			expectError: false,
		},
		{
			name:     "Multiple bboxes via FormatCollection",
			settings: OutputSettings{FormatType: FormatGeoJsonl, GeojsonIndent: 0},
			boxes: []core.Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
				{Left: 5.0, Bottom: 6.0, Right: 7.0, Top: 8.0},
			},
			expected:    "{\"type\":\"Polygon\",\"coordinates\":[[[1,2],[3,2],[3,4],[1,4],[1,2]]]}\n{\"type\":\"Polygon\",\"coordinates\":[[[5,6],[7,6],[7,8],[5,8],[5,6]]]}",
			expectError: false,
		},
		{
			name:        "Error with indentation via FormatCollection",
			settings:    OutputSettings{FormatType: FormatGeoJsonl, GeojsonIndent: 2},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}},
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FormatCollection(tc.boxes, tc.settings)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError && err == nil {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestJsonFormatCollection(t *testing.T) {
	tests := []struct {
		name        string
		settings    OutputSettings
		boxes       []core.Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Single bbox via FormatCollection",
			settings:    OutputSettings{FormatType: FormatJson},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}},
			expected:    "[[1,2,3,4]]",
			expectError: false,
		},
		{
			name:     "Multiple bboxes via FormatCollection",
			settings: OutputSettings{FormatType: FormatJson},
			boxes: []core.Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
				{Left: 5.0, Bottom: 6.0, Right: 7.0, Top: 8.0},
			},
			expected:    "[[1,2,3,4],[5,6,7,8]]",
			expectError: false,
		},
		{
			name:        "Empty collection via FormatCollection",
			settings:    OutputSettings{FormatType: FormatJson},
			boxes:       []core.Bbox{},
			expected:    "[]",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := JsonFormatCollection(tc.settings, tc.boxes)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError && err == nil {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestJsonlFormatCollection(t *testing.T) {
	tests := []struct {
		name        string
		settings    OutputSettings
		boxes       []core.Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Empty collection",
			settings:    OutputSettings{},
			boxes:       []core.Bbox{},
			expected:    "",
			expectError: false,
		},
		{
			name:        "Single bbox",
			settings:    OutputSettings{},
			boxes:       []core.Bbox{{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}},
			expected:    "[1,2,3,4]",
			expectError: false,
		},
		{
			name:     "Multiple bboxes",
			settings: OutputSettings{},
			boxes: []core.Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
				{Left: 5.0, Bottom: 6.0, Right: 7.0, Top: 8.0},
			},
			expected:    "[1,2,3,4]\n[5,6,7,8]",
			expectError: false,
		},
		{
			name:     "Multiple bboxes with decimal coordinates",
			settings: OutputSettings{},
			boxes: []core.Bbox{
				{Left: 1.5, Bottom: 2.5, Right: 3.5, Top: 4.5},
				{Left: 10.25, Bottom: 20.75, Right: 30.125, Top: 40.875},
				{Left: -1.0, Bottom: -2.0, Right: -0.5, Top: -1.5},
			},
			expected:    "[1.5,2.5,3.5,4.5]\n[10.25,20.75,30.125,40.875]\n[-1,-2,-0.5,-1.5]",
			expectError: false,
		},
		{
			name:        "Zero coordinates",
			settings:    OutputSettings{},
			boxes:       []core.Bbox{{Left: 0, Bottom: 0, Right: 0, Top: 0}},
			expected:    "[0,0,0,0]",
			expectError: false,
		},
		{
			name:        "Negative coordinates",
			settings:    OutputSettings{},
			boxes:       []core.Bbox{{Left: -10, Bottom: -20, Right: -5, Top: -15}},
			expected:    "[-10,-20,-5,-15]",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := JsonlFormatCollection(tc.settings, tc.boxes)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Only check result if we don't expect an error
			if !tc.expectError && err == nil {
				if result != tc.expected {
					t.Errorf("Expected %q but got %q", tc.expected, result)
				}
			}
		})
	}
}

func TestWkbhexFormatCollection(t *testing.T) {
	tests := []struct {
		name     string
		boxes    []core.Bbox
		settings OutputSettings
		expected string
	}{
		{
			name:     "empty collection",
			boxes:    []core.Bbox{},
			settings: OutputSettings{},
			expected: "010700000000000000",
		},
		{
			name: "single bbox",
			boxes: []core.Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			},
			settings: OutputSettings{},
			expected: "01070000000100000001030000000100000005000000000000000000F03F00000000000000400000000000000840000000000000004000000000000008400000000000001040000000000000F03F0000000000001040000000000000F03F0000000000000040",
		},
		{
			name: "two bboxes",
			boxes: []core.Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
				{Left: 5.0, Bottom: 6.0, Right: 7.0, Top: 8.0},
			},
			settings: OutputSettings{},
			expected: "01070000000200000001030000000100000005000000000000000000F03F00000000000000400000000000000840000000000000004000000000000008400000000000001040000000000000F03F0000000000001040000000000000F03F000000000000004001030000000100000005000000000000000000144000000000000018400000000000001C4000000000000018400000000000001C4000000000000020400000000000001440000000000000204000000000000014400000000000001840",
		},
		{
			name: "three bboxes with decimal coordinates",
			boxes: []core.Bbox{
				{Left: 1.5, Bottom: 2.5, Right: 3.5, Top: 4.5},
				{Left: 10.25, Bottom: 20.75, Right: 30.125, Top: 40.875},
				{Left: -1.0, Bottom: -2.0, Right: -0.5, Top: -1.5},
			},
			settings: OutputSettings{},
			expected: "01070000000300000001030000000100000005000000000000000000F83F00000000000004400000000000000C4000000000000004400000000000000C400000000000001240000000000000F83F0000000000001240000000000000F83F00000000000004400103000000010000000500000000000000008024400000000000C034400000000000203E400000000000C034400000000000203E4000000000007044400000000000802440000000000070444000000000008024400000000000C0344001030000000100000005000000000000000000F0BF00000000000000C0000000000000E0BF00000000000000C0000000000000E0BF000000000000F8BF000000000000F0BF000000000000F8BF000000000000F0BF00000000000000C0",
		},
		{
			name: "bbox with zero coordinates",
			boxes: []core.Bbox{
				{Left: 0.0, Bottom: 0.0, Right: 1.0, Top: 1.0},
			},
			settings: OutputSettings{},
			expected: "0107000000010000000103000000010000000500000000000000000000000000000000000000000000000000F03F0000000000000000000000000000F03F000000000000F03F0000000000000000000000000000F03F00000000000000000000000000000000",
		},
		{
			name: "bbox with negative coordinates",
			boxes: []core.Bbox{
				{Left: -10.0, Bottom: -20.0, Right: -5.0, Top: -15.0},
			},
			settings: OutputSettings{},
			expected: "0107000000010000000103000000010000000500000000000000000024C000000000000034C000000000000014C000000000000034C000000000000014C00000000000002EC000000000000024C00000000000002EC000000000000024C000000000000034C0",
		},
		{
			name: "bbox with very large coordinates",
			boxes: []core.Bbox{
				{Left: 1000000.0, Bottom: 2000000.0, Right: 3000000.0, Top: 4000000.0},
			},
			settings: OutputSettings{},
			expected: "010700000001000000010300000001000000050000000000000080842E410000000080843E410000000060E346410000000080843E410000000060E346410000000080844E410000000080842E410000000080844E410000000080842E410000000080843E41",
		},
		{
			name: "bbox with very small coordinates",
			boxes: []core.Bbox{
				{Left: 0.000001, Bottom: 0.000002, Right: 0.000003, Top: 0.000004},
			},
			settings: OutputSettings{},
			expected: "010700000001000000010300000001000000050000008DEDB5A0F7C6B03E8DEDB5A0F7C6C03E54E41071732AC93E8DEDB5A0F7C6C03E54E41071732AC93E8DEDB5A0F7C6D03E8DEDB5A0F7C6B03E8DEDB5A0F7C6D03E8DEDB5A0F7C6B03E8DEDB5A0F7C6C03E",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WkbhexFormatCollection(tt.settings, tt.boxes)
			if err != nil {
				t.Errorf("WkbhexFormatCollection() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("WkbhexFormatCollection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWkbhexFormatCollectionSettings(t *testing.T) {
	// Test that the function works with different OutputSettings
	boxes := []core.Bbox{
		{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
	}

	settings := OutputSettings{
		FormatType:    FormatWkbhex,
		GeojsonIndent: 2, // This should be ignored for WKB format
	}

	result, err := WkbhexFormatCollection(settings, boxes)
	if err != nil {
		t.Errorf("WkbhexFormatCollection() error = %v", err)
		return
	}

	expected := "01070000000100000001030000000100000005000000000000000000F03F00000000000000400000000000000840000000000000004000000000000008400000000000001040000000000000F03F0000000000001040000000000000F03F0000000000000040"
	if result != expected {
		t.Errorf("WkbhexFormatCollection() = %v, want %v", result, expected)
	}
}

func TestGetCollectionFormatterWkbhex(t *testing.T) {
	// Test that the WkbhexFormatCollection function is properly registered
	formatter, err := GetCollectionFormatter(FormatWkbhex)
	if err != nil {
		t.Errorf("GetCollectionFormatter(FormatWkbhex) error = %v", err)
		return
	}
	if formatter == nil {
		t.Errorf("GetCollectionFormatter(FormatWkbhex) returned nil formatter")
		return
	}

	// Test that it actually works with a simple case
	boxes := []core.Bbox{
		{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
	}
	settings := OutputSettings{}

	result, err := formatter(settings, boxes)
	if err != nil {
		t.Errorf("formatter(settings, boxes) error = %v", err)
		return
	}

	expected := "01070000000100000001030000000100000005000000000000000000F03F00000000000000400000000000000840000000000000004000000000000008400000000000001040000000000000F03F0000000000001040000000000000F03F0000000000000040"
	if result != expected {
		t.Errorf("formatter(settings, boxes) = %v, want %v", result, expected)
	}
}

func TestFormatCollectionWkbhex(t *testing.T) {
	tests := []struct {
		name        string
		settings    OutputSettings
		boxes       []core.Bbox
		expected    string
		expectError bool
	}{
		{
			name:        "Single bbox via FormatCollection",
			settings:    OutputSettings{FormatType: FormatWkbhex},
			boxes:       []core.Bbox{{Left: 1, Bottom: 2, Right: 3, Top: 4}},
			expected:    "01070000000100000001030000000100000005000000000000000000F03F00000000000000400000000000000840000000000000004000000000000008400000000000001040000000000000F03F0000000000001040000000000000F03F0000000000000040",
			expectError: false,
		},
		{
			name:     "Multiple bboxes via FormatCollection",
			settings: OutputSettings{FormatType: FormatWkbhex},
			boxes: []core.Bbox{
				{Left: 1, Bottom: 2, Right: 3, Top: 4},
				{Left: 5, Bottom: 6, Right: 7, Top: 8},
			},
			expected:    "01070000000200000001030000000100000005000000000000000000F03F00000000000000400000000000000840000000000000004000000000000008400000000000001040000000000000F03F0000000000001040000000000000F03F000000000000004001030000000100000005000000000000000000144000000000000018400000000000001C4000000000000018400000000000001C4000000000000020400000000000001440000000000000204000000000000014400000000000001840",
			expectError: false,
		},
		{
			name:        "Empty collection via FormatCollection",
			settings:    OutputSettings{FormatType: FormatWkbhex},
			boxes:       []core.Bbox{},
			expected:    "010700000000000000",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatCollection(tt.boxes, tt.settings)
			if (err != nil) != tt.expectError {
				t.Errorf("FormatCollection() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if result != tt.expected {
				t.Errorf("FormatCollection() = %v, want %v", result, tt.expected)
			}
		})
	}
}
