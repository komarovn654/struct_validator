package struct_validator

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateIntField(t *testing.T) {
	type in struct {
		tag   string
		value []int
		name  string
	}
	tests := []struct {
		name  string
		input in
		err   error
	}{
		{
			name:  "validate in",
			input: in{tag: "in:0,11", value: []int{11}, name: "in"},
			err:   nil,
		},
		{
			name:  "validate err in",
			input: in{tag: "in:0,11", value: []int{12}, name: "in"},
			err: ValidationErrors{
				ValidationError{Field: "in", Err: ErrValidationIntIn},
			},
		},
		{
			name:  "validate max",
			input: in{tag: "max:50", value: []int{40}, name: "max"},
			err:   nil,
		},
		{
			name:  "validate err max",
			input: in{tag: "max:50", value: []int{60}, name: "max"},
			err: ValidationErrors{
				ValidationError{Field: "max", Err: ErrValidationIntMax},
			},
		},
		{
			name:  "validate min",
			input: in{tag: "min:60", value: []int{70}, name: "min"},
			err:   nil,
		},
		{
			name:  "validate err min",
			input: in{tag: "min:60", value: []int{50}, name: "min"},
			err: ValidationErrors{
				ValidationError{Field: "min", Err: ErrValidationIntMin},
			},
		},
		{
			name:  "validate multi conditions",
			input: in{tag: "in:0,5,11|max:6|min:4", value: []int{5}, name: "multi"},
			err:   nil,
		},
		{
			name:  "validate err multi conditions",
			input: in{tag: "in:0,11|max:6|min:13", value: []int{12}, name: "multi"},
			err: ValidationErrors{
				ValidationError{Field: "multi", Err: ErrValidationIntIn},
				ValidationError{Field: "multi", Err: ErrValidationIntMax},
				ValidationError{Field: "multi", Err: ErrValidationIntMin},
			},
		},
		{
			name:  "validate slice multi conditions",
			input: in{tag: "in:0,5,6,11|max:7|min:4", value: []int{5, 6}, name: "multi"},
			err:   nil,
		},
		{
			name:  "validate slice err multi conditions",
			input: in{tag: "in:0,11|max:6|min:13", value: []int{12, 5}, name: "multi"},
			err: ValidationErrors{
				ValidationError{Field: "multi", Err: ErrValidationIntIn},
				ValidationError{Field: "multi", Err: ErrValidationIntMax},
				ValidationError{Field: "multi", Err: ErrValidationIntMin},
				ValidationError{Field: "multi", Err: ErrValidationIntIn},
				ValidationError{Field: "multi", Err: ErrValidationIntMin},
			},
		},
		{
			name:  "strconv err",
			input: in{tag: "in:a,b", value: []int{0}, name: "err"},
			err:   strconv.ErrSyntax,
		},
		{
			name:  "unsupported condition",
			input: in{tag: "unsup:cond", value: []int{5, 6}, name: "err"},
			err:   ErrUnsupCondition,
		},
	}

	var verr ValidationErrors
	var serr *strconv.NumError
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIntField(tc.input.tag, tc.input.value, tc.input.name)
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr)
				return
			}
			if errors.As(err, &serr) {
				require.Equal(t, tc.err, serr.Err)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}

func TestValidateIntIn(t *testing.T) {
	tests := []struct {
		name string
		v    int
		in   string
		err  error
	}{
		{name: "strconv err", v: 0, in: "", err: strconv.ErrSyntax},
		{name: "strconv err", v: 0, in: "foo,bar", err: strconv.ErrSyntax},
		{name: "in validate", v: 55, in: "50,60,55", err: nil},
		{name: "in validate err", v: 40, in: "50,60", err: ErrValidationIntIn},
	}

	var verr ValidationErrors
	var serr *strconv.NumError
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIntIn(tc.v, tc.in, tc.name)
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr[0].Err)
				return
			}
			if errors.As(err, &serr) {
				require.Equal(t, tc.err, serr.Err)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}

func TestValidateIntMax(t *testing.T) {
	tests := []struct {
		name string
		v    int
		max  string
		err  error
	}{
		{name: "strconv err", v: 0, max: "foo", err: strconv.ErrSyntax},
		{name: "equal", v: 50, max: "50", err: nil},
		{name: "smaller", v: 40, max: "50", err: nil},
		{name: "greater", v: 60, max: "50", err: ErrValidationIntMax},
	}

	var verr ValidationErrors
	var serr *strconv.NumError
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIntMax(tc.v, tc.max, tc.name)
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr[0].Err)
				return
			}
			if errors.As(err, &serr) {
				require.Equal(t, tc.err, serr.Err)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}

func TestValidateIntMin(t *testing.T) {
	tests := []struct {
		name string
		v    int
		min  string
		err  error
	}{
		{name: "strconv err", v: 0, min: "foo", err: strconv.ErrSyntax},
		{name: "equal", v: 2, min: "2", err: nil},
		{name: "smaller", v: 2, min: "10", err: ErrValidationIntMin},
		{name: "greater", v: 2, min: "1", err: nil},
	}

	var verr ValidationErrors
	var serr *strconv.NumError
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIntMin(tc.v, tc.min, tc.name)
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr[0].Err)
				return
			}
			if errors.As(err, &serr) {
				require.Equal(t, tc.err, serr.Err)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}
