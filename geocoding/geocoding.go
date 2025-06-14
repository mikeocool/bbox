package geocoding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mikeocool/bbox/core"
)

type Geocoder string

const (
	GeocoderPhotonDefault    Geocoder = "photon"
	GeocoderNominatimDefault Geocoder = "nominatim"
)

type GeocodeResult struct {
	Type      string
	FullName  string
	LocationX float64
	LocationY float64
	Extent    *core.Bbox
}

// HTTPClient interface allows for dependency injection and testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type resultResponse struct {
	Features []resultFeature `json:"features"`
}

type resultFeature struct {
	Geometry   resultGeometry         `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
	Bbox       []float64              `json:"bbox,omitempty"` // Nominatim returns bbox at feature level
}

type resultGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// parseBboxFromInterface parses a bbox from an interface{} array of 4 float64 values
func parseBboxFromInterface(val interface{}) *core.Bbox {
	if bboxArray, ok := val.([]interface{}); ok && len(bboxArray) == 4 {
		// Convert bbox values to float64
		var bboxFloats [4]float64
		allValid := true
		for i, val := range bboxArray {
			if floatVal, ok := val.(float64); ok {
				bboxFloats[i] = floatVal
			} else {
				allValid = false
				break
			}
		}

		if allValid {
			return &core.Bbox{
				Left:   bboxFloats[0],
				Bottom: bboxFloats[1],
				Right:  bboxFloats[2],
				Top:    bboxFloats[3],
			}
		}
	}
	return nil
}

func GeocodePlace(geocoder Geocoder, query string, headers []string) (*GeocodeResult, error) {
	var url string
	switch geocoder {
	case GeocoderPhotonDefault:
		url = "https://photon.komoot.io/api?q=%s&limit=1"
	case GeocoderNominatimDefault:
		url = "https://nominatim.openstreetmap.org/search?q=%s&format=geojson&limit=1"
	default:
		return nil, fmt.Errorf("unsupported geocoder: %s", geocoder)
	}
	return GeocodePlaceWithClient(url, query, http.DefaultClient, headers)
}

func GeocodePlaceWithURL(customURL, query string, headers []string) (*GeocodeResult, error) {
	return GeocodePlaceWithClient(customURL, query, http.DefaultClient, headers)
}

// GeocodePlaceWithClient allows dependency injection for testing
func GeocodePlaceWithClient(geocoderURL, query string, client HTTPClient, headers []string) (*GeocodeResult, error) {
	if geocoderURL == "" {
		return nil, fmt.Errorf("geocoder URL is required")
	}
	
	requestURL := fmt.Sprintf(geocoderURL, url.QueryEscape(query))

	// Create HTTP request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Header.Set(name, value)
		}
	}

	// Make the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request geocoding: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 response
	if resp.StatusCode != 200 {
		// Read up to 500 characters of the response body for error details
		bodyBytes := make([]byte, 500)
		n, _ := resp.Body.Read(bodyBytes)
		bodyPreview := string(bodyBytes[:n])
		if n > 0 {
			return nil, fmt.Errorf("geocoding request failed with status %d: %s", resp.StatusCode, bodyPreview)
		}
		return nil, fmt.Errorf("geocoding request failed with status %d", resp.StatusCode)
	}

	// Parse the JSON response
	var geocodeResp resultResponse
	if err := json.NewDecoder(resp.Body).Decode(&geocodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	// Check if we got any results
	if len(geocodeResp.Features) == 0 {
		return nil, fmt.Errorf("Could not find place matching: \"%s\"", query)
	}

	feature := geocodeResp.Features[0]

	// Get the center coordinates
	if len(feature.Geometry.Coordinates) < 2 {
		return nil, fmt.Errorf("invalid coordinates in geocoding response")
	}

	locationX := feature.Geometry.Coordinates[0]
	locationY := feature.Geometry.Coordinates[1]

	// Extract properties
	var placeType string
	if typeVal, ok := feature.Properties["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			placeType = typeStr
		}
	}

	// Build full name from name, state, and country
	var nameParts []string
	if name, ok := feature.Properties["name"]; ok {
		if nameStr, ok := name.(string); ok {
			nameParts = append(nameParts, nameStr)
		}
	}
	if state, ok := feature.Properties["state"]; ok {
		if stateStr, ok := state.(string); ok {
			nameParts = append(nameParts, stateStr)
		}
	}
	if country, ok := feature.Properties["country"]; ok {
		if countryStr, ok := country.(string); ok {
			nameParts = append(nameParts, countryStr)
		}
	}
	fullName := strings.Join(nameParts, ", ")

	// Check for bbox/extent in properties first, then at feature level
	var extent *core.Bbox
	
	// Check for bbox in properties first
	if bboxVal, hasBbox := feature.Properties["bbox"]; hasBbox {
		extent = parseBboxFromInterface(bboxVal)
	}
	
	// If no bbox in properties, check for extent property (Photon format)
	if extent == nil {
		if extentVal, hasExtent := feature.Properties["extent"]; hasExtent {
			extent = parseBboxFromInterface(extentVal)
		}
	}

	// If no bbox/extent in properties, check for bbox at feature level (Nominatim format)
	if extent == nil && len(feature.Bbox) == 4 {
		extent = &core.Bbox{
			Left:   feature.Bbox[0],
			Bottom: feature.Bbox[1],
			Right:  feature.Bbox[2],
			Top:    feature.Bbox[3],
		}
	}

	return &GeocodeResult{
		Type:      placeType,
		FullName:  fullName,
		LocationX: locationX,
		LocationY: locationY,
		Extent:    extent,
	}, nil
}
