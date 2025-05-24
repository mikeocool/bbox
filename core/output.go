package core

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplatedFormat formats a Bbox using a given template string.
// The template can reference any of the Bbox fields using {{.FieldName}} syntax.
// For example: "{{.MinX}} {{.MinY}} {{.MaxX}} {{.MaxY}}"
func TemplatedFormat(templateStr string, bbox Bbox) (string, error) {
	tmpl, err := template.New("bbox").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, bbox); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX,MinY,MaxX,MaxY".
func CommaFormat(bbox Bbox) (string, error) {
	return TemplatedFormat("{{.MinX}},{{.MinY}},{{.MaxX}},{{.MaxY}}", bbox)
}

// CommaFormat formats a Bbox as a comma-separated string of its coordinates.
// The returned string will be in the format "MinX MinY MaxX MaxY".
func SpaceFormat(bbox Bbox) (string, error) {
	return TemplatedFormat("{{.MinX}} {{.MinY}} {{.MaxX}} {{.MaxY}}", bbox)
}

// Format type constants
const (
	FormatComma = "comma"
	FormatSpace = "space"
)

// FormatFunctions maps format type constants to their corresponding format functions
var outputFormatters = map[string]func(Bbox) (string, error){
	FormatComma: CommaFormat,
	FormatSpace: SpaceFormat,
}

// GetFormatter returns the format function for the given format type.
func GetFormatter(formatType string) func(Bbox) (string, error) {
	return outputFormatters[formatType]
}

// Format formats a Bbox using the specified format type.
func Format(bbox Bbox, formatType string) (string, error) {
	formatter := GetFormatter(formatType)
	if formatter == nil {
		return "", fmt.Errorf("unknown output format: %s", formatType)
	}
	return formatter(bbox)
}
