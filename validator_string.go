package hw09structvalidator

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrValidationStrLen    = errors.New("validation error, string's length is not as expected")
	ErrValidationStrRegexp = errors.New("validation error, string doesn't match the regexp")
	ErrValidationStrIn     = errors.New("validation error, string doesn't match the substring")
)

func validateStringField(tag string, value []string, fieldName string) error {
	var verr ValidationErrors
	var err error
	ve := ValidationErrors{}

	cond, err := parseConditions(tag)
	if err != nil {
		return err
	}

	for _, v := range value {
		for _, c := range cond {
			switch c.operator {
			case "len":
				err = validateStringLen(v, c.operand, fieldName)
			case "regexp":
				err = validateStringRegexp(v, c.operand, fieldName)
			case "in":
				err = validateStringIn(v, c.operand, fieldName)
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
		return ve
	}
	return err
}

func validateStringLen(field string, elen string, name string) error {
	expLen, err := strconv.Atoi(elen)
	if err != nil {
		return err
	}
	if len(field) != expLen {
		return ValidationErrors{ValidationError{Field: name, Err: ErrValidationStrLen}}
	}
	return nil
}

func validateStringRegexp(field string, regxp string, name string) error {
	re, err := regexp.Compile(regxp)
	if err != nil {
		return err
	}
	if !re.MatchString(field) {
		return ValidationErrors{ValidationError{Field: name, Err: ErrValidationStrRegexp}}
	}
	return nil
}

func validateStringIn(field string, in string, name string) error {
	if field == "" && in == "" {
		return nil
	}
	for _, is := range strings.Split(in, inSplitSymbol) {
		if field == is {
			return nil
		}
	}
	return ValidationErrors{ValidationError{Field: name, Err: ErrValidationStrIn}}
}
