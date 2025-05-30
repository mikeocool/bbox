package core

import (
	"fmt"
	"math"
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

// Bounds returns the bounding box as a list of the bounds
func (b Bbox) Bounds() []float64 {
	return []float64{
		b.Left,
		b.Bottom, // bottom-left
		b.Right,
		b.Top, // top-right
	}
}

// Center returns the center point of the bounding box.
func (b Bbox) Center() [2]float64 {
	return [2]float64{
		(b.Left + b.Right) / 2,
		(b.Bottom + b.Top) / 2,
	}
}

// IsZero returns true if the bounding box has zero coordinates.
func (b Bbox) IsZero() bool {
	return b.Left == 0 && b.Bottom == 0 && b.Right == 0 && b.Top == 0
}

func (b Bbox) Union(other Bbox) Bbox {
	return Bbox{
		Left:   math.Min(b.Left, other.Left),
		Bottom: math.Min(b.Bottom, other.Bottom),
		Right:  math.Max(b.Right, other.Right),
		Top:    math.Max(b.Top, other.Top),
	}
}
