package core

import (
	"fmt"
)


type Bbox struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}

// Validate checks if the Bbox has valid coordinates.
// A valid bounding box requires MaxX > MinX and MaxY > MinY.
func (b Bbox) Validate() error {
	if b.MaxX <= b.MinX {
		return fmt.Errorf("invalid bbox: MaxX (%f) must be greater than MinX (%f)", b.MaxX, b.MinX)
	}
	if b.MaxY <= b.MinY {
		return fmt.Errorf("invalid bbox: MaxY (%f) must be greater than MinY (%f)", b.MaxY, b.MinY)
	}

	// TODO ensure the box is somewhere on the earth for the given projection
	return nil
}
