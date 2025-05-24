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
