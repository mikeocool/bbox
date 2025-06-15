package core

// IsValidWgs84 checks if the bounding box coordinates are within valid WGS84 ranges.
// Valid WGS84 coordinates: longitude [-180, 180], latitude [-90, 90]
func IsValidWgs84(b Bbox) bool {
	// Check longitude bounds (Left and Right)
	if b.Left < -180 || b.Left > 180 {
		return false
	}
	if b.Right < -180 || b.Right > 180 {
		return false
	}

	// Check latitude bounds (Bottom and Top)
	if b.Bottom < -90 || b.Bottom > 90 {
		return false
	}
	if b.Top < -90 || b.Top > 90 {
		return false
	}

	return true
}
