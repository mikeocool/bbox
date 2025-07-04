package input

import (
	"encoding/hex"
	"math"
	"testing"

	"github.com/mikeocool/bbox/core"
	tu "github.com/mikeocool/bbox/test_utils"
)

func TestParseWKBBounds(t *testing.T) {
	tests := []struct {
		name     string
		wkbHex   string
		wantMinX float64
		wantMinY float64
		wantMaxX float64
		wantMaxY float64
		wantErr  bool
	}{
		{
			name:     "Point",
			wkbHex:   "0101000000000000000000F03F0000000000000040", // POINT(1 2)
			wantMinX: 1.0,
			wantMinY: 2.0,
			wantMaxX: 1.0,
			wantMaxY: 2.0,
			wantErr:  false,
		},
		{
			name:     "LineString",
			wkbHex:   "010200000002000000000000000000F03F00000000000000400000000000000840000000000000F03F", // LINESTRING(1 2, 3 1)
			wantMinX: 1.0,
			wantMinY: 1.0,
			wantMaxX: 3.0,
			wantMaxY: 2.0,
			wantErr:  false,
		},
		{
			name:     "Polygon",
			wkbHex:   "01030000000100000005000000000000000000000000000000000000000000000000000000000000000000F03F000000000000F03F000000000000F03F000000000000F03F000000000000000000000000000000000000000000000000", // POLYGON((0 0, 0 1, 1 1, 1 0, 0 0))
			wantMinX: 0.0,
			wantMinY: 0.0,
			wantMaxX: 1.0,
			wantMaxY: 1.0,
			wantErr:  false,
		},
		{
			name:     "MultiPoint",
			wkbHex:   "0104000000020000000101000000000000000000F03F000000000000004001010000000000000000000840000000000000104000", // MULTIPOINT((1 2), (3 4))
			wantMinX: 1.0,
			wantMinY: 2.0,
			wantMaxX: 3.0,
			wantMaxY: 4.0,
			wantErr:  false,
		},
		{
			name:     "Empty WKB",
			wkbHex:   "",
			wantMinX: 0,
			wantMinY: 0,
			wantMaxX: 0,
			wantMaxY: 0,
			wantErr:  true,
		},
		{
			name:     "Too short WKB",
			wkbHex:   "0101",
			wantMinX: 0,
			wantMinY: 0,
			wantMaxX: 0,
			wantMaxY: 0,
			wantErr:  true,
		},
		{
			name:     "Big endian Point",
			wkbHex:   "00000000013FF00000000000004000000000000000", // POINT(1 2) in big endian
			wantMinX: 1.0,
			wantMinY: 2.0,
			wantMaxX: 1.0,
			wantMaxY: 2.0,
			wantErr:  false,
		},
		{
			name:     "Polygon with hole",
			wkbHex:   "010300000002000000050000000000000000000000000000000000000000000000000000000000000000001440000000000000144000000000000014400000000000001440000000000000000000000000000000000000000000000000050000000000000000000040000000000000004000000000000000400000000000001040000000000000104000000000000010400000000000001040000000000000004000000000000000400000000000000040", // POLYGON with outer ring (0 0, 0 5, 5 5, 5 0, 0 0) and hole (2 2, 2 4, 4 4, 4 2, 2 2)
			wantMinX: 0.0,
			wantMinY: 0.0,
			wantMaxX: 5.0,
			wantMaxY: 5.0,
			wantErr:  false,
		},
		{
			name:     "Invalid geometry type",
			wkbHex:   "01FF000000", // Invalid type 255
			wantMinX: 0,
			wantMinY: 0,
			wantMaxX: 0,
			wantMaxY: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wkb, err := hex.DecodeString(tt.wkbHex)
			if err != nil && !tt.wantErr {
				t.Fatalf("Failed to decode test hex: %v", err)
			}

			minX, minY, maxX, maxY, err := ParseWKBBounds(wkb)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWKBBounds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				const epsilon = 1e-10
				if math.Abs(minX-tt.wantMinX) > epsilon {
					t.Errorf("ParseWKBBounds() minX = %v, want %v", minX, tt.wantMinX)
				}
				if math.Abs(minY-tt.wantMinY) > epsilon {
					t.Errorf("ParseWKBBounds() minY = %v, want %v", minY, tt.wantMinY)
				}
				if math.Abs(maxX-tt.wantMaxX) > epsilon {
					t.Errorf("ParseWKBBounds() maxX = %v, want %v", maxX, tt.wantMaxX)
				}
				if math.Abs(maxY-tt.wantMaxY) > epsilon {
					t.Errorf("ParseWKBBounds() maxY = %v, want %v", maxY, tt.wantMaxY)
				}
			}
		})
	}
}

func TestParseWKBBoundsMultiLineString(t *testing.T) {
	// MULTILINESTRING((0 0, 1 1), (2 2, 3 3))
	wkbHex := "01050000000200000001020000000200000000000000000000000000000000000000000000000000F03F000000000000F03F0102000000020000000000000000000040000000000000004000000000000008400000000000000840"
	wkb, err := hex.DecodeString(wkbHex)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	minX, minY, maxX, maxY, err := ParseWKBBounds(wkb)
	if err != nil {
		t.Fatalf("ParseWKBBounds() unexpected error: %v", err)
	}

	const epsilon = 1e-10
	if math.Abs(minX-0.0) > epsilon || math.Abs(minY-0.0) > epsilon ||
		math.Abs(maxX-3.0) > epsilon || math.Abs(maxY-3.0) > epsilon {
		t.Errorf("ParseWKBBounds() = (%v, %v, %v, %v), want (0, 0, 3, 3)",
			minX, minY, maxX, maxY)
	}
}

func TestParseWKBBoundsMultiPolygon(t *testing.T) {
	// MULTIPOLYGON(((0 0, 0 1, 1 1, 1 0, 0 0)), ((2 2, 2 3, 3 3, 3 2, 2 2)))
	wkbHex := "01060000000200000001030000000100000005000000000000000000000000000000000000000000000000000000000000000000F03F000000000000F03F000000000000F03F000000000000F03F000000000000000000000000000000000000000000000000010300000001000000050000000000000000000040000000000000004000000000000000400000000000000840000000000000084000000000000008400000000000000840000000000000004000000000000000400000000000000040"
	wkb, err := hex.DecodeString(wkbHex)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	minX, minY, maxX, maxY, err := ParseWKBBounds(wkb)
	if err != nil {
		t.Fatalf("ParseWKBBounds() unexpected error: %v", err)
	}

	const epsilon = 1e-10
	if math.Abs(minX-0.0) > epsilon || math.Abs(minY-0.0) > epsilon ||
		math.Abs(maxX-3.0) > epsilon || math.Abs(maxY-3.0) > epsilon {
		t.Errorf("ParseWKBBounds() = (%v, %v, %v, %v), want (0, 0, 3, 3)",
			minX, minY, maxX, maxY)
	}
}

func TestParseWKBBoundsGeometryCollection(t *testing.T) {
	// GEOMETRYCOLLECTION(POINT(1 1), LINESTRING(0 0, 2 2))
	wkbHex := "01070000000200000001010000000000000000000F3F000000000000F03F0102000000020000000000000000000000000000000000000000000000000000400000000000000040"
	wkb, err := hex.DecodeString(wkbHex)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	minX, minY, maxX, maxY, err := ParseWKBBounds(wkb)
	if err != nil {
		t.Fatalf("ParseWKBBounds() unexpected error: %v", err)
	}

	const epsilon = 1e-10
	if math.Abs(minX-0.0) > epsilon || math.Abs(minY-0.0) > epsilon ||
		math.Abs(maxX-2.0) > epsilon || math.Abs(maxY-2.0) > epsilon {
		t.Errorf("ParseWKBBounds() = (%v, %v, %v, %v), want (0, 0, 2, 2)",
			minX, minY, maxX, maxY)
	}
}

func TestParseWKBBoundsWithSRID(t *testing.T) {
	// EWKB Point with SRID=4326; geometry type will have SRID flag (0x20000001)
	wkbHex := "0101000020E6100000000000000000F03F0000000000000040" // SRID=4326;POINT(1 2)
	wkb, err := hex.DecodeString(wkbHex)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	// Test ParseWKBBounds (should work but ignore SRID)
	minX, minY, maxX, maxY, err := ParseWKBBounds(wkb)
	if err != nil {
		t.Fatalf("ParseWKBBounds() unexpected error: %v", err)
	}

	const epsilon = 1e-10
	if math.Abs(minX-1.0) > epsilon || math.Abs(minY-2.0) > epsilon ||
		math.Abs(maxX-1.0) > epsilon || math.Abs(maxY-2.0) > epsilon {
		t.Errorf("ParseWKBBounds() = (%v, %v, %v, %v), want (1, 2, 1, 2)",
			minX, minY, maxX, maxY)
	}

	// Test ParseWKBBoundsWithSRID (should extract SRID)
	minX2, minY2, maxX2, maxY2, srid, err := ParseWKBBoundsWithSRID(wkb)
	if err != nil {
		t.Fatalf("ParseWKBBoundsWithSRID() unexpected error: %v", err)
	}

	if math.Abs(minX2-1.0) > epsilon || math.Abs(minY2-2.0) > epsilon ||
		math.Abs(maxX2-1.0) > epsilon || math.Abs(maxY2-2.0) > epsilon {
		t.Errorf("ParseWKBBoundsWithSRID() bounds = (%v, %v, %v, %v), want (1, 2, 1, 2)",
			minX2, minY2, maxX2, maxY2)
	}

	if srid != 4326 {
		t.Errorf("ParseWKBBoundsWithSRID() srid = %v, want 4326", srid)
	}
}

func TestParseWKBToBboxWithSRID(t *testing.T) {
	tests := []struct {
		name     string
		wkbHex   string
		wantBox  core.Bbox
		wantErr  bool
	}{
		{
			name:   "EWKB Point with SRID=4326",
			wkbHex: "0101000020E6100000000000000000F03F0000000000000040", // SRID=4326;POINT(1 2)
			wantBox: core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
				Srid:   4326,
			},
			wantErr: false,
		},
		{
			name:   "EWKB LineString with SRID=3857",
			wkbHex: "0102000020110F000002000000000000000000F03F00000000000000400000000000000840000000000000F03F", // SRID=3857;LINESTRING(1 2, 3 1)
			wantBox: core.Bbox{
				Left:   1.0,
				Bottom: 1.0,
				Right:  3.0,
				Top:    2.0,
				Srid:   3857,
			},
			wantErr: false,
		},
		{
			name:   "Regular WKB Point (no SRID)",
			wkbHex: "0101000000000000000000F03F0000000000000040", // POINT(1 2)
			wantBox: core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
				Srid:   0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wkb, err := hex.DecodeString(tt.wkbHex)
			if err != nil && !tt.wantErr {
				t.Fatalf("Failed to decode test hex: %v", err)
			}

			got, err := ParseWKBToBbox(wkb)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWKBToBbox() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tu.AssertBboxEqual(t, tt.wantBox, got)
			}
		})
	}
}

func TestParseHexWKB(t *testing.T) {
	tests := []struct {
		name    string
		hexStr  string
		want    []byte
		wantErr bool
	}{
		{
			name:    "Valid hex string",
			hexStr:  "0101000000",
			want:    []byte{0x01, 0x01, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Empty string",
			hexStr:  "",
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "Uppercase hex",
			hexStr:  "DEADBEEF",
			want:    []byte{0xDE, 0xAD, 0xBE, 0xEF},
			wantErr: false,
		},
		{
			name:    "Mixed case hex",
			hexStr:  "DeAdBeEf",
			want:    []byte{0xDE, 0xAD, 0xBE, 0xEF},
			wantErr: false,
		},
		{
			name:    "Odd length hex string",
			hexStr:  "01010",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid hex characters",
			hexStr:  "01GH",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid hex characters mixed",
			hexStr:  "01XY23",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Non-hex string",
			hexStr:  "hello",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Special characters",
			hexStr:  "01@#",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Hex with spaces (invalid)",
			hexStr:  "DE AD BE EF",
			wantErr: true,
		},
		{
			name:    "Hex with newlines (invalid)",
			hexStr:  "DEAD\nBEEF",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHexWKB(tt.hexStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHexWKB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseHexWKB() length = %v, want %v", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("parseHexWKB()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestParseWKBToBbox(t *testing.T) {
	tests := []struct {
		name    string
		wkbHex  string
		wantBox core.Bbox
		wantErr bool
	}{
		{
			name:   "Point",
			wkbHex: "0101000000000000000000F03F0000000000000040", // POINT(1 2)
			wantBox: core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
			wantErr: false,
		},
		{
			name:   "LineString",
			wkbHex: "010200000002000000000000000000F03F00000000000000400000000000000840000000000000F03F", // LINESTRING(1 2, 3 1)
			wantBox: core.Bbox{
				Left:   1.0,
				Bottom: 1.0,
				Right:  3.0,
				Top:    2.0,
			},
			wantErr: false,
		},
		{
			name:   "Polygon",
			wkbHex: "01030000000100000005000000000000000000000000000000000000000000000000000000000000000000F03F000000000000F03F000000000000F03F000000000000F03F000000000000000000000000000000000000000000000000", // POLYGON((0 0, 0 1, 1 1, 1 0, 0 0))
			wantBox: core.Bbox{
				Left:   0.0,
				Bottom: 0.0,
				Right:  1.0,
				Top:    1.0,
			},
			wantErr: false,
		},
		{
			name:    "Invalid WKB",
			wkbHex:  "01FF000000", // Invalid type
			wantBox: core.Bbox{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wkb, err := hex.DecodeString(tt.wkbHex)
			if err != nil && !tt.wantErr {
				t.Fatalf("Failed to decode test hex: %v", err)
			}

			got, err := ParseWKBToBbox(wkb)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWKBToBbox() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				const epsilon = 1e-10
				if math.Abs(got.Left-tt.wantBox.Left) > epsilon ||
					math.Abs(got.Bottom-tt.wantBox.Bottom) > epsilon ||
					math.Abs(got.Right-tt.wantBox.Right) > epsilon ||
					math.Abs(got.Top-tt.wantBox.Top) > epsilon {
					t.Errorf("ParseWKBToBbox() = %+v, want %+v", got, tt.wantBox)
				}
			}
		})
	}
}

func TestParseHexWKBToBbox(t *testing.T) {
	tests := []struct {
		name    string
		hexStr  string
		wantBox core.Bbox
		wantErr bool
	}{
		{
			name:   "Point hex",
			hexStr: "0101000000000000000000F03F0000000000000040", // POINT(1 2)
			wantBox: core.Bbox{
				Left:   1.0,
				Bottom: 2.0,
				Right:  1.0,
				Top:    2.0,
			},
			wantErr: false,
		},
		{
			name:   "LineString hex with whitespace",
			hexStr: " 010200000002000000000000000000F03F00000000000000400000000000000840000000000000F03F ", // LINESTRING(1 2, 3 1)
			wantBox: core.Bbox{
				Left:   1.0,
				Bottom: 1.0,
				Right:  3.0,
				Top:    2.0,
			},
			wantErr: false,
		},
		{
			name:    "Invalid hex string",
			hexStr:  "01GH000000",
			wantBox: core.Bbox{},
			wantErr: true,
		},
		{
			name:    "Empty hex string",
			hexStr:  "",
			wantBox: core.Bbox{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHexWKBToBbox(tt.hexStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHexWKBToBbox() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				const epsilon = 1e-10
				if math.Abs(got.Left-tt.wantBox.Left) > epsilon ||
					math.Abs(got.Bottom-tt.wantBox.Bottom) > epsilon ||
					math.Abs(got.Right-tt.wantBox.Right) > epsilon ||
					math.Abs(got.Top-tt.wantBox.Top) > epsilon {
					t.Errorf("ParseHexWKBToBbox() = %+v, want %+v", got, tt.wantBox)
				}
			}
		})
	}
}

func TestSniffWkbHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid hex WKB formats
		{
			name:     "Valid point hex",
			input:    "0101000000000000000000f03f0000000000000040",
			expected: true,
		},
		{
			name:     "Valid hex uppercase",
			input:    "0101000000000000000000F03F0000000000000040",
			expected: true,
		},
		{
			name:     "Valid hex mixed case",
			input:    "0101000000000000000000f03F0000000000000040",
			expected: true,
		},
		{
			name:     "Valid hex with whitespace",
			input:    "  0101000000000000000000f03f0000000000000040  ",
			expected: true,
		},
		{
			name:     "Minimal valid hex (10 chars)",
			input:    "0102030405",
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
			input:    "01020304",
			expected: false,
		},
		{
			name:     "Odd length",
			input:    "010203040",
			expected: false,
		},
		{
			name:     "Invalid hex characters",
			input:    "0101000g00000000000000f03f0000000000000040",
			expected: false,
		},
		{
			name:     "Non-hex string",
			input:    "POINT(1 2)",
			expected: false,
		},
		{
			name:     "JSON data",
			input:    `{"type": "Point", "coordinates": [1, 2]}`,
			expected: false,
		},
		{
			name:     "Plain coordinates",
			input:    "10 20 30 40",
			expected: false,
		},
		{
			name:     "Hex with spaces inside",
			input:    "0101 0000 0000 0000",
			expected: false,
		},
		{
			name:     "Hex with newlines",
			input:    "0101000000\n000000000000f03f",
			expected: false,
		},
		{
			name:     "Binary data",
			input:    "\x01\x01\x00\x00\x00",
			expected: false,
		},
		{
			name:     "Whitespace only",
			input:    "   \n\t   ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SniffWkbHex([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("SniffWkbHex(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
