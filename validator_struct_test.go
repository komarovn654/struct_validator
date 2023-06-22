package struct_validator

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateStructField(t *testing.T) {
	tests := []struct {
		name  string
		tag   string
		input interface{}
		err   error
	}{
		{
			name: "validate len err",
			tag:  "nested",
			input: struct {
				Name string `validate:"len:11"`
			}{
				Name: "Validation Error",
			},
			err: ValidationErrors{
				ValidationError{Field: "Name", Err: ErrValidationStrLen},
			},
		},
		{
			name: "validate len err",
			tag:  "nested",
			input: struct {
				Name string `validate:"len:8"`
			}{
				Name: "No Error",
			},
			err: nil,
		},
		{
			name: "unsupported condition",
			tag:  "unsupported",
			input: struct {
				Name string `validate:"len:8"`
			}{
				Name: "err",
			},
			err: ErrUnsupCondition,
		},
	}

	var verr ValidationErrors
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStructField(tc.tag, reflect.ValueOf(tc.input))
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}
