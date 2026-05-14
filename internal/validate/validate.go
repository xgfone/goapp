// Copyright 2026 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validate

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/xgfone/go-validation"
)

func Validate(structptr any) error {
	return ValidateStruct(structptr, validateField)
}

func validateField(field reflect.StructField, value reflect.Value, rule string) error {
	err := validation.Validate(value.Interface(), rule)
	if err != nil {
		err = FieldError{
			Name: getFieldName(field),
			Err:  err,
		}
	}
	return err
}

var tags = []string{"json", "header", "query", "form"}

func getFieldName(field reflect.StructField) string {
	for _, tag := range tags {
		if name := field.Tag.Get(tag); name != "" {
			return name
		}
	}
	return field.Name
}

type FieldError struct {
	Name string
	Err  error
}

func (e FieldError) Error() string {
	return strings.Join([]string{e.Name, e.Err.Error()}, ": ")
}

func (e FieldError) Unwrap() error {
	return e.Err
}

/// ------------------------------------------------------------------------ ///

type FieldValidateFunc func(field reflect.StructField, value reflect.Value, rule string) error

// ValidateStruct validates the struct fields of the given pointer using the
// provided validateField function.
//
// It only inspects the struct fields that have a "validate" struct tag,
// traversing nested structs and pointer fields recursively.
//
// If the structorptr is nil or points to a nil pointer, it returns nil.
func ValidateStruct(value any, validateField FieldValidateFunc) (err error) {
	if value == nil {
		return nil
	}

	rtype := reflect.TypeOf(value)
	rvalue := reflect.ValueOf(value)

	if rtype.Kind() == reflect.Pointer {
		if rvalue.IsNil() {
			return nil
		}

		rtype = rtype.Elem()
		rvalue = rvalue.Elem()
	}

	if rtype.Kind() != reflect.Struct {
		return errors.New("value is not a struct or a pointer to struct")
	}

	var errs []error
	for _, f := range parseStructWithCache(rtype).Fields {
		if field, ok := f.Get(rvalue); ok {
			err = validateField(f.Field, field, f.Rule)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

type (
	FieldGetter func(root reflect.Value) (reflect.Value, bool)

	Field struct {
		Field reflect.StructField
		Rule  string
		Get   FieldGetter
	}

	Struct struct {
		Fields []Field
	}
)

var structmaps sync.Map // map[reflect.Type]*Struct

func parseStructWithCache(t reflect.Type) (s *Struct) {
	if v, ok := structmaps.Load(t); ok {
		return v.(*Struct)
	}

	actual, _ := structmaps.LoadOrStore(t, parseStruct(t))
	return actual.(*Struct)
}

func parseStruct(t reflect.Type) (s *Struct) {
	fields := parseStructWithParent(t, nil)
	return &Struct{Fields: fields}
}

func parseStructWithParent(t reflect.Type, parent []int) (fields []Field) {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		ft := sf.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		// Always recurse into nested structs to find child fields with validate tags.
		if ft.Kind() == reflect.Struct && (sf.Anonymous || sf.IsExported()) {
			index := appendIndex(parent, i)
			fields = append(fields, parseStructWithParent(ft, index)...)
			continue
		}

		if !sf.IsExported() {
			continue
		}

		rule := sf.Tag.Get("validate")
		if rule == "" {
			continue
		}

		index := appendIndex(parent, i)
		fields = append(fields, Field{
			Field: sf,
			Rule:  rule,
			Get:   makeFieldGetter(index),
		})
	}

	return
}

func appendIndex(parent []int, i int) []int {
	index := make([]int, len(parent)+1)
	copy(index, parent)
	index[len(parent)] = i
	return index
}

func makeFieldGetter(index []int) FieldGetter {
	return func(root reflect.Value) (reflect.Value, bool) {
		return fieldByIndex(root, index)
	}
}

func fieldByIndex(v reflect.Value, index []int) (reflect.Value, bool) {
	cur := v

	for i, x := range index {
		f := cur.Field(x)
		if i == len(index)-1 {
			return f, true
		}

		if f.Kind() == reflect.Pointer {
			if f.Type().Elem().Kind() != reflect.Struct {
				panic("non-struct pointer in field path")
			}

			if f.IsNil() {
				return reflect.Value{}, false
			}

			cur = f.Elem()
			continue
		}

		if f.Kind() != reflect.Struct {
			panic("invalid field path")
		}

		cur = f
	}

	panic("empty field index")
}
