package input

import (
	"bytes"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestParseGeojson(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Bbox
		wantErr bool
	}{
		{
			name: "FeatureCollection with single polygon",
			input: `{
				"type": "FeatureCollection",
				"features": [
					{
						"type": "Feature",
						"geometry": {
							"type": "Polygon",
							"coordinates": [[[0,0],[0,1],[1,1],[1,0],[0,0]]]
						}
					}
				]
			}`,
			want: core.Bbox{
				Left:   0,
				Bottom: 0,
				Right:  1,
				Top:    1,
			},
			wantErr: false,
		},
		{
			name: "FeatureCollection with multiple features",
			input: `{
				"type": "FeatureCollection",
				"features": [
					{
						"type": "Feature",
						"geometry": {
							"type": "Polygon",
							"coordinates": [[[0,0],[0,1],[1,1],[1,0],[0,0]]]
						}
					},
					{
						"type": "Feature",
						"geometry": {
							"type": "Point",
							"coordinates": [2, 2]
						}
					}
				]
			}`,
			want: core.Bbox{
				Left:   0,
				Bottom: 0,
				Right:  2,
				Top:    2,
			},
			wantErr: false,
		},
		{
			name: "Array of Features",
			input: `[
				{
					"type": "Feature",
					"geometry": {
						"type": "Polygon",
						"coordinates": [[[-1,-1],[-1,0],[0,0],[0,-1],[-1,-1]]]
					}
				},
				{
					"type": "Feature",
					"geometry": {
						"type": "LineString",
						"coordinates": [[0,0],[2,2]]
					}
				}
			]`,
			want: core.Bbox{
				Left:   -1,
				Bottom: -1,
				Right:  2,
				Top:    2,
			},
			wantErr: false,
		},
		{
			name: "Single Feature",
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
			name: "Single Polygon",
			input: `{
				"type": "Polygon",
				"coordinates": [[[100.0, 0.0], [101.0, 0.0], [101.0, 1.0], [100.0, 1.0], [100.0, 0.0]]]
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  101.0,
				Top:    1.0,
			},
			wantErr: false,
		},
		{
			name:  "Raw coordinates array",
			input: `[[[0,0],[0,1],[1,1],[1,0],[0,0]]]`,
			want: core.Bbox{
				Left:   0,
				Bottom: 0,
				Right:  1,
				Top:    1,
			},
			wantErr: false,
		},
		{
			name:  "Raw 2D coordinates array (single ring)",
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
			name:  "Bounds array",
			input: `[0,0,1,1]`,
			want: core.Bbox{
				Left:   0,
				Bottom: 0,
				Right:  1,
				Top:    1,
			},
			wantErr: false,
		},
		{
			name:  "Raw 2D coordinates with negative values",
			input: `[[-10,-5],[-10,5],[10,5],[10,-5],[-10,-5]]`,
			want: core.Bbox{
				Left:   -10,
				Bottom: -5,
				Right:  10,
				Top:    5,
			},
			wantErr: false,
		},
		{
			name:  "Raw 2D coordinates with decimal values",
			input: `[[100.5, 0.5], [101.5, 0.5], [101.5, 1.5], [100.5, 1.5], [100.5, 0.5]]`,
			want: core.Bbox{
				Left:   100.5,
				Bottom: 0.5,
				Right:  101.5,
				Top:    1.5,
			},
			wantErr: false,
		},
		{
			name: "Feature with Point geometry",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "Point",
					"coordinates": [125.6, 10.1]
				}
			}`,
			want: core.Bbox{
				Left:   125.6,
				Bottom: 10.1,
				Right:  125.6,
				Top:    10.1,
			},
			wantErr: false,
		},
		{
			name: "Feature with LineString geometry",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "LineString",
					"coordinates": [[100.0, 0.0], [101.0, 1.0], [102.0, 0.5]]
				}
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  102.0,
				Top:    1.0,
			},
			wantErr: false,
		},
		{
			name: "Feature with MultiPoint geometry",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "MultiPoint",
					"coordinates": [[100.0, 0.0], [101.0, 1.0]]
				}
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  101.0,
				Top:    1.0,
			},
			wantErr: false,
		},
		{
			name: "Feature with MultiLineString geometry",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "MultiLineString",
					"coordinates": [
						[[100.0, 0.0], [101.0, 1.0]],
						[[102.0, 2.0], [103.0, 3.0]]
					]
				}
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  103.0,
				Top:    3.0,
			},
			wantErr: false,
		},
		{
			name: "Feature with MultiPolygon geometry",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "MultiPolygon",
					"coordinates": [
						[[[102.0, 2.0], [103.0, 2.0], [103.0, 3.0], [102.0, 3.0], [102.0, 2.0]]],
						[[[100.0, 0.0], [101.0, 0.0], [101.0, 1.0], [100.0, 1.0], [100.0, 0.0]]]
					]
				}
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  103.0,
				Top:    3.0,
			},
			wantErr: false,
		},
		{
			name: "Polygon with hole",
			input: `{
				"type": "Polygon",
				"coordinates": [
					[[100.0, 0.0], [101.0, 0.0], [101.0, 1.0], [100.0, 1.0], [100.0, 0.0]],
					[[100.2, 0.2], [100.8, 0.2], [100.8, 0.8], [100.2, 0.8], [100.2, 0.2]]
				]
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  101.0,
				Top:    1.0,
			},
			wantErr: false,
		},
		{
			name: "geojsonl with features",
			input: `{"type": "Feature","geometry": {"type": "Polygon","coordinates": [[[-10,-5],[-10,5],[10,5],[10,-5],[-10,-5]]]}}
			{"type": "Feature","geometry": {"type": "Polygon","coordinates": [[[-10,-6],[-10,5],[10,5],[10,-6],[-10,-6]]]}}
			`,
			want: core.Bbox{
				Left:   -10,
				Bottom: -6,
				Right:  10,
				Top:    5,
			},
			wantErr: false,
		},
		{
			name: "geojsonl with invalid features",
			input: `{"type": "Feature","geometry": {"type": "Polygon","coordinates": [[[-10,-5],[-10,5],[10,5],[10,-5],[-10,-5]]]}}
			{"type": "Feature","geometry": {"type": "Invalid","coordinates": [[[-10,-6],[-10,5],[10,5],[10,-6],[-10,-6]]]}}
			`,
			wantErr: true,
		},
		{
			name:  "geojsonl with bounds array",
			input: "[-10,-5,10,5]\n[-10,-6,10,5]",
			want: core.Bbox{
				Left:   -10,
				Bottom: -6,
				Right:  10,
				Top:    5,
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
		{
			name: "Empty FeatureCollection",
			input: `{
				"type": "FeatureCollection",
				"features": []
			}`,
			wantErr: true,
		},
		{
			name: "Invalid feature type in array",
			input: `[
				{
					"type": "NotAFeature",
					"geometry": {}
				}
			]`,
			wantErr: true,
		},
		{
			name: "Plain object (not GeoJSON)",
			input: `{
				"foo": "bar"
			}`,
			wantErr: true,
		},
		{
			name: "Feature with missing geometry",
			input: `{
				"type": "Feature"
			}`,
			wantErr: true,
		},
		{
			name: "Feature with null geometry",
			input: `{
				"type": "Feature",
				"geometry": null
			}`,
			wantErr: true,
		},
		{
			name: "Feature with invalid geometry type",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "InvalidType",
					"coordinates": [[0,0]]
				}
			}`,
			wantErr: true,
		},
		{
			name: "Point with insufficient coordinates",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "Point",
					"coordinates": [100.0]
				}
			}`,
			wantErr: true,
		},
		{
			name: "Point with empty coordinates",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "Point",
					"coordinates": []
				}
			}`,
			wantErr: true,
		},
		{
			name: "Polygon with empty coordinates",
			input: `{
				"type": "Polygon",
				"coordinates": []
			}`,
			wantErr: true,
		},
		{
			name: "Polygon with empty ring",
			input: `{
				"type": "Polygon",
				"coordinates": [[]]
			}`,
			wantErr: true,
		},
		{
			name: "LineString with single coordinate",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "LineString",
					"coordinates": [[100.0, 0.0]]
				}
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  100.0,
				Top:    0.0,
			},
			wantErr: false,
		},
		{
			name: "Coordinates with non-numeric values",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "Point",
					"coordinates": ["string", "values"]
				}
			}`,
			wantErr: true,
		},
		{
			name: "Mixed valid and invalid features",
			input: `{
				"type": "FeatureCollection",
				"features": [
					{
						"type": "Feature",
						"geometry": {
							"type": "Point",
							"coordinates": [1, 1]
						}
					},
					{
						"type": "Feature",
						"geometry": {
							"type": "InvalidGeometry",
							"coordinates": []
						}
					}
				]
			}`,
			want: core.Bbox{
				Left:   1,
				Bottom: 1,
				Right:  1,
				Top:    1,
			},
			wantErr: false,
		},
		{
			name: "FeatureCollection with non-feature objects",
			input: `{
				"type": "FeatureCollection",
				"features": [
					{
						"type": "NotAFeature",
						"geometry": {
							"type": "Point",
							"coordinates": [0, 0]
						}
					}
				]
			}`,
			wantErr: true,
		},
		{
			name:    "Raw coordinates with insufficient dimensions",
			input:   `[[[0],[1],[1],[0]]]`,
			wantErr: true,
		},
		{
			name:    "Raw 2D coordinates with insufficient dimensions",
			input:   `[[0],[1],[1],[0]]`,
			wantErr: true,
		},
		{
			name:    "Empty 2D coordinates array",
			input:   `[]`,
			wantErr: true,
		},
		{
			name:  "2D array with single coordinate",
			input: `[[100.0, 50.0]]`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 50.0,
				Right:  100.0,
				Top:    50.0,
			},
			wantErr: false,
		},
		{
			name: "Feature with coordinates as object instead of array",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "Point",
					"coordinates": {"x": 100, "y": 0}
				}
			}`,
			wantErr: true,
		},
		{
			name:    "Null input",
			input:   `null`,
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   ``,
			wantErr: true,
		},
		{
			name: "Array of non-features",
			input: `[
				{"type": "Point", "coordinates": [0, 0]},
				{"type": "LineString", "coordinates": [[0, 0], [1, 1]]}
			]`,
			wantErr: true,
		},
		{
			name: "Feature with geometry missing type",
			input: `{
				"type": "Feature",
				"geometry": {
					"coordinates": [100.0, 0.0]
				}
			}`,
			wantErr: true,
		},
		{
			name: "Feature with geometry missing coordinates",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "Point"
				}
			}`,
			wantErr: true,
		},
		{
			name: "MultiPolygon with mixed valid/invalid polygons",
			input: `{
				"type": "Feature",
				"geometry": {
					"type": "MultiPolygon",
					"coordinates": [
						[[[102.0, 2.0], [103.0, 2.0], [103.0, 3.0], [102.0, 3.0], [102.0, 2.0]]],
						[[[]]],
						[[[100.0, 0.0], [101.0, 0.0], [101.0, 1.0], [100.0, 1.0], [100.0, 0.0]]]
					]
				}
			}`,
			want: core.Bbox{
				Left:   100.0,
				Bottom: 0.0,
				Right:  103.0,
				Top:    3.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGeojson(bytes.NewReader([]byte(tt.input)))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGeojson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Left != tt.want.Left || got.Bottom != tt.want.Bottom ||
					got.Right != tt.want.Right || got.Top != tt.want.Top {
					t.Errorf("ParseGeojson() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestParseGeojsonNegativeCoordinates(t *testing.T) {
	input := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"geometry": {
					"type": "Polygon",
					"coordinates": [[[-180,-90],[-180,90],[180,90],[180,-90],[-180,-90]]]
				}
			}
		]
	}`

	want := core.Bbox{
		Left:   -180,
		Bottom: -90,
		Right:  180,
		Top:    90,
	}

	got, err := ParseGeojson(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Errorf("ParseGeojson() unexpected error = %v", err)
		return
	}

	if got.Left != want.Left || got.Bottom != want.Bottom ||
		got.Right != want.Right || got.Top != want.Top {
		t.Errorf("ParseGeojson() = %v, want %v", got, want)
	}
}

func TestParseGeojsonInvalidByteInput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Invalid UTF-8 bytes",
			input: []byte{0xFF, 0xFE, 0xFD},
		},
		{
			name:  "Truncated JSON",
			input: []byte(`{"type": "Feature", "geometry": {"type": "Poin`),
		},
		{
			name:  "Binary data",
			input: []byte{0x00, 0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseGeojson(bytes.NewReader(tt.input))
			if err == nil {
				t.Errorf("ParseGeojson() expected error for invalid byte input, got nil")
			}
		})
	}
}

func TestParseGeojsonLargeCoordinates(t *testing.T) {
	// Test with very large coordinate values
	input := `{
		"type": "Feature",
		"geometry": {
			"type": "Point",
			"coordinates": [1e308, -1e308]
		}
	}`

	got, err := ParseGeojson(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Errorf("ParseGeojson() unexpected error = %v", err)
		return
	}

	if got.Left != 1e308 || got.Bottom != -1e308 ||
		got.Right != 1e308 || got.Top != -1e308 {
		t.Errorf("ParseGeojson() failed to handle large coordinates correctly")
	}
}

func TestParseGeojsonEmptyGeometryCollection(t *testing.T) {
	input := `{
		"type": "Feature",
		"geometry": {
			"type": "GeometryCollection",
			"geometries": []
		}
	}`

	_, err := ParseGeojson(bytes.NewReader([]byte(input)))
	if err == nil {
		t.Errorf("ParseGeojson() expected error for empty GeometryCollection, got nil")
	}
}

func TestSniffGeojson(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "Empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "Only whitespace",
			data: []byte("   \t\n\r "),
			want: false,
		},
		{
			name: "Starts with invalid character",
			data: []byte("invalid json"),
			want: false,
		},
		{
			name: "Valid FeatureCollection",
			data: []byte(`{"type": "FeatureCollection", "features": []}`),
			want: true,
		},
		{
			name: "Valid Feature",
			data: []byte(`{"type": "Feature", "geometry": {"type": "Point"}}`),
			want: true,
		},
		{
			name: "Valid Polygon",
			data: []byte(`{"type": "Polygon", "coordinates": [[[0,0],[0,1],[1,1],[1,0],[0,0]]]}`),
			want: true,
		},
		{
			name: "Contains type keyword",
			data: []byte(`{"type": "something"}`),
			want: true,
		},
		{
			name: "Contains geometry keyword",
			data: []byte(`{"geometry": {"type": "Point"}}`),
			want: true,
		},
		{
			name: "Contains coordinates keyword",
			data: []byte(`{"coordinates": [0, 0]}`),
			want: true,
		},
		{
			name: "Contains features keyword",
			data: []byte(`{"features": []}`),
			want: true,
		},
		{
			name: "Array format with type",
			data: []byte(`[{"type": "Feature"}]`),
			want: true,
		},
		{
			name: "Raw coordinate array 3D",
			data: []byte(`[[[0,0],[0,1],[1,1],[1,0],[0,0]]]`),
			want: true,
		},
		{
			name: "Raw coordinate array 2D",
			data: []byte(`[[0,0],[0,1],[1,1],[1,0],[0,0]]`),
			want: true,
		},
		{
			name: "Raw coordinate array with decimals",
			data: []byte(`[[100.5, 0.5], [101.5, 1.5]]`),
			want: true,
		},
		{
			name: "Raw coordinate array with negative values",
			data: []byte(`[[-10,-5],[10,5]]`),
			want: true,
		},
		{
			name: "Raw coordinate array with whitespace",
			data: []byte(`[ [ 0 , 0 ] , [ 1 , 1 ] ]`),
			want: true,
		},
		{
			name: "Raw coordinate array with tabs and newlines",
			data: []byte("[\n\t[0,0],\n\t[1,1]\n]"),
			want: true,
		},
		{
			name: "Coordinate array with null bytes",
			data: []byte("[[0,0\x00],[1,1]]"),
			want: true,
		},
		{
			name: "Bounds array",
			data: []byte("[[0,0,1,1]]"),
			want: true,
		},
		{
			name: "Features in Geojsonl",
			data: []byte(`{"type": "Feature","geometry": {"type": "Polygon","coordinates": [[[-10,-5],[-10,5],[10,5],[10,-5],[-10,-5]]]}}
			{"type": "Feature","geometry": {"type": "Polygon","coordinates": [[[-10,-6],[-10,5],[10,5],[10,-6],[-10,-6]]]}}
			`),
			want: true,
		},
		{
			name: "Features in Geojsonl with whitespace",
			data: []byte(`{ "type": "Feature", "properties": { "name": "Abinodji Lake #1", "ref": 1, "tourism": "camp_site", "backcountry": "yes", "tents": "yes", "capacity:persons": "9", "openfire": "yes", "toilets": "yes", "toilets:disposal": "pitlatrine", "operator": "Superior National Forest", "operator:short": "USFS" }, "geometry": { "type": "Point", "coordinates": [ -91.341759857475424, 48.013553783013343 ] } }
{ "type": "Feature", "properties": { "name": "Adams Lake #1", "ref": 1, "tourism": "camp_site", "backcountry": "yes", "tents": "yes", "capacity:persons": "9", "openfire": "yes", "toilets": "yes", "toilets:disposal": "pitlatrine", "operator": "Superior National Forest", "operator:short": "USFS" }, "geometry": { "type": "Point", "coordinates": [ -91.147944441173721, 47.997554133858252 ] } }
{ "type": "Feature", "properties": { "name": "Adams Lake #2", "ref": 2, "tourism": "camp_site", "backcountry": "yes", "tents": "yes", "capacity:persons": "9", "openfire": "yes", "toilets": "yes", "toilets:disposal": "pitlatrine", "operator": "Superior National Forest", "operator:short": "USFS" }, "geometry": { "type": "Point", "coordinates": [ -91.147412362626596, 48.0008072084795 ] } }
			`),
			want: true,
		},
		{
			name: "Invalid characters in coordinate array",
			data: []byte(`[[0,0],[a,b]]`),
			want: false,
		},
		{
			name: "Contains letters mixed with coordinates",
			data: []byte(`[[0,0],[1,1]] some text`),
			want: false,
		},
		{
			name: "Plain JSON object without GeoJSON keywords",
			data: []byte(`{"name": "test", "value": 123}`),
			want: false,
		},
		{
			name: "Array of strings",
			data: []byte(`["hello", "world"]`),
			want: false,
		},
		{
			name: "Array of numbers only",
			data: []byte(`[1, 2, 3, 4]`),
			want: true,
		},
		{
			name: "Incomplete JSON with type",
			data: []byte(`{"type": "Feature", "geom`),
			want: true,
		},
		{
			name: "Incomplete coordinate array",
			data: []byte(`[[0,0],[1,`),
			want: true,
		},
		{
			name: "Case insensitive type matching",
			data: []byte(`{"TYPE": "Feature"}`),
			want: true,
		},
		{
			name: "Case insensitive geometry matching",
			data: []byte(`{"GEOMETRY": {"type": "Point"}}`),
			want: true,
		},
		{
			name: "Case insensitive coordinates matching",
			data: []byte(`{"COORDINATES": [0, 0]}`),
			want: true,
		},
		{
			name: "Case insensitive features matching",
			data: []byte(`{"FEATURES": []}`),
			want: true,
		},
		{
			name: "Nested type keyword",
			data: []byte(`{"data": {"type": "something"}}`),
			want: true,
		},
		{
			name: "Type keyword in string value",
			data: []byte(`{"description": "This type is not GeoJSON"}`),
			want: false,
		},
		{
			name: "Single coordinate pair",
			data: []byte(`[0, 0]`),
			want: true,
		},
		{
			name: "Single coordinate with decimals",
			data: []byte(`[100.123, -45.678]`),
			want: true,
		},
		{
			name: "Empty object",
			data: []byte(`{}`),
			want: false,
		},
		{
			name: "Empty array",
			data: []byte(`[]`),
			want: true,
		},
		{
			name: "Starts with object, no GeoJSON keywords",
			data: []byte(`{"foo": "bar", "baz": 123}`),
			want: false,
		},
		{
			name: "Starts with array, not coordinates",
			data: []byte(`["string1", "string2"]`),
			want: false,
		},
		{
			name: "Mixed valid coordinate characters with invalid",
			data: []byte(`[[0,0],[1,1]]xyz`),
			want: false,
		},
		{
			name: "Only valid coordinate characters",
			data: []byte(`0123456789[],. -\t\n\r`),
			want: false,
		},
		{
			name: "Coordinate array with extra whitespace and null bytes",
			data: []byte("[\n\t[0.0, 0.0\x00],\r\n\t[1.0, 1.0]\n]"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SniffGeojson(tt.data)
			if got != tt.want {
				t.Errorf("SniffGeojson() = %v, want %v", got, tt.want)
			}
		})
	}
}
