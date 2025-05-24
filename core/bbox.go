package core

import (
	"fmt"
)

type Bbox struct {
	Left   float64 `json:"left"`
	Bottom float64 `json:"bottom"`
	Right  float64 `json:"right"`
	Top    float64 `json:"top"`
}

// Validate checks if the Bbox has valid coordinates.
// A valid bounding box requires Right > Left and Top > Bottom.
func (b Bbox) Validate() error {
	if b.Right <= b.Left {
		return fmt.Errorf("invalid bbox: Right (%f) must be greater than Left (%f)", b.Right, b.Left)
	}
	if b.Top <= b.Bottom {
		return fmt.Errorf("invalid bbox: Top (%f) must be greater than Bottom (%f)", b.Top, b.Bottom)
	}

	// TODO ensure the box is somewhere on the earth for the given projection
	return nil
}

// Polygon returns the corner points of the bounding box as a closed polygon.
// The points are returned in counter-clockwise order starting from the bottom-left corner,
// with the first point repeated at the end to close the polygon.
func (b Bbox) Polygon() [][2]float64 {
	return [][2]float64{
		{b.Left, b.Bottom},  // bottom-left
		{b.Right, b.Bottom}, // bottom-right
		{b.Right, b.Top},    // top-right
		{b.Left, b.Top},     // top-left
		{b.Left, b.Bottom},  // bottom-left (close the polygon)
	}
}
