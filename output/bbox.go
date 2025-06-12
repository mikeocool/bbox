package output

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/mikeocool/bbox/geojson"
)

// TemplatedFormat formats a Bbox using a given template string.
// The template can reference any of the Bbox fields using {{.FieldName}} syntax.
// For example: "{{.MinX}} {{.MinY}} {{.MaxX}} {{.MaxY}}"
func TemplatedFormat(settings OutputSettings, bbox core.Bbox) (string, error) {
	return FormatWithTemplate(settings.FormatDetails, bbox)
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX,MinY,MaxX,MaxY".
func CommaFormat(_ OutputSettings, bbox core.Bbox) (string, error) {
	return FormatWithTemplate("{{.Left}},{{.Bottom}},{{.Right}},{{.Top}}", bbox)
}

// SpaceFormat formats a Bbox as a space-separated string of its coordinates.
// The returned string will be in the format "MinX MinY MaxX MaxY".
func SpaceFormat(_ OutputSettings, bbox core.Bbox) (string, error) {
	return FormatWithTemplate("{{.Left}} {{.Bottom}} {{.Right}} {{.Top}}", bbox)
}

// TabFormat formats a Bbox as a tab-separated string of its coordinates.
// The returned string will be in the format "MinX\tMinY\tMaxX\tMaxY".
func TabFormat(_ OutputSettings, bbox core.Bbox) (string, error) {
	return FormatWithTemplate("{{.Left}}\t{{.Bottom}}\t{{.Right}}\t{{.Top}}", bbox)
}

// GeojsonFormat formats a Bbox as a GeoJSON Polygon geometry.
// The returned string will be a complete GeoJSON Polygon representing the bounding box.
func GeojsonFormat(settings OutputSettings, bbox core.Bbox) (string, error) {
	geojsonType := strings.ToLower(settings.GeojsonType)

	geom := []geojson.Geometry{
		geojson.PolygonGeometry([][][2]float64{bbox.Polygon()}),
	}

	return geojson.Format(geom, geojsonType, settings.GeojsonIndent)
}

// WktFormat formats a Bbox as a WKT (Well-Known Text) Polygon geometry.
// The returned string will be in the format "POLYGON((x1 y1, x2 y2, x3 y3, x4 y4, x1 y1))".
func WktFormat(_ OutputSettings, bbox core.Bbox) (string, error) {
	coords := bbox.Polygon()

	// Build WKT polygon string
	wkt := "POLYGON(("
	for i, coord := range coords {
		if i > 0 {
			wkt += ", "
		}
		wkt += fmt.Sprintf("%g %g", coord[0], coord[1])
	}
	wkt += "))"

	return wkt, nil
}

// WkbhexFormat formats a Bbox as a WKB (Well-Known Binary) Polygon geometry encoded as hexadecimal.
// The returned string will be the hexadecimal representation of the WKB binary data.
func WkbhexFormat(_ OutputSettings, bbox core.Bbox) (string, error) {
	coords := bbox.Polygon()

	// Create buffer for WKB data
	buf := new(bytes.Buffer)

	// Write byte order (little endian)
	binary.Write(buf, binary.LittleEndian, uint8(1))

	// Write geometry type (polygon = 3)
	binary.Write(buf, binary.LittleEndian, uint32(3))

	// Write number of rings (always 1 for a simple polygon)
	binary.Write(buf, binary.LittleEndian, uint32(1))

	// Write number of points in the ring
	binary.Write(buf, binary.LittleEndian, uint32(len(coords)))

	// Write each coordinate pair
	for _, coord := range coords {
		binary.Write(buf, binary.LittleEndian, coord[0])
		binary.Write(buf, binary.LittleEndian, coord[1])
	}

	// Convert to hex string
	return strings.ToUpper(hex.EncodeToString(buf.Bytes())), nil
}

// UrlFormat formats a Bbox as a URL to visualize it on various mapping services.
func UrlFormat(settings OutputSettings, bbox core.Bbox) (string, error) {
	urlType := settings.FormatDetails
	if urlType == "" {
		return "", fmt.Errorf("no url type specified")
	}

	switch strings.ToLower(urlType) {
	case "openstreetmap.org", "openstreetmap.com", "osm":
		return FormatWithTemplate("https://www.openstreetmap.org/?box=yes&minlon={{.Left}}&minlat={{.Bottom}}&maxlon={{.Right}}&maxlat={{.Top}}", bbox)
	case "geojson.io":
		return GeojsonIoUrl(bbox)
	default:
		return "", fmt.Errorf("Unknown url type: %s", urlType)
	}
}

// GeojsonIoUrl creates a URL to visualize the bbox on geojson.io.
func GeojsonIoUrl(bbox core.Bbox) (string, error) {
	geojson, err := GeojsonFormat(OutputSettings{}, bbox)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://geojson.io/#data=data:application/json,%s", url.QueryEscape(geojson)), nil
}

// bboxOutputFormatters maps format type constants to their corresponding format functions
var bboxOutputFormatters = map[string]func(OutputSettings, core.Bbox) (string, error){
	FormatGoTpl:   TemplatedFormat,
	FormatComma:   CommaFormat,
	FormatSpace:   SpaceFormat,
	FormatTab:     TabFormat,
	FormatGeoJson: GeojsonFormat,
	FormatWkt:     WktFormat,
	FormatWkbhex:  WkbhexFormat,
	FormatUrl:     UrlFormat,
}

// GetBboxFormatter returns the format function for the given format type.
func GetBboxFormatter(formatType string) func(OutputSettings, core.Bbox) (string, error) {
	return bboxOutputFormatters[formatType]
}

// FormatBbox formats a Bbox using the specified format type.
func FormatBbox(bbox core.Bbox, settings OutputSettings) (string, error) {
	formatter := GetBboxFormatter(settings.FormatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", settings.FormatType)
	}
	return formatter(settings, bbox)
}
