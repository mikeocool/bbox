package geojson

import "encoding/json"

// GeoJSON type definitions
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type     string   `json:"type"`
	Geometry Geometry `json:"geometry"`
}

type Geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type Polygon struct {
	Type        string         `json:"type"`
	Coordinates [][][2]float64 `json:"coordinates"`
}

func PolygonGeometry(coords [][][2]float64) Geometry {
	coordsData, _ := json.Marshal(coords)
	return Geometry{
		Type:        "Polygon",
		Coordinates: json.RawMessage(coordsData),
	}
}
