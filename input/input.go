package input

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/mikeocool/bbox/geocoding"
)

type InputParams struct {
	Left        *float64
	Bottom      *float64
	Right       *float64
	Top         *float64
	Center      []float64 // a pair of floats representing the center coordinates
	Width       string
	Height      string
	Raw         []byte
	File        []string
	Place       string
	GeocoderURL string
	Buffer      float64
}

func (params *InputParams) HasWidth() bool  { return params.Width != "" }
func (params *InputParams) HasHeight() bool { return params.Height != "" }

func (params *InputParams) HasAnyCoordinates() bool {
	return params.Left != nil || params.Bottom != nil || params.Right != nil || params.Top != nil
}

func (params *InputParams) GetBbox() (core.Bbox, error) {
	builders := []BboxBuilder{
		RawBuilder,
		PlaceBuilder,
		FileBuilder,
		CenterBuilder,
		BoundsBuilder,
	}

	for i := range builders {
		if builders[i].IsUsable(params) {
			return buildBbox(builders[i], params)
		}
	}

	// dont allow the buffer parameter if there is no GetBbox
	if params.Buffer != 0 {
		return core.Bbox{}, fmt.Errorf("Cannot specify buffer without a bounding box")
	}

	return core.Bbox{}, NoUsableBuilderError{}
}

func (p *InputParams) getSetFields() []string {
	t := reflect.TypeOf(*p)
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if !isFieldEmpty(p, fieldName) {
			fields = append(fields, fieldName)
		}
	}
	return fields
}

func buildBbox(builder BboxBuilder, params *InputParams) (core.Bbox, error) {
	usedFieldsSet := make(map[string]bool)
	for _, field := range builder.UsedFields {
		usedFieldsSet[field] = true
	}

	setFields := params.getSetFields()
	for _, field := range setFields {
		// buffer is a global used field TODO make this better
		if !usedFieldsSet[field] && field != "Buffer" {
			return core.Bbox{}, fmt.Errorf("Unexpected argument: %s with %s", field, builder.Name)
		}
	}

	if err := builder.ValidateParams(params); err != nil {
		return core.Bbox{}, err
	}
	bbox, err := builder.Build(params)
	if err != nil {
		return core.Bbox{}, err
	}

	if params.Buffer != 0 {
		bbox, err = bbox.Buffer(params.Buffer)
		if err != nil {
			return core.Bbox{}, err
		}
	}

	return bbox, nil
}

type InputValidationError struct {
	Field   string
	Message string
}

func (e InputValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type NoUsableBuilderError struct{}

func (e NoUsableBuilderError) Error() string {
	return "no usable builder for the provided parameters"
}

type BboxBuilder struct {
	Name           string
	IsUsable       func(*InputParams) bool
	ValidateParams func(*InputParams) error
	UsedFields     []string
	Build          func(*InputParams) (core.Bbox, error)
}

var RawBuilder = BboxBuilder{
	IsUsable: func(params *InputParams) bool {
		return params.Raw != nil
	},
	ValidateParams: func(params *InputParams) error {
		return nil
	},
	UsedFields: []string{"Raw"},
	Build: func(params *InputParams) (core.Bbox, error) {
		return ParseRaw(params.Raw)
	},
}

var FileBuilder = BboxBuilder{
	IsUsable: func(params *InputParams) bool {
		return len(params.File) > 0
	},
	ValidateParams: func(params *InputParams) error {
		// Filter out blank and whitespace values
		var validFiles []string
		for _, file := range params.File {
			trimmed := strings.TrimSpace(file)
			if trimmed != "" {
				validFiles = append(validFiles, trimmed)
			}
		}

		if len(validFiles) == 0 {
			return InputValidationError{Field: "File", Message: "no valid file paths provided"}
		}

		return nil
	},
	UsedFields: []string{"File"},
	Build: func(params *InputParams) (core.Bbox, error) {
		var bbox *core.Bbox

		for _, file := range params.File {
			if file == "" {
				continue
			}
			fbox, err := LoadFile(file)
			if err == ErrNoFeaturesFound {
				continue
			} else if err != nil {
				return core.Bbox{}, err
			}

			if bbox == nil {
				bbox = &fbox
			} else {
				updated_bbox := bbox.Union(fbox)
				bbox = &updated_bbox
			}
		}
		if bbox == nil {
			return core.Bbox{}, ErrNoFeaturesFound
		}
		return *bbox, nil
	},
}

var PlaceBuilder = BboxBuilder{
	Name: "place",
	IsUsable: func(params *InputParams) bool {
		return params.Place != ""
	},
	ValidateParams: func(params *InputParams) error {
		// Validate that geocoder-url contains %s placeholder
		if params.GeocoderURL != "" && !strings.Contains(params.GeocoderURL, "%s") {
			return InputValidationError{Field: "geocoder-url", Message: "must contain %s placeholder for place name"}
		}
		return nil
	},
	UsedFields: []string{"Place", "GeocoderURL", "Width", "Height"},
	Build: func(params *InputParams) (core.Bbox, error) {
		var result *geocoding.GeocodeResult
		var err error
		
		// Geocode the place
		if params.GeocoderURL != "" {
			result, err = geocoding.GeocodePlaceWithURL(params.GeocoderURL, params.Place)
		} else {
			result, err = geocoding.GeocodePlace(geocoding.GeocoderPhotonKamoot, params.Place)
		}
		if err != nil {
			return core.Bbox{}, err
		}

		log.Printf("Geocoder matched %s: %s\n", result.Type, result.FullName)

		// If width and height are specified, create bounds around the center
		if params.HasWidth() && params.HasHeight() {
			width, err := strconv.ParseFloat(params.Width, 64)
			if err != nil {
				return core.Bbox{}, fmt.Errorf("invalid width: %w", err)
			}

			height, err := strconv.ParseFloat(params.Height, 64)
			if err != nil {
				return core.Bbox{}, fmt.Errorf("invalid height: %w", err)
			}

			return core.Bbox{
				Left:   result.LocationX - width/2,
				Bottom: result.LocationY - height/2,
				Right:  result.LocationX + width/2,
				Top:    result.LocationY + height/2,
			}, nil
		}

		// Use extent if available
		if result.Extent != nil {
			return *result.Extent, nil
		}

		// No extent and no width/height
		return core.Bbox{}, fmt.Errorf("Geocoder did not return extent for '%s', please a width and height", params.Place)
	},
}

var CenterBuilder = BboxBuilder{
	Name: "center",
	IsUsable: func(params *InputParams) bool {
		return len(params.Center) > 0
	},
	ValidateParams: func(params *InputParams) error {
		if len(params.Center) != 2 {
			return InputValidationError{Field: "center", Message: "invalid center coordinates"}
		}
		if !params.HasWidth() {
			return InputValidationError{Field: "width", Message: "width required"}
		}
		if !params.HasHeight() {
			return InputValidationError{Field: "height", Message: "height required"}
		}
		return nil
	},
	UsedFields: []string{"Center", "Width", "Height"},
	Build: func(params *InputParams) (core.Bbox, error) {
		width, err := strconv.ParseFloat(params.Width, 64)
		if err != nil {
			return core.Bbox{}, err
		}

		height, err := strconv.ParseFloat(params.Height, 64)
		if err != nil {
			return core.Bbox{}, err
		}

		return core.Bbox{
			Left:   params.Center[0] - width/2,
			Bottom: params.Center[1] - height/2,
			Right:  params.Center[0] + width/2,
			Top:    params.Center[1] + height/2,
		}, nil // TODO
	},
}

var BoundsBuilder = BboxBuilder{
	Name: "bounds",
	IsUsable: func(params *InputParams) bool {
		return params.HasAnyCoordinates()
	},
	ValidateParams: func(params *InputParams) error {
		if err := validateBoundsPair(params.Left, params.Right, params.Width); err != nil {
			return err
		}
		if err := validateBoundsPair(params.Bottom, params.Top, params.Height); err != nil {
			return err
		}
		return nil
	},
	UsedFields: []string{"Left", "Bottom", "Right", "Top", "Width", "Height"},
	Build: func(params *InputParams) (core.Bbox, error) {
		left, right := getBoundsPair(params.Left, params.Right, params.Width)
		bottom, top := getBoundsPair(params.Bottom, params.Top, params.Height)

		return core.Bbox{
			Left:   left,
			Right:  right,
			Bottom: bottom,
			Top:    top,
		}, nil // TODO
	},
}

func validateBoundsPair(min, max *float64, length string) error {
	hasMin := min != nil
	hasMax := max != nil
	hasLength := length != ""

	if !hasMin && !hasMax {
		return InputValidationError{Field: "", Message: "Must specify two of: min, max, or length"} // TODO better error
	}

	if hasMin && !hasMax && !hasLength {
		return fmt.Errorf("min specified without max or length") //TOOD
	}
	if hasMax && !hasMin && !hasLength {
		return fmt.Errorf("max specified without min or legnth") // TOOD
	}
	if hasMin && hasMax && hasLength {
		return fmt.Errorf("must specify only two of: min, max, and length")
	}

	return nil
}

func getBoundsPair(min, max *float64, length string) (float64, float64) {
	if min != nil && max != nil {
		return *min, *max
	}

	lengthVal := 0.0
	if length != "" {
		// TODO handle errors
		lengthVal, _ = strconv.ParseFloat(length, 64)
	}

	if min != nil && length != "" {
		return *min, *min + lengthVal
	}
	if max != nil && length != "" {
		return *max - lengthVal, *max
	}
	return 0, 0 // TODO
}

func isFieldEmpty(p *InputParams, fieldName string) bool {
	v := reflect.ValueOf(*p)
	field := v.FieldByName(fieldName)

	// fields does exist on struct
	if !field.IsValid() {
		return true
	}

	switch field.Kind() {
	case reflect.Ptr:
		return field.IsNil()
	case reflect.String:
		return field.String() == ""
	case reflect.Slice:
		return field.Len() == 0
	default:
		// For other types, check if it's the zero value
		return field.IsZero()
	}
}

func IsInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
