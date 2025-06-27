package input

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/mikeocool/bbox/core"
)

// SniffWkbHex checks if the data looks like hexadecimal Well-Known Binary (WKB) format
func SniffWkbHex(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Trim whitespace and convert to string for analysis
	s := strings.TrimSpace(string(data))

	// Must have even length for valid hex
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}

	// Must be at least 10 characters (5 bytes) for minimal WKB geometry
	if len(s) < 10 {
		return false
	}

	// Check if all characters are valid hexadecimal
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

const (
	wkbPoint              = 1
	wkbLineString         = 2
	wkbPolygon            = 3
	wkbMultiPoint         = 4
	wkbMultiLineString    = 5
	wkbMultiPolygon       = 6
	wkbGeometryCollection = 7
)

type wkbReader struct {
	r         io.Reader
	byteOrder binary.ByteOrder
}

// parseHexWKB converts hex-encoded WKB string to bytes
func ParseHexWKB(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("invalid hex string length")
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		bytes[i/2] = b
	}
	return bytes, nil
}

// ParseWKBBounds extracts the bounding box from WKB data without fully parsing the geometry
func ParseWKBBounds(wkb []byte) (minX, minY, maxX, maxY float64, err error) {
	if len(wkb) < 5 {
		return 0, 0, 0, 0, fmt.Errorf("WKB data too short")
	}

	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)

	reader := &wkbReader{r: bytes.NewReader(wkb)}

	// Read byte order
	var byteOrderFlag uint8
	if err := binary.Read(reader.r, binary.LittleEndian, &byteOrderFlag); err != nil {
		return 0, 0, 0, 0, err
	}

	if byteOrderFlag == 0 {
		reader.byteOrder = binary.BigEndian
	} else {
		reader.byteOrder = binary.LittleEndian
	}

	// Read geometry type
	var geomType uint32
	if err := binary.Read(reader.r, reader.byteOrder, &geomType); err != nil {
		return 0, 0, 0, 0, err
	}

	// Extract bounds based on geometry type
	if err := reader.extractBounds(geomType, &minX, &minY, &maxX, &maxY); err != nil {
		return 0, 0, 0, 0, err
	}

	// Check if we found any valid coordinates
	if math.IsInf(minX, 1) || math.IsInf(minY, 1) || math.IsInf(maxX, -1) || math.IsInf(maxY, -1) {
		return 0, 0, 0, 0, fmt.Errorf("no valid coordinates found in WKB")
	}

	return minX, minY, maxX, maxY, nil
}

func (w *wkbReader) extractBounds(geomType uint32, minX, minY, maxX, maxY *float64) error {
	// Handle EWKB SRID flag
	if geomType&0x20000000 != 0 {
		// Read and ignore SRID
		var srid uint32
		if err := binary.Read(w.r, w.byteOrder, &srid); err != nil {
			return err
		}
	}

	// Strip any flags from geometry type (like SRID flag)
	baseType := geomType & 0xff

	switch baseType {
	case wkbPoint:
		return w.readPointBounds(minX, minY, maxX, maxY)
	case wkbLineString:
		return w.readLineStringBounds(minX, minY, maxX, maxY)
	case wkbPolygon:
		return w.readPolygonBounds(minX, minY, maxX, maxY)
	case wkbMultiPoint:
		return w.readMultiPointBounds(minX, minY, maxX, maxY)
	case wkbMultiLineString:
		return w.readMultiLineStringBounds(minX, minY, maxX, maxY)
	case wkbMultiPolygon:
		return w.readMultiPolygonBounds(minX, minY, maxX, maxY)
	case wkbGeometryCollection:
		return w.readGeometryCollectionBounds(minX, minY, maxX, maxY)
	default:
		return fmt.Errorf("unsupported geometry type: %d", baseType)
	}
}

func (w *wkbReader) updateBounds(x, y float64, minX, minY, maxX, maxY *float64) {
	if x < *minX {
		*minX = x
	}
	if x > *maxX {
		*maxX = x
	}
	if y < *minY {
		*minY = y
	}
	if y > *maxY {
		*maxY = y
	}
}

func (w *wkbReader) readPointBounds(minX, minY, maxX, maxY *float64) error {
	var x, y float64
	if err := binary.Read(w.r, w.byteOrder, &x); err != nil {
		return err
	}
	if err := binary.Read(w.r, w.byteOrder, &y); err != nil {
		return err
	}
	w.updateBounds(x, y, minX, minY, maxX, maxY)
	return nil
}

func (w *wkbReader) readLineStringBounds(minX, minY, maxX, maxY *float64) error {
	var numPoints uint32
	if err := binary.Read(w.r, w.byteOrder, &numPoints); err != nil {
		return err
	}

	for i := range numPoints {
		_ = i // silence unused variable warning
		var x, y float64
		if err := binary.Read(w.r, w.byteOrder, &x); err != nil {
			return err
		}
		if err := binary.Read(w.r, w.byteOrder, &y); err != nil {
			return err
		}
		w.updateBounds(x, y, minX, minY, maxX, maxY)
	}
	return nil
}

func (w *wkbReader) readPolygonBounds(minX, minY, maxX, maxY *float64) error {
	var numRings uint32
	if err := binary.Read(w.r, w.byteOrder, &numRings); err != nil {
		return err
	}

	for ring := range numRings {
		_ = ring // silence unused variable warning
		if err := w.readLineStringBounds(minX, minY, maxX, maxY); err != nil {
			return err
		}
	}
	return nil
}

func (w *wkbReader) readMultiPointBounds(minX, minY, maxX, maxY *float64) error {
	var numPoints uint32
	if err := binary.Read(w.r, w.byteOrder, &numPoints); err != nil {
		return err
	}

	for i := range numPoints {
		_ = i // silence unused variable warning
		// Each point has its own byte order and type header
		var byteOrderFlag uint8
		if err := binary.Read(w.r, binary.LittleEndian, &byteOrderFlag); err != nil {
			return err
		}

		if byteOrderFlag == 0 {
			w.byteOrder = binary.BigEndian
		} else {
			w.byteOrder = binary.LittleEndian
		}

		var geomType uint32
		if err := binary.Read(w.r, w.byteOrder, &geomType); err != nil {
			return err
		}

		if err := w.readPointBounds(minX, minY, maxX, maxY); err != nil {
			return err
		}
	}
	return nil
}

func (w *wkbReader) readMultiLineStringBounds(minX, minY, maxX, maxY *float64) error {
	var numLineStrings uint32
	if err := binary.Read(w.r, w.byteOrder, &numLineStrings); err != nil {
		return err
	}

	for i := range numLineStrings {
		_ = i // silence unused variable warning
		// Each linestring has its own byte order and type header
		var byteOrderFlag uint8
		if err := binary.Read(w.r, binary.LittleEndian, &byteOrderFlag); err != nil {
			return err
		}

		if byteOrderFlag == 0 {
			w.byteOrder = binary.BigEndian
		} else {
			w.byteOrder = binary.LittleEndian
		}

		var geomType uint32
		if err := binary.Read(w.r, w.byteOrder, &geomType); err != nil {
			return err
		}

		if err := w.readLineStringBounds(minX, minY, maxX, maxY); err != nil {
			return err
		}
	}
	return nil
}

func (w *wkbReader) readMultiPolygonBounds(minX, minY, maxX, maxY *float64) error {
	var numPolygons uint32
	if err := binary.Read(w.r, w.byteOrder, &numPolygons); err != nil {
		return err
	}

	for i := range numPolygons {
		_ = i // silence unused variable warning
		// Each polygon has its own byte order and type header
		var byteOrderFlag uint8
		if err := binary.Read(w.r, binary.LittleEndian, &byteOrderFlag); err != nil {
			return err
		}

		if byteOrderFlag == 0 {
			w.byteOrder = binary.BigEndian
		} else {
			w.byteOrder = binary.LittleEndian
		}

		var geomType uint32
		if err := binary.Read(w.r, w.byteOrder, &geomType); err != nil {
			return err
		}

		if err := w.readPolygonBounds(minX, minY, maxX, maxY); err != nil {
			return err
		}
	}
	return nil
}

func (w *wkbReader) readGeometryCollectionBounds(minX, minY, maxX, maxY *float64) error {
	var numGeometries uint32
	if err := binary.Read(w.r, w.byteOrder, &numGeometries); err != nil {
		return err
	}

	for i := range numGeometries {
		_ = i // silence unused variable warning
		// Each geometry has its own byte order and type header
		var byteOrderFlag uint8
		if err := binary.Read(w.r, binary.LittleEndian, &byteOrderFlag); err != nil {
			return err
		}

		if byteOrderFlag == 0 {
			w.byteOrder = binary.BigEndian
		} else {
			w.byteOrder = binary.LittleEndian
		}

		var geomType uint32
		if err := binary.Read(w.r, w.byteOrder, &geomType); err != nil {
			return err
		}

		if err := w.extractBounds(geomType, minX, minY, maxX, maxY); err != nil {
			return err
		}
	}
	return nil
}

// ParseWKBToBbox parses WKB binary data and returns a core.Bbox
func ParseWKBToBbox(wkb []byte) (core.Bbox, error) {
	minX, minY, maxX, maxY, err := ParseWKBBounds(wkb)
	if err != nil {
		return core.Bbox{}, err
	}

	bbox := core.Bbox{
		Left:   minX,
		Bottom: minY,
		Right:  maxX,
		Top:    maxY,
	}

	if err := bbox.Validate(); err != nil {
		return core.Bbox{}, err
	}

	return bbox, nil
}

// ParseHexWKBToBbox parses hex-encoded WKB data and returns a core.Bbox
func ParseHexWKBToBbox(hexStr string) (core.Bbox, error) {
	hexStr = strings.TrimSpace(hexStr)
	wkb, err := ParseHexWKB(hexStr)
	if err != nil {
		return core.Bbox{}, err
	}

	return ParseWKBToBbox(wkb)
}
