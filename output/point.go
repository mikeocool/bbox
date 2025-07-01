package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/mikeocool/bbox/geojson"
)

func TemplatedFormatPoint(settings OutputSettings, point [2]float64) (string, error) {
	// Give the point values names
	pt := struct {
		X float64
		Y float64
	}{
		X: point[0],
		Y: point[1],
	}

	return FormatWithTemplate(settings.FormatDetails, pt)
}

// CommaFormatPoint formats a point as a comma-separated string of its coordinates.
// The returned string will be in the format "X,Y".
func CommaFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("%g,%g", point[0], point[1]), nil
}

// SpaceFormatPoint formats a point as a space-separated string of its coordinates.
// The returned string will be in the format "X Y".
func SpaceFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("%g %g", point[0], point[1]), nil
}

// TabFormatPoint formats a point as a tab-separated string of its coordinates.
// The returned string will be in the format "X\tY".
func TabFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("%g\t%g", point[0], point[1]), nil
}

// WktFormatPoint formats a point as a WKT (Well-Known Text) Point geometry.
// The returned string will be in the format "POINT (X Y)".
func WktFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("POINT (%g %g)", point[0], point[1]), nil
}

// WkbhexFormatPoint formats a point as a WKB (Well-Known Binary) Point geometry encoded as hexadecimal.
// The returned string will be the hexadecimal representation of the WKB binary data.
func WkbhexFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	// Create buffer for WKB data
	buf := new(bytes.Buffer)
	WkbPoint(buf, point)
	return WkbHex(buf), nil
}

func JsonFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	data, err := json.Marshal(point)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GeojsonFormatPoint formats a point as a GeoJSON Point geometry.
// The returned string will be a complete GeoJSON Point representing the coordinates.
func GeojsonFormatPoint(settings OutputSettings, coords [2]float64) (string, error) {
	geojsonType := strings.ToLower(settings.GeojsonType)

	geom := []geojson.Geometry{
		geojson.PointGeometry(coords[0], coords[1]),
	}

	return geojson.Format(geom, geojsonType, settings.GeojsonIndent)
}

func GeojsonlFormatPoint(settings OutputSettings, coords [2]float64) (string, error) {
	if settings.GeojsonIndent != 0 {
		return "", fmt.Errorf("geojsonl format does not support indentation")
	}

	return GeojsonFormatPoint(settings, coords)
}

// UrlFormatPoint formats a point as a URL to visualize it on various mapping services.
func UrlFormatPoint(settings OutputSettings, point [2]float64) (string, error) {
	urlType := settings.FormatDetails
	if urlType == "" {
		return "", fmt.Errorf("no url type specified")
	}

	var urlStr string
	var err error

	switch strings.ToLower(urlType) {
	case "openstreetmap.org", "openstreetmap.com", "osm":
		urlStr = fmt.Sprintf("https://www.openstreetmap.org/?mlat=%g&mlon=%g&zoom=16", point[1], point[0])
	case "maps.google.com", "google-maps":
		urlStr = fmt.Sprintf("https://maps.google.com/maps?q=%f,%f", point[1], point[0])
	case "geojson.io":
		urlStr, err = GeojsonIoPointUrl(point)
	default:
		return "", fmt.Errorf("unknown url type: %s", urlType)
	}

	if err != nil {
		return "", err
	}

	return urlStr, nil
}

func GeojsonIoPointUrl(point [2]float64) (string, error) {
	geojson, err := GeojsonFormatPoint(OutputSettings{}, point)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://geojson.io/#data=data:application/json,%s", url.QueryEscape(geojson)), nil
}

// pointOutputFormatters maps format type constants to their corresponding format functions
var pointOutputFormatters = map[string]func(OutputSettings, [2]float64) (string, error){
	FormatGoTpl:    TemplatedFormatPoint,
	FormatComma:    CommaFormatPoint,
	FormatSpace:    SpaceFormatPoint,
	FormatTab:      TabFormatPoint,
	FormatJson:     JsonFormatPoint,
	FormatJsonl:    JsonFormatPoint,
	FormatGeoJson:  GeojsonFormatPoint,
	FormatGeoJsonl: GeojsonlFormatPoint,
	FormatWkt:      WktFormatPoint,
	FormatWkbhex:   WkbhexFormatPoint,
	// TODO dublincore
	FormatUrl: UrlFormatPoint,
}

// GetPointFormatter returns the format function for the given format type.
func GetPointFormatter(formatType string) func(OutputSettings, [2]float64) (string, error) {
	return pointOutputFormatters[formatType]
}

// FormatPoint formats a Point using the specified format type.
func FormatPoint(point [2]float64, settings OutputSettings) (string, error) {
	formatter := GetPointFormatter(settings.FormatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", settings.FormatType)
	}
	return formatter(settings, point)
}
