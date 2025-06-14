package input

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/mikeocool/bbox/geojson"
)

var ErrCouldNotParseGeoJSON = errors.New("unable to parse input as valid GeoJSON format")
var ErrNoFeaturesFound = errors.New("no features found")

func LoadGeojsonFile(filename string) (core.Bbox, error) {
	file, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, err
	}
	defer file.Close()
	return ParseGeojson(file)
}

// Check if a fragment of the file looks like GeoJSON
func SniffGeojson(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false
	}

	// Must start with { or [
	if trimmed[0] != '{' && trimmed[0] != '[' {
		return false
	}

	dataStr := strings.ToLower(string(data))
	// Look for common JSON patterns even if incomplete
	if strings.Contains(dataStr, `"type"`) ||
		strings.Contains(dataStr, `"geometry"`) ||
		strings.Contains(dataStr, `"coordinates"`) ||
		strings.Contains(dataStr, `"features"`) {
		return true
	}

	// Check if it's just a list of coordinates
	allowedChars := "[],. -0123456789\t\n\r\x00"
	for _, char := range dataStr {
		if !strings.ContainsRune(allowedChars, char) {
			return false
		}
	}

	return true
}

// ParseGeojson parses various GeoJSON formats and returns the bounding box of all features.
// Supported formats:
// - FeatureCollection containing one or more features
// - JSON list of Features
// - Single Feature
// - Single Polygon
// - 3D coordinate array (polygon with rings): [[[0,0],[0,1],[1,1],[1,0],[0,0]]]
// - 2D coordinate array (single ring): [[0,0],[0,1],[1,1],[1,0],[0,0]]
func ParseGeojson(r io.Reader) (core.Bbox, error) {
	var bbox core.Bbox

	input, err := io.ReadAll(r)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to read GeoJSON data: %w", err)
	}

	// Try parsing as FeatureCollection
	var featureCollection geojson.FeatureCollection
	if err := json.Unmarshal(input, &featureCollection); err == nil && featureCollection.Type == "FeatureCollection" {
		return calculateBboxFromFeatures(featureCollection.Features)
	}

	// Try parsing as array of Features
	var features []geojson.Feature
	if err := json.Unmarshal(input, &features); err == nil && len(features) > 0 {
		// Verify it's actually an array of features
		if isValidFeatureArray(features) {
			return calculateBboxFromFeatures(features)
		}
	}

	// Try parsing as single Feature
	var feature geojson.Feature
	if err := json.Unmarshal(input, &feature); err == nil && feature.Type == "Feature" {
		return calculateBboxFromFeatures([]geojson.Feature{feature})
	}

	// Try parsing as Polygon
	var polygon geojson.Polygon
	if err := json.Unmarshal(input, &polygon); err == nil && polygon.Type == "Polygon" {
		return calculateBboxFromCoordinates(polygon.Coordinates)
	}

	// Try parsing as raw coordinates (3D array for polygon)
	if coords, err := parseRaw3DCoordinates(input); err == nil {
		return calculateBboxFromCoordinates(coords)
	}

	// Try parsing as 2D array (single ring)
	if coords, err := parseRaw2DCoordinates(input); err == nil {
		// Wrap in an additional array to make it a 3D array
		return calculateBboxFromCoordinates([][][2]float64{coords})
	}

	return bbox, ErrCouldNotParseGeoJSON
}

// isValidFeatureArray checks if the array contains at least one valid feature
func isValidFeatureArray(features []geojson.Feature) bool {
	for _, f := range features {
		if f.Type == "Feature" {
			return true
		}
	}
	return false
}

// calculateBboxFromFeatures calculates bounding box from an array of features
func calculateBboxFromFeatures(features []geojson.Feature) (core.Bbox, error) {
	if len(features) == 0 {
		return core.Bbox{}, ErrNoFeaturesFound
	}

	minLon := math.Inf(1)
	minLat := math.Inf(1)
	maxLon := math.Inf(-1)
	maxLat := math.Inf(-1)
	hasValidCoordinates := false

	for _, feature := range features {
		// Skip non-Feature objects
		if feature.Type != "Feature" {
			continue
		}

		// Skip features with missing or empty geometry
		if feature.Geometry.Type == "" || len(feature.Geometry.Coordinates) == 0 {
			continue
		}

		switch feature.Geometry.Type {
		case "Point":
			var coords []float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			if len(coords) >= 2 {
				updateBounds(&minLon, &minLat, &maxLon, &maxLat, coords[0], coords[1])
				hasValidCoordinates = true
			}

		case "LineString":
			var coords [][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			for _, coord := range coords {
				if len(coord) >= 2 {
					updateBounds(&minLon, &minLat, &maxLon, &maxLat, coord[0], coord[1])
					hasValidCoordinates = true
				}
			}

		case "Polygon":
			var coords [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			for _, ring := range coords {
				for _, coord := range ring {
					if len(coord) >= 2 {
						updateBounds(&minLon, &minLat, &maxLon, &maxLat, coord[0], coord[1])
						hasValidCoordinates = true
					}
				}
			}

		case "MultiPoint":
			var coords [][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			for _, coord := range coords {
				if len(coord) >= 2 {
					updateBounds(&minLon, &minLat, &maxLon, &maxLat, coord[0], coord[1])
					hasValidCoordinates = true
				}
			}

		case "MultiLineString":
			var coords [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			for _, line := range coords {
				for _, coord := range line {
					if len(coord) >= 2 {
						updateBounds(&minLon, &minLat, &maxLon, &maxLat, coord[0], coord[1])
						hasValidCoordinates = true
					}
				}
			}

		case "MultiPolygon":
			var coords [][][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			for _, polygon := range coords {
				for _, ring := range polygon {
					for _, coord := range ring {
						if len(coord) >= 2 {
							updateBounds(&minLon, &minLat, &maxLon, &maxLat, coord[0], coord[1])
							hasValidCoordinates = true
						}
					}
				}
			}
		}
	}

	if !hasValidCoordinates || math.IsInf(minLon, 0) || math.IsInf(minLat, 0) || math.IsInf(maxLon, 0) || math.IsInf(maxLat, 0) {
		return core.Bbox{}, fmt.Errorf("no valid coordinates found")
	}

	return core.Bbox{
		Left:   minLon,
		Bottom: minLat,
		Right:  maxLon,
		Top:    maxLat,
	}, nil
}

// calculateBboxFromCoordinates calculates bounding box from polygon coordinates
func calculateBboxFromCoordinates(coords [][][2]float64) (core.Bbox, error) {
	if len(coords) == 0 {
		return core.Bbox{}, fmt.Errorf("no coordinates found")
	}

	minLon := math.Inf(1)
	minLat := math.Inf(1)
	maxLon := math.Inf(-1)
	maxLat := math.Inf(-1)
	hasValidCoordinates := false

	for _, ring := range coords {
		for _, coord := range ring {
			if len(coord) >= 2 {
				updateBounds(&minLon, &minLat, &maxLon, &maxLat, coord[0], coord[1])
				hasValidCoordinates = true
			}
		}
	}

	if !hasValidCoordinates || math.IsInf(minLon, 0) || math.IsInf(minLat, 0) || math.IsInf(maxLon, 0) || math.IsInf(maxLat, 0) {
		return core.Bbox{}, fmt.Errorf("no valid coordinates found")
	}

	return core.Bbox{
		Left:   minLon,
		Bottom: minLat,
		Right:  maxLon,
		Top:    maxLat,
	}, nil
}

// parseRaw3DCoordinates validates and parses 3D coordinate arrays like [[[0,0],[0,1],[1,1],[1,0],[0,0]]]
func parseRaw3DCoordinates(input []byte) ([][][2]float64, error) {
	// First unmarshal into a flexible structure to validate dimensions
	var rawCoords [][][]interface{}
	if err := json.Unmarshal(input, &rawCoords); err != nil {
		return nil, err
	}
	
	if len(rawCoords) == 0 {
		return nil, fmt.Errorf("empty coordinate array")
	}
	
	// Validate that each coordinate has at least 2 dimensions
	for _, ring := range rawCoords {
		for _, coord := range ring {
			if len(coord) < 2 {
				return nil, fmt.Errorf("coordinate has insufficient dimensions: %v", coord)
			}
		}
	}
	
	// Now unmarshal into the typed structure
	var coordinates [][][2]float64
	if err := json.Unmarshal(input, &coordinates); err != nil {
		return nil, err
	}
	
	return coordinates, nil
}

// parseRaw2DCoordinates validates and parses 2D coordinate arrays like [[0,0],[0,1],[1,1],[1,0],[0,0]]
func parseRaw2DCoordinates(input []byte) ([][2]float64, error) {
	// First unmarshal into a flexible structure to validate dimensions
	var rawCoords [][]interface{}
	if err := json.Unmarshal(input, &rawCoords); err != nil {
		return nil, err
	}
	
	if len(rawCoords) == 0 {
		return nil, fmt.Errorf("empty coordinate array")
	}
	
	// Validate that each coordinate has at least 2 dimensions
	for _, coord := range rawCoords {
		if len(coord) < 2 {
			return nil, fmt.Errorf("coordinate has insufficient dimensions: %v", coord)
		}
	}
	
	// Now unmarshal into the typed structure
	var coordinates [][2]float64
	if err := json.Unmarshal(input, &coordinates); err != nil {
		return nil, err
	}
	
	return coordinates, nil
}

// updateBounds updates the min/max bounds with the given coordinate
func updateBounds(minLon, minLat, maxLon, maxLat *float64, lon, lat float64) {
	if lon < *minLon {
		*minLon = lon
	}
	if lon > *maxLon {
		*maxLon = lon
	}
	if lat < *minLat {
		*minLat = lat
	}
	if lat > *maxLat {
		*maxLat = lat
	}
}
