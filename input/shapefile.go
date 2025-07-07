package input

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mikeocool/bbox/core"
)

const (
	shpHeaderSize    = 100
	shpHeaderVersion = 1000
	shpFileCode      = 9994
)

func SniffShapefile(data []byte) bool {
	if len(data) < shpHeaderSize {
		return false
	}

	// Shapefile main file (.shp) has a specific header structure
	// File code should be 9994 (0x270A) in big-endian at bytes 0-3
	if len(data) >= 4 {
		fileCode := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		if fileCode == shpFileCode {
			return true
		}
	}

	return false
}

func LoadShapefile(filename string) (core.Bbox, error) {
	r, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, err
	}
	defer r.Close()
	
	bbox, err := ParseShapefile(r)
	if err != nil {
		return core.Bbox{}, err
	}
	
	// Look for corresponding .prj file and parse projection
	prjPath := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".prj"
	srid := parseProjectionFile(prjPath)
	bbox.Srid = srid
	
	return bbox, nil
}

func ParseShapefile(r io.Reader) (core.Bbox, error) {
	header := make([]byte, shpHeaderSize)
	_, err := r.Read(header)
	if err != nil {
		return core.Bbox{}, err
	}

	if len(header) < shpHeaderSize {
		return core.Bbox{}, fmt.Errorf("shapefile does not have valid header")
	}

	if headerFileCode := binary.BigEndian.Uint32(header[:4]); headerFileCode != shpFileCode {
		return core.Bbox{}, errors.New("invalid file code")
	}
	if headerVersion := binary.LittleEndian.Uint32(header[28:32]); headerVersion != shpHeaderVersion {
		return core.Bbox{}, errors.New("invalid header version")
	}

	// TODO we're reading the bounds from the header -- which isn't guaranteed to reflect
	// the bounds of the geometries in the file -- read the geometries to confirm
	minX := math.Float64frombits(binary.LittleEndian.Uint64(header[36:44]))
	minY := math.Float64frombits(binary.LittleEndian.Uint64(header[44:52]))
	maxX := math.Float64frombits(binary.LittleEndian.Uint64(header[52:60]))
	maxY := math.Float64frombits(binary.LittleEndian.Uint64(header[60:68]))

	// check if any values represent no data
	if minX <= -1e38 {
		minX = math.Inf(1)
	}
	if minY <= -1e38 {
		minY = math.Inf(1)
	}
	if maxX <= -1e38 {
		maxX = math.Inf(-1)
	}
	if maxY <= -1e38 {
		maxY = math.Inf(-1)
	}

	return core.Bbox{
		Left:   minX,
		Bottom: minY,
		Right:  maxX,
		Top:    maxY,
		Srid:   core.UnknownCrs,
	}, nil
}

// Common projection name to EPSG code mappings
var projectionLookup = map[string]int{
	// Geographic coordinate systems
	"GCS_WGS_1984":           core.Wgs84,
	"WGS_1984":               core.Wgs84,
	"WGS84":                  core.Wgs84,
	"GCS_North_American_1983": core.Nad83,
	"NAD_1983":               core.Nad83,
	"NAD83":                  core.Nad83,
	
	// Web Mercator variants
	"WGS_1984_Web_Mercator":           3857,
	"WGS_1984_Web_Mercator_Auxiliary_Sphere": 3857,
	"Popular_Visualisation_CRS":       3857,
	
	// Common UTM zones for NAD83
	"NAD_1983_UTM_Zone_10N": 26910,
	"NAD_1983_UTM_Zone_11N": 26911,
	"NAD_1983_UTM_Zone_12N": 26912,
	"NAD_1983_UTM_Zone_13N": 26913,
	"NAD_1983_UTM_Zone_14N": 26914,
	"NAD_1983_UTM_Zone_15N": 26915,
	"NAD_1983_UTM_Zone_16N": 26916,
	"NAD_1983_UTM_Zone_17N": 26917,
	"NAD_1983_UTM_Zone_18N": 26918,
	"NAD_1983_UTM_Zone_19N": 26919,
	
	// Common UTM zones for WGS84
	"WGS_1984_UTM_Zone_10N": 32610,
	"WGS_1984_UTM_Zone_11N": 32611,
	"WGS_1984_UTM_Zone_12N": 32612,
	"WGS_1984_UTM_Zone_13N": 32613,
	"WGS_1984_UTM_Zone_14N": 32614,
	"WGS_1984_UTM_Zone_15N": 32615,
	"WGS_1984_UTM_Zone_16N": 32616,
	"WGS_1984_UTM_Zone_17N": 32617,
	"WGS_1984_UTM_Zone_18N": 32618,
	"WGS_1984_UTM_Zone_19N": 32619,
}

// parseProjectionWKT attempts to extract SRID from projection WKT
func parseProjectionWKT(prjContent string) int {
	// Clean up the content
	content := strings.TrimSpace(prjContent)
	if content == "" {
		return core.UnknownCrs
	}

	// First try to find AUTHORITY["EPSG","xxxx"] pattern
	authorityRegex := regexp.MustCompile(`AUTHORITY\["EPSG","(\d+)"\]`)
	matches := authorityRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		if srid, err := strconv.Atoi(matches[1]); err == nil {
			return srid
		}
	}

	// Fall back to projection name lookup
	// Extract projection name from PROJCS["name",...] or GEOGCS["name",...]
	projRegex := regexp.MustCompile(`(?:PROJCS|GEOGCS)\["([^"]+)"`)
	nameMatches := projRegex.FindStringSubmatch(content)
	if len(nameMatches) > 1 {
		projName := nameMatches[1]
		if srid, exists := projectionLookup[projName]; exists {
			return srid
		}
	}

	// Try to match UTM zone patterns for cases where the exact name isn't in our lookup
	utmRegex := regexp.MustCompile(`UTM_Zone_(\d+)([NS])`)
	utmMatches := utmRegex.FindStringSubmatch(content)
	if len(utmMatches) > 2 {
		zone, err := strconv.Atoi(utmMatches[1])
		if err == nil && zone >= 1 && zone <= 60 {
			hemisphere := utmMatches[2]
			
			// Determine if it's NAD83 or WGS84 based on content
			if strings.Contains(content, "NAD_1983") || strings.Contains(content, "North_American_1983") {
				if hemisphere == "N" {
					return 26900 + zone // NAD83 UTM North zones start at 26901
				}
				// NAD83 UTM South zones would be 269xx range, but less common
			} else if strings.Contains(content, "WGS_1984") || strings.Contains(content, "WGS84") {
				if hemisphere == "N" {
					return 32600 + zone // WGS84 UTM North zones start at 32601
				} else if hemisphere == "S" {
					return 32700 + zone // WGS84 UTM South zones start at 32701
				}
			}
		}
	}

	return core.UnknownCrs
}

// parseProjectionFile reads and parses a .prj file to extract SRID
func parseProjectionFile(prjPath string) int {
	content, err := os.ReadFile(prjPath)
	if err != nil {
		// File doesn't exist or can't be read, return unknown
		return core.UnknownCrs
	}
	
	return parseProjectionWKT(string(content))
}
