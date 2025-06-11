package geojson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	// Helper function to create test geometries
	pointGeom := func(x, y float64) Geometry {
		coords, _ := json.Marshal([]float64{x, y})
		return Geometry{
			Type:        "Point",
			Coordinates: json.RawMessage(coords),
		}
	}

	polygonGeom := func(coords [][][2]float64) Geometry {
		coordsData, _ := json.Marshal(coords)
		return Geometry{
			Type:        "Polygon",
			Coordinates: json.RawMessage(coordsData),
		}
	}

	lineStringGeom := func(coords [][2]float64) Geometry {
		coordsData, _ := json.Marshal(coords)
		return Geometry{
			Type:        "LineString",
			Coordinates: json.RawMessage(coordsData),
		}
	}

	tests := []struct {
		name            string
		geoms           []Geometry
		outputType      string
		indent          int
		wantContains    []string // Strings that should be present in output
		wantNotContains []string // Strings that should NOT be present in output
		wantErr         bool
	}{
		{
			name:       "Single geometry - default output type (geometry)",
			geoms:      []Geometry{pointGeom(1.0, 2.0)},
			outputType: "",
			indent:     0,
			wantContains: []string{
				`"type":"Point"`,
				`"coordinates":[1,2]`,
			},
			wantNotContains: []string{
				`"type":"Feature"`,
				`"type":"FeatureCollection"`,
			},
		},
		{
			name:       "Multiple geometries - default output type (feature-collection)",
			geoms:      []Geometry{pointGeom(1.0, 2.0), pointGeom(3.0, 4.0)},
			outputType: "",
			indent:     0,
			wantContains: []string{
				`"type":"FeatureCollection"`,
				`"features":[`,
				`"type":"Feature"`,
				`"type":"Point"`,
				`[1,2]`,
				`[3,4]`,
			},
		},
		{
			name:       "Single geometry - coordinates output",
			geoms:      []Geometry{pointGeom(5.5, 6.5)},
			outputType: "coordinates",
			indent:     0,
			wantContains: []string{
				`[5.5,6.5]`,
			},
			wantNotContains: []string{
				`"type"`,
				`"geometry"`,
				`"coordinates":`,
			},
		},
		{
			name:       "Multiple geometries - coordinates output",
			geoms:      []Geometry{pointGeom(1.0, 2.0), pointGeom(3.0, 4.0)},
			outputType: "coordinates",
			indent:     0,
			wantContains: []string{
				`[[1,2],[3,4]]`,
			},
			wantNotContains: []string{
				`"type"`,
				`"geometry"`,
			},
		},
		{
			name:       "Single geometry - geometry output",
			geoms:      []Geometry{pointGeom(7.0, 8.0)},
			outputType: "geometry",
			indent:     0,
			wantContains: []string{
				`"type":"Point"`,
				`"coordinates":[7,8]`,
			},
			wantNotContains: []string{
				`"type":"Feature"`,
				`"type":"FeatureCollection"`,
			},
		},
		{
			name:       "Multiple geometries - geometry output",
			geoms:      []Geometry{pointGeom(1.0, 2.0), pointGeom(3.0, 4.0)},
			outputType: "geometry",
			indent:     0,
			wantContains: []string{
				`[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]`,
			},
			wantNotContains: []string{
				`"type":"Feature"`,
				`"type":"FeatureCollection"`,
			},
		},
		{
			name:       "Single geometry - feature output",
			geoms:      []Geometry{pointGeom(9.0, 10.0)},
			outputType: "feature",
			indent:     0,
			wantContains: []string{
				`"type":"Feature"`,
				`"geometry":{"type":"Point","coordinates":[9,10]}`,
			},
			wantNotContains: []string{
				`"type":"FeatureCollection"`,
				`"features"`,
			},
		},
		{
			name:       "Multiple geometries - feature output",
			geoms:      []Geometry{pointGeom(1.0, 2.0), pointGeom(3.0, 4.0)},
			outputType: "feature",
			indent:     0,
			wantContains: []string{
				`[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]}},{"type":"Feature","geometry":{"type":"Point","coordinates":[3,4]}}]`,
			},
			wantNotContains: []string{
				`"type":"FeatureCollection"`,
			},
		},
		{
			name:       "Feature collection output",
			geoms:      []Geometry{pointGeom(1.0, 2.0), pointGeom(3.0, 4.0)},
			outputType: "feature-collection",
			indent:     0,
			wantContains: []string{
				`"type":"FeatureCollection"`,
				`"features":[`,
				`"type":"Feature"`,
			},
		},
		{
			name:       "Single geometry with indentation",
			geoms:      []Geometry{pointGeom(1.0, 2.0)},
			outputType: "geometry",
			indent:     2,
			wantContains: []string{
				"{\n  \"type\": \"Point\",\n  \"coordinates\": [\n    1,\n    2\n  ]\n}",
			},
		},
		{
			name:       "Feature collection with indentation",
			geoms:      []Geometry{pointGeom(1.0, 2.0)},
			outputType: "feature-collection",
			indent:     4,
			wantContains: []string{
				"{\n    \"type\": \"FeatureCollection\",",
				"    \"features\": [",
				"        {",
				"            \"type\": \"Feature\",",
			},
		},
		{
			name: "Complex polygon geometry",
			geoms: []Geometry{
				polygonGeom([][][2]float64{
					{{0, 0}, {0, 1}, {1, 1}, {1, 0}, {0, 0}},
				}),
			},
			outputType: "geometry",
			indent:     0,
			wantContains: []string{
				`"type":"Polygon"`,
				`"coordinates":[[[0,0],[0,1],[1,1],[1,0],[0,0]]]`,
			},
		},
		{
			name: "LineString geometry",
			geoms: []Geometry{
				lineStringGeom([][2]float64{{0, 0}, {1, 1}, {2, 2}}),
			},
			outputType: "feature",
			indent:     0,
			wantContains: []string{
				`"type":"Feature"`,
				`"type":"LineString"`,
				`"coordinates":[[0,0],[1,1],[2,2]]`,
			},
		},
		{
			name: "Mixed geometry types",
			geoms: []Geometry{
				pointGeom(0, 0),
				lineStringGeom([][2]float64{{1, 1}, {2, 2}}),
				polygonGeom([][][2]float64{
					{{3, 3}, {3, 4}, {4, 4}, {4, 3}, {3, 3}},
				}),
			},
			outputType: "feature-collection",
			indent:     0,
			wantContains: []string{
				`"type":"FeatureCollection"`,
				`"type":"Point"`,
				`"type":"LineString"`,
				`"type":"Polygon"`,
				`[0,0]`,
				`[[1,1],[2,2]]`,
				`[[[3,3],[3,4],[4,4],[4,3],[3,3]]]`,
			},
		},
		{
			name:       "Empty geometry array - coordinates",
			geoms:      []Geometry{},
			outputType: "coordinates",
			indent:     0,
			wantContains: []string{
				`[]`,
			},
		},
		{
			name:       "Empty geometry array - geometry",
			geoms:      []Geometry{},
			outputType: "geometry",
			indent:     0,
			wantContains: []string{
				`[]`,
			},
		},
		{
			name:       "Empty geometry array - feature",
			geoms:      []Geometry{},
			outputType: "feature",
			indent:     0,
			wantContains: []string{
				`[]`,
			},
		},
		{
			name:       "Empty geometry array - feature-collection",
			geoms:      []Geometry{},
			outputType: "feature-collection",
			indent:     0,
			wantContains: []string{
				`"type":"FeatureCollection"`,
				`"features":[]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.geoms, tt.outputType, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check that all required strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("Format() output missing expected content.\nWant to contain: %s\nGot: %s", want, got)
				}
			}

			// Check that unwanted strings are not present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("Format() output contains unexpected content.\nDid not want: %s\nGot: %s", notWant, got)
				}
			}

			// Verify that the output is valid JSON
			var jsonCheck interface{}
			if err := json.Unmarshal([]byte(got), &jsonCheck); err != nil {
				t.Errorf("Format() output is not valid JSON: %v\nOutput: %s", err, got)
			}
		})
	}
}

func TestFormatEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		geoms      []Geometry
		outputType string
		indent     int
		wantErr    bool
	}{
		{
			name: "Invalid coordinates in geometry",
			geoms: []Geometry{
				{
					Type:        "Point",
					Coordinates: json.RawMessage(`invalid json`),
				},
			},
			outputType: "coordinates",
			indent:     0,
			wantErr:    true,
		},
		{
			name: "Very large indent value",
			geoms: []Geometry{
				{
					Type:        "Point",
					Coordinates: json.RawMessage(`[1,2]`),
				},
			},
			outputType: "geometry",
			indent:     100,
			wantErr:    false,
		},
		{
			name: "Negative indent value (should work like 0)",
			geoms: []Geometry{
				{
					Type:        "Point",
					Coordinates: json.RawMessage(`[1,2]`),
				},
			},
			outputType: "geometry",
			indent:     -5,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.geoms, tt.outputType, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == "" {
				t.Errorf("Format() returned empty string but no error")
			}
		})
	}
}

func TestMarshalGeojson(t *testing.T) {
	tests := []struct {
		name     string
		geojson  interface{}
		indent   int
		wantErr  bool
		wantJSON bool // Whether the output should be valid JSON
	}{
		{
			name: "Simple object with no indent",
			geojson: map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{1, 2},
			},
			indent:   0,
			wantErr:  false,
			wantJSON: true,
		},
		{
			name: "Simple object with indent",
			geojson: map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{1, 2},
			},
			indent:   2,
			wantErr:  false,
			wantJSON: true,
		},
		{
			name:     "Nil input",
			geojson:  nil,
			indent:   0,
			wantErr:  false,
			wantJSON: true,
		},
		{
			name:     "Function input (should error)",
			geojson:  func() {},
			indent:   0,
			wantErr:  true,
			wantJSON: false,
		},
		{
			name:     "Channel input (should error)",
			geojson:  make(chan int),
			indent:   0,
			wantErr:  true,
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalGeojson(tt.geojson, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("marshalGeojson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.wantJSON {
				var jsonCheck interface{}
				if err := json.Unmarshal([]byte(got), &jsonCheck); err != nil {
					t.Errorf("marshalGeojson() output is not valid JSON: %v\nOutput: %s", err, got)
				}
			}

			// Check indentation behavior
			if tt.indent > 0 {
				if !strings.Contains(got, "\n") {
					t.Errorf("marshalGeojson() with indent %d should contain newlines, got: %s", tt.indent, got)
				}
				expectedIndent := strings.Repeat(" ", tt.indent)
				if !strings.Contains(got, expectedIndent) {
					t.Errorf("marshalGeojson() with indent %d should contain %d spaces, got: %s", tt.indent, tt.indent, got)
				}
			}
		})
	}
}

func TestFormatOutputTypeValidation(t *testing.T) {
	pointGeom := Geometry{
		Type:        "Point",
		Coordinates: json.RawMessage(`[1,2]`),
	}

	validOutputTypes := []string{
		"",
		"coordinates",
		"geometry",
		"feature",
		"feature-collection",
	}

	for _, outputType := range validOutputTypes {
		t.Run("Valid output type: "+outputType, func(t *testing.T) {
			_, err := Format([]Geometry{pointGeom}, outputType, 0)
			if err != nil {
				t.Errorf("Format() with valid outputType %q should not error, got: %v", outputType, err)
			}
		})
	}

	// Test that any invalid output type defaults to feature-collection behavior
	invalidOutputTypes := []string{
		"invalid",
		"point",
		"polygon",
		"geojson",
		"collection",
	}

	for _, outputType := range invalidOutputTypes {
		t.Run("Invalid output type defaults to feature-collection: "+outputType, func(t *testing.T) {
			got, err := Format([]Geometry{pointGeom}, outputType, 0)
			if err != nil {
				t.Errorf("Format() with invalid outputType %q should not error, got: %v", outputType, err)
				return
			}

			// Should default to feature-collection behavior
			if !strings.Contains(got, `"type":"FeatureCollection"`) {
				t.Errorf("Format() with invalid outputType %q should default to FeatureCollection, got: %s", outputType, got)
			}
		})
	}
}

func TestFormatJSONStructure(t *testing.T) {
	pointGeom := Geometry{
		Type:        "Point",
		Coordinates: json.RawMessage(`[1,2]`),
	}

	polygonGeom := Geometry{
		Type:        "Polygon",
		Coordinates: json.RawMessage(`[[[0,0],[0,1],[1,1],[1,0],[0,0]]]`),
	}

	tests := []struct {
		name           string
		geoms          []Geometry
		outputType     string
		validateStruct func(t *testing.T, jsonStr string)
	}{
		{
			name:       "Single geometry output structure",
			geoms:      []Geometry{pointGeom},
			outputType: "geometry",
			validateStruct: func(t *testing.T, jsonStr string) {
				var geom Geometry
				if err := json.Unmarshal([]byte(jsonStr), &geom); err != nil {
					t.Errorf("Failed to unmarshal as Geometry: %v", err)
				}
				if geom.Type != "Point" {
					t.Errorf("Expected Point geometry, got %s", geom.Type)
				}
			},
		},
		{
			name:       "Feature output structure",
			geoms:      []Geometry{pointGeom},
			outputType: "feature",
			validateStruct: func(t *testing.T, jsonStr string) {
				var feature Feature
				if err := json.Unmarshal([]byte(jsonStr), &feature); err != nil {
					t.Errorf("Failed to unmarshal as Feature: %v", err)
				}
				if feature.Type != "Feature" {
					t.Errorf("Expected Feature type, got %s", feature.Type)
				}
				if feature.Geometry.Type != "Point" {
					t.Errorf("Expected Point geometry in feature, got %s", feature.Geometry.Type)
				}
			},
		},
		{
			name:       "FeatureCollection output structure",
			geoms:      []Geometry{pointGeom, polygonGeom},
			outputType: "feature-collection",
			validateStruct: func(t *testing.T, jsonStr string) {
				var fc FeatureCollection
				if err := json.Unmarshal([]byte(jsonStr), &fc); err != nil {
					t.Errorf("Failed to unmarshal as FeatureCollection: %v", err)
				}
				if fc.Type != "FeatureCollection" {
					t.Errorf("Expected FeatureCollection type, got %s", fc.Type)
				}
				if len(fc.Features) != 2 {
					t.Errorf("Expected 2 features, got %d", len(fc.Features))
				}
				if fc.Features[0].Geometry.Type != "Point" {
					t.Errorf("Expected first feature to be Point, got %s", fc.Features[0].Geometry.Type)
				}
				if fc.Features[1].Geometry.Type != "Polygon" {
					t.Errorf("Expected second feature to be Polygon, got %s", fc.Features[1].Geometry.Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.geoms, tt.outputType, 0)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			tt.validateStruct(t, got)
		})
	}
}
