package geojson

import (
	"encoding/json"
	"strings"
)

func Format(geoms []Geometry, outputType string, indent int) (string, error) {
	// TODO ensure outputType is a valid geojson type
	if outputType == "" {
		if len(geoms) == 1 {
			outputType = "geometry"
		} else {
			outputType = "feature-collection"
		}
	}

	if outputType == "coordinates" {
		if len(geoms) == 1 {
			return marshalGeojson(geoms[0].Coordinates, indent)
		} else {
			coords := make([]json.RawMessage, len(geoms))
			for i, geom := range geoms {
				coords[i] = geom.Coordinates
			}
			return marshalGeojson(coords, indent)
		}
	}

	if outputType == "geometry" {
		if len(geoms) == 1 {
			return marshalGeojson(geoms[0], indent)
		} else {
			return marshalGeojson(geoms, indent)
		}
	}

	features := make([]Feature, len(geoms))
	for i, geom := range geoms {
		features[i] = Feature{
			Type:     "Feature",
			Geometry: geom,
		}
	}

	if outputType == "feature" {
		if len(features) == 1 {
			return marshalGeojson(features[0], indent)
		} else {
			return marshalGeojson(features, indent)
		}
	}

	collection := FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	return marshalGeojson(collection, indent)
}

func marshalGeojson(geojson any, indent int) (string, error) {
	var data []byte
	var err error
	if indent > 0 {
		data, err = json.MarshalIndent(geojson, "", strings.Repeat(" ", indent))
	} else {
		data, err = json.Marshal(geojson)
	}
	if err != nil {
		return "", err
	}

	return string(data), nil
}
