package test_utils

import (
	"math"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func BboxAlmostEqual(a, b core.Bbox) bool {
	const epsilon = 1e-5
	return math.Abs(a.Left-b.Left) < epsilon &&
		math.Abs(a.Bottom-b.Bottom) < epsilon &&
		math.Abs(a.Right-b.Right) < epsilon &&
		math.Abs(a.Top-b.Top) < epsilon
}

func AssertBboxEqual(t *testing.T, expected, actual core.Bbox) {
	if !BboxAlmostEqual(expected, actual) {
		t.Errorf("Expected bbox %v but got %v", expected, actual)
	}
}
