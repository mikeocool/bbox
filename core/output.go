package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/mikeocool/bbox/geojson"
)

type OutputSettings struct {
	FormatType    string
	FormatDetails string
	GeojsonIndent int
	GeojsonType   string
}

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

func FormatWithTemplate(templateStr string, bbox Bbox) (string, error) {
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

// TemplatedFormat formats a Bbox using a given template string.
// The template can reference any of the Bbox fields using {{.FieldName}} syntax.
// For example: "{{.MinX}} {{.MinY}} {{.MaxX}} {{.MaxY}}"
func TemplatedFormat(settings OutputSettings, bbox Bbox) (string, error) {
	return FormatWithTemplate(settings.FormatDetails, bbox)
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX,MinY,MaxX,MaxY".
func CommaFormat(_ OutputSettings, bbox Bbox) (string, error) {
	return FormatWithTemplate("{{.Left}},{{.Bottom}},{{.Right}},{{.Top}}", bbox)
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX MinY MaxX MaxY".
func SpaceFormat(_ OutputSettings, bbox Bbox) (string, error) {
	return FormatWithTemplate("{{.Left}} {{.Bottom}} {{.Right}} {{.Top}}", bbox)
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX\tMinY\tMaxX\tMaxY".
func TabFormat(_ OutputSettings, bbox Bbox) (string, error) {
	return FormatWithTemplate("{{.Left}}\t{{.Bottom}}\t{{.Right}}\t{{.Top}}", bbox)
}

// GeojsonFormat formats a Bbox as a GeoJSON Polygon geometry.
// The returned string will be a complete GeoJSON Polygon representing the bounding box.
func GeojsonFormat(settings OutputSettings, bbox Bbox) (string, error) {
	geojsonType := strings.ToLower(settings.GeojsonType)
	// TODO ensure this is a valid geojson type

	coords := bbox.Polygon()
	if geojsonType == "coordinates" {
		return marshalGeojson(coords, settings)
	}

	geom := geojson.Geometry{
		Type:        "Polygon",
		Coordinates: json.RawMessage(func() []byte {
			data, _ := json.Marshal([][][2]float64{coords})
			return data
		}()),
	}

	if geojsonType == "geometry" {
		return marshalGeojson(geom, settings)
	}

	feature := geojson.Feature{
		Type:     "Feature",
		Geometry: geom,
	}

	if geojsonType == "feature" {
		return marshalGeojson(feature, settings)
	}

	collection := geojson.FeatureCollection{
		Type: "FeatureCollection",
		Features: []geojson.Feature{feature},
	}

	return marshalGeojson(collection, settings)
}

func marshalGeojson(geojson any, settings OutputSettings) (string, error) {
	var data []byte
	var err error
	if settings.GeojsonIndent > 0 {
		data, err = json.MarshalIndent(geojson, "", strings.Repeat(" ", settings.GeojsonIndent))
	} else {
		data, err = json.Marshal(geojson)
	}
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// WktFormat formats a Bbox as a WKT (Well-Known Text) Polygon geometry.
// The returned string will be in the format "POLYGON((x1 y1, x2 y2, x3 y3, x4 y4, x1 y1))".
func WktFormat(_ OutputSettings, bbox Bbox) (string, error) {
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

func UrlFormat(settings OutputSettings, bbox Bbox) (string, error) {
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

func GeojsonIoUrl(bbox Bbox) (string, error) {
	geojson, err := GeojsonFormat(OutputSettings{}, bbox)
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
var bboxOutputFormatters = map[string]func(OutputSettings, Bbox) (string, error){
	FormatGoTpl:   TemplatedFormat,
	FormatComma:   CommaFormat,
	FormatSpace:   SpaceFormat,
	FormatTab:     TabFormat,
	FormatGeoJson: GeojsonFormat,
	FormatWkt:     WktFormat,
	FormatUrl:     UrlFormat,
}

// GetFormatter returns the format function for the given format type.
func GetBboxFormatter(formatType string) func(OutputSettings, Bbox) (string, error) {
	return bboxOutputFormatters[formatType]
}

// Format formats a Bbox using the specified format type.
func FormatBbox(bbox Bbox, settings OutputSettings) (string, error) {

	formatter := GetBboxFormatter(settings.FormatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", settings.FormatType)
	}
	return formatter(settings, bbox)
}

// Point Format functions
func CommaFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("%g,%g", point[0], point[1]), nil
}

func SpaceFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("%g %g", point[0], point[1]), nil
}

func TabFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("%g\t%g", point[0], point[1]), nil
}

func WktFormatPoint(_ OutputSettings, point [2]float64) (string, error) {
	return fmt.Sprintf("POINT (%g %g)", point[0], point[1]), nil
}

func GeojsonFormatPoint(_ OutputSettings, coords [2]float64) (string, error) {
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

var pointOutputFormatters = map[string]func(OutputSettings, [2]float64) (string, error){
	FormatComma:   CommaFormatPoint,
	FormatSpace:   SpaceFormatPoint,
	FormatTab:     TabFormatPoint,
	FormatWkt:     WktFormatPoint,
	FormatGeoJson: GeojsonFormatPoint,
	// TODO url?
}

// GetFormatter returns the format function for the given format type.
func GetPointFormatter(formatType string) func(OutputSettings, [2]float64) (string, error) {
	return pointOutputFormatters[formatType]
}

// Format formats a Point using the specified format type.
func FormatPoint(point [2]float64, settings OutputSettings) (string, error) {
	formatter := GetPointFormatter(settings.FormatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", settings.FormatType)
	}
	return formatter(settings, point)
}

// Collections
// JoinedFormatCollection formats a collection of bboxes using the provided formatter function
// and joins the results with newlines
func JoinedFormatCollection(formatter func(OutputSettings, Bbox) (string, error), boxes []Bbox) (string, error) {
	out := make([]string, len(boxes))
	for i, box := range boxes {
		// TODO pass through settings?
		val, err := formatter(OutputSettings{}, box)
		if err != nil {
			return "", err
		}
		out[i] = val
	}
	return strings.Join(out, "\n"), nil
}

func SpaceFormatCollection(_ OutputSettings, boxes []Bbox) (string, error) {
	return JoinedFormatCollection(SpaceFormat, boxes)
}

func CommaFormatCollection(_ OutputSettings, boxes []Bbox) (string, error) {
	return JoinedFormatCollection(CommaFormat, boxes)
}

func TabFormatCollection(_ OutputSettings, boxes []Bbox) (string, error) {
	return JoinedFormatCollection(TabFormat, boxes)
}

var colletionOutputFormatters = map[string]func(OutputSettings, []Bbox) (string, error){
	// TOOD
	FormatComma: CommaFormatCollection,
	FormatSpace: SpaceFormatCollection,
	FormatTab:   TabFormatCollection,
	// FormatWkt:     WktFormatPoint,
	// FormatGeoJson: GeojsonFormatPoint,
	// TODO url?
}

// GetFormatter returns the format function for the given format type.
func GetCollectionFormatter(formatType string) (func(OutputSettings, []Bbox) (string, error), error) {
	formatter, exists := colletionOutputFormatters[formatType]
	if !exists {
		return nil, fmt.Errorf("unknown output format: %s", formatType)
	}
	return formatter, nil
}

func FormatCollection(boxes []Bbox, settings OutputSettings) (string, error) {
	formatter, err := GetCollectionFormatter(settings.FormatType)
	if err != nil {
		return "", err
	}
	return formatter(settings, boxes)
}
