package output

import (
	"bytes"
	"html/template"
	"strings"
)

type OutputSettings struct {
	FormatType    string
	FormatDetails string
	GeojsonIndent int
	GeojsonType   string
}

// ParseFormat parses a format string into format type and details.
// Format strings can be in the form "type" or "type=details".
func ParseFormat(formatStr string) (string, string) {
	formatType := formatStr
	formatDetails := ""
	parts := strings.SplitN(formatStr, "=", 2)
	if len(parts) > 1 {
		formatType = parts[0]
		formatDetails = parts[1]
	}

	return formatType, formatDetails
}

// FormatWithTemplate formats a Bbox using a given template string.
// The template can reference any of the Bbox fields using {{.FieldName}} syntax.
func FormatWithTemplate(templateStr string, geom any) (string, error) {
	tmpl, err := template.New("bbox").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, geom); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Format type constants
const (
	FormatGoTpl      = "go-template"
	FormatComma      = "comma"
	FormatSpace      = "space"
	FormatTab        = "tab"
	FormatGeoJson    = "geojson"
	FormatWkt        = "wkt"
	FormatWkbhex     = "wkbhex"
	FormatDublinCore = "dcsv"
	FormatUrl        = "url"
)
