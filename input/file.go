package input

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeocool/bbox/core"
)

func LoadFile(filename string) (core.Bbox, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	// TODO support shapefiles from zip
	case ".shp":
		return LoadShapefile(filename)
	case ".geojson", ".json", ".geojsonl", ".jsonl":
		return LoadGeojsonFile(filename)
	case ".parquet":
		return LoadGeoparquetFile(filename)
	case ".osm", ".pbf":
		return LoadOSMFile(filename)
	default:
		return ParseFileData(filename)
	}
}

func ParseFileData(filename string) (core.Bbox, error) {
	file, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, err
	}
	defer file.Close()
	return ParseData(file)
}

var ErrUnrecognizedDataFormat = fmt.Errorf("Input does not appear to be a valid format")

// Attempt to auto-detect the format and parse the data
func ParseData(r io.Reader) (core.Bbox, error) {
	var buf bytes.Buffer
	// as we read through the original reader, copy the bytes to the buffer
	teeReader := io.TeeReader(r, &buf)

	detectionBuf := make([]byte, 8192)
	fmt.Printf("ParseData db: %v", detectionBuf)
	_, err := teeReader.Read(detectionBuf)
	if err != nil && err != io.EOF {
		return core.Bbox{}, fmt.Errorf("failed to read data: %w", err)
	}

	// reader that contains the detection buffer and the rest of the reader
	fullReader := io.MultiReader(&buf, r)

	if SniffGeojson(detectionBuf) {
		box, err := ParseGeojson(fullReader)
		if err == nil {
			return box, nil
		} else if errors.Is(ErrNoFeaturesFound, err) {
			// sucessfully parsed geojson but found not features
			return core.Bbox{}, err
		}
	}

	if SniffOsmXml(detectionBuf) {
		return ParseOsmXML(fullReader)
	}

	if SniffOsmPbf(detectionBuf) {
		return ParseOsmPbf(fullReader)
	}

	if SniffShapefile(detectionBuf) {
		return ParseShapefile(fullReader)
	}

	// Note: GeoParquet files require random access and cannot be read from a stream
	if SniffGeoparquet(detectionBuf) {
		return core.Bbox{}, fmt.Errorf("GeoParquet files must be read from disk, not from stdin or streams")
	}

	return ParseSimpleRaw(fullReader)
}
