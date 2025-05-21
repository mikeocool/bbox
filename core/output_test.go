package core

import (
	"strings"
	"testing"
)

func TestTemplatedFormat(t *testing.T) {
	// Test with zero value Bbox
	t.Run("Zero value bbox", func(t *testing.T) {
		result, err := TemplatedFormat("{{.MinX}} {{.MinY}} {{.MaxX}} {{.MaxY}}", Bbox{})
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
			template:    "{{.MinX}} {{.MinY}} {{.MaxX}} {{.MaxY}}",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0},
			expected:    "1 2 3 4",
			expectError: false,
		},
		{
			name:        "JSON format",
			template:    "{\"min_x\":{{.MinX}},\"min_y\":{{.MinY}},\"max_x\":{{.MaxX}},\"max_y\":{{.MaxY}}}",
			bbox:        Bbox{MinX: 10.5, MinY: 20.5, MaxX: 30.5, MaxY: 40.5},
			expected:    "{\"min_x\":10.5,\"min_y\":20.5,\"max_x\":30.5,\"max_y\":40.5}",
			expectError: false,
		},
		{
			name:        "With missing function",
			template:    "Width: {{.MaxX}} - {{.MinX}} = {{sub .MaxX .MinX}}, Height: {{.MaxY}} - {{.MinY}} = {{sub .MaxY .MinY}}",
			bbox:        Bbox{MinX: 10, MinY: 20, MaxX: 30, MaxY: 50},
			expected:    "",
			expectError: true, // Will error because the "sub" function is not defined
		},
		{
			name:        "Template execution error",
			template:    "{{if .NonExistentMethod.Call}}This will fail at execution time{{end}}",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0},
			expected:    "",
			expectError: true, // Will error during execution, not parsing
		},
		{
			name:        "Invalid template syntax",
			template:    "{{if .MinX}}Only one closing bracket",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Malformed template",
			template:    "{{.MinX} {{.MinY}}",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0},
			expected:    "",
			expectError: true,
		},
		{
			name:        "With special characters",
			template:    "<bbox min=\"{{.MinX}},{{.MinY}}\" max=\"{{.MaxX}},{{.MaxY}}\"/>",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0},
			expected:    "<bbox min=\"1,2\" max=\"3,4\"/>",
			expectError: false,
		},
		{
			name:        "Mixed text and fields",
			template:    "Min: ({{.MinX}}, {{.MinY}}), Max: ({{.MaxX}}, {{.MaxY}})",
			bbox:        Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0},
			expected:    "Min: (1, 2), Max: (3, 4)",
			expectError: false,
		},
		{
			name:        "With formatting",
			template:    "{{printf \"%.2f\" .MinX}} {{printf \"%.2f\" .MinY}} {{printf \"%.2f\" .MaxX}} {{printf \"%.2f\" .MaxY}}",
			bbox:        Bbox{MinX: 1.123, MinY: 2.456, MaxX: 3.789, MaxY: 4.012},
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
		result, err := TemplatedFormat("", Bbox{MinX: 1.0, MinY: 2.0, MaxX: 3.0, MaxY: 4.0})
		if err != nil {
			t.Errorf("Unexpected error with empty template: %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty result for empty template, got %q", result)
		}
	})
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
			bbox:        Bbox{MinX: 1, MinY: 2, MaxX: 3, MaxY: 4},
			expected:    "1,2,3,4",
			expectError: false,
		},
		{
			name:        "Decimal coordinates",
			bbox:        Bbox{MinX: 10.5, MinY: 20.25, MaxX: 30.75, MaxY: 40.125},
			expected:    "10.5,20.25,30.75,40.125",
			expectError: false,
		},
		{
			name:        "Negative coordinates",
			bbox:        Bbox{MinX: -10, MinY: -20, MaxX: -5, MaxY: -15},
			expected:    "-10,-20,-5,-15",
			expectError: false,
		},
		{
			name:        "Mixed sign coordinates",
			bbox:        Bbox{MinX: -10.5, MinY: 20.25, MaxX: -5.75, MaxY: 15.125},
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