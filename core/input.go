package core

import (
	"fmt"
	"reflect"
	"strconv"
)

type InputParams struct {
	MinX   *float64
	MinY   *float64
	MaxX   *float64
	MaxY   *float64
	Center []float64 // a pair of floats representing the center coordinates
	Width  string
	Height string
	Raw    string
	Place  string
}

func (params *InputParams) HasWidth() bool  { return params.Width != "" }
func (params *InputParams) HasHeight() bool { return params.Height != "" }

func (params *InputParams) HasAnyCoordinates() bool {
	return params.MinX != nil || params.MinY != nil || params.MaxX != nil || params.MaxY != nil
}

func (params *InputParams) GetBbox() (Bbox, error) {
	builders := []BboxBuilder{
		RawBuilder,
		PlaceBuilder,
		CenterBuilder,
		BoundsBuilder,
	}

	for i := range builders {
		if builders[i].IsUsable(params) {
			return buildBbox(builders[i], params)
		}
	}
	return Bbox{}, fmt.Errorf("no usable builder") // TODO show usage
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

func buildBbox(builder BboxBuilder, params *InputParams) (Bbox, error) {
	usedFieldsSet := make(map[string]bool)
	for _, field := range builder.UsedFields {
		usedFieldsSet[field] = true
	}

	setFields := params.getSetFields()
	for _, field := range setFields {
		if !usedFieldsSet[field] {
			return Bbox{}, fmt.Errorf("Unexpected argument: %s with %s", field, builder.Name)
		}
	}

	if err := builder.ValidateParams(params); err != nil {
		return Bbox{}, err
	}
	return builder.Build(params)
}

type InputValidationError struct {
	Field   string
	Message string
}

func (e InputValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type BboxBuilder struct {
	Name           string
	IsUsable       func(*InputParams) bool
	ValidateParams func(*InputParams) error
	UsedFields     []string
	Build          func(*InputParams) (Bbox, error)
}

var RawBuilder = BboxBuilder{
	IsUsable: func(params *InputParams) bool {
		return params.Raw != ""
	},
	ValidateParams: func(params *InputParams) error {
		return nil
	},
	UsedFields: []string{"Raw"},
	Build: func(params *InputParams) (Bbox, error) {
		return Bbox{}, nil // TODO
	},
}

var PlaceBuilder = BboxBuilder{
	Name: "place",
	IsUsable: func(params *InputParams) bool {
		return params.Place != ""
	},
	ValidateParams: func(params *InputParams) error {
		if !params.HasWidth() {
			return InputValidationError{Field: "width", Message: "width required"}
		}
		if !params.HasHeight() {
			return InputValidationError{Field: "height", Message: "height required"}
		}
		return nil
	},
	UsedFields: []string{"Place", "Width", "Height"},
	Build: func(params *InputParams) (Bbox, error) {
		return Bbox{}, nil // TODO
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
	Build: func(params *InputParams) (Bbox, error) {
		width, err := strconv.ParseFloat(params.Width, 64)
		if err != nil {
			return Bbox{}, err
		}

		height, err := strconv.ParseFloat(params.Height, 64)
		if err != nil {
			return Bbox{}, err
		}

		return Bbox{
			MinX: params.Center[0] - width/2,
			MinY: params.Center[1] - height/2,
			MaxX: params.Center[0] + width/2,
			MaxY: params.Center[1] + height/2,
		}, nil // TODO
	},
}

var BoundsBuilder = BboxBuilder{
	Name: "bounds",
	IsUsable: func(params *InputParams) bool {
		return params.HasAnyCoordinates()
	},
	ValidateParams: func(params *InputParams) error {
		if err := validateBoundsPair(params.MinX, params.MaxX, params.Width); err != nil {
			return err
		}
		if err := validateBoundsPair(params.MinY, params.MaxY, params.Height); err != nil {
			return err
		}
		return nil
	},
	UsedFields: []string{"MinX", "MinY", "MaxX", "MaxY", "Width", "Height"},
	Build: func(params *InputParams) (Bbox, error) {
		minX, maxX := getBoundsPair(params.MinX, params.MaxX, params.Width)
		minY, maxY := getBoundsPair(params.MinY, params.MaxY, params.Height)

		return Bbox{
			MinX: minX,
			MaxX: maxX,
			MinY: minY,
			MaxY: maxY,
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
