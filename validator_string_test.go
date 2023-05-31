package hw09structvalidator

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateStringField(t *testing.T) {
	type in struct {
		tag   string
		value []string
		name  string
	}
	tests := []struct {
		name  string
		input in
		err   error
	}{
		{
			name:  "validate len",
			input: in{tag: "len:6", value: []string{"foobar"}, name: "len"},
			err:   nil,
		},
		{
			name:  "validate err len",
			input: in{tag: "len:2", value: []string{"foo"}, name: "len"},
			err: ValidationErrors{
				ValidationError{Field: "len", Err: ErrValidationStrLen},
			},
		},
		{
			name:  "validate regexp",
			input: in{tag: "regexp:^\\w+@\\w+.com$", value: []string{"kom@gmail.com"}, name: "regexp"},
			err:   nil,
		},
		{
			name:  "validate err regexp",
			input: in{tag: "regexp:regarg", value: []string{"^\\w+@\\w+.com$"}, name: "regexp"},
			err: ValidationErrors{
				ValidationError{Field: "regexp", Err: ErrValidationStrRegexp},
			},
		},
		{
			name:  "validate in",
			input: in{tag: "in:foo,bar", value: []string{"foo"}, name: "in"},
			err:   nil,
		},
		{
			name:  "validate err in",
			input: in{tag: "in:foo", value: []string{"bar"}, name: "in"},
			err: ValidationErrors{
				ValidationError{Field: "in", Err: ErrValidationStrIn},
			},
		},
		{
			name:  "validate multi conditions",
			input: in{tag: "in:foo|regexp:foo|len:3", value: []string{"foo"}, name: "multi"},
			err:   nil,
		},
		{
			name:  "validate err multi conditions",
			input: in{tag: "in:bar|regexp:bar|len:6", value: []string{"foo"}, name: "multi"},
			err: ValidationErrors{
				ValidationError{Field: "multi", Err: ErrValidationStrIn},
				ValidationError{Field: "multi", Err: ErrValidationStrRegexp},
				ValidationError{Field: "multi", Err: ErrValidationStrLen},
			},
		},
		{
			name:  "validate slice multi conditions",
			input: in{tag: "in:foo,bar|regexp:^\\w+$|len:3", value: []string{"foo", "bar"}, name: "multi"},
			err:   nil,
		},
		{
			name:  "validate err slice multi conditions",
			input: in{tag: "in:tmp|regexp:^\\d+$|len:6", value: []string{"foo", "bar"}, name: "multi"},
			err: ValidationErrors{
				ValidationError{Field: "multi", Err: ErrValidationStrIn},
				ValidationError{Field: "multi", Err: ErrValidationStrRegexp},
				ValidationError{Field: "multi", Err: ErrValidationStrLen},
				ValidationError{Field: "multi", Err: ErrValidationStrIn},
				ValidationError{Field: "multi", Err: ErrValidationStrRegexp},
				ValidationError{Field: "multi", Err: ErrValidationStrLen},
			},
		},
		{
			name:  "strconv err",
			input: in{tag: "len: ", value: []string{""}, name: "err"},
			err:   strconv.ErrSyntax,
		},
		{
			name:  "unsupported condition",
			input: in{tag: "unsup:cond", value: []string{"str"}, name: "err"},
			err:   ErrUnsupCondition,
		},
	}

	var verr ValidationErrors
	var serr *strconv.NumError
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStringField(tc.input.tag, tc.input.value, tc.input.name)
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

func TestValidateStringLen(t *testing.T) {
	tests := []struct {
		name        string
		testString  string
		expectedLen string
		err         error
	}{
		{name: "empty strings", testString: "", expectedLen: "", err: strconv.ErrSyntax},
		{name: "empty string, str len", testString: "", expectedLen: "asd", err: strconv.ErrSyntax},
		{name: "empty string", testString: "", expectedLen: "0", err: nil},
		{name: "common case", testString: "string len", expectedLen: "10", err: nil},
		{name: "validate error", testString: "string len", expectedLen: "9", err: ErrValidationStrLen},
	}

	var verr ValidationErrors
	var serr *strconv.NumError
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStringLen(tc.testString, tc.expectedLen, tc.name)
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

func TestValidateStringRegexp(t *testing.T) {
	tests := []struct {
		name       string
		testString string
		regexp     string
		err        error
	}{
		{name: "empty strings", testString: "", regexp: "", err: nil},
		{name: "empty string", testString: "", regexp: "^\\d+$", err: ErrValidationStrRegexp},
		{name: "empty regexp", testString: "foo", regexp: "", err: nil},
		{name: "common case", testString: "123456", regexp: "^\\d+$", err: nil},
		{name: "validate error", testString: "error case", regexp: "foo", err: ErrValidationStrRegexp},
	}

	var verr ValidationErrors
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStringRegexp(tc.testString, tc.regexp, tc.name)
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr[0].Err)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}

func TestValidateStringIn(t *testing.T) {
	tests := []struct {
		name       string
		testString string
		subString  string
		err        error
	}{
		{name: "empty strings", testString: "", subString: "", err: nil},
		{name: "empty substring", testString: "foo", subString: "", err: ErrValidationStrIn},
		{name: "empty string", testString: "", subString: "foo", err: ErrValidationStrIn},
		{name: "common case", testString: "foo", subString: "foo,bar", err: nil},
		{name: "common case", testString: "bar", subString: "foo,bar", err: nil},
		{name: "validate error", testString: "error case", subString: "а", err: ErrValidationStrIn},
	}

	var verr ValidationErrors
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStringIn(tc.testString, tc.subString, tc.name)
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr[0].Err)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}
