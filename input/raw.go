package input

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mikeocool/bbox/core"
)

func ParseRaw(input []byte) (core.Bbox, error) {
	bbox, err := ParseGeojson(input)
	if err != nil {
		if !errors.Is(err, ErrCouldNotParseGeoJSON) {
			fmt.Println("Failed to parse GeoJSON")
			return core.Bbox{}, err
		}
		// Continue to try other parsing methods
	} else {
		return bbox, nil
	}

	// Check if input matches 4 floats separated by spaces and/or commas
	parts := strings.FieldsFunc(string(input), func(c rune) bool {
		return c == ' ' || c == ',' || c == '\t'
	})

	// Filter out empty strings
	var validParts []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			validParts = append(validParts, part)
		}
	}

	if len(validParts) == 4 {
		var floats [4]float64
		for i, part := range validParts {
			val, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return core.Bbox{}, fmt.Errorf("invalid float at position %d: %s", i+1, part)
			}
			floats[i] = val
		}

		return core.Bbox{
			Left:   floats[0],
			Bottom: floats[1],
			Right:  floats[2],
			Top:    floats[3],
		}, nil
	}

	return core.Bbox{}, fmt.Errorf("invalid input")
}
