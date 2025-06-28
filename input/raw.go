package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mikeocool/bbox/core"
)

// LoadFilePaths attempts to load all args as file paths and returns combined bbox
func LoadFilePaths(paths []string) (core.Bbox, bool, error) {
	// TOOD attempt to determine if the paths appear to be file paths
	// (whether or not that are paths to exiting files)
	// And return errors accordingly

	// Check if all args are valid file paths
	var cleanedPaths []string
	for _, arg := range paths {
		trimmed := strings.TrimSpace(arg)
		if trimmed == "" {
			continue
		}

		if _, err := os.Stat(trimmed); err != nil {
			return core.Bbox{}, false, nil
		}

		cleanedPaths = append(cleanedPaths, trimmed)
	}

	if len(cleanedPaths) == 0 {
		return core.Bbox{}, false, nil
	}

	// Load all files and combine their bboxes
	var combinedBbox core.Bbox = core.EmptyBbox()
	for _, filename := range cleanedPaths {
		bbox, err := LoadFile(filename)
		if err != nil {
			if errors.Is(err, ErrNoFeaturesFound) {
				continue
			}
			return core.Bbox{}, true, fmt.Errorf("error loading file %s: %w", filename, err)
		}

		combinedBbox = combinedBbox.Union(bbox)
	}

	if combinedBbox.Validate() == nil {
		return combinedBbox, true, nil
	}

	return core.Bbox{}, true, ErrNoFeaturesFound
}

func ParseRawArgs(args []string) (core.Bbox, error) {
	if len(args) == 0 {
		return core.Bbox{}, fmt.Errorf("no arguments provided")
	}

	fmt.Printf("ParseRawArgs: %v", args)

	// Try loading args as file paths first
	if bbox, isFiles, err := LoadFilePaths(args); isFiles {
		return bbox, err
	}

	// Join args into a single string for format detection
	joinedArgs := strings.Join(args, " ")
	inputBytes := []byte(joinedArgs)

	// Try hex WKB first (single arg only)
	inputStr := strings.TrimSpace(args[0])
	if SniffWkbHex([]byte(inputStr)) {
		bbox, err := ParseHexWKBToBbox(inputStr)
		if err == nil {
			return bbox, nil
		}
	}

	// Try WKT format
	if SniffWkt(inputBytes) {
		reader := strings.NewReader(joinedArgs)
		return ParseWkt(reader)
	}

	// Try GeoJSON format
	if SniffGeojson(inputBytes) {
		reader := strings.NewReader(joinedArgs)
		return ParseGeojson(reader)
	}

	// Fall back to simple coordinate format
	reader := strings.NewReader(joinedArgs)
	return ParseSimpleRaw(reader)
}

// Parse simple bbox formats -- 4 ints or 2 ints separated by
// space, tab, or comma
func ParseSimpleRaw(r io.Reader) (core.Bbox, error) {
	var rbbox *core.Bbox

	expectedLineVals := 0 // unset value
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// TODO try ParseGeojson, incase it's geojsonl

		lineVals, err := parseLine(line)
		if err != nil {
			return core.Bbox{}, err
		}

		// ensure # of vals remains consistent
		var lineBbox core.Bbox
		if expectedLineVals != 0 && len(lineVals) != expectedLineVals {
			return core.Bbox{}, fmt.Errorf("invalid input")
		}

		expectedLineVals = len(lineVals)
		if len(lineVals) == 4 {
			lineBbox = core.Bbox{
				Left:   lineVals[0],
				Bottom: lineVals[1],
				Right:  lineVals[2],
				Top:    lineVals[3],
			}
		} else if len(lineVals) == 2 {
			lineBbox = core.Bbox{
				Left:   lineVals[0],
				Bottom: lineVals[1],
				Right:  lineVals[0],
				Top:    lineVals[1],
			}
		} else {
			return core.Bbox{}, fmt.Errorf("invalid input")
		}

		if err := lineBbox.Validate(); err != nil {
			return core.Bbox{}, err
		}

		if rbbox == nil {
			rbbox = &lineBbox
		} else {
			updated_bbox := rbbox.Union(lineBbox)
			rbbox = &updated_bbox
		}

	}

	if rbbox == nil {
		return core.Bbox{}, ErrUnrecognizedDataFormat
	}

	return *rbbox, nil
}

func parseLine(line string) ([]float64, error) {
	parts := strings.FieldsFunc(line, func(c rune) bool {
		return c == ' ' || c == ',' || c == '\t'
	})

	// Filter out empty strings
	var floats []float64
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			val, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil, fmt.Errorf("could not parse value: %s", part)
			}
			floats = append(floats, val)
		}
	}

	return floats[:], nil
}
