package output

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/mikeocool/bbox/geojson"
)

func TemplatedFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	return FormatWithTemplate(settings.FormatDetails, boxes)
}

// JoinedFormatCollection formats a collection of bboxes using the provided formatter function
// and joins the results with newlines
func JoinedFormatCollection(formatter func(OutputSettings, core.Bbox) (string, error), boxes []core.Bbox, settings OutputSettings) (string, error) {
	out := make([]string, len(boxes))
	for i, box := range boxes {
		// TODO pass through settings?
		val, err := formatter(settings, box)
		if err != nil {
			return "", err
		}
		out[i] = val
	}
	return strings.Join(out, "\n"), nil
}

// SpaceFormatCollection formats a collection of bboxes as space-separated coordinates.
func SpaceFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	return JoinedFormatCollection(SpaceFormat, boxes, settings)
}

// CommaFormatCollection formats a collection of bboxes as comma-separated coordinates.
func CommaFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	return JoinedFormatCollection(CommaFormat, boxes, settings)
}

// TabFormatCollection formats a collection of bboxes as tab-separated coordinates.
func TabFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	return JoinedFormatCollection(TabFormat, boxes, settings)
}

func JsonFormatCollection(_ OutputSettings, boxes []core.Bbox) (string, error) {
	bounds := make([][]float64, len(boxes))
	for i, box := range boxes {
		bounds[i] = box.Bounds()
	}

	data, err := json.Marshal(bounds)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func JsonlFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	return JoinedFormatCollection(JsonFormat, boxes, settings)
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

func GeojsonlFormatCollection(settings OutputSettings, boxes []core.Bbox) (string, error) {
	if settings.GeojsonIndent != 0 {
		return "", fmt.Errorf("GeoJSONL format does not support indentation")
	}

	return JoinedFormatCollection(GeojsonFormat, boxes, settings)
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

// WkbhexFormatCollection formats a collection of bboxes as a WKB GEOMETRYCOLLECTION encoded as hexadecimal.
func WkbhexFormatCollection(_ OutputSettings, boxes []core.Bbox) (string, error) {
	// Create buffer for WKB data
	buf := new(bytes.Buffer)

	// Write byte order (little endian)
	binary.Write(buf, binary.LittleEndian, uint8(1))

	// Write geometry type (GeometryCollection = 7)
	binary.Write(buf, binary.LittleEndian, uint32(7))

	// Write number of geometries
	binary.Write(buf, binary.LittleEndian, uint32(len(boxes)))

	// Write each polygon geometry
	for _, box := range boxes {
		coords := box.Polygon()

		// Write byte order for this geometry (little endian)
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
	}

	// Convert to hex string
	return strings.ToUpper(hex.EncodeToString(buf.Bytes())), nil
}

// collectionOutputFormatters maps format type constants to their corresponding format functions
var collectionOutputFormatters = map[string]func(OutputSettings, []core.Bbox) (string, error){
	FormatGoTpl:    TemplatedFormatCollection,
	FormatComma:    CommaFormatCollection,
	FormatSpace:    SpaceFormatCollection,
	FormatTab:      TabFormatCollection,
	FormatJson:     JsonFormatCollection,
	FormatJsonl:    JsonlFormatCollection,
	FormatGeoJson:  GeojsonFormatCollection,
	FormatGeoJsonl: GeojsonlFormatCollection,
	FormatWkt:      WktFormatCollection,
	FormatWkbhex:   WkbhexFormatCollection,
	// dublin core
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
