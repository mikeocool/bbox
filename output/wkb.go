package output

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/mikeocool/bbox/core"
)

// write a a bbox to wkb
func WkbPolygon(buf *bytes.Buffer, bbox core.Bbox) {
	// Write byte order (little endian)
	binary.Write(buf, binary.LittleEndian, uint8(1))

	// Write geometry type (polygon = 3)
	binary.Write(buf, binary.LittleEndian, uint32(3))

	// Write number of rings (always 1 for a simple polygon)
	binary.Write(buf, binary.LittleEndian, uint32(1))

	// Write number of points in the ring
	binary.Write(buf, binary.LittleEndian, uint32(len(bbox.Polygon())))

	// Write each coordinate pair
	for _, coord := range bbox.Polygon() {
		binary.Write(buf, binary.LittleEndian, coord[0])
		binary.Write(buf, binary.LittleEndian, coord[1])
	}
}

func WkbPoint(buf *bytes.Buffer, point [2]float64) {
	// Write byte order (little endian)
	binary.Write(buf, binary.LittleEndian, uint8(1))

	// Write geometry type (point = 1)
	binary.Write(buf, binary.LittleEndian, uint32(1))

	// Write X coordinate
	binary.Write(buf, binary.LittleEndian, point[0])

	// Write Y coordinate
	binary.Write(buf, binary.LittleEndian, point[1])
}

// Write a feature collection of bboxes
func WkbFeatureCollection(buf *bytes.Buffer, boxes []core.Bbox) {
	// Write byte order (little endian)
	binary.Write(buf, binary.LittleEndian, uint8(1))

	// Write geometry type (GeometryCollection = 7)
	binary.Write(buf, binary.LittleEndian, uint32(7))

	// Write number of geometries
	binary.Write(buf, binary.LittleEndian, uint32(len(boxes)))

	for _, box := range boxes {
		WkbPolygon(buf, box)
	}
}

func WkbHex(buf *bytes.Buffer) string {
	return strings.ToUpper(hex.EncodeToString(buf.Bytes()))
}
