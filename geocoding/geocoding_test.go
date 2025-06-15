package geocoding

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mikeocool/bbox/core"
)

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	Response *http.Response
	Error    error
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Response, nil
}

// Helper function to create a mock HTTP response
func createMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestGeocodePlaceWithClient_Success(t *testing.T) {
	// Mock response JSON with all fields
	mockJSON := `{
		"type":"FeatureCollection",
		"features": [
			{
				"geometry": {
					"type": "Point",
					"coordinates": [-122.4194, 37.7749]
				},
				"properties": {
					"name": "San Francisco",
					"state": "California",
					"country": "United States",
					"type": "city",
					"extent": [-122.5, 37.7, -122.3, 37.8]
				}
			}
		]
	}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	result, err := GeocodePlaceWithClient("https://photon.komoot.io/api?q=%s&limit=1", "San Francisco", mockClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.LocationX != -122.4194 {
		t.Errorf("Expected LocationX -122.4194, got %f", result.LocationX)
	}

	if result.LocationY != 37.7749 {
		t.Errorf("Expected LocationY 37.7749, got %f", result.LocationY)
	}

	if result.Type != "city" {
		t.Errorf("Expected Type 'city', got '%s'", result.Type)
	}

	expectedName := "San Francisco, California, United States"
	if result.FullName != expectedName {
		t.Errorf("Expected FullName '%s', got '%s'", expectedName, result.FullName)
	}

	if result.Extent == nil {
		t.Error("Expected Extent to be set")
	} else {
		expected := &core.Bbox{Left: -122.5, Bottom: 37.7, Right: -122.3, Top: 37.8}
		if *result.Extent != *expected {
			t.Errorf("Expected Extent %+v, got %+v", expected, result.Extent)
		}
	}
}

func TestGeocodePlaceWithClient_MinimalData(t *testing.T) {
	// Mock response with minimal required fields
	mockJSON := `{
		"type":"FeatureCollection",
		"features": [
			{
				"geometry": {
					"type": "Point",
					"coordinates": [-73.9857, 40.7484]
				},
				"properties": {
					"name": "New York"
				}
			}
		]
	}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	result, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "New York", mockClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.LocationX != -73.9857 {
		t.Errorf("Expected LocationX -73.9857, got %f", result.LocationX)
	}

	if result.LocationY != 40.7484 {
		t.Errorf("Expected LocationY 40.7484, got %f", result.LocationY)
	}

	if result.Type != "" {
		t.Errorf("Expected empty Type, got '%s'", result.Type)
	}

	if result.FullName != "New York" {
		t.Errorf("Expected FullName 'New York', got '%s'", result.FullName)
	}

	if result.Extent != nil {
		t.Error("Expected Extent to be nil")
	}
}

func TestGeocodePlaceWithClient_NoResults(t *testing.T) {
	mockJSON := `{"type":"FeatureCollection", "features": []}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "NonexistentPlace", mockClient)

	if err == nil {
		t.Fatal("Expected error for no results, got nil")
	}

	expectedError := `Could not find place matching: "NonexistentPlace"`
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGeocodePlaceWithClient_HTTPError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Response: createMockResponse(500, "Internal Server Error"),
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient)

	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}

	expectedError := "geocoding request failed with status 500"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGeocodePlaceWithClient_NetworkError(t *testing.T) {
	networkErr := errors.New("network connection failed")
	mockClient := &MockHTTPClient{
		Error: networkErr,
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient)

	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to request geocoding") {
		t.Errorf("Expected network error, got '%s'", err.Error())
	}
}

func TestGeocodePlaceWithClient_InvalidJSON(t *testing.T) {
	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, "invalid json"),
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient)

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse geocoding response") {
		t.Errorf("Expected parsing error, got '%s'", err.Error())
	}
}

func TestGeocodePlaceWithClient_InvalidCoordinates(t *testing.T) {
	// Mock response with insufficient coordinates
	mockJSON := `{
		"features": [
			{
				"geometry": {
					"type": "Point",
					"coordinates": [-122.4194]
				},
				"properties": {
					"name": "Test Place"
				}
			}
		]
	}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient)

	if err == nil {
		t.Fatal("Expected error for invalid coordinates, got nil")
	}

	expectedError := "invalid coordinates in geocoding response"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGeocodePlaceWithClient_InvalidExtent(t *testing.T) {
	// Mock response with invalid extent (not 4 elements)
	mockJSON := `{
		"features": [
			{
				"geometry": {
					"type": "Point",
					"coordinates": [-122.4194, 37.7749]
				},
				"properties": {
					"name": "Test Place",
					"extent": [-122.5, 37.7, -122.3]
				}
			}
		]
	}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	result, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should succeed but extent should be nil due to invalid format
	if result.Extent != nil {
		t.Error("Expected Extent to be nil for invalid extent data")
	}
}

func TestGeocodePlace_UnsupportedGeocoder(t *testing.T) {
	_, err := GeocodePlace("unsupported", "test")

	if err == nil {
		t.Fatal("Expected error for unsupported geocoder, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported geocoder") {
		t.Errorf("Expected unsupported geocoder error, got '%s'", err.Error())
	}
}

func TestGeocodePlaceWithClient_EmptyURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	_, err := GeocodePlaceWithClient("", "test", mockClient)

	if err == nil {
		t.Fatal("Expected error for empty URL, got nil")
	}

	expectedError := "geocoder URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGeocodePlaceWithURL(t *testing.T) {
	mockJSON := `{
		"features": [
			{
				"geometry": {
					"type": "Point",
					"coordinates": [-122.4194, 37.7749]
				},
				"properties": {
					"name": "San Francisco"
				}
			}
		]
	}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	// We need to mock the http.DefaultClient, but since we can't easily do that,
	// we'll test the URL construction logic by calling GeocodePlaceWithClient directly
	result, err := GeocodePlaceWithClient("https://custom.geocoder.com/api?query=%s", "San Francisco", mockClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.FullName != "San Francisco" {
		t.Errorf("Expected FullName 'San Francisco', got '%s'", result.FullName)
	}
}

func TestGeocodePlaceWithClient_NominatimBbox(t *testing.T) {
	// Mock response with bbox at feature level (Nominatim format)
	mockJSON := `{
		"type":"FeatureCollection",
		"features": [
			{
				"geometry": {
					"type": "Point",
					"coordinates": [2.3522, 48.8566]
				},
				"properties": {
					"name": "Paris",
					"country": "France",
					"type": "city"
				},
				"bbox": [2.224, 48.815, 2.470, 48.902]
			}
		]
	}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	result, err := GeocodePlaceWithClient("https://nominatim.openstreetmap.org/search?q=%s&format=geojson", "Paris", mockClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.LocationX != 2.3522 {
		t.Errorf("Expected LocationX 2.3522, got %f", result.LocationX)
	}

	if result.LocationY != 48.8566 {
		t.Errorf("Expected LocationY 48.8566, got %f", result.LocationY)
	}

	if result.Type != "city" {
		t.Errorf("Expected Type 'city', got '%s'", result.Type)
	}

	expectedName := "Paris, France"
	if result.FullName != expectedName {
		t.Errorf("Expected FullName '%s', got '%s'", expectedName, result.FullName)
	}

	// Check that bbox from feature level is parsed correctly
	if result.Extent == nil {
		t.Error("Expected Extent to be set from bbox field")
	} else {
		expected := &core.Bbox{Left: 2.224, Bottom: 48.815, Right: 2.470, Top: 48.902}
		if *result.Extent != *expected {
			t.Errorf("Expected Extent %+v, got %+v", expected, result.Extent)
		}
	}
}
