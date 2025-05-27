package input

import (
	"os"

	"github.com/mikeocool/bbox/core"
)

func LoadFile(filename string) (core.Bbox, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return core.Bbox{}, err
	}
	return ParseGeojson(data)
}
