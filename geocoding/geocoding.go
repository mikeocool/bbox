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

const GeocoderPhotonKamoot Geocoder = "photon_kamoot"

type GeocodeResult struct {
	Type      string
	FullName  string
	LocationX float64
	LocationY float64
	Extent    *core.Bbox
}

type photonResponse struct {
	Features []photonFeature `json:"features"`
}

type photonFeature struct {
	Geometry   photonGeometry         `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type photonGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func GeocodePlace(geocoder Geocoder, query string) (*GeocodeResult, error) {
	switch geocoder {
	case GeocoderPhotonKamoot:
		return geocodePhotonKamoot(query)
	default:
		return nil, fmt.Errorf("unsupported geocoder: %s", geocoder)
	}
}

func geocodePhotonKamoot(query string) (*GeocodeResult, error) {
	// Build the URL with query parameter
	baseURL := "https://photon.komoot.io/api"
	params := url.Values{}
	params.Add("q", query)
	params.Add("limit", "1")
	requestURL := baseURL + "?" + params.Encode()

	// Make the HTTP request
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request geocoding: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 response
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("geocoding request failed with status %d", resp.StatusCode)
	}

	// Parse the JSON response
	var photonResp photonResponse
	if err := json.NewDecoder(resp.Body).Decode(&photonResp); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	// Check if we got any results
	if len(photonResp.Features) == 0 {
		return nil, fmt.Errorf("Could not find place matching: \"%s\"", query)
	}

	feature := photonResp.Features[0]

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

	// Check for extent property
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

	return &GeocodeResult{
		Type:      placeType,
		FullName:  fullName,
		LocationX: locationX,
		LocationY: locationY,
		Extent:    extent,
	}, nil
}
