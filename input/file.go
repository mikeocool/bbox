package input

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/twpayne/go-shapefile"
)

func LoadFile(filename string) (core.Bbox, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	// TODO support shapefiles from zip
	case ".shp":
		return LoadShapefile(filename)
	case ".geojson", ".json":
		return LoadGeojson(filename)
	default:
		return core.Bbox{}, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func LoadShapefile(filename string) (core.Bbox, error) {
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	opts := shapefile.ReadShapefileOptions{
		DBF: &shapefile.ReadDBFOptions{
			SkipBrokenFields: true,
		},
	}
	shp, err := shapefile.Read(filename, &opts)
	if err != nil {
		return core.Bbox{}, fmt.Errorf("failed to read shapefile: %w", err)
	}
	if shp.SHP == nil {
		return core.Bbox{}, fmt.Errorf("unexpected error reading shapefile")
	}

	// TOOD potentially actually look at geoms since this bounds could be wrong
	bounds := shp.SHP.Bounds
	// TODO if empty return error

	return core.Bbox{
		Left:   bounds.Min(0),
		Bottom: bounds.Min(1),
		Right:  bounds.Max(0),
		Top:    bounds.Max(1),
	}, nil
}
