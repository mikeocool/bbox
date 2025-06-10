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

// Slice divides the bounding box into a grid of columns and rows,
// returning a slice of bounding boxes for each cell in the grid.
// The boxes are returned in row-major order (left to right, top to bottom).
func (b Bbox) Slice(columns, rows int) []Bbox {
	if columns <= 0 || rows <= 0 {
		return []Bbox{}
	}

	totalWidth := b.Right - b.Left
	totalHeight := b.Top - b.Bottom
	cellWidth := totalWidth / float64(columns)
	cellHeight := totalHeight / float64(rows)

	boxes := make([]Bbox, 0, columns*rows)

	for row := range rows {
		for col := range columns {
			left := b.Left + float64(col)*cellWidth
			right := b.Left + float64(col+1)*cellWidth
			bottom := b.Bottom + float64(rows-row-1)*cellHeight
			top := b.Bottom + float64(rows-row)*cellHeight

			boxes = append(boxes, Bbox{
				Left:   left,
				Bottom: bottom,
				Right:  right,
				Top:    top,
			})
		}
	}

	return boxes
}
