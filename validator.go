package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrValidationFormat = errors.New("invalid validation format, expected [operator[:operand]]")
	ErrType             = errors.New("invalid type, expected struct")
	ErrUnsupCondition   = errors.New("unsupported condition")
	ErrUnsupType        = errors.New("unsupported type")
)

const (
	splitSymbol   = ":"
	inSplitSymbol = ","
	andSymbol     = "|"
	validationTag = "validate"
)

type ValidationError struct {
	Field string
	Err   error
}

type Condition struct {
	operator string
	operand  string
}

type Field struct {
	name  string
	value reflect.Value
	tag   string
	kind  reflect.Kind
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var errors strings.Builder
	for _, err := range v {
		errors.WriteString(fmt.Sprintf("%v: %v", err.Field, err.Err.Error()))
	}
	return errors.String()
}

func parseConditions(tag string) ([]Condition, error) {
	cond := []Condition{}
	conditions := strings.Split(tag, andSymbol)

	for _, c := range conditions {
		splitedCond := strings.Split(c, splitSymbol)
		if len(splitedCond) > 2 {
			return nil, ErrValidationFormat
		}
		if len(splitedCond) == 2 {
			cond = append(cond, Condition{operator: splitedCond[0], operand: splitedCond[1]})
			continue
		}
		cond = append(cond, Condition{operator: splitedCond[0], operand: ""})
	}
	return cond, nil
}

func parseFieldByTag(strct interface{}, expectedTag string) ([]Field, error) {
	st := reflect.TypeOf(strct)
	fields := []Field{}
	if st.Kind() != reflect.Struct {
		return nil, ErrType
	}

	for i := 0; i < st.NumField(); i++ {
		if !reflect.ValueOf(strct).Field(i).CanInterface() {
			continue //  Unexported field
		}
		tag, ok := st.Field(i).Tag.Lookup(expectedTag)
		if ok {
			fields = append(fields, Field{
				name:  st.Field(i).Name,
				value: reflect.ValueOf(strct).Field(i),
				tag:   tag,
				kind:  st.Field(i).Type.Kind(),
			})
		}
	}
	return fields, nil
}

func validateField(field Field) error {
	switch field.kind { //nolint:exhaustive
	case reflect.Struct:
		return validateStructField(field.tag, field.value)
	case reflect.String:
		return validateStringField(field.tag, []string{field.value.String()}, field.name)
	case reflect.Int:
		return validateIntField(field.tag, []int{int(field.value.Int())}, field.name)
	case reflect.Slice:
		switch field.value.Index(0).Kind() { //nolint:exhaustive
		case reflect.String:
			v := make([]string, field.value.Len())
			for i := 0; i < len(v); i++ {
				v[i] = field.value.Index(i).String()
			}
			return validateStringField(field.tag, v, field.name)
		case reflect.Int:
			v := make([]int, field.value.Len())
			for i := 0; i < len(v); i++ {
				v[i] = int(field.value.Index(i).Int())
			}
			return validateIntField(field.tag, v, field.name)
		default:
		}
	default:
	}
	return ErrUnsupType
}

func Validate(v interface{}) error {
	var verr ValidationErrors
	ve := ValidationErrors{}

	fields, err := parseFieldByTag(v, validationTag)
	if err != nil {
		return err
	}

	for _, f := range fields {
		err = validateField(f)
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
