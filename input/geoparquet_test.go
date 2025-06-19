package input

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestSniffGeoparquet(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "Valid Parquet magic header",
			data: []byte("PAR1somedata"),
			want: true,
		},
		{
			name: "Exact PAR1 header",
			data: []byte("PAR1"),
			want: true,
		},
		{
			name: "Invalid header",
			data: []byte("INVALID"),
			want: false,
		},
		{
			name: "Too short data",
			data: []byte("PA"),
			want: false,
		},
		{
			name: "Empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "Different magic header",
			data: []byte("GIF89a"),
			want: false,
		},
		{
			name: "Case sensitive check",
			data: []byte("par1"),
			want: false,
		},
		{
			name: "PAR1 in middle of data",
			data: []byte("somePAR1data"),
			want: false,
		},
		{
			name: "Binary data with PAR1",
			data: append([]byte("PAR1"), []byte{0x00, 0x01, 0x02, 0x03}...),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SniffGeoparquet(tt.data)
			if got != tt.want {
				t.Errorf("SniffGeoparquet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadGeoparquetFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected core.Bbox
		wantErr  bool
	}{
		{
			name:     "Valid multipolygon file",
			filename: "../integration_tests/data/multipolygon-encoding_wkb.parquet",
			expected: core.Bbox{Left: 5, Bottom: 5, Right: 45, Top: 45},
			wantErr:  false,
		},
		// {
		// 	name:     "Valid point file",
		// 	filename: "../integration_tests/data/point-encoding_native.parquet",
		// 	expected: core.Bbox{Left: -180, Right: 180, Bottom: -90, Top: 90},
		// 	wantErr:  false,
		// },
		// TODO parquet file with geom col, but no metadata
		{
			name:     "file without geocolumn",
			filename: "../integration_tests/data/nongeo.parquet",
			expected: core.Bbox{},
			wantErr:  true,
		},
		{
			name:     "Non-existent file",
			filename: "nonexistent.parquet",
			expected: core.Bbox{},
			wantErr:  true,
		},
		{
			name:     "Empty filename",
			filename: "",
			expected: core.Bbox{},
			wantErr:  true,
		},
		{
			name:     "Directory instead of file",
			filename: ".",
			expected: core.Bbox{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box, err := LoadGeoparquetFile(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGeoparquetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !box.Equals(tt.expected) {
					t.Errorf("LoadGeoparquetFile() = %v, want %v", box, tt.expected)
				}
			}
		})
	}
}

func TestLoadGeoparquetFile_InvalidParquetFile(t *testing.T) {
	// Create a temporary file with invalid parquet content
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.parquet")

	// Write some non-parquet data
	err := os.WriteFile(invalidFile, []byte("this is not a parquet file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadGeoparquetFile(invalidFile)
	if err == nil {
		t.Error("LoadGeoparquetFile() expected error for invalid parquet file, got nil")
	}
}

func TestLoadGeoparquetFile_EmptyFile(t *testing.T) {
	// Create an empty temporary file
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.parquet")

	err := os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadGeoparquetFile(emptyFile)
	if err == nil {
		t.Error("LoadGeoparquetFile() expected error for empty file, got nil")
	}
}

// Additional comprehensive edge case tests
func TestSniffGeoparquet_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "PAR1 with null bytes",
			data: []byte{'P', 'A', 'R', '1', 0x00, 0x00, 0x00, 0x00},
			want: true,
		},
		{
			name: "PAR1 with high-bit characters",
			data: []byte{'P', 'A', 'R', '1', 0xFF, 0xFE, 0xFD, 0xFC},
			want: true,
		},
		{
			name: "Almost PAR1 - one character off",
			data: []byte("PAR2"),
			want: false,
		},
		{
			name: "Unicode that looks like PAR1",
			data: []byte("ＰＡＲ１"), // Full-width characters
			want: false,
		},
		{
			name: "Empty followed by PAR1",
			data: []byte(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SniffGeoparquet(tt.data)
			if got != tt.want {
				t.Errorf("SniffGeoparquet() = %v, want %v", got, tt.want)
			}
		})
	}
}
