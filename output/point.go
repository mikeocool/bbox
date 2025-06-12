package output

import (
	"fmt"
	"strings"

	"github.com/mikeocool/bbox/geojson"
)

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

// GeojsonFormatPoint formats a point as a GeoJSON Point geometry.
// The returned string will be a complete GeoJSON Point representing the coordinates.
func GeojsonFormatPoint(settings OutputSettings, coords [2]float64) (string, error) {
	geojsonType := strings.ToLower(settings.GeojsonType)

	geom := []geojson.Geometry{
		geojson.PointGeometry(coords[0], coords[1]),
	}

	return geojson.Format(geom, geojsonType, settings.GeojsonIndent)
}

// pointOutputFormatters maps format type constants to their corresponding format functions
var pointOutputFormatters = map[string]func(OutputSettings, [2]float64) (string, error){
	// TODO go template
	FormatComma:   CommaFormatPoint,
	FormatSpace:   SpaceFormatPoint,
	FormatTab:     TabFormatPoint,
	FormatWkt:     WktFormatPoint,
	FormatGeoJson: GeojsonFormatPoint,
	// TODO url?
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
