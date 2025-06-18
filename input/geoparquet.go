package input

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/parquet/file"
	"github.com/apache/arrow/go/v15/parquet/pqarrow"
	"github.com/mikeocool/bbox/core"
)

const (
	defaultChunkSize   = 10000
	parquetMagicHeader = "PAR1"
	parquetMagicFooter = "PAR1"
)

// GeoParquetMetadata represents the geo metadata stored in the parquet file
type GeoParquetMetadata struct {
	Version       string                    `json:"version"`
	PrimaryColumn string                    `json:"primary_column"`
	Columns       map[string]GeometryColumn `json:"columns"`
}

type GeometryColumn struct {
	Encoding      string      `json:"encoding"`
	GeometryTypes []string    `json:"geometry_types,omitempty"`
	CRS           interface{} `json:"crs,omitempty"`
	Bbox          []float64   `json:"bbox,omitempty"`
}

// SniffGeoparquet checks if the data looks like a Parquet file
func SniffGeoparquet(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// Check for PAR1 magic bytes at the beginning
	return string(data[0:4]) == parquetMagicHeader
}

// LoadGeoparquetFile loads a GeoParquet file and returns its bounding box
func LoadGeoparquetFile(filename string) (core.Bbox, error) {
	// Open parquet file
	reader, err := file.OpenParquetFile(filename, false)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to open parquet file: %w", err)
	}
	defer reader.Close()

	// Find geometry column from metadata
	geomColumn, err := findGeometryColumn(reader)
	if err != nil {
		return core.Bbox{}, err
	}

	// Read file in chunks and accumulate bounds
	return readGeoparquetChunked(reader, geomColumn)
}

// findGeometryColumn reads the GeoParquet metadata to find the geometry column
func findGeometryColumn(reader *file.Reader) (string, error) {
	// Get file metadata
	metadata := reader.MetaData()
	if metadata == nil {
		return "", fmt.Errorf("no metadata found in parquet file")
	}

	// Look for geo metadata in key-value pairs
	keyValueMeta := metadata.KeyValueMetadata()
	if keyValueMeta == nil {
		return "", fmt.Errorf("no key-value metadata found")
	}

	// Find the "geo" key
	geoJSON := ""
	keys := keyValueMeta.Keys()
	values := keyValueMeta.Values()
	for i := range len(keys) {
		if keys[i] == "geo" {
			geoJSON = values[i]
			break
		}
	}

	if geoJSON == "" {
		return "", fmt.Errorf("no 'geo' metadata found - this may not be a GeoParquet file")
	}

	// Parse the geo metadata
	var geoMeta GeoParquetMetadata
	if err := json.Unmarshal([]byte(geoJSON), &geoMeta); err != nil {
		return "", fmt.Errorf("failed to parse geo metadata: %w", err)
	}

	// Use primary column if specified
	if geoMeta.PrimaryColumn != "" {
		return geoMeta.PrimaryColumn, nil
	}

	// Otherwise, use the first geometry column
	for colName := range geoMeta.Columns {
		return colName, nil
	}

	// TODO fallback to looking for binary colums with common geonames

	return "", fmt.Errorf("no geometry column found in geo metadata")
}

// readGeoparquetChunked reads the parquet file in chunks to handle large files
func readGeoparquetChunked(reader *file.Reader, geomColumn string) (core.Bbox, error) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	hasValidGeometry := false

	// Create arrow file reader
	pool := memory.NewGoAllocator()
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, pool)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	// Get schema to find column index
	schema, err := arrowReader.Schema()
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to get schema: %w", err)
	}

	// Find geometry column index
	geomColIndex := -1
	for i, field := range schema.Fields() {
		if field.Name == geomColumn {
			geomColIndex = i
			break
		}
	}

	if geomColIndex == -1 {
		return core.Bbox{}, fmt.Errorf("geometry column '%s' not found in schema", geomColumn)
	}

	// Process each row group
	numRowGroups := reader.NumRowGroups()
	for rgIdx := range numRowGroups {
		// Read row group
		rgReader := arrowReader.RowGroup(rgIdx)

		// Read the geometry column for this row group
		colReader := rgReader.Column(geomColIndex)
		chunkedCol, err := colReader.Read(context.Background())
		if err != nil {
			continue
		}
		defer chunkedCol.Release()

		// Process the chunked column
		if chunkedCol.Len() > 0 {
			if err := processBinaryColumn(chunkedCol, &minX, &minY, &maxX, &maxY, &hasValidGeometry); err != nil {
				// Log error but continue processing
				fmt.Printf("Warning: error processing chunk: %v\n", err)
			}
		}
	}

	if !hasValidGeometry {
		return core.Bbox{}, fmt.Errorf("no valid geometries found")
	}

	return core.Bbox{
		Left:   minX,
		Bottom: minY,
		Right:  maxX,
		Top:    maxY,
	}, nil
}

// processBinaryColumn processes a binary column containing WKB geometries
func processBinaryColumn(col *arrow.Chunked, minX, minY, maxX, maxY *float64, hasValidGeometry *bool) error {
	for chunkIdx := range col.Len() {
		chunk := col.Chunk(chunkIdx)

		// Handle different binary array types
		switch arr := chunk.(type) {
		case *array.Binary:
			processBinaryArray(arr, minX, minY, maxX, maxY, hasValidGeometry)
		case *array.LargeBinary:
			processLargeBinaryArray(arr, minX, minY, maxX, maxY, hasValidGeometry)
		case *array.String:
			// Sometimes geometry might be stored as string-encoded WKB
			processStringArray(arr, minX, minY, maxX, maxY, hasValidGeometry)
		default:
			return fmt.Errorf("unsupported array type for geometry column: %T", arr)
		}
	}
	return nil
}

func processBinaryArray(arr *array.Binary, minX, minY, maxX, maxY *float64, hasValidGeometry *bool) {
	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		wkb := arr.Value(i)
		x1, y1, x2, y2, err := ParseWKBBounds(wkb)
		if err != nil {
			continue
		}

		updateParquetBounds(minX, minY, maxX, maxY, x1, y1, x2, y2)
		*hasValidGeometry = true
	}
}

func processLargeBinaryArray(arr *array.LargeBinary, minX, minY, maxX, maxY *float64, hasValidGeometry *bool) {
	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		wkb := arr.Value(i)
		x1, y1, x2, y2, err := ParseWKBBounds(wkb)
		if err != nil {
			continue
		}

		updateParquetBounds(minX, minY, maxX, maxY, x1, y1, x2, y2)
		*hasValidGeometry = true
	}
}

func processStringArray(arr *array.String, minX, minY, maxX, maxY *float64, hasValidGeometry *bool) {
	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		// Try to parse as hex-encoded WKB
		hexStr := arr.Value(i)
		wkb, err := ParseHexWKB(hexStr)
		if err != nil {
			continue
		}

		x1, y1, x2, y2, err := ParseWKBBounds(wkb)
		if err != nil {
			continue
		}

		updateParquetBounds(minX, minY, maxX, maxY, x1, y1, x2, y2)
		*hasValidGeometry = true
	}
}

func updateParquetBounds(minX, minY, maxX, maxY *float64, x1, y1, x2, y2 float64) {
	if x1 < *minX {
		*minX = x1
	}
	if y1 < *minY {
		*minY = y1
	}
	if x2 > *maxX {
		*maxX = x2
	}
	if y2 > *maxY {
		*maxY = y2
	}
}
