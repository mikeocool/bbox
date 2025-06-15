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
	Get(url string) (*http.Response, error)
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

func GeocodePlace(geocoder Geocoder, query string) (*GeocodeResult, error) {
	var url string
	switch geocoder {
	case GeocoderPhotonDefault:
		url = "https://photon.komoot.io/api?q=%s&limit=1"
	case GeocoderNominatimDefault:
		url = "https://nominatim.openstreetmap.org/search?q=%s&format=geojson&limit=1"
	default:
		return nil, fmt.Errorf("unsupported geocoder: %s", geocoder)
	}
	return GeocodePlaceWithClient(url, query, http.DefaultClient)
}

func GeocodePlaceWithURL(customURL, query string) (*GeocodeResult, error) {
	return GeocodePlaceWithClient(customURL, query, http.DefaultClient)
}

// GeocodePlaceWithClient allows dependency injection for testing
func GeocodePlaceWithClient(geocoderURL, query string, client HTTPClient) (*GeocodeResult, error) {
	if geocoderURL == "" {
		return nil, fmt.Errorf("geocoder URL is required")
	}
	
	requestURL := fmt.Sprintf(geocoderURL, url.QueryEscape(query))

	// Make the HTTP request
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request geocoding: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 response
	if resp.StatusCode != 200 {
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

	// Check for extent property in properties (Photon format)
	var extent *core.Bbox
	if extentVal, hasExtent := feature.Properties["extent"]; hasExtent {
		if extentArray, ok := extentVal.([]interface{}); ok && len(extentArray) == 4 {
			// Convert extent values to float64
			var extentFloats [4]float64
			allValid := true
			for i, val := range extentArray {
				if floatVal, ok := val.(float64); ok {
					extentFloats[i] = floatVal
				} else {
					allValid = false
					break
				}
			}

			if allValid {
				extent = &core.Bbox{
					Left:   extentFloats[0],
					Bottom: extentFloats[1],
					Right:  extentFloats[2],
					Top:    extentFloats[3],
				}
			}
		}
	}

	// If no extent in properties, check for bbox at feature level (Nominatim format)
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
