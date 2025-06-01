package input

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/twpayne/go-shapefile"
)

func LoadFile(filename string) (core.Bbox, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	// TODO support shapefiles from zip
	case ".shp":
		return LoadShapefile(filename)
	case ".geojson", ".json":
		return LoadGeojsonFile(filename)
	default:
		return ParseFile(filename)
	}
}

func ParseFile(filename string) (core.Bbox, error) {
	file, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, err
	}
	defer file.Close()
	return ParseData(file)
}

// Attempt to auto-detect the format and parse the data
func ParseData(r io.Reader) (core.Bbox, error) {
	var buf bytes.Buffer
	// as we read through the original reader, copy the bytes to the buffer
	teeReader := io.TeeReader(r, &buf)

	detectionBuf := make([]byte, 8192)
	_, err := teeReader.Read(detectionBuf)
	if err != nil && err != io.EOF {
		return core.Bbox{}, fmt.Errorf("failed to read data: %w", err)
	}

	// reader that contains the detection buffer and the rest of the reader
	fullReader := io.MultiReader(&buf, r)

	if SniffGeojson(detectionBuf) {
		box, err := ParseGeojson(fullReader)
		// TODO on certain error keep trying
		if err == nil {
			return box, nil
		}
	}

	if SniffShapefile(detectionBuf) {
		box, err := ParseShapefile(fullReader)
		if err == nil {
			return box, nil
		}
	}

	return core.Bbox{}, fmt.Errorf("Input does not appear to be a valid format")
}

func SniffShapefile(data []byte) bool {
	if len(data) < 100 {
		return false
	}

	// Shapefile main file (.shp) has a specific header structure
	// File code should be 9994 (0x270A) in big-endian at bytes 0-3
	if len(data) >= 4 {
		fileCode := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		if fileCode == 9994 {
			return true
		}
	}

	return false
}

func LoadShapefile(filename string) (core.Bbox, error) {
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	opts := shapefile.ReadShapefileOptions{
		DBF: &shapefile.ReadDBFOptions{
			SkipBrokenFields: true,
		},
	}
	shp, err := shapefile.Read(filename, &opts)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to read shapefile: %w", err)
	}
	if shp.SHP == nil {
		return core.Bbox{}, fmt.Errorf("unexpected error reading shapefile")
	}

	// TOOD potentially actually look at geoms since this bounds could be wrong
	bounds := shp.SHP.Bounds
	// TODO if empty return error

	return core.Bbox{
		Left:   bounds.Min(0),
		Bottom: bounds.Min(1),
		Right:  bounds.Max(0),
		Top:    bounds.Max(1),
	}, nil
}

func ParseShapefile(r io.Reader) (core.Bbox, error) {
	// the shapefile lib seems to just use the filelength to verify that the file is longer than
	// the headers -- so just passing in the header length here, since we've already
	// verified it's longer
	shp, err := shapefile.ReadSHP(r, 100, nil)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("Error reading shapefile: %s", err)
	}

	if shp == nil {
		return core.Bbox{}, fmt.Errorf("unexpected error reading shapefile")
	}

	bounds := shp.Bounds
	// TODO if empty return error

	return core.Bbox{
		Left:   bounds.Min(0),
		Bottom: bounds.Min(1),
		Right:  bounds.Max(0),
		Top:    bounds.Max(1),
	}, nil
}
