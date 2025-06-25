package input

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/mikeocool/bbox/core"
	tu "github.com/mikeocool/bbox/test_utils"
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
					tu.AssertBboxEqual(t, *tt.expectBbox, bbox)
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
				tu.AssertBboxEqual(t, *tt.expectBbox, bbox)
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

func TestSniffWkt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid WKT formats
		{
			name:     "Simple POINT",
			input:    "POINT (10 20)",
			expected: true,
		},
		{
			name:     "POINT with whitespace",
			input:    "POINT   (10 20)",
			expected: true,
		},
		{
			name:     "POINT with leading whitespace",
			input:    "  POINT (10 20)",
			expected: true,
		},
		{
			name:     "Lowercase point",
			input:    "point (10 20)",
			expected: true,
		},
		{
			name:     "Mixed case point",
			input:    "Point (10 20)",
			expected: true,
		},
		{
			name:     "LINESTRING",
			input:    "LINESTRING (0 0, 10 10)",
			expected: true,
		},
		{
			name:     "POLYGON",
			input:    "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			expected: true,
		},
		{
			name:     "MULTIPOINT",
			input:    "MULTIPOINT ((10 20), (30 40))",
			expected: true,
		},
		{
			name:     "MULTILINESTRING",
			input:    "MULTILINESTRING ((0 0, 10 10), (20 20, 30 30))",
			expected: true,
		},
		{
			name:     "MULTIPOLYGON",
			input:    "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)))",
			expected: true,
		},
		{
			name:     "GEOMETRYCOLLECTION",
			input:    "GEOMETRYCOLLECTION (POINT (10 20))",
			expected: true,
		},
		{
			name:     "POINT with tabs",
			input:    "POINT\t(10 20)",
			expected: true,
		},
		{
			name:     "POINT with newlines",
			input:    "POINT\n(10 20)",
			expected: true,
		},

		// Invalid formats
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Too short",
			input:    "POIN",
			expected: false,
		},
		{
			name:     "Invalid geometry type",
			input:    "INVALID (10 20)",
			expected: false,
		},
		{
			name:     "POINT without space or parenthesis",
			input:    "POINTABC",
			expected: false,
		},
		{
			name:     "JSON data",
			input:    `{"type": "Point", "coordinates": [10, 20]}`,
			expected: false,
		},
		{
			name:     "Plain coordinates",
			input:    "10 20 30 40",
			expected: false,
		},
		{
			name:     "CSV data",
			input:    "x,y,z\n10,20,30",
			expected: false,
		},
		{
			name:     "POINT in middle of string",
			input:    "Some text POINT (10 20)",
			expected: false,
		},
		{
			name:     "Whitespace only",
			input:    "   \n\t   ",
			expected: false,
		},
		{
			name:     "Partial geometry type",
			input:    "POI (10 20)",
			expected: false,
		},

		// Edge cases
		{
			name:     "POINT exactly 5 chars",
			input:    "POINT",
			expected: false, // No space or parenthesis after
		},
		{
			name:     "POINT with immediate parenthesis",
			input:    "POINT(10 20)",
			expected: true,
		},
		{
			name:     "Long geometry type",
			input:    "GEOMETRYCOLLECTION (POINT (1 2))",
			expected: true,
		},
		{
			name:     "Binary data with POINT",
			input:    "\x00\x01POINT (10 20)",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SniffWkt([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("SniffWkt(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSniffWktWithLargeData(t *testing.T) {
	// Test with data larger than typical detection buffers
	largeWkt := "LINESTRING ("
	for i := 0; i < 1000; i++ {
		if i > 0 {
			largeWkt += ", "
		}
		largeWkt += fmt.Sprintf("%d %d", i, i)
	}
	largeWkt += ")"

	if !SniffWkt([]byte(largeWkt)) {
		t.Error("SniffWkt should detect large WKT data")
	}

	// Test with large non-WKT data
	largeNonWkt := strings.Repeat("not wkt data ", 1000)
	if SniffWkt([]byte(largeNonWkt)) {
		t.Error("SniffWkt should not detect large non-WKT data")
	}
}

func TestParseWktMultipleGeometries(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectBbox *core.Bbox
	}{
		{
			name:  "Two POINT geometries",
			input: "POINT(1 2) POINT(3 4)",
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:  "POINT and LINESTRING",
			input: "POINT(0 0) LINESTRING(5 5, 10 10)",
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  10.0,
				Top:    10.0,
			},
		},
		{
			name:  "Multiple geometries with newlines",
			input: "POINT(1 2)\nPOINT(3 4)\nLINESTRING(0 0, 5 5)",
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  5.0,
				Top:    5.0,
			},
		},
		{
			name:  "Multiple geometries with various whitespace",
			input: "POINT(1 1)   \n\t  POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))",
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  2.0,
				Top:    2.0,
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
				tu.AssertBboxEqual(t, *tt.expectBbox, bbox)
			}
		})
	}
}

// Helper functions
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
