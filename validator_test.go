package struct_validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Nested struct {
		User     User `validate:"nested"`
		Intfield int  `validate:"in:200,404"`
		App      App  `validate:"nested"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "7f0e3265-ca96-4b33-8858-fef9696cc71b",
				Name:   "Name",
				Age:    30,
				Email:  "somemail@gmail.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
				meta:   []byte{12},
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "7f0e3265-ca96-4b33-8858-fef9",
				Name:   "Name",
				Age:    1,
				Email:  "somemailgmail.com",
				Role:   "role",
				Phones: []string{"12345678901234567890"},
				meta:   []byte{12},
			},
			expectedErr: ValidationErrors{
				ValidationError{Field: "ID", Err: ErrValidationStrLen},
				ValidationError{Field: "Age", Err: ErrValidationIntMin},
				ValidationError{Field: "Email", Err: ErrValidationStrRegexp},
				ValidationError{Field: "Role", Err: ErrValidationStrIn},
				ValidationError{Field: "Phones", Err: ErrValidationStrLen},
			},
		},
		{
			in: App{
				Version: "debug",
			},
			expectedErr: nil,
		},
		{
			in: App{
				Version: "release",
			},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Version", Err: ErrValidationStrLen},
			},
		},
		{
			in: Token{
				Header:    []byte{1, 2, 3},
				Payload:   []byte{1, 2, 3},
				Signature: []byte{1, 2, 3},
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 213,
				Body: "body",
			},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Code", Err: ErrValidationIntIn},
			},
		},
		{
			in: Nested{
				User: User{
					ID:     "7f0e3265-ca96-4b33-8858-fef9696cc71b",
					Name:   "Name",
					Age:    30,
					Email:  "error",
					Role:   "admin",
					Phones: []string{"12345678901"},
					meta:   []byte{12},
				},
				Intfield: 100,
				App: App{
					Version: "release",
				},
			},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Email", Err: ErrValidationStrRegexp},
				ValidationError{Field: "Intfield", Err: ErrValidationIntIn},
				ValidationError{Field: "Version", Err: ErrValidationStrLen},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			require.Equal(t, tt.expectedErr, Validate(tt.in))
		})
	}
}

func TestValidateField(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		err   error
	}{
		{
			name: "string",
			field: Field{
				name:  "StrTag",
				value: reflect.ValueOf("foo"),
				tag:   "in:foo|regexp:foo|len:3",
				kind:  reflect.String,
			},
			err: nil,
		},
		{
			name: "string slice",
			field: Field{
				name:  "StrTag",
				value: reflect.ValueOf([]string{"foo", "bar"}),
				tag:   "in:foo|regexp:foo|len:3",
				kind:  reflect.Slice,
			},
			err: ValidationErrors{
				ValidationError{Field: "StrTag", Err: ErrValidationStrIn},
				ValidationError{Field: "StrTag", Err: ErrValidationStrRegexp},
			},
		},
		{
			name: "string, validation error",
			field: Field{
				name:  "StrTag",
				value: reflect.ValueOf("bar"),
				tag:   "in:foo|regexp:foo|len:6",
				kind:  reflect.String,
			},
			err: ValidationErrors{
				ValidationError{Field: "StrTag", Err: ErrValidationStrIn},
				ValidationError{Field: "StrTag", Err: ErrValidationStrRegexp},
				ValidationError{Field: "StrTag", Err: ErrValidationStrLen},
			},
		},
		{
			name: "int",
			field: Field{
				name:  "IntTag",
				value: reflect.ValueOf(25),
				tag:   "in:20,25,30|max:30|min:20",
				kind:  reflect.Int,
			},
			err: nil,
		},
		{
			name: "int slice",
			field: Field{
				name:  "IntTag",
				value: reflect.ValueOf([]int{10, 20}),
				tag:   "in:5,15|max:16|min:4",
				kind:  reflect.Slice,
			},
			err: ValidationErrors{
				ValidationError{Field: "IntTag", Err: ErrValidationIntIn},
				ValidationError{Field: "IntTag", Err: ErrValidationIntIn},
				ValidationError{Field: "IntTag", Err: ErrValidationIntMax},
			},
		},
		{
			name: "int, validation error",
			field: Field{
				name:  "IntTag",
				value: reflect.ValueOf(50),
				tag:   "in:20,30|max:30|min:60",
				kind:  reflect.Int,
			},
			err: ValidationErrors{
				ValidationError{Field: "IntTag", Err: ErrValidationIntIn},
				ValidationError{Field: "IntTag", Err: ErrValidationIntMax},
				ValidationError{Field: "IntTag", Err: ErrValidationIntMin},
			},
		},
		{
			name: "unsupported type",
			field: Field{
				name:  "BoolTag",
				value: reflect.ValueOf(true),
				tag:   "validate:tag",
				kind:  reflect.Bool,
			},
			err: ErrUnsupType,
		},
	}

	var verr ValidationErrors
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateField(tc.field)
			if errors.As(err, &verr) {
				require.Equal(t, tc.err, verr)
				return
			}
			require.Equal(t, tc.err, err)
		})
	}
}

func TestParseFieldByTag(t *testing.T) {
	testStruct := struct {
		StrTag     string `validate:"in:foo|regexp:foo|len:3"`
		IntTag     int    `validate:"in:10,20|max:20|min:30"`
		unexpField int    `validate:"in:10,20|max:20|min:30"`
		JSONTag    string `json:"id"`
		NoTag      string
	}{"string", 0, 0, "json", "notag"}

	tests := []struct {
		name  string
		st    interface{}
		tag   string
		field []Field
		err   error
	}{
		{
			name: "common case",
			st:   testStruct,
			tag:  validationTag,
			field: []Field{
				{"StrTag", reflect.ValueOf(testStruct.StrTag), "in:foo|regexp:foo|len:3", reflect.String},
				{"IntTag", reflect.ValueOf(testStruct.IntTag), "in:10,20|max:20|min:30", reflect.Int},
			},
			err: nil,
		},
		{
			name:  "type error",
			st:    "string",
			tag:   validationTag,
			field: nil,
			err:   ErrType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fields, err := parseFieldByTag(tc.st, tc.tag)
			require.Equal(t, tc.err, err)
			for i, f := range fields {
				require.Equal(t, tc.field[i].name, f.name)
				require.Equal(t, tc.field[i].value.Interface(), f.value.Interface())
				require.Equal(t, tc.field[i].tag, f.tag)
				require.Equal(t, tc.field[i].kind, f.kind)
			}
		})
	}
}

func TestParseConditions(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		cond []Condition
		err  error
	}{
		{
			name: "length tag",
			tag:  "len:5",
			cond: []Condition{{"len", "5"}},
			err:  nil,
		},
		{
			name: "regexp tag",
			tag:  "regexp:regarg",
			cond: []Condition{{"regexp", "regarg"}},
			err:  nil,
		},
		{
			name: "in string tag",
			tag:  "in:first,second",
			cond: []Condition{{"in", "first,second"}},
			err:  nil,
		},
		{
			name: "min tag",
			tag:  "min:10",
			cond: []Condition{{"min", "10"}},
			err:  nil,
		},
		{
			name: "max tag",
			tag:  "max:20",
			cond: []Condition{{"max", "20"}},
			err:  nil,
		},
		{
			name: "in int tag",
			tag:  "in:15,16",
			cond: []Condition{{"in", "15,16"}},
			err:  nil,
		},
		{
			name: "multi string tag",
			tag:  "len:5|regexp:regarg|in:first",
			cond: []Condition{{"len", "5"}, {"regexp", "regarg"}, {"in", "first"}},
			err:  nil,
		},
		{
			name: "multi int tag",
			tag:  "min:5|max:20|in:15",
			cond: []Condition{{"min", "5"}, {"max", "20"}, {"in", "15"}},
			err:  nil,
		},
		{
			name: "without operand",
			tag:  "nested",
			cond: []Condition{{"nested", ""}},
			err:  nil,
		},
		{
			name: "validation format error",
			tag:  "validation:format:error",
			cond: nil,
			err:  ErrValidationFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := parseConditions(tc.tag)
			require.Equal(t, c, tc.cond)
			require.Equal(t, err, tc.err)
		})
	}
}
