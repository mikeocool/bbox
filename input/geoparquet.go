package input

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/parquet-go/parquet-go"
)

const (
	defaultChunkSize   = 10000
	parquetMagicHeader = "PAR1"
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
	if filename == "" {
		return core.Bbox{}, fmt.Errorf("empty filename")
	}

	// Check if file exists and is not a directory
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to access file: %w", err)
	}
	if fileInfo.IsDir() {
		return core.Bbox{}, fmt.Errorf("path is a directory, not a file")
	}

	// Open the parquet file
	f, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	file, err := parquet.OpenFile(f, fileInfo.Size())
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to open parquet file: %w", err)
	}

	// Try to find geometry column name
	geoColumn, encoding := findGeoColumn(file)

	// If no metadata found, look for common geo column names
	if geoColumn == "" {
		geoColumn = findCommonGeoColumn(file.Schema())
		if geoColumn != "" {
			encoding = "wkb" // Default assumption for binary columns
		}
	}

	if geoColumn == "" {
		return core.Bbox{}, fmt.Errorf("no geometry column found in parquet file")
	}

	// Extract SRID from metadata
	srid := extractSRIDFromGeoParquet(file, geoColumn)

	// Read geometry data and calculate bounds
	return calculateBoundsFromParquet(file, geoColumn, encoding, srid)
}

// findGeoColumn looks for GeoParquet metadata and returns the primary geometry column
func findGeoColumn(file *parquet.File) (column string, encoding string) {
	// Check file metadata for geo metadata
	metadata := file.Metadata()
	if metadata == nil || metadata.KeyValueMetadata == nil {
		return "", ""
	}

	for _, kv := range metadata.KeyValueMetadata {
		if kv.Key == "" || kv.Value == "" {
			continue
		}

		// Look for "geo" key in metadata
		if kv.Key == "geo" {
			var geoMeta GeoParquetMetadata
			if err := json.Unmarshal([]byte(kv.Value), &geoMeta); err == nil {
				// Return primary column if specified
				if geoMeta.PrimaryColumn != "" {
					if col, exists := geoMeta.Columns[geoMeta.PrimaryColumn]; exists {
						return geoMeta.PrimaryColumn, col.Encoding
					}
				}
				// Otherwise return first geometry column found
				for name, col := range geoMeta.Columns {
					return name, col.Encoding
				}
			}
		}
	}

	return "", ""
}

// extractSRIDFromGeoParquet extracts SRID from GeoParquet metadata
func extractSRIDFromGeoParquet(file *parquet.File, geoColumn string) int {
	// Check file metadata for geo metadata
	metadata := file.Metadata()
	if metadata == nil || metadata.KeyValueMetadata == nil {
		return core.UnknownCrs
	}

	for _, kv := range metadata.KeyValueMetadata {
		if kv.Key == "" || kv.Value == "" {
			continue
		}

		// Look for "geo" key in metadata
		if kv.Key == "geo" {
			var geoMeta GeoParquetMetadata
			if err := json.Unmarshal([]byte(kv.Value), &geoMeta); err == nil {
				// Look for our specific geometry column
				if col, exists := geoMeta.Columns[geoColumn]; exists {
					return extractSRIDFromCRS(col.CRS)
				}
			}
		}
	}

	return core.UnknownCrs
}

// extractSRIDFromCRS extracts SRID from various CRS formats
func extractSRIDFromCRS(crs interface{}) int {
	if crs == nil {
		return core.UnknownCrs
	}

	// Handle different CRS formats
	switch v := crs.(type) {
	case map[string]interface{}:
		// Handle CRS as object (GeoJSON-style or PROJ-style)
		if srid, exists := v["srid"]; exists {
			if sridFloat, ok := srid.(float64); ok {
				return int(sridFloat)
			}
		}
		// Handle EPSG codes
		if properties, exists := v["properties"]; exists {
			if propsMap, ok := properties.(map[string]interface{}); ok {
				if code, exists := propsMap["code"]; exists {
					if codeFloat, ok := code.(float64); ok {
						return int(codeFloat)
					}
				}
			}
		}
		// Handle authority codes
		if authority, exists := v["authority"]; exists {
			if authMap, ok := authority.(map[string]interface{}); ok {
				if code, exists := authMap["code"]; exists {
					if codeFloat, ok := code.(float64); ok {
						return int(codeFloat)
					}
				}
			}
		}
		// Handle id field
		if id, exists := v["id"]; exists {
			if idMap, ok := id.(map[string]interface{}); ok {
				if code, exists := idMap["code"]; exists {
					if codeFloat, ok := code.(float64); ok {
						return int(codeFloat)
					}
				}
			}
		}
	case string:
		// Handle CRS as string (e.g., "EPSG:4326")
		if strings.HasPrefix(strings.ToUpper(v), "EPSG:") {
			epsgCode := strings.TrimPrefix(strings.ToUpper(v), "EPSG:")
			if code, err := strconv.Atoi(epsgCode); err == nil {
				return code
			}
		}
	case float64:
		// Handle CRS as direct numeric SRID
		return int(v)
	}

	return core.UnknownCrs
}


// findCommonGeoColumn searches for columns with common geometry names
func findCommonGeoColumn(schema *parquet.Schema) string {
	commonNames := []string{
		"geometry", "geom", "wkb_geometry", "shape", "the_geom", "geom_wkb",
		"GEOMETRY", "GEOM", "WKB_GEOMETRY", "SHAPE", "THE_GEOM", "GEOM_WKB",
	}

	// Get all column names from schema
	for _, field := range schema.Fields() {
		fieldName := field.Name()
		for _, commonName := range commonNames {
			if strings.EqualFold(fieldName, commonName) {
				return fieldName
			}
		}
	}

	return ""
}

// calculateBoundsFromParquet reads the geometry column and calculates overall bounds
func calculateBoundsFromParquet(file *parquet.File, geoColumn string, encoding string, srid int) (core.Bbox, error) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	rowGroups := file.RowGroups()
	geometryFound := false

	// Process each row group
	for _, rg := range rowGroups {
		// Find the column chunk for our geometry column
		columnChunk := findColumnChunk(rg, geoColumn)
		if columnChunk == nil {
			continue
		}

		// Read pages from the column chunk
		pages := columnChunk.Pages()
		defer pages.Close()

		// Create a buffer for reading values
		values := make([]parquet.Value, defaultChunkSize)

		for {
			page, err := pages.ReadPage()

			if err != nil && err.Error() != "EOF" {
				return core.Bbox{}, fmt.Errorf("error reading values: %w", err)
			}

			if page == nil {
				break
			}

			vr := page.Values()
			n, err := vr.ReadValues(values)
			if err != nil && err != io.EOF {
				return core.Bbox{}, fmt.Errorf("error reading page values: %w", err)
			}
			if n == 0 {
				break
			}

			// Process each value
			for i := range n {
				if values[i].IsNull() {
					continue
				}

				var wkbData []byte

				// Handle different data types
				// parquet.Value stores binary/string data as ByteArray
				bytes := values[i].ByteArray()
				if bytes == nil {
					continue
				}

				// First try to parse as hex-encoded WKB (common for string columns)
				parsed, err := ParseHexWKB(string(bytes))
				if err == nil {
					wkbData = parsed
				} else {
					// If not hex, use as raw WKB bytes
					wkbData = bytes
				}

				if len(wkbData) == 0 {
					continue
				}

				// Parse WKB bounds
				geomMinX, geomMinY, geomMaxX, geomMaxY, err := ParseWKBBounds(wkbData)
				if err != nil {
					// Skip invalid geometries
					continue
				}

				geometryFound = true

				// Update overall bounds
				if geomMinX < minX {
					minX = geomMinX
				}
				if geomMinY < minY {
					minY = geomMinY
				}
				if geomMaxX > maxX {
					maxX = geomMaxX
				}
				if geomMaxY > maxY {
					maxY = geomMaxY
				}
			}
		}
	}

	if !geometryFound {
		return core.Bbox{}, ErrNoFeaturesFound
	}

	// Create and return the bounding box with SRID
	bbox := core.Bbox{Left: minX, Bottom: minY, Right: maxX, Top: maxY, Srid: srid}

	return bbox, nil
}

// findColumnChunk finds the column chunk for a given column name
func findColumnChunk(rg parquet.RowGroup, columnName string) parquet.ColumnChunk {
	schema := rg.Schema()
	for i, field := range schema.Fields() {
		if field.Name() == columnName {
			return rg.ColumnChunks()[i]
		}
	}
	return nil
}
