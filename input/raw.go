package input

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mikeocool/bbox/core"
)

func ParseRawArgs(args []string) (core.Bbox, error) {
	// TODO attempt to validate if args are file paths
	// Call LoadFiles

	// TOOD accept WKT, GeoJSON, HexWKB
	// NOTE these should also be accepted by ParseData
	// -- maybe break them out into a function called here and ParseData
	// accept hex wkb on cli
	inputStr := strings.TrimSpace(args[0])
	if SniffWkbHex([]byte(inputStr)) {
		bbox, err := ParseHexWKBToBbox(inputStr)
		if err == nil {
			return bbox, nil
		}
	}

	// Join args into a single string and try parsing as simple format
	joinedArgs := strings.Join(args, " ")
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

