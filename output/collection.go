package output

import (
	"fmt"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/mikeocool/bbox/geojson"
)

// JoinedFormatCollection formats a collection of bboxes using the provided formatter function
// and joins the results with newlines
func JoinedFormatCollection(formatter func(OutputSettings, core.Bbox) (string, error), boxes []core.Bbox) (string, error) {
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

// SpaceFormatCollection formats a collection of bboxes as space-separated coordinates.
func SpaceFormatCollection(_ OutputSettings, boxes []core.Bbox) (string, error) {
	return JoinedFormatCollection(SpaceFormat, boxes)
}

// CommaFormatCollection formats a collection of bboxes as comma-separated coordinates.
func CommaFormatCollection(_ OutputSettings, boxes []core.Bbox) (string, error) {
	return JoinedFormatCollection(CommaFormat, boxes)
}

// TabFormatCollection formats a collection of bboxes as tab-separated coordinates.
func TabFormatCollection(_ OutputSettings, boxes []core.Bbox) (string, error) {
	return JoinedFormatCollection(TabFormat, boxes)
}

// GeojsonFormatCollection formats a collection of bboxes as a GeoJSON FeatureCollection or GeometryCollection.
func GeojsonFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	geojsonType := strings.ToLower(settings.GeojsonType)

	geoms := make([]geojson.Geometry, len(boxes))
	for i, box := range boxes {
		geoms[i] = geojson.PolygonGeometry([][][2]float64{box.Polygon()})
	}

	return geojson.Format(geoms, geojsonType, settings.GeojsonIndent)
}

// WktFormatCollection formats a collection of bboxes as a WKT GEOMETRYCOLLECTION.
func WktFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	polys := make([]string, len(boxes))
	for i, box := range boxes {
		poly, _ := WktFormat(settings, box)
		polys[i] = poly
	}
	val := fmt.Sprintf("GEOMETRYCOLLECTION(%s)", strings.Join(polys, ",\n"))
	return val, nil
}

// collectionOutputFormatters maps format type constants to their corresponding format functions
var collectionOutputFormatters = map[string]func(OutputSettings, []core.Bbox) (string, error){
	// TODO go template (test with for loops)
	FormatComma:   CommaFormatCollection,
	FormatSpace:   SpaceFormatCollection,
	FormatTab:     TabFormatCollection,
	FormatWkt:     WktFormatCollection,
	FormatGeoJson: GeojsonFormatCollection,
	// TODO url?
}

// GetCollectionFormatter returns the format function for the given format type.
func GetCollectionFormatter(formatType string) (func(OutputSettings, []core.Bbox) (string, error), error) {
	formatter, exists := collectionOutputFormatters[formatType]
	if !exists {
		return nil, fmt.Errorf("unknown output format: %s", formatType)
	}
	return formatter, nil
}

// FormatCollection formats a collection of bboxes using the specified format type.
func FormatCollection(boxes []core.Bbox, settings OutputSettings) (string, error) {
	formatter, err := GetCollectionFormatter(settings.FormatType)
	if err != nil {
		return "", err
	}
	return formatter(settings, boxes)
}
