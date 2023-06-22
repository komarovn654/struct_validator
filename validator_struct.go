package struct_validator

import (
	"errors"
	"reflect"
)

func validateStructField(tag string, value reflect.Value) error {
	var verr ValidationErrors
	var err error
	ve := ValidationErrors{}

	cond, err := parseConditions(tag)
	if err != nil {
		return err
	}

	for _, c := range cond {
		switch c.operator {
		case "nested":
			err = Validate(value.Interface())
		default:
			return ErrUnsupCondition
		}
		if err != nil {
			if !errors.As(err, &verr) {
				return err
			}
			ve = append(ve, verr...)
		}
	}
	if len(ve) != 0 {
		return ve
	}
	return nil
}
