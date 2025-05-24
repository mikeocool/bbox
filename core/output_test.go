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