package goform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

const tag = "form"

var (
	stringer  reflect.Type = reflect.TypeOf((FromStringer)(nil))
	validator reflect.Type = reflect.TypeOf((Validator)(nil))
)

type FormValuer interface {
	FormValue(string) string
}

type FromStringer interface {
	FromString(string) (interface{}, error)
}

type Validator interface {
	Validate(interface{}) error
}

type ValidationError []error

func (v ValidationError) MarshalJSON() ([]byte, error) {
	errmap := map[string][]error{
		"validation-errors": v,
	}

	return json.Marshal(errmap)
}

func (v ValidationError) Error() string {
	buf := &bytes.Buffer{}
	for i, err := range v {
		fmt.Fprintf(buf, "[%d] %s\n", i, err)
	}
	return buf.String()
}

type ImplementationError string

func (m ImplementationError) Error() string {
	return fmt.Sprintf("Does Not Implement %s", m)
}

func Unmarshal(f FormValuer, v interface{}) error {
	st := reflect.TypeOf(v)
	if st.Kind() != reflect.Ptr || st.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("Unmarshal expects a pointer to a struct")
	}

	st = st.Elem()

	errs := ValidationError{}
	target := reflect.ValueOf(v).Elem()
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)

		tagv := field.Tag.Get(tag)
		// ignore fields tagged with "-" and unexported fields
		if tagv == "-" || field.PkgPath != "" {
			continue
		}

		if tagv == "" {
			tagv = field.Name
		}

		val := f.FormValue(tagv)

		parse := parse
		if field.Type.Implements(stringer) {
			parse = callFromString
		}

		parsed, err := parse(field.Type, val)
		if err != nil {
			return err
		}

		if field.Type.Implements(validator) {
			err := validate(field.Type, parsed)
			if _, ok := err.(ImplementationError); ok {
				return err
			}

			errs = append(errs, err)
		}

		target.Field(i).Set(reflect.ValueOf(parsed))
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func callFromString(t reflect.Type, s string) (interface{}, error) {
	m, ok := t.MethodByName("FromString")
	if !ok {
		return nil, errors.New("Does Not Implement FromStringer")
	}

	vals := m.Func.Call([]reflect.Value{reflect.ValueOf(s)})
	return vals[0].Interface(), (vals[1].Interface()).(error)
}

func validate(t reflect.Type, parsed interface{}) error {
	m, ok := t.MethodByName("Validate")
	if !ok {
		return errors.New("Does Not Implement Validate")
	}

	return (m.Func.Call([]reflect.Value{reflect.ValueOf(parsed)})[0].Interface()).(error)
}

func parse(t reflect.Type, val string) (interface{}, error) {
	if val == "" {
		return nil, nil
	}

	return nil, nil
}
