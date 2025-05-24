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

func TestBboxPolygon(t *testing.T) {
	tests := []struct {
		name     string
		bbox     Bbox
		expected [][2]float64
	}{
		{
			name: "Basic rectangle",
			bbox: Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expected: [][2]float64{
				{1.0, 2.0}, // bottom-left
				{3.0, 2.0}, // bottom-right
				{3.0, 4.0}, // top-right
				{1.0, 4.0}, // top-left
				{1.0, 2.0}, // bottom-left (closed)
			},
		},
		{
			name: "Zero value bbox",
			bbox: Bbox{},
			expected: [][2]float64{
				{0.0, 0.0}, // bottom-left
				{0.0, 0.0}, // bottom-right
				{0.0, 0.0}, // top-right
				{0.0, 0.0}, // top-left
				{0.0, 0.0}, // bottom-left (closed)
			},
		},
		{
			name: "Negative coordinates",
			bbox: Bbox{Left: -10.0, Bottom: -20.0, Right: -5.0, Top: -15.0},
			expected: [][2]float64{
				{-10.0, -20.0}, // bottom-left
				{-5.0, -20.0},  // bottom-right
				{-5.0, -15.0},  // top-right
				{-10.0, -15.0}, // top-left
				{-10.0, -20.0}, // bottom-left (closed)
			},
		},
		{
			name: "Mixed positive/negative coordinates",
			bbox: Bbox{Left: -1.5, Bottom: -2.5, Right: 1.5, Top: 2.5},
			expected: [][2]float64{
				{-1.5, -2.5}, // bottom-left
				{1.5, -2.5},  // bottom-right
				{1.5, 2.5},   // top-right
				{-1.5, 2.5},  // top-left
				{-1.5, -2.5}, // bottom-left (closed)
			},
		},
		{
			name: "Decimal coordinates",
			bbox: Bbox{Left: 10.25, Bottom: 20.75, Right: 30.125, Top: 40.875},
			expected: [][2]float64{
				{10.25, 20.75},  // bottom-left
				{30.125, 20.75}, // bottom-right
				{30.125, 40.875}, // top-right
				{10.25, 40.875},  // top-left
				{10.25, 20.75},   // bottom-left (closed)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.bbox.Polygon()

			// Check that we get exactly 5 points (4 corners + closure)
			if len(result) != 5 {
				t.Errorf("Expected 5 coordinates, got %d", len(result))
			}

			// Check each coordinate
			for i, coord := range result {
				if coord != tc.expected[i] {
					t.Errorf("Coordinate %d: expected %v, got %v", i, tc.expected[i], coord)
				}
			}

			// Verify the polygon is properly closed (first and last points are the same)
			if len(result) > 0 && result[0] != result[len(result)-1] {
				t.Errorf("Polygon is not closed: first point %v != last point %v", result[0], result[len(result)-1])
			}
		})
	}
}

func TestBboxPolygonProperties(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0}
	coords := bbox.Polygon()

	t.Run("Polygon is closed", func(t *testing.T) {
		if len(coords) < 2 {
			t.Fatal("Polygon must have at least 2 coordinates")
		}
		if coords[0] != coords[len(coords)-1] {
			t.Errorf("Polygon is not closed: first %v != last %v", coords[0], coords[len(coords)-1])
		}
	})

	t.Run("Has correct number of coordinates", func(t *testing.T) {
		// Should have 5 coordinates: 4 corners + 1 to close
		expected := 5
		if len(coords) != expected {
			t.Errorf("Expected %d coordinates, got %d", expected, len(coords))
		}
	})

	t.Run("Counter-clockwise order", func(t *testing.T) {
		if len(coords) < 4 {
			t.Fatal("Not enough coordinates to check order")
		}
		
		// Check the order: bottom-left -> bottom-right -> top-right -> top-left
		bottomLeft := coords[0]
		bottomRight := coords[1]
		topRight := coords[2]
		topLeft := coords[3]

		// Bottom-left should have minimum x and y
		if bottomLeft[0] != bbox.Left || bottomLeft[1] != bbox.Bottom {
			t.Errorf("Bottom-left coordinate incorrect: expected [%f, %f], got %v", bbox.Left, bbox.Bottom, bottomLeft)
		}

		// Bottom-right should have maximum x and minimum y
		if bottomRight[0] != bbox.Right || bottomRight[1] != bbox.Bottom {
			t.Errorf("Bottom-right coordinate incorrect: expected [%f, %f], got %v", bbox.Right, bbox.Bottom, bottomRight)
		}

		// Top-right should have maximum x and y
		if topRight[0] != bbox.Right || topRight[1] != bbox.Top {
			t.Errorf("Top-right coordinate incorrect: expected [%f, %f], got %v", bbox.Right, bbox.Top, topRight)
		}

		// Top-left should have minimum x and maximum y
		if topLeft[0] != bbox.Left || topLeft[1] != bbox.Top {
			t.Errorf("Top-left coordinate incorrect: expected [%f, %f], got %v", bbox.Left, bbox.Top, topLeft)
		}
	})
}