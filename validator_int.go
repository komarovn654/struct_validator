package hw09structvalidator

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrValidationIntMin = errors.New("validation error, value less than expected")
	ErrValidationIntMax = errors.New("validation error, value greater than expected")
	ErrValidationIntIn  = errors.New("validation error, value dosen't match a subset of int")
)

func validateIntField(tag string, value []int, fieldName string) error {
	var verr ValidationErrors
	var err error
	ve := []ValidationError{}

	cond, err := parseConditions(tag)
	if err != nil {
		return err
	}

	for _, v := range value {
		for _, c := range cond {
			switch c.operator {
			case "min":
				err = validateIntMin(v, c.operand, fieldName)
			case "max":
				err = validateIntMax(v, c.operand, fieldName)
			case "in":
				err = validateIntIn(v, c.operand, fieldName)
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
	}
	if len(ve) > 0 {
		return ValidationErrors(ve)
	}
	return err
}

func validateIntMin(field int, min string, name string) error {
	m, err := strconv.Atoi(min)
	if err != nil {
		return err
	}
	if field < m {
		return ValidationErrors{ValidationError{Field: name, Err: ErrValidationIntMin}}
	}
	return nil
}

func validateIntMax(field int, max string, name string) error {
	m, err := strconv.Atoi(max)
	if err != nil {
		return err
	}
	if field > m {
		return ValidationErrors{ValidationError{Field: name, Err: ErrValidationIntMax}}
	}
	return nil
}

func validateIntIn(field int, in string, name string) error {
	for _, is := range strings.Split(in, inSplitSymbol) {
		if _, err := strconv.Atoi(is); err != nil {
			return err
		}

		if strconv.Itoa(field) == is {
			return nil
		}
	}

	return ValidationErrors{ValidationError{Field: name, Err: ErrValidationIntIn}}
}
