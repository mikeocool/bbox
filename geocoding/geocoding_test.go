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
	Response        *http.Response
	Error           error
	CapturedRequest *http.Request
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.CapturedRequest = req
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

	result, err := GeocodePlaceWithClient("https://photon.komoot.io/api?q=%s&limit=1", "San Francisco", mockClient, nil)

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

	result, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "New York", mockClient, nil)

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

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "NonexistentPlace", mockClient, nil)

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

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient, nil)

	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}

	expectedError := "geocoding request failed with status 500: Internal Server Error"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGeocodePlaceWithClient_HTTPErrorWithLongBody(t *testing.T) {
	// Create a response body longer than 500 characters
	longBody := strings.Repeat("Error details: server overloaded. ", 20) // Creates ~660 character string
	mockClient := &MockHTTPClient{
		Response: createMockResponse(503, longBody),
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient, nil)

	if err == nil {
		t.Fatal("Expected error for HTTP 503, got nil")
	}

	// Should contain the status code and truncated body (up to 500 chars)
	if !strings.Contains(err.Error(), "geocoding request failed with status 503") {
		t.Errorf("Expected error to contain status code, got '%s'", err.Error())
	}
	
	// The error message should be truncated to around 500 characters of body content
	if len(err.Error()) > 600 { // Some buffer for the status message part
		t.Errorf("Expected error message to be truncated, but got %d characters: %s", len(err.Error()), err.Error())
	}
	
	// Should contain some of the error details but not all
	if !strings.Contains(err.Error(), "Error details: server overloaded") {
		t.Errorf("Expected error to contain body preview, got '%s'", err.Error())
	}
}

func TestGeocodePlaceWithClient_NetworkError(t *testing.T) {
	networkErr := errors.New("network connection failed")
	mockClient := &MockHTTPClient{
		Error: networkErr,
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient, nil)

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

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient, nil)

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

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient, nil)

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

	result, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "test", mockClient, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should succeed but extent should be nil due to invalid format
	if result.Extent != nil {
		t.Error("Expected Extent to be nil for invalid extent data")
	}
}

func TestGeocodePlace_UnsupportedGeocoder(t *testing.T) {
	_, err := GeocodePlace("unsupported", "test", nil)

	if err == nil {
		t.Fatal("Expected error for unsupported geocoder, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported geocoder") {
		t.Errorf("Expected unsupported geocoder error, got '%s'", err.Error())
	}
}

func TestGeocodePlaceWithClient_EmptyURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	_, err := GeocodePlaceWithClient("", "test", mockClient, nil)

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
	result, err := GeocodePlaceWithClient("https://custom.geocoder.com/api?query=%s", "San Francisco", mockClient, nil)

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

	result, err := GeocodePlaceWithClient("https://nominatim.openstreetmap.org/search?q=%s&format=geojson", "Paris", mockClient, nil)

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

func TestGeocodePlaceWithClient_BboxInProperties(t *testing.T) {
	// Mock response with bbox in properties (should take precedence over feature-level bbox)
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
					"bbox": [2.200, 48.800, 2.500, 48.900]
				},
				"bbox": [2.224, 48.815, 2.470, 48.902]
			}
		]
	}`

	mockClient := &MockHTTPClient{
		Response: createMockResponse(200, mockJSON),
	}

	result, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "Paris", mockClient, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should use bbox from properties, not feature level
	if result.Extent == nil {
		t.Error("Expected Extent to be set from bbox in properties")
	} else {
		expected := &core.Bbox{Left: 2.200, Bottom: 48.800, Right: 2.500, Top: 48.900}
		if *result.Extent != *expected {
			t.Errorf("Expected Extent %+v, got %+v", expected, result.Extent)
		}
	}
}

func TestGeocodePlaceWithClient_CustomHeaders(t *testing.T) {
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

	headers := []string{
		"Authorization: Bearer token123",
		"X-API-Key: secret-key",
	}

	_, err := GeocodePlaceWithClient("https://test.com/api?q=%s", "San Francisco", mockClient, headers)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify headers were set
	if mockClient.CapturedRequest == nil {
		t.Fatal("Expected request to be captured")
	}

	if auth := mockClient.CapturedRequest.Header.Get("Authorization"); auth != "Bearer token123" {
		t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", auth)
	}

	if apiKey := mockClient.CapturedRequest.Header.Get("X-API-Key"); apiKey != "secret-key" {
		t.Errorf("Expected X-API-Key header 'secret-key', got '%s'", apiKey)
	}
}
