package input

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/paulmach/osm/osmxml"
)

// SniffOsmXml checks if the data looks like an OSM XML file
func SniffOsmXml(data []byte) bool {
	if len(data) < 10 {
		return false
	}

	// Check for XML format with OSM root element
	dataStr := string(data[:min(200, len(data))])
	return strings.Contains(dataStr, "<?xml") && strings.Contains(dataStr, "<osm")
}

// SniffOsmPbf checks if the data looks like an OSM PBF file
func SniffOsmPbf(data []byte) bool {
	if len(data) < 20 {
		return false
	}

	// PBF files start with a blob header containing:
	// - 4 bytes: blob header size (big-endian)
	// - blob header containing type field
	// Look for typical PBF patterns in the first few bytes

	// Check if it starts with reasonable blob header size (typically 13-20 bytes)
	blobHeaderSize := (uint32(data[0]) << 24) | (uint32(data[1]) << 16) | (uint32(data[2]) << 8) | uint32(data[3])
	if blobHeaderSize < 9 || blobHeaderSize > 64*1024 {
		return false
	}

	// Check if we have enough data to examine the blob header
	if len(data) < int(4+blobHeaderSize) {
		return false
	}

	// Look for "OSMHeader" or "OSMData" type strings in the blob header
	// These are protobuf encoded, but we can search for the string patterns
	blobHeader := data[4 : 4+blobHeaderSize]
	blobHeaderStr := string(blobHeader)

	return strings.Contains(blobHeaderStr, "OSMHeader") || strings.Contains(blobHeaderStr, "OSMData")
}

// ParseOsmXML parses OSM XML data from a reader and returns its bounding box
func ParseOsmXML(r io.Reader) (core.Bbox, error) {
	scanner := osmxml.New(context.Background(), r)
	defer scanner.Close()
	return calculateBoundsFromScanner(scanner)
}

// ParseOsmPbf parses OSM PBF data from a reader and returns its bounding box
func ParseOsmPbf(r io.Reader) (core.Bbox, error) {
	scanner := osmpbf.New(context.Background(), r, 3)
	defer scanner.Close()
	return calculateBoundsFromScanner(scanner)
}

// calculateBoundsFromScanner iterates through an OSM scanner and calculates the bounding box
func calculateBoundsFromScanner(scanner osm.Scanner) (core.Bbox, error) {
	minLat, minLon := math.Inf(1), math.Inf(1)
	maxLat, maxLon := math.Inf(-1), math.Inf(-1)
	nodeFound := false

	// Scan through all objects, but only process nodes for bounding box
	for scanner.Scan() {
		obj := scanner.Object()

		// Only care about nodes for bounding box calculation
		if node, ok := obj.(*osm.Node); ok {
			nodeFound = true

			// Update bounds
			if node.Lat < minLat {
				minLat = node.Lat
			}
			if node.Lat > maxLat {
				maxLat = node.Lat
			}
			if node.Lon < minLon {
				minLon = node.Lon
			}
			if node.Lon > maxLon {
				maxLon = node.Lon
			}
		}
		// Skip ways and relations - they just reference nodes
	}

	if err := scanner.Err(); err != nil {
		return core.Bbox{}, fmt.Errorf("error scanning OSM file: %w", err)
	}

	if !nodeFound {
		return core.Bbox{}, fmt.Errorf("no nodes found in OSM file")
	}

	// Create and return the bounding box
	// OSM uses lat/lon (WGS84), core.Bbox uses Left/Bottom/Right/Top
	bbox := core.Bbox{
		Left:   minLon, // Left = minimum longitude
		Bottom: minLat, // Bottom = minimum latitude
		Right:  maxLon, // Right = maximum longitude
		Top:    maxLat, // Top = maximum latitude
	}

	return bbox, nil
}

// LoadOSMFile loads an OSM file (XML or PBF) and returns its bounding box
func LoadOSMFile(filename string) (core.Bbox, error) {
	file, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, err
	}
	defer file.Close()

	// Call appropriate parser based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".pbf" {
		return ParseOsmPbf(file)
	} else {
		// Default to XML parser for .osm files or when extension is unclear
		return ParseOsmXML(file)
	}
}
