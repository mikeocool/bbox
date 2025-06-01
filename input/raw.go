package input

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mikeocool/bbox/core"
)

func ParseRaw(input []byte) (core.Bbox, error) {
	// TODO integrate ParseData here

	// attempt to parse as a GeoJSON document
	bbox, err := ParseGeojson(bytes.NewReader(input))
	if err != nil {
		if !errors.Is(err, ErrCouldNotParseGeoJSON) {
			fmt.Println("Failed to parse GeoJSON")
			return core.Bbox{}, err
		}
		// Continue to try other parsing methods
	} else {
		return bbox, nil
	}

	var rbbox *core.Bbox

	expectedLineVals := 0 // unset value
	scanner := bufio.NewScanner(bytes.NewReader(input))
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

		// TODO ensure # of vals remains consistent
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

		if rbbox == nil {
			rbbox = &lineBbox
		} else {
			updated_bbox := rbbox.Union(lineBbox)
			rbbox = &updated_bbox
		}

	}

	if rbbox == nil {
		return core.Bbox{}, fmt.Errorf("invalid input")
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
