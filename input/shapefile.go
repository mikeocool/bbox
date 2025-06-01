package input

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/mikeocool/bbox/core"
)

const (
	shpHeaderSize    = 100
	shpHeaderVersion = 1000
	shpFileCode      = 9994
)

func SniffShapefile(data []byte) bool {
	if len(data) < shpHeaderSize {
		return false
	}

	// Shapefile main file (.shp) has a specific header structure
	// File code should be 9994 (0x270A) in big-endian at bytes 0-3
	if len(data) >= 4 {
		fileCode := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		if fileCode == shpFileCode {
			return true
		}
	}

	return false
}

func LoadShapefile(filename string) (core.Bbox, error) {
	r, err := os.Open(filename)
	if err != nil {
		return core.Bbox{}, err
	}
	defer r.Close()
	return ParseShapefile(r)
}

func ParseShapefile(r io.Reader) (core.Bbox, error) {
	header := make([]byte, shpHeaderSize)
	_, err := r.Read(header)
	if err != nil {
		return core.Bbox{}, err
	}

	if len(header) < shpHeaderSize {
		return core.Bbox{}, fmt.Errorf("shapefile does not have valid header")
	}

	if headerFileCode := binary.BigEndian.Uint32(header[:4]); headerFileCode != shpFileCode {
		return core.Bbox{}, errors.New("invalid file code")
	}
	if headerVersion := binary.LittleEndian.Uint32(header[28:32]); headerVersion != shpHeaderVersion {
		return core.Bbox{}, errors.New("invalid header version")
	}

	// TODO we're reading the bounds from the header -- which isn't guaranteed to reflect
	// the bounds of the geometries in the file -- read the geometries to confirm
	minX := math.Float64frombits(binary.LittleEndian.Uint64(header[36:44]))
	minY := math.Float64frombits(binary.LittleEndian.Uint64(header[44:52]))
	maxX := math.Float64frombits(binary.LittleEndian.Uint64(header[52:60]))
	maxY := math.Float64frombits(binary.LittleEndian.Uint64(header[60:68]))

	// check if any values represent no data
	if minX <= -1e38 {
		minX = math.Inf(1)
	}
	if minY <= -1e38 {
		minY = math.Inf(1)
	}
	if maxX <= -1e38 {
		maxX = math.Inf(-1)
	}
	if maxY <= -1e38 {
		maxY = math.Inf(-1)
	}

	return core.Bbox{
		Left:   minX,
		Bottom: minY,
		Right:  maxX,
		Top:    maxY,
	}, nil
}
