package input

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/paulmach/osm/osmxml"
)

// SniffOSM checks if the data looks like an OSM XML file
func SniffOSM(data []byte) bool {
	if len(data) < 10 {
		return false
	}

	// Check for XML format with OSM root element
	dataStr := string(data[:min(200, len(data))])
	return strings.Contains(dataStr, "<?xml") && strings.Contains(dataStr, "<osm")
}

// LoadOSMFile loads an OSM file (XML or PBF) and returns its bounding box
func LoadOSMFile(filename string) (core.Bbox, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create OSM scanner based on file extension
	var scanner osm.Scanner

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".pbf" {
		scanner = osmpbf.New(context.Background(), file, 3)
	} else {
		// Default to XML scanner for .osm files or when extension is unclear
		scanner = osmxml.New(context.Background(), file)
	}
	defer scanner.Close()

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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
