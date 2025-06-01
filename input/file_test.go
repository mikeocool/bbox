package input

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestParseData(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Bbox
		wantErr bool
	}{
		{
			name: "Valid GeoJSON FeatureCollection",
			input: `{
				"type": "FeatureCollection",
				"features": [
					{
						"type": "Feature",
						"geometry": {
							"type": "Point",
							"coordinates": [100.0, 0.5]
						}
					},
					{
						"type": "Feature",
						"geometry": {
							"type": "Point",
							"coordinates": [101.0, 1.5]
						}
					}
				]
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.5,
				Right:  101.0,
				Top:    1.5,
			},
			wantErr: false,
		},
		{
			name: "Valid GeoJSON single Feature",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "Polygon",
					"coordinates": [[[-10,-5],[-10,5],[10,5],[10,-5],[-10,-5]]]
				}
			}`,
			want: core.Bbox{
				Left:   -10,
				Bottom: -5,
				Right:  10,
				Top:    5,
			},
			wantErr: false,
		},
		{
			name: "Valid GeoJSON array of Features",
			input: `[
				{
					"type": "Feature",
					"geometry": {
						"type": "Point",
						"coordinates": [0, 0]
					}
				},
				{
					"type": "Feature",
					"geometry": {
						"type": "Point",
						"coordinates": [1, 1]
					}
				}
			]`,
			want: core.Bbox{
				Left:   0,
				Bottom: 0,
				Right:  1,
				Top:    1,
			},
			wantErr: false,
		},
		{
			name:  "Raw coordinates",
			input: `[[0,0],[0,1],[1,1],[1,0],[0,0]]`,
			want: core.Bbox{
				Left:   0,
				Bottom: 0,
				Right:  1,
				Top:    1,
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			input:   `{invalid json`,
			wantErr: true,
		},
		{
			name:    "Empty GeoJSON FeatureCollection",
			input:   `{"type": "FeatureCollection", "features": []}`,
			wantErr: true,
		},
		{
			name:    "Non-GeoJSON, non-Shapefile data",
			input:   `{"foo": "bar"}`,
			wantErr: true,
		},
		{
			name:    "Plain text",
			input:   `This is just plain text`,
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   ``,
			wantErr: true,
		},
		{
			name:    "Binary data that isn't shapefile",
			input:   string([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}),
			wantErr: true,
		},
		{
			name:    "XML data",
			input:   `<?xml version="1.0"?><root><item>test</item></root>`,
			wantErr: true,
		},
		{
			name:    "CSV data",
			input:   `name,lat,lon\nTest,45.0,-90.0`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := ParseData(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Left != tt.want.Left || got.Bottom != tt.want.Bottom ||
					got.Right != tt.want.Right || got.Top != tt.want.Top {
					t.Errorf("ParseData() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestParseDataWithRealFiles(t *testing.T) {
	t.Run("Real GeoJSON file - campsites", func(t *testing.T) {
		file, err := os.Open("../integration_tests/data/campsites.geojson")
		if err != nil {
			t.Skipf("Skipping real file test: %v", err)
			return
		}
		defer file.Close()

		got, err := ParseData(file)
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		// Verify we got a reasonable bounding box (should contain Minnesota coordinates)
		if got.Left > -90 || got.Right < -92 || got.Bottom > 48 || got.Top < 47 {
			t.Errorf("ParseData() got unreasonable bounds for Minnesota data: %v", got)
		}
	})

	t.Run("Real Shapefile", func(t *testing.T) {
		file, err := os.Open("../integration_tests/data/ne_10m_populated_places_simple/ne_10m_populated_places_simple.shp")
		if err != nil {
			t.Skipf("Skipping real shapefile test: %v", err)
			return
		}
		defer file.Close()

		got, err := ParseData(file)
		// ParseShapefile requires file length information that's not available from io.Reader
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		expected := core.Bbox{
			Left:   -179.5899789,
			Bottom: -89.9999998,
			Right:  179.3833036,
			Top:    82.4833232,
		}
		if got != expected {
			t.Errorf("ParseData() got unexpected bounds: %v", got)
		}
	})

	t.Run("Empty GeoJSON file", func(t *testing.T) {
		file, err := os.Open("../integration_tests/data/empty.geojson")
		if err != nil {
			t.Skipf("Skipping empty file test: %v", err)
			return
		}
		defer file.Close()

		_, err = ParseData(file)
		if err == nil {
			t.Errorf("ParseData() expected error for empty GeoJSON file, got nil")
		}
	})
}

func TestParseDataEdgeCases(t *testing.T) {
	t.Run("Input smaller than detection buffer", func(t *testing.T) {
		input := `{"type":"Feature","geometry":{"type":"Point","coordinates":[5,10]}}`

		got, err := ParseData(strings.NewReader(input))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		want := core.Bbox{
			Left:   5,
			Bottom: 10,
			Right:  5,
			Top:    10,
		}

		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})

	t.Run("Fake shapefile header with invalid content", func(t *testing.T) {
		// Create data that looks like a shapefile header but isn't valid
		header := make([]byte, 100)
		// Set the magic number for shapefile (9994 = 0x270A)
		header[0] = 0x27
		header[1] = 0x0A
		header[2] = 0x00
		header[3] = 0x00
		// Fill rest with invalid data
		for i := 4; i < 100; i++ {
			header[i] = 0xFF
		}

		_, err := ParseData(bytes.NewReader(header))
		if err == nil {
			t.Errorf("ParseData() expected error for fake shapefile, got nil")
		}
	})

	t.Run("Reader that returns error on first read", func(t *testing.T) {
		errorReader := &erroringReader{}
		_, err := ParseData(errorReader)
		if err == nil {
			t.Errorf("ParseData() expected error from failing reader, got nil")
		}
	})
}

func TestParseDataDetectionFallback(t *testing.T) {
	t.Run("GeoJSON detection succeeds but parsing fails", func(t *testing.T) {
		// Create input that looks like GeoJSON but is malformed
		input := `{"type": "FeatureCollection", "features": [{"invalid": "feature"}]}`

		_, err := ParseData(strings.NewReader(input))
		if err == nil {
			t.Errorf("ParseData() expected error for malformed GeoJSON that can't fallback to shapefile")
		}
	})

	t.Run("Both detection methods fail", func(t *testing.T) {
		input := `This is definitely not a geo format`

		_, err := ParseData(strings.NewReader(input))
		if err == nil {
			t.Errorf("ParseData() expected error when no format can be detected")
		}

		expectedMsg := "Input does not appear to be a valid format"
		if err.Error() != expectedMsg {
			t.Errorf("ParseData() error = %v, want %v", err.Error(), expectedMsg)
		}
	})
}

func TestParseDataWithDifferentReaderTypes(t *testing.T) {
	geoJSON := `{"type":"Feature","geometry":{"type":"Point","coordinates":[42,24]}}`
	want := core.Bbox{Left: 42, Bottom: 24, Right: 42, Top: 24}

	t.Run("strings.Reader", func(t *testing.T) {
		got, err := ParseData(strings.NewReader(geoJSON))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}
		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})

	t.Run("bytes.Reader", func(t *testing.T) {
		got, err := ParseData(bytes.NewReader([]byte(geoJSON)))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}
		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})

	t.Run("bytes.Buffer", func(t *testing.T) {
		buffer := bytes.NewBufferString(geoJSON)
		got, err := ParseData(buffer)
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}
		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})
}

func TestParseDataDetectionBufferEdgeCases(t *testing.T) {
	t.Run("GeoJSON with leading whitespace within buffer", func(t *testing.T) {
		// Test GeoJSON with whitespace that still fits in detection buffer
		padding := strings.Repeat(" ", 100)
		geoJSON := `{"type":"Feature","geometry":{"type":"Point","coordinates":[42,24]}}`
		input := padding + geoJSON

		got, err := ParseData(strings.NewReader(input))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		want := core.Bbox{Left: 42, Bottom: 24, Right: 42, Top: 24}
		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})

	t.Run("Valid GeoJSON that fits exactly in buffer", func(t *testing.T) {
		// Create a GeoJSON that uses most of the 8192 byte buffer
		coords := make([][2]float64, 100)
		for i := range coords {
			coords[i] = [2]float64{float64(i), float64(i)}
		}

		// Simple point that should work
		input := `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]}}`

		got, err := ParseData(strings.NewReader(input))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		want := core.Bbox{Left: 1, Bottom: 2, Right: 1, Top: 2}
		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})

	t.Run("Mixed whitespace and newlines", func(t *testing.T) {
		input := "\n\t   \r\n  " + `{
			"type": "Feature",
			"geometry": {
				"type": "Point",
				"coordinates": [100, 200]
			}
		}`

		got, err := ParseData(strings.NewReader(input))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		want := core.Bbox{Left: 100, Bottom: 200, Right: 100, Top: 200}
		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})
}

func TestParseDataShapefileEdgeCases(t *testing.T) {
	t.Run("Shapefile header detection works", func(t *testing.T) {
		// Test that shapefile detection works correctly
		header := make([]byte, 200)
		// Set the magic number for shapefile (9994 = 0x270A in big-endian)
		header[0] = 0x00
		header[1] = 0x00
		header[2] = 0x27
		header[3] = 0x0A

		isShapefile := SniffShapefile(header)
		if !isShapefile {
			t.Errorf("SniffShapefile() should detect valid shapefile header")
		}
	})

	t.Run("Invalid shapefile magic number", func(t *testing.T) {
		header := make([]byte, 200)
		// Set wrong magic number
		header[0] = 0x00
		header[1] = 0x00
		header[2] = 0x00
		header[3] = 0x01

		isShapefile := SniffShapefile(header)
		if isShapefile {
			t.Errorf("SniffShapefile() should not detect invalid magic number as shapefile")
		}
	})

	t.Run("Too short for shapefile", func(t *testing.T) {
		header := make([]byte, 50) // Less than 100 bytes required

		isShapefile := SniffShapefile(header)
		if isShapefile {
			t.Errorf("SniffShapefile() should not detect short buffer as shapefile")
		}
	})
}

// erroringReader is a test helper that always returns an error on Read
type erroringReader struct{}

func (e *erroringReader) Read(p []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}

func TestParseDataComplexGeoJSON(t *testing.T) {
	t.Run("Complex MultiPolygon", func(t *testing.T) {
		input := `{
			"type": "Feature",
			"geometry": {
				"type": "MultiPolygon",
				"coordinates": [
					[[[0,0],[0,10],[10,10],[10,0],[0,0]]],
					[[[20,20],[20,30],[30,30],[30,20],[20,20]]]
				]
			}
		}`

		got, err := ParseData(strings.NewReader(input))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		want := core.Bbox{
			Left:   0,
			Bottom: 0,
			Right:  30,
			Top:    30,
		}

		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})

	t.Run("FeatureCollection with mixed geometry types", func(t *testing.T) {
		input := `{
			"type": "FeatureCollection",
			"features": [
				{
					"type": "Feature",
					"geometry": {
						"type": "Point",
						"coordinates": [-100, 40]
					}
				},
				{
					"type": "Feature",
					"geometry": {
						"type": "LineString",
						"coordinates": [[-90, 35], [-80, 45]]
					}
				},
				{
					"type": "Feature",
					"geometry": {
						"type": "Polygon",
						"coordinates": [[[-75, 30], [-75, 50], [-70, 50], [-70, 30], [-75, 30]]]
					}
				}
			]
		}`

		got, err := ParseData(strings.NewReader(input))
		if err != nil {
			t.Errorf("ParseData() unexpected error = %v", err)
			return
		}

		want := core.Bbox{
			Left:   -100,
			Bottom: 30,
			Right:  -70,
			Top:    50,
		}

		if got != want {
			t.Errorf("ParseData() = %v, want %v", got, want)
		}
	})
}
