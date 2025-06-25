package input

import (
	"strings"
	"testing"

	"github.com/mikeocool/bbox/core"
	tu "github.com/mikeocool/bbox/test_utils"
)

// TODO move geosjon and wkb stuff out
func TestParseRawArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expectError bool
		errorMsg    string
		expectBbox  *core.Bbox
	}{
		// Valid inputs
		{
			name:        "Valid input - separate args",
			input:       []string{"1.0", "2.0", "3.0", "4.0"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:        "Valid input - space separated",
			input:       []string{"1.0 2.0 3.0 4.0"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:        "Valid input - multiline",
			input:       []string{"1.0 2.0 3.0 4.0\n0.0 3.0 2.0 5.0"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    5.0,
			},
		},

		{
			name:        "Valid GeoJSON - Point feature",
			input:       []string{`{"type":"Feature","geometry":{"type":"Point","coordinates":[1.0,2.0]}}`},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
		},
		{
			name:        "geojson - empty FeatureCollection",
			input:       []string{`{"type":"FeatureCollection","features":[]}`},
			expectError: true,
			errorMsg:    "no features found",
		},

		{
			name:        "Valid JSON",
			input:       []string{`[1.0, 2.0, 3.0, 4.0]`},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},

		// WKB tests
		{
			name:        "Valid WKB hex - Point",
			input:       []string{"0101000000000000000000F03F0000000000000040"}, // POINT(1 2)
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
		},
		{
			name:        "Valid WKB hex - LineString",
			input:       []string{"010200000002000000000000000000F03F00000000000000400000000000000840000000000000F03F"}, // LINESTRING(1 2, 3 1)
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 1.0,
				Right:  3.0,
				Top:    2.0,
			},
		},
		{
			name:        "Valid WKB hex - Polygon",
			input:       []string{"01030000000100000005000000000000000000000000000000000000000000000000000000000000000000F03F000000000000F03F000000000000F03F000000000000F03F000000000000000000000000000000000000000000000000"}, // POLYGON((0 0, 0 1, 1 1, 1 0, 0 0))
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  1.0,
				Top:    1.0,
			},
		},
		{
			name:        "Valid WKB hex with whitespace",
			input:       []string{" 0101000000000000000000F03F0000000000000040 "}, // POINT(1 2) with whitespace
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
		},
		{
			name:        "Invalid WKB hex - bad characters",
			input:       []string{"01GH000000000000000000F03F0000000000000040"},
			expectError: true,
		},

		// WKT
		{
			name:        "Valid WKT - Point",
			input:       []string{"POINT(1 2)"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
		},
		{
			name:        "Valid WKT - LineString",
			input:       []string{"LINESTRING(1 2, 3 1)"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 1.0,
				Right:  3.0,
				Top:    2.0,
			},
		},
		{
			name:        "Valid WKT - Polygon",
			input:       []string{"POLYGON((0 0, 0 1, 1 1, 1 0, 0 0))"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  1.0,
				Top:    1.0,
			},
		},
		{
			name:        "Valid WKT with whitespace",
			input:       []string{" POINT(1 2) "},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
		},
		{
			name:        "Valid WKT - nultiple lines",
			input:       []string{"POINT(1 2)\nPOINT(3 4)"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:        "Invalid WKT - bad type",
			input:       []string{"PONT(1 2)"},
			expectError: true,
		},

		// file paths
		{
			name:        "Valid file",
			input:       []string{"../integration_tests/data/campsites.geojson"},
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   -92.429194,
				Bottom: 47.776398,
				Right:  -90.035484,
				Top:    48.355011,
			},
		},
		{
			name:        "Single valid file",
			input:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson")},
			expectError: false,
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99755413385825, Right: -91.14794444117372, Top: 48.01355378301334},
		},
		{
			name:        "Multiple valid files - union",
			input:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson"), getTestDataPath(t, "../integration_tests/data/subset_b.geojson")},
			expectError: false,
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99067253859491, Right: -90.92072645384923, Top: 48.07394149630552},
		},
		{
			name:        "Empty file",
			input:       []string{getTestDataPath(t, "../integration_tests/data/empty.geojson")},
			expectError: true, // Assuming empty file causes an error
			errorMsg:    "no features found",
		},
		{
			name:        "Non-existent file",
			input:       []string{getTestDataPath(t, "non_existent_file.geojson")},
			expectError: true,
		},
		{
			name:        "Mixed valid and empty files",
			input:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson"), getTestDataPath(t, "../integration_tests/data/empty.geojson")},
			expectError: false,
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99755413385825, Right: -91.14794444117372, Top: 48.01355378301334},
		},
		{
			name:        "Empty string in file list",
			input:       []string{getTestDataPath(t, "../integration_tests/data/subset_a.geojson"), "", getTestDataPath(t, "../integration_tests/data/subset_b.geojson")},
			expectError: false, // Empty strings should be skipped
			expectBbox:  &core.Bbox{Left: -91.34175985747542, Bottom: 47.99067253859491, Right: -90.92072645384923, Top: 48.07394149630552},
		},
		{
			name:        "Empty args list",
			input:       []string{},
			expectError: true,
			errorMsg:    "no arguments provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bbox, err := ParseRawArgs(tc.input)

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// If expecting an error, verify the error message
			if tc.expectError && err != nil {
				if tc.errorMsg != "" && err.Error() != tc.errorMsg {
					t.Errorf("Expected error message %q but got %q", tc.errorMsg, err.Error())
				}
				return
			}

			// If not expecting an error, verify the bbox
			if tc.expectBbox != nil {
				tu.AssertBboxEqual(t, *tc.expectBbox, bbox)
			}
		})
	}
}

func TestParseSimpleRaw(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
		expectBbox  *core.Bbox
	}{
		// Valid inputs
		{
			name:        "Valid input - space separated",
			input:       "1.0 2.0 3.0 4.0",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:        "Valid input - comma separated",
			input:       "1.5,2.5,3.5,4.5",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.5,
				Bottom: 2.5,
				Right:  3.5,
				Top:    4.5,
			},
		},
		{
			name:        "Valid input - tab separated",
			input:       "10\t20\t30\t40",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   10.0,
				Bottom: 20.0,
				Right:  30.0,
				Top:    40.0,
			},
		},
		{
			name:        "Valid input - ending in new line",
			input:       "10 20 30 40\n",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   10.0,
				Bottom: 20.0,
				Right:  30.0,
				Top:    40.0,
			},
		},
		{
			name:        "Valid input - mixed separators",
			input:       "1.0, 2.0\t3.0 4.0",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:        "Valid input - extra whitespace",
			input:       "  1.0  ,  2.0  ,  3.0  ,  4.0  ",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:        "Valid input - negative numbers",
			input:       "-1.0 -2.0 3.0 4.0",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   -1.0,
				Bottom: -2.0,
				Right:  3.0,
				Top:    4.0,
			},
		},
		{
			name:        "Valid input - zero values",
			input:       "0 0 0 0",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  0.0,
				Top:    0.0,
			},
		},
		{
			name:        "Valid input - scientific notation",
			input:       "1e2 4e-1 3.0E3 2.5e1",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   100.0,
				Bottom: 0.4,
				Right:  3000.0,
				Top:    25.0,
			},
		},
		{
			name:        "Valid input - list of points",
			input:       "1.0 1.0\n2.0 4.0\n2.0 6.0\n3.0 8.0\n",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 1.0,
				Right:  3.0,
				Top:    8.0,
			},
		},
		{
			name:        "Valid input - list of points comma separated",
			input:       "1.0,1.0\n2.0,4.0\n2.0,6.0\n3.0,8.0",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 1.0,
				Right:  3.0,
				Top:    8.0,
			},
		},

		// Invalid inputs - parsing errors
		{
			name:        "Invalid float at position 2",
			input:       "1.0 xyz 3.0 4.0",
			expectError: true,
			errorMsg:    "could not parse value: xyz",
		},
		{
			name:        "Invalid box order",
			input:       "4.0 5.0 1.0 4.0",
			expectError: true,
			errorMsg:    "invalid bbox: Right (1.000000) must be greater than Left (4.000000)",
		},
		{
			name:        "Too few numbers - 3 values",
			input:       "1.0 2.0 3.0",
			expectError: true,
			errorMsg:    "invalid input",
		},
		{
			name:        "Too many numbers - 5 values",
			input:       "1.0 2.0 3.0 4.0 5.0",
			expectError: true,
			errorMsg:    "invalid input",
		},
		{
			name:        "Single number",
			input:       "1.0",
			expectError: true,
			errorMsg:    "invalid input",
		},
		{
			name:        "Empty input",
			input:       "",
			expectError: true,
			errorMsg:    ErrUnrecognizedDataFormat.Error(),
		},
		{
			name:        "Only whitespace",
			input:       "   \t  \n  ",
			expectError: true,
			errorMsg:    ErrUnrecognizedDataFormat.Error(),
		},
		{
			name:        "Only separators",
			input:       ", , ,",
			expectError: true,
			errorMsg:    "invalid input",
		},
		{
			name:        "Mixed valid and empty parts",
			input:       "1.0, , 3.0, 4.0",
			expectError: true,
			errorMsg:    "invalid input",
		},
		{
			name: "Lines with 2 and 4 values",
			input: `
				1.0 2.0
				3.0 4.0 5.0 6.0
			`,
			expectError: true,
			errorMsg:    "invalid input",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bbox, err := ParseSimpleRaw(strings.NewReader(tc.input))

			// Check error status
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// If expecting an error, verify the error message
			if tc.expectError && err != nil {
				if tc.errorMsg != "" && err.Error() != tc.errorMsg {
					t.Errorf("Expected error message %q but got %q", tc.errorMsg, err.Error())
				}
				return
			}

			// If not expecting an error, verify the bbox
			if tc.expectBbox != nil {
				tu.AssertBboxEqual(t, *tc.expectBbox, bbox)
			}
		})
	}
}
