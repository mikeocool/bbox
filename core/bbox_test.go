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

func TestBboxSlice(t *testing.T) {
	tests := []struct {
		name     string
		bbox     Bbox
		columns  int
		rows     int
		expected []Bbox
	}{
		{
			name:    "2x2 grid",
			bbox:    Bbox{Left: 0.0, Bottom: 0.0, Right: 4.0, Top: 4.0},
			columns: 2,
			rows:    2,
			expected: []Bbox{
				{Left: 0.0, Bottom: 2.0, Right: 2.0, Top: 4.0}, // top-left
				{Left: 2.0, Bottom: 2.0, Right: 4.0, Top: 4.0}, // top-right
				{Left: 0.0, Bottom: 0.0, Right: 2.0, Top: 2.0}, // bottom-left
				{Left: 2.0, Bottom: 0.0, Right: 4.0, Top: 2.0}, // bottom-right
			},
		},
		{
			name:    "1x1 grid (single box)",
			bbox:    Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			columns: 1,
			rows:    1,
			expected: []Bbox{
				{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			},
		},
		{
			name:    "3x1 grid (3 columns, 1 row)",
			bbox:    Bbox{Left: 0.0, Bottom: 0.0, Right: 6.0, Top: 2.0},
			columns: 3,
			rows:    1,
			expected: []Bbox{
				{Left: 0.0, Bottom: 0.0, Right: 2.0, Top: 2.0},
				{Left: 2.0, Bottom: 0.0, Right: 4.0, Top: 2.0},
				{Left: 4.0, Bottom: 0.0, Right: 6.0, Top: 2.0},
			},
		},
		{
			name:    "1x3 grid (1 column, 3 rows)",
			bbox:    Bbox{Left: 0.0, Bottom: 0.0, Right: 2.0, Top: 6.0},
			columns: 1,
			rows:    3,
			expected: []Bbox{
				{Left: 0.0, Bottom: 4.0, Right: 2.0, Top: 6.0}, // top
				{Left: 0.0, Bottom: 2.0, Right: 2.0, Top: 4.0}, // middle
				{Left: 0.0, Bottom: 0.0, Right: 2.0, Top: 2.0}, // bottom
			},
		},
		{
			name:    "Non-square bbox with 2x2 grid",
			bbox:    Bbox{Left: -1.0, Bottom: -2.0, Right: 3.0, Top: 4.0},
			columns: 2,
			rows:    2,
			expected: []Bbox{
				{Left: -1.0, Bottom: 1.0, Right: 1.0, Top: 4.0},  // top-left
				{Left: 1.0, Bottom: 1.0, Right: 3.0, Top: 4.0},   // top-right
				{Left: -1.0, Bottom: -2.0, Right: 1.0, Top: 1.0}, // bottom-left
				{Left: 1.0, Bottom: -2.0, Right: 3.0, Top: 1.0},  // bottom-right
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.bbox.Slice(tc.columns, tc.rows)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d boxes, got %d", len(tc.expected), len(result))
				return
			}

			for i, box := range result {
				expected := tc.expected[i]
				if box.Left != expected.Left || box.Bottom != expected.Bottom ||
					box.Right != expected.Right || box.Top != expected.Top {
					t.Errorf("Box %d: expected %+v, got %+v", i, expected, box)
				}
			}
		})
	}
}

func TestBboxSliceEdgeCases(t *testing.T) {
	bbox := Bbox{Left: 0.0, Bottom: 0.0, Right: 4.0, Top: 4.0}

	t.Run("Zero columns", func(t *testing.T) {
		result := bbox.Slice(0, 2)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for zero columns, got %d boxes", len(result))
		}
	})

	t.Run("Zero rows", func(t *testing.T) {
		result := bbox.Slice(2, 0)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for zero rows, got %d boxes", len(result))
		}
	})

	t.Run("Negative columns", func(t *testing.T) {
		result := bbox.Slice(-1, 2)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for negative columns, got %d boxes", len(result))
		}
	})

	t.Run("Negative rows", func(t *testing.T) {
		result := bbox.Slice(2, -1)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for negative rows, got %d boxes", len(result))
		}
	})

	t.Run("Large number of divisions", func(t *testing.T) {
		result := bbox.Slice(100, 100)
		expectedCount := 100 * 100
		if len(result) != expectedCount {
			t.Errorf("Expected %d boxes for 100x100 grid, got %d", expectedCount, len(result))
		}

		// Check that all boxes are valid and within bounds
		for i, box := range result {
			if err := box.Validate(); err != nil {
				t.Errorf("Box %d is invalid: %v", i, err)
			}
			if box.Left < bbox.Left || box.Right > bbox.Right ||
				box.Bottom < bbox.Bottom || box.Top > bbox.Top {
				t.Errorf("Box %d is outside original bounds: %+v", i, box)
			}
		}
	})

	t.Run("Very small bbox", func(t *testing.T) {
		smallBbox := Bbox{Left: 0.0, Bottom: 0.0, Right: 0.001, Top: 0.001}
		result := smallBbox.Slice(2, 2)

		if len(result) != 4 {
			t.Errorf("Expected 4 boxes, got %d", len(result))
		}

		// Check that all boxes are valid despite small size
		for i, box := range result {
			if err := box.Validate(); err != nil {
				t.Errorf("Small box %d is invalid: %v", i, err)
			}
		}
	})
}

func TestBboxSliceProperties(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 5.0, Top: 6.0}
	result := bbox.Slice(2, 3)

	t.Run("Correct number of boxes", func(t *testing.T) {
		expected := 2 * 3
		if len(result) != expected {
			t.Errorf("Expected %d boxes, got %d", expected, len(result))
		}
	})

	t.Run("All boxes are valid", func(t *testing.T) {
		for i, box := range result {
			if err := box.Validate(); err != nil {
				t.Errorf("Box %d is invalid: %v", i, err)
			}
		}
	})

	t.Run("Boxes tile the original bbox", func(t *testing.T) {
		// Union of all boxes should equal the original bbox
		if len(result) == 0 {
			t.Fatal("No boxes to test")
		}

		union := result[0]
		for i := 1; i < len(result); i++ {
			union = union.Union(result[i])
		}

		if union.Left != bbox.Left || union.Bottom != bbox.Bottom ||
			union.Right != bbox.Right || union.Top != bbox.Top {
			t.Errorf("Union of sliced boxes %+v does not match original bbox %+v", union, bbox)
		}
	})

	t.Run("Boxes are within original bounds", func(t *testing.T) {
		for i, box := range result {
			if box.Left < bbox.Left || box.Right > bbox.Right ||
				box.Bottom < bbox.Bottom || box.Top > bbox.Top {
				t.Errorf("Box %d is outside original bounds: %+v", i, box)
			}
		}
	})
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
				{10.25, 20.75},   // bottom-left
				{30.125, 20.75},  // bottom-right
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

func TestBboxBuffer(t *testing.T) {
	tests := []struct {
		name        string
		bbox        Bbox
		radius      float64
		expected    Bbox
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Positive buffer",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			radius:      1.0,
			expected:    Bbox{Left: 0.0, Bottom: 1.0, Right: 4.0, Top: 5.0},
			expectError: false,
		},
		{
			name:        "Zero buffer",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			radius:      0.0,
			expected:    Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 4.0},
			expectError: false,
		},
		{
			name:        "Negative buffer",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 5.0, Top: 6.0},
			radius:      -0.5,
			expected:    Bbox{Left: 1.5, Bottom: 2.5, Right: 4.5, Top: 5.5},
			expectError: false,
		},
		{
			name:        "Error: Negative buffer too large for width",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 3.0, Top: 6.0},
			radius:      -1.0,
			expectError: true,
			errorMsg:    fmt.Sprintf("cannot shrink box with width %f by %f", 2.0, -1.0),
		},
		{
			name:        "Error: Negative buffer too large for height",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 5.0, Top: 3.0},
			radius:      -0.6,
			expectError: true,
			errorMsg:    fmt.Sprintf("cannot shrink box with height %f by %f", 1.0, -0.6),
		},
		{
			name:        "Error: Negative buffer too large for both dimensions",
			bbox:        Bbox{Left: 1.0, Bottom: 2.0, Right: 2.0, Top: 3.0},
			radius:      -0.5,
			expectError: true,
			errorMsg:    fmt.Sprintf("cannot shrink box with width %f by %f", 1.0, -0.5),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.bbox.Buffer(tc.radius)

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

			// If not expecting an error, verify the result
			if !tc.expectError {
				if result.Left != tc.expected.Left || result.Bottom != tc.expected.Bottom ||
					result.Right != tc.expected.Right || result.Top != tc.expected.Top {
					t.Errorf("Expected %+v, got %+v", tc.expected, result)
				}
			}
		})
	}
}

func TestBboxBufferProperties(t *testing.T) {
	bbox := Bbox{Left: 1.0, Bottom: 2.0, Right: 5.0, Top: 6.0}

	t.Run("Buffer maintains proportions", func(t *testing.T) {
		radius := 2.0
		result, err := bbox.Buffer(radius)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		originalWidth := bbox.Width()
		originalHeight := bbox.Height()
		newWidth := result.Width()
		newHeight := result.Height()

		expectedWidth := originalWidth + 2*radius
		expectedHeight := originalHeight + 2*radius

		if newWidth != expectedWidth {
			t.Errorf("Expected width %f, got %f", expectedWidth, newWidth)
		}
		if newHeight != expectedHeight {
			t.Errorf("Expected height %f, got %f", expectedHeight, newHeight)
		}
	})

	t.Run("Center remains the same", func(t *testing.T) {
		radius := 3.0
		result, err := bbox.Buffer(radius)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		originalCenter := bbox.Center()
		newCenter := result.Center()

		if originalCenter != newCenter {
			t.Errorf("Expected center to remain %v, got %v", originalCenter, newCenter)
		}
	})

	t.Run("Buffer with zero radius is identity", func(t *testing.T) {
		result, err := bbox.Buffer(0.0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Left != bbox.Left || result.Bottom != bbox.Bottom ||
			result.Right != bbox.Right || result.Top != bbox.Top {
			t.Errorf("Expected unchanged bbox %+v, got %+v", bbox, result)
		}
	})
}
