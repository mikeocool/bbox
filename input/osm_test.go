package input

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeocool/bbox/core"
)

func TestSniffOSM(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "Valid OSM XML header",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?><osm version="0.6">`),
			want: true,
		},
		{
			name: "OSM XML with whitespace",
			data: []byte(`<?xml version="1.0"?>
<osm version="0.6" generator="Overpass API">`),
			want: true,
		},
		{
			name: "OSM XML minimal",
			data: []byte(`<?xml><osm>`),
			want: true,
		},
		{
			name: "Invalid XML without osm tag",
			data: []byte(`<?xml version="1.0"?><root>`),
			want: false,
		},
		{
			name: "OSM without XML declaration",
			data: []byte(`<osm version="0.6">`),
			want: false,
		},
		{
			name: "Too short data",
			data: []byte(`<?xml`),
			want: false,
		},
		{
			name: "Empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "Binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03},
			want: false,
		},
		{
			name: "Case sensitive XML",
			data: []byte(`<?XML version="1.0"?><OSM>`),
			want: false,
		},
		{
			name: "Long valid OSM header",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6" generator="Overpass API 0.7.59.8" note="The data included in this document is from www.openstreetmap.org. The data is made available under ODbL." copyright="OpenStreetMap and contributors">
<note>This is a test file</note>`),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SniffOSM(tt.data)
			if got != tt.want {
				t.Errorf("SniffOSM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadOSMFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected core.Bbox
		wantErr  bool
	}{
		{
			name:     "Non-existent file",
			filename: "nonexistent.osm",
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
			box, err := LoadOSMFile(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadOSMFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !box.Equals(tt.expected) {
					t.Errorf("LoadOSMFile() = %v, want %v", box, tt.expected)
				}
			}
		})
	}
}

func TestLoadOSMFile_ValidXML(t *testing.T) {
	// Create a temporary OSM XML file with known bounds
	tmpDir := t.TempDir()
	osmFile := filepath.Join(tmpDir, "test.osm")

	osmContent := `<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6" generator="test">
  <node id="1" lat="50.0" lon="8.0"/>
  <node id="2" lat="52.0" lon="10.0"/>
  <node id="3" lat="51.0" lon="9.0"/>
  <way id="1">
    <nd ref="1"/>
    <nd ref="2"/>
    <nd ref="3"/>
  </way>
</osm>`

	err := os.WriteFile(osmFile, []byte(osmContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	expected := core.Bbox{
		Left:   8.0,  // min longitude
		Bottom: 50.0, // min latitude
		Right:  10.0, // max longitude
		Top:    52.0, // max latitude
	}

	box, err := LoadOSMFile(osmFile)
	if err != nil {
		t.Errorf("LoadOSMFile() unexpected error: %v", err)
		return
	}

	if !box.Equals(expected) {
		t.Errorf("LoadOSMFile() = %v, want %v", box, expected)
	}
}

func TestLoadOSMFile_SingleNode(t *testing.T) {
	// Test with a single node
	tmpDir := t.TempDir()
	osmFile := filepath.Join(tmpDir, "single_node.osm")

	osmContent := `<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6">
  <node id="1" lat="45.5" lon="-73.5"/>
</osm>`

	err := os.WriteFile(osmFile, []byte(osmContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	expected := core.Bbox{
		Left:   -73.5,
		Bottom: 45.5,
		Right:  -73.5,
		Top:    45.5,
	}

	box, err := LoadOSMFile(osmFile)
	if err != nil {
		t.Errorf("LoadOSMFile() unexpected error: %v", err)
		return
	}

	if !box.Equals(expected) {
		t.Errorf("LoadOSMFile() = %v, want %v", box, expected)
	}
}

func TestLoadOSMFile_NoNodes(t *testing.T) {
	// Test with OSM file containing no nodes
	tmpDir := t.TempDir()
	osmFile := filepath.Join(tmpDir, "no_nodes.osm")

	osmContent := `<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6">
  <relation id="1">
    <member type="way" ref="1" role="outer"/>
  </relation>
</osm>`

	err := os.WriteFile(osmFile, []byte(osmContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadOSMFile(osmFile)
	if err == nil {
		t.Error("LoadOSMFile() expected error for file with no nodes, got nil")
	}
}

func TestLoadOSMFile_InvalidXML(t *testing.T) {
	// Create a temporary file with invalid XML content
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.osm")

	// Write some invalid XML data
	err := os.WriteFile(invalidFile, []byte("this is not valid XML"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadOSMFile(invalidFile)
	if err == nil {
		t.Error("LoadOSMFile() expected error for invalid XML file, got nil")
	}
}

func TestLoadOSMFile_EmptyFile(t *testing.T) {
	// Create an empty temporary file
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.osm")

	err := os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadOSMFile(emptyFile)
	if err == nil {
		t.Error("LoadOSMFile() expected error for empty file, got nil")
	}
}

func TestLoadOSMFile_WorldwideCoordinates(t *testing.T) {
	// Test with coordinates spanning the globe
	tmpDir := t.TempDir()
	osmFile := filepath.Join(tmpDir, "worldwide.osm")

	osmContent := `<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6">
  <node id="1" lat="-90.0" lon="-180.0"/>
  <node id="2" lat="90.0" lon="180.0"/>
  <node id="3" lat="0.0" lon="0.0"/>
</osm>`

	err := os.WriteFile(osmFile, []byte(osmContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	expected := core.Bbox{
		Left:   -180.0,
		Bottom: -90.0,
		Right:  180.0,
		Top:    90.0,
	}

	box, err := LoadOSMFile(osmFile)
	if err != nil {
		t.Errorf("LoadOSMFile() unexpected error: %v", err)
		return
	}

	if !box.Equals(expected) {
		t.Errorf("LoadOSMFile() = %v, want %v", box, expected)
	}
}

func TestLoadOSMFile_PrecisionTest(t *testing.T) {
	// Test with high precision coordinates
	tmpDir := t.TempDir()
	osmFile := filepath.Join(tmpDir, "precision.osm")

	osmContent := `<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6">
  <node id="1" lat="50.123456789" lon="8.987654321"/>
  <node id="2" lat="50.123456790" lon="8.987654322"/>
</osm>`

	err := os.WriteFile(osmFile, []byte(osmContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	expected := core.Bbox{
		Left:   8.987654321,
		Bottom: 50.123456789,
		Right:  8.987654322,
		Top:    50.123456790,
	}

	box, err := LoadOSMFile(osmFile)
	if err != nil {
		t.Errorf("LoadOSMFile() unexpected error: %v", err)
		return
	}

	if !box.Equals(expected) {
		t.Errorf("LoadOSMFile() = %v, want %v", box, expected)
	}
}

func TestSniffOSM_RealPBFFile(t *testing.T) {
	// Test with real Monaco PBF file
	pbfFile := "../integration_tests/data/monaco-latest.osm.pbf"

	// Check if file exists
	if _, err := os.Stat(pbfFile); os.IsNotExist(err) {
		t.Skip("Monaco PBF file not found, skipping test")
		return
	}

	// Read first few bytes to test sniffing
	file, err := os.Open(pbfFile)
	if err != nil {
		t.Fatalf("Failed to open PBF file: %v", err)
	}
	defer file.Close()

	header := make([]byte, 100)
	n, err := file.Read(header)
	if err != nil {
		t.Fatalf("Failed to read PBF file header: %v", err)
	}

	// Note: SniffOSM() is designed for XML detection, so it should return false for PBF
	if SniffOSM(header[:n]) {
		t.Error("SniffOSM() should return false for PBF files (it's XML-specific)")
	}
}

func TestLoadOSMFile_RealPBFFile(t *testing.T) {
	// Test with real Monaco PBF file
	pbfFile := "../integration_tests/data/monaco-latest.osm.pbf"

	// Check if file exists
	if _, err := os.Stat(pbfFile); os.IsNotExist(err) {
		t.Skip("Monaco PBF file not found, skipping test")
		return
	}

	box, err := LoadOSMFile(pbfFile)
	if err != nil {
		t.Errorf("LoadOSMFile() failed on real PBF file: %v", err)
		return
	}

	// Monaco OSM extract from Geofabrik includes surrounding areas
	// Based on actual data: Longitude ~7.20-7.62, Latitude ~37.27-43.76
	// This is larger than Monaco proper due to the way OSM extracts work

	// Verify that we got reasonable bounds for the Monaco region
	if box.Left < 7.0 || box.Left > 8.0 {
		t.Errorf("Monaco region Left boundary seems wrong: %f (expected ~7.20)", box.Left)
	}
	if box.Right < 7.0 || box.Right > 8.0 {
		t.Errorf("Monaco region Right boundary seems wrong: %f (expected ~7.62)", box.Right)
	}
	if box.Bottom < 35.0 || box.Bottom > 45.0 {
		t.Errorf("Monaco region Bottom boundary seems wrong: %f (expected ~37.27)", box.Bottom)
	}
	if box.Top < 35.0 || box.Top > 45.0 {
		t.Errorf("Monaco region Top boundary seems wrong: %f (expected ~43.76)", box.Top)
	}

	// Ensure the bounding box makes sense (right > left, top > bottom)
	if box.Right <= box.Left {
		t.Errorf("Invalid bounding box: Right (%f) should be > Left (%f)", box.Right, box.Left)
	}
	if box.Top <= box.Bottom {
		t.Errorf("Invalid bounding box: Top (%f) should be > Bottom (%f)", box.Top, box.Bottom)
	}

	t.Logf("Monaco bounding box: %+v", box)
}

func TestSniffOSM_RealXMLFile(t *testing.T) {
	// Test with real Munich OSM XML file
	xmlFile := "../integration_tests/data/map.osm"

	// Read first few bytes to test sniffing
	file, err := os.Open(xmlFile)
	if err != nil {
		t.Fatalf("Failed to open OSM XML file: %v", err)
	}
	defer file.Close()

	header := make([]byte, 200)
	n, err := file.Read(header)
	if err != nil {
		t.Fatalf("Failed to read OSM XML file header: %v", err)
	}

	// Test content-based sniffing
	if !SniffOSM(header[:n]) {
		t.Error("SniffOSM() should detect OSM XML files")
	}

}

func TestLoadOSMFile_RealXMLFile(t *testing.T) {
	// Test with real Munich OSM XML file
	xmlFile := "../integration_tests/data/map.osm"

	box, err := LoadOSMFile(xmlFile)
	if err != nil {
		t.Errorf("LoadOSMFile() failed on real OSM XML file: %v", err)
		return
	}

	expected := core.Bbox{Left: -93.3121726, Bottom: 46.9726411, Right: -92.5814836, Top: 47.263262}
	if !box.Equals(expected) {
		t.Errorf("Box: %v does not match expected: %v", box, expected)
	}
}
