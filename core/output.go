package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"text/template"
)

func ParseFormat(formatStr string) (string, string) {
	formatType := formatStr
	formatDetails := ""
	parts := strings.SplitN(formatStr, "=", 2)
	if len(parts) > 1 {
		formatType = parts[0]
		formatDetails = parts[1]
	}

	return formatType, formatDetails
}

// TemplatedFormat formats a Bbox using a given template string.
// The template can reference any of the Bbox fields using {{.FieldName}} syntax.
// For example: "{{.MinX}} {{.MinY}} {{.MaxX}} {{.MaxY}}"
func TemplatedFormat(templateStr string, bbox Bbox) (string, error) {
	tmpl, err := template.New("bbox").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, bbox); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX,MinY,MaxX,MaxY".
func CommaFormat(_ string, bbox Bbox) (string, error) {
	return TemplatedFormat("{{.Left}},{{.Bottom}},{{.Right}},{{.Top}}", bbox)
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX MinY MaxX MaxY".
func SpaceFormat(_ string, bbox Bbox) (string, error) {
	return TemplatedFormat("{{.Left}} {{.Bottom}} {{.Right}} {{.Top}}", bbox)
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX\tMinY\tMaxX\tMaxY".
func TabFormat(_ string, bbox Bbox) (string, error) {
	return TemplatedFormat("{{.Left}}\t{{.Bottom}}\t{{.Right}}\t{{.Top}}", bbox)
}

// GeojsonFormat formats a Bbox as a GeoJSON Polygon geometry.
// The returned string will be a complete GeoJSON Polygon representing the bounding box.
func GeojsonFormat(_ string, bbox Bbox) (string, error) {
	coords := bbox.Polygon()

	geojson := struct {
		Type        string         `json:"type"`
		Coordinates [][][2]float64 `json:"coordinates"`
	}{
		Type:        "Polygon",
		Coordinates: [][][2]float64{coords},
	}

	data, err := json.Marshal(geojson)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// WktFormat formats a Bbox as a WKT (Well-Known Text) Polygon geometry.
// The returned string will be in the format "POLYGON((x1 y1, x2 y2, x3 y3, x4 y4, x1 y1))".
func WktFormat(_ string, bbox Bbox) (string, error) {
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

func UrlFormat(urlType string, bbox Bbox) (string, error) {
	if urlType == "" {
		return "", fmt.Errorf("no url type specified")
	}

	switch strings.ToLower(urlType) {
	case "openstreetmap.org", "openstreetmap.com", "osm":
		return TemplatedFormat("https://www.openstreetmap.org/?box=yes&minlon={{.Left}}&minlat={{.Bottom}}&maxlon={{.Right}}&maxlat={{.Top}}", bbox)
	case "geojson.io":
		return GeojsonIoUrl(bbox)
	default:
		return "", fmt.Errorf("Unknown url type: %s", urlType)
	}
}

func GeojsonIoUrl(bbox Bbox) (string, error) {
	geojson, err := GeojsonFormat("", bbox)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://geojson.io/#data=data:application/json,%s", url.QueryEscape(geojson)), nil
}

// Format type constants
const (
	FormatGoTpl   = "go-template"
	FormatComma   = "comma"
	FormatSpace   = "space"
	FormatTab     = "tab"
	FormatGeoJson = "geojson"
	FormatWkt     = "wkt"
	FormatUrl     = "url"
)

// FormatFunctions maps format type constants to their corresponding format functions
var bboxOutputFormatters = map[string]func(string, Bbox) (string, error){
	FormatGoTpl:   TemplatedFormat,
	FormatComma:   CommaFormat,
	FormatSpace:   SpaceFormat,
	FormatTab:     TabFormat,
	FormatGeoJson: GeojsonFormat,
	FormatWkt:     WktFormat,
	FormatUrl:     UrlFormat,
}

// GetFormatter returns the format function for the given format type.
func GetBboxFormatter(formatType string) func(string, Bbox) (string, error) {
	return bboxOutputFormatters[formatType]
}

// Format formats a Bbox using the specified format type.
func FormatBbox(bbox Bbox, format string) (string, error) {
	formatType, formatDetails := ParseFormat(format)

	formatter := GetBboxFormatter(formatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", formatType)
	}
	return formatter(formatDetails, bbox)
}

// Point Format functions
func CommaFormatPoint(point [2]float64) (string, error) {
	return fmt.Sprintf("%g,%g", point[0], point[1]), nil
}

func SpaceFormatPoint(point [2]float64) (string, error) {
	return fmt.Sprintf("%g %g", point[0], point[1]), nil
}

func TabFormatPoint(point [2]float64) (string, error) {
	return fmt.Sprintf("%g\t%g", point[0], point[1]), nil
}

func WktFormatPoint(point [2]float64) (string, error) {
	return fmt.Sprintf("POINT (%g %g)", point[0], point[1]), nil
}

func GeojsonFormatPoint(coords [2]float64) (string, error) {
	geojson := struct {
		Type        string     `json:"type"`
		Coordinates [2]float64 `json:"coordinates"`
	}{
		Type:        "Point",
		Coordinates: coords,
	}

	data, err := json.Marshal(geojson)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

var pointOutputFormatters = map[string]func([2]float64) (string, error){
	FormatComma:   CommaFormatPoint,
	FormatSpace:   SpaceFormatPoint,
	FormatTab:     TabFormatPoint,
	FormatWkt:     WktFormatPoint,
	FormatGeoJson: GeojsonFormatPoint,
	// TODO url?
}

// GetFormatter returns the format function for the given format type.
func GetPointFormatter(formatType string) func([2]float64) (string, error) {
	return pointOutputFormatters[formatType]
}

// Format formats a Point using the specified format type.
func FormatPoint(point [2]float64, formatType string) (string, error) {
	formatter := GetPointFormatter(formatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", formatType)
	}
	return formatter(point)
}

// Collections
func SpaceFormatCollection(_ string, boxes []Bbox) (string, error) {
	out := make([]string, len(boxes))
	for _, box := range boxes {
		val, err := SpaceFormat("", box)
		if err != nil {
			return "", err
		}
		out = append(out, val)
	}

	return strings.Join(out, "\n"), nil
}

var colletionOutputFormatters = map[string]func(string, []Bbox) (string, error){
	// TOOD
	//FormatComma: CommaFormatCollection,
	FormatSpace: SpaceFormatCollection,
	// FormatTab:     TabFormatPoint,
	// FormatWkt:     WktFormatPoint,
	// FormatGeoJson: GeojsonFormatPoint,
	// TODO url?
}

// GetFormatter returns the format function for the given format type.
func GetCollectionFormatter(formatType string) func(string, []Bbox) (string, error) {
	return colletionOutputFormatters[formatType]
}

func FormatCollection(boxes []Bbox, format string) (string, error) {
	formatType, formatDetails := ParseFormat(format)

	formatter := GetCollectionFormatter(formatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", formatType)
	}
	return formatter(formatDetails, boxes)
}
