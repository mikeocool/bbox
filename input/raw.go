package input

import (
	"bbox/core"
	"fmt"
	"strconv"
	"strings"
)

func ParseRaw(input string) (core.Bbox, error) {
	// Check if input matches 4 floats separated by spaces and/or commas
	parts := strings.FieldsFunc(input, func(c rune) bool {
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
	return core.Bbox{}, nil // TODO
}
