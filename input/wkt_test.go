package input

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestParseWkt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
		expectBbox  *core.Bbox
	}{
		// POINT tests
		{
			name:        "Simple POINT",
			input:       "POINT (10 20)",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   10.0,
				Bottom: 20.0,
				Right:  10.0,
				Top:    20.0,
			},
		},
		{
			name:        "POINT with negative coordinates",
			input:       "POINT (-10.5 -20.7)",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   -10.5,
				Bottom: -20.7,
				Right:  -10.5,
				Top:    -20.7,
			},
		},
		{
			name:        "POINT with extra whitespace",
			input:       "POINT   (  10.0   20.0  ) ",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   10.0,
				Bottom: 20.0,
				Right:  10.0,
				Top:    20.0,
			},
		},

		// LINESTRING tests
		{
			name:        "Simple LINESTRING",
			input:       "LINESTRING (0 0, 10 10, 20 0)",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  20.0,
				Top:    10.0,
			},
		},
		{
			name:        "LINESTRING with many points",
			input:       "LINESTRING (0 0, 5 10, 10 5, 15 15, 20 0)",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  20.0,
				Top:    15.0,
			},
		},

		// POLYGON tests
		{
			name:        "Simple POLYGON",
			input:       "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  10.0,
				Top:    10.0,
			},
		},
		{
			name:        "POLYGON with hole",
			input:       "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  20.0,
				Top:    20.0,
			},
		},
		{
			name:        "POLYGON with multiple holes",
			input:       "POLYGON ((0 0, 30 0, 30 30, 0 30, 0 0), (5 5, 10 5, 10 10, 5 10, 5 5), (20 20, 25 20, 25 25, 20 25, 20 20))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  30.0,
				Top:    30.0,
			},
		},

		// MULTIPOINT tests
		{
			name:        "Simple MULTIPOINT",
			input:       "MULTIPOINT ((10 20), (30 40))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   10.0,
				Bottom: 20.0,
				Right:  30.0,
				Top:    40.0,
			},
		},
		{
			name:        "MULTIPOINT alternative syntax",
			input:       "MULTIPOINT (10 20, 30 40, 50 10)",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   10.0,
				Bottom: 10.0,
				Right:  50.0,
				Top:    40.0,
			},
		},

		// MULTILINESTRING tests
		{
			name:        "Simple MULTILINESTRING",
			input:       "MULTILINESTRING ((0 0, 10 10), (20 20, 30 30))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  30.0,
				Top:    30.0,
			},
		},
		{
			name:        "MULTILINESTRING with multiple segments",
			input:       "MULTILINESTRING ((0 0, 10 0, 10 10), (20 0, 30 0, 30 10, 20 10))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  30.0,
				Top:    10.0,
			},
		},

		// MULTIPOLYGON tests
		{
			name:        "Simple MULTIPOLYGON",
			input:       "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 20, 30 20, 30 30, 20 30, 20 20)))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  30.0,
				Top:    30.0,
			},
		},
		{
			name:        "MULTIPOLYGON with holes",
			input:       "MULTIPOLYGON (((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5)), ((30 30, 50 30, 50 50, 30 50, 30 30)))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  50.0,
				Top:    50.0,
			},
		},

		// GEOMETRYCOLLECTION tests
		{
			name:        "Simple GEOMETRYCOLLECTION",
			input:       "GEOMETRYCOLLECTION (POINT (10 20), LINESTRING (0 0, 30 30))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  30.0,
				Top:    30.0,
			},
		},
		{
			name:        "Complex GEOMETRYCOLLECTION",
			input:       "GEOMETRYCOLLECTION (POINT (10 20), POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0)), MULTIPOINT ((30 30), (40 40)))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  40.0,
				Top:    40.0,
			},
		},
		{
			name:        "Nested GEOMETRYCOLLECTION",
			input:       "GEOMETRYCOLLECTION (POINT (5 5), GEOMETRYCOLLECTION (POINT (10 10), LINESTRING (0 0, 20 20)))",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  20.0,
				Top:    20.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			bbox, err := ParseWkt(reader)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if tt.expectBbox != nil {
					if !bboxEqual(bbox, *tt.expectBbox) {
						t.Errorf("Expected bbox %+v, got %+v", *tt.expectBbox, bbox)
					}
				}
			}
		})
	}
}

func TestParseWktEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		// Empty and invalid inputs
		{
			name:        "Empty input",
			input:       "",
			expectError: true,
			errorMsg:    "empty",
		},
		{
			name:        "Whitespace only",
			input:       "   \n\t   ",
			expectError: true,
			errorMsg:    "empty",
		},
		{
			name:        "Invalid geometry type",
			input:       "INVALID (10 20)",
			expectError: true,
			errorMsg:    "unknown geometry type",
		},
		{
			name:        "Missing geometry type",
			input:       "(10 20)",
			expectError: true,
			errorMsg:    "missing geometry type",
		},

		// Malformed coordinates
		{
			name:        "Invalid coordinate format",
			input:       "POINT (abc def)",
			expectError: true,
			errorMsg:    "invalid coordinate",
		},
		{
			name:        "Missing Y coordinate",
			input:       "POINT (10)",
			expectError: true,
			errorMsg:    "incomplete coordinate",
		},
		{
			name:        "Too many coordinates in POINT",
			input:       "POINT (10 20 30)",
			expectError: true,
			errorMsg:    "unexpected coordinate",
		},

		// Malformed parentheses
		{
			name:        "Missing opening parenthesis",
			input:       "POINT 10 20)",
			expectError: true,
			errorMsg:    "expected opening parenthesis",
		},
		{
			name:        "Missing closing parenthesis",
			input:       "POINT (10 20",
			expectError: true,
			errorMsg:    "expected closing parenthesis",
		},
		{
			name:        "Mismatched parentheses in POLYGON",
			input:       "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0)",
			expectError: true,
			errorMsg:    "mismatched parentheses",
		},

		// Empty geometries
		{
			name:        "POINT EMPTY",
			input:       "POINT EMPTY",
			expectError: true,
			errorMsg:    "empty geometry",
		},
		{
			name:        "LINESTRING EMPTY",
			input:       "LINESTRING EMPTY",
			expectError: true,
			errorMsg:    "empty geometry",
		},
		{
			name:        "POLYGON EMPTY",
			input:       "POLYGON EMPTY",
			expectError: true,
			errorMsg:    "empty geometry",
		},

		// Invalid geometry structures
		{
			name:        "LINESTRING with single point",
			input:       "LINESTRING (10 20)",
			expectError: true,
			errorMsg:    "linestring requires at least 2 points",
		},
		{
			name:        "POLYGON with unclosed ring",
			input:       "POLYGON ((0 0, 10 0, 10 10, 0 10))",
			expectError: true,
			errorMsg:    "polygon ring not closed",
		},
		{
			name:        "POLYGON with insufficient points",
			input:       "POLYGON ((0 0, 10 0, 0 0))",
			expectError: true,
			errorMsg:    "polygon requires at least 4 points",
		},

		// Extra characters and formatting issues
		{
			name:        "Extra characters after valid WKT",
			input:       "POINT (10 20) EXTRA",
			expectError: true,
			errorMsg:    "unexpected characters",
		},
		{
			name:        "Missing comma between coordinates",
			input:       "LINESTRING (0 0 10 10)",
			expectError: true,
			errorMsg:    "expected comma",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			_, err := ParseWkt(reader)

			if !tt.expectError {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestParseWktStreamingBehavior(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectBbox *core.Bbox
	}{
		{
			name:  "Large LINESTRING with many points",
			input: generateLargeLineString(1000),
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  999.0,
				Top:    999.0,
			},
		},
		{
			name:  "Large GEOMETRYCOLLECTION",
			input: generateLargeGeometryCollection(100),
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  99.0,
				Top:    99.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			bbox, err := ParseWkt(reader)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tt.expectBbox != nil {
				if !bboxEqual(bbox, *tt.expectBbox) {
					t.Errorf("Expected bbox %+v, got %+v", *tt.expectBbox, bbox)
				}
			}
		})
	}
}

func TestParseWktIOErrors(t *testing.T) {
	t.Run("Reader error", func(t *testing.T) {
		reader := &errorReader{error: io.ErrUnexpectedEOF}
		_, err := ParseWkt(reader)
		if err == nil {
			t.Error("Expected error from reader but got none")
		}
	})
}

// Helper functions

func bboxEqual(a, b core.Bbox) bool {
	const epsilon = 1e-9
	return abs(a.Left-b.Left) < epsilon &&
		abs(a.Bottom-b.Bottom) < epsilon &&
		abs(a.Right-b.Right) < epsilon &&
		abs(a.Top-b.Top) < epsilon
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func generateLargeLineString(numPoints int) string {
	var sb strings.Builder
	sb.WriteString("LINESTRING (")
	for i := 0; i < numPoints; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%d %d", i, i))
	}
	sb.WriteString(")")
	return sb.String()
}

func generateLargeGeometryCollection(numGeometries int) string {
	var sb strings.Builder
	sb.WriteString("GEOMETRYCOLLECTION (")
	for i := 0; i < numGeometries; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("POINT (%d %d)", i, i))
	}
	sb.WriteString(")")
	return sb.String()
}

type errorReader struct {
	error error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.error
}
