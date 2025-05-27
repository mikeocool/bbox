package input

import (
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestParseRaw(t *testing.T) {
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
			input:       "1e2 2.5e1 3.0E0 4e-1",
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   100.0,
				Bottom: 25.0,
				Right:  3.0,
				Top:    0.4,
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

		{
			name:        "Valid GeoJSON - Point feature",
			input:       `{"type":"Feature","geometry":{"type":"Point","coordinates":[1.0,2.0]}}`,
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
		},

		// Invalid inputs - parsing errors
		{
			name:        "Valid JSON but invalid GeoJSON - empty FeatureCollection",
			input:       `{"type":"FeatureCollection","features":[]}`,
			expectError: true,
			errorMsg:    "no features found",
		},
		{
			name:        "Invalid float at position 2",
			input:       "1.0 xyz 3.0 4.0",
			expectError: true,
			errorMsg:    "could not parse value: xyz",
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
			errorMsg:    "invalid input",
		},
		{
			name:        "Only whitespace",
			input:       "   \t  \n  ",
			expectError: true,
			errorMsg:    "invalid input",
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
		{
			name: "Lines with 2 and 4 values",
			input: `
				1.0 2.0
				3.0 4.0 5.0 6.0
			`,
			expectError: false,
			expectBbox: &core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  5.0,
				Top:    6.0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bbox, err := ParseRaw([]byte(tc.input))

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
				if bbox.Left != tc.expectBbox.Left {
					t.Errorf("Expected Left %f but got %f", tc.expectBbox.Left, bbox.Left)
				}
				if bbox.Bottom != tc.expectBbox.Bottom {
					t.Errorf("Expected Bottom %f but got %f", tc.expectBbox.Bottom, bbox.Bottom)
				}
				if bbox.Right != tc.expectBbox.Right {
					t.Errorf("Expected Right %f but got %f", tc.expectBbox.Right, bbox.Right)
				}
				if bbox.Top != tc.expectBbox.Top {
					t.Errorf("Expected Top %f but got %f", tc.expectBbox.Top, bbox.Top)
				}
			}
		})
	}
}
