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
	"fmt"
	"reflect"
	"testing"
)

// validateFunc tracks which rule strings are passed to it.
func validateFunc(calls map[string]int) func(reflect.StructField, reflect.Value, string) error {
	return func(_ reflect.StructField, _ reflect.Value, rule string) error {
		calls[rule]++
		return nil
	}
}

func TestVisitsOnlyTaggedExportedFields(t *testing.T) {
	type Inner struct {
		Tagged   string `validate:"r1"` // exported + tagged → visited
		Untagged string // exported + untagged → NOT visited
	}

	type Root struct {
		Tagged   string `validate:"r2"` // exported + tagged → visited
		Untagged string // exported + untagged → NOT visited
		private  string `validate:"r3"` // unexported + tagged → NOT visited
		Inner           // anonymous struct → recursed into
		Named    Inner  // named exported struct → recursed into
		P        *Inner // non-nil pointer → recursed into
		NP       *Inner // nil pointer → NOT visited at runtime
	}

	s := Root{
		Tagged: "a",
		Inner:  Inner{Tagged: "b"},
		Named:  Inner{Tagged: "c"},
		P:      &Inner{Tagged: "d"},
	}

	calls := make(map[string]int)
	if err := ValidateStruct(&s, validateFunc(calls)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// r2 is Root.Tagged (visited once)
	if got := calls["r2"]; got != 1 {
		t.Errorf("r2: want 1 call, got %d", got)
	}

	// r1 is visited 3 times: anonymous Inner, named Inner, pointer Inner
	if got := calls["r1"]; got != 3 {
		t.Errorf("r1: want 3 calls, got %d", got)
	}

	// r3 (unexported) must NOT be visited
	if _, ok := calls["r3"]; ok {
		t.Error("r3 (unexported field) should not have been visited")
	}

	// No unexpected rules
	delete(calls, "r1")
	delete(calls, "r2")
	if len(calls) > 0 {
		t.Errorf("unexpected rules visited: %v", calls)
	}
}

func TestValidateFieldCorrectness(t *testing.T) {
	t.Run("passes correct field info", func(t *testing.T) {
		type S struct {
			Name string `validate:"req"`
		}
		var gotField *reflect.StructField
		var gotVal string
		var gotRule string
		s := &S{Name: "hello"}

		ValidateStruct(s, func(field reflect.StructField, value reflect.Value, rule string) error {
			gotField = &field
			gotVal = value.String()
			gotRule = rule
			return nil
		})

		if gotField == nil || gotField.Name != "Name" {
			t.Errorf("expected field Name, got %v", gotField)
		}
		if gotVal != "hello" {
			t.Errorf("expected value hello, got %s", gotVal)
		}
		if gotRule != "req" {
			t.Errorf("expected rule req, got %s", gotRule)
		}
	})

	t.Run("passes correct type for int field", func(t *testing.T) {
		type S struct {
			Age int `validate:"gt0"`
		}
		var gotKind reflect.Kind
		var gotVal int64
		s := &S{Age: 25}
		ValidateStruct(s, func(_ reflect.StructField, value reflect.Value, rule string) error {
			gotKind = value.Kind()
			gotVal = value.Int()
			return nil
		})
		if gotKind != reflect.Int {
			t.Errorf("expected int kind, got %s", gotKind)
		}
		if gotVal != 25 {
			t.Errorf("expected 25, got %d", gotVal)
		}
	})

	t.Run("single validation error", func(t *testing.T) {
		s := &struct {
			Name string `validate:"required"`
		}{}
		err := ValidateStruct(s, func(_ reflect.StructField, v reflect.Value, rule string) error {
			if rule == "required" && v.IsZero() {
				return errors.New("name is required")
			}
			return nil
		})
		if err == nil || err.Error() != "name is required" {
			t.Errorf("expected 'name is required', got %v", err)
		}
	})

	t.Run("multiple validation errors aggregated", func(t *testing.T) {
		s := &struct {
			A string `validate:"r1"`
			B string `validate:"r2"`
		}{}
		err := ValidateStruct(s, func(_ reflect.StructField, _ reflect.Value, rule string) error {
			switch rule {
			case "r1":
				return errors.New("err1")
			case "r2":
				return errors.New("err2")
			}
			return nil
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Should contain both errors
		msg := err.Error()
		if !contains(msg, "err1") || !contains(msg, "err2") {
			t.Errorf("expected both err1 and err2, got %s", msg)
		}
	})

	t.Run("passing field returns nil", func(t *testing.T) {
		s := &struct {
			Name string `validate:"required"`
		}{Name: "ok"}
		err := ValidateStruct(s, func(_ reflect.StructField, v reflect.Value, rule string) error {
			if rule == "required" && v.IsZero() {
				return errors.New("required")
			}
			return nil
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestInputValidation(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		if err := ValidateStruct(nil, nil); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var s *struct{ A string }
		if err := ValidateStruct(s, nil); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("non struct value", func(t *testing.T) {
		if err := ValidateStruct("not a struct", nil); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("value struct (not pointer)", func(t *testing.T) {
		s := struct {
			A string `validate:"r"`
		}{A: "ok"}
		if err := ValidateStruct(s, func(_ reflect.StructField, _ reflect.Value, _ string) error {
			return nil
		}); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestNestedStructRecursion(t *testing.T) {
	t.Run("named nested struct", func(t *testing.T) {
		type Inner struct {
			X string `validate:"r1"`
		}
		type Outer struct {
			Inner Inner
		}
		calls := make(map[string]int)
		ValidateStruct(&Outer{Inner: Inner{X: "v"}}, validateFunc(calls))
		if calls["r1"] != 1 {
			t.Errorf("r1: want 1, got %d", calls["r1"])
		}
	})

	t.Run("anonymous nested struct", func(t *testing.T) {
		type Inner struct {
			X string `validate:"r1"`
		}
		type Outer struct {
			Inner
		}
		calls := make(map[string]int)
		ValidateStruct(&Outer{Inner: Inner{X: "v"}}, validateFunc(calls))
		if calls["r1"] != 1 {
			t.Errorf("r1: want 1, got %d", calls["r1"])
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type Inner struct {
			X string `validate:"r1"`
		}
		type Outer struct {
			P *Inner
		}
		calls := make(map[string]int)
		ValidateStruct(&Outer{P: &Inner{X: "v"}}, validateFunc(calls))
		if calls["r1"] != 1 {
			t.Errorf("r1: want 1, got %d", calls["r1"])
		}
	})

	t.Run("nil pointer to struct skips children", func(t *testing.T) {
		type Inner struct {
			X string `validate:"r1"`
		}
		type Outer struct {
			P *Inner
		}
		calls := make(map[string]int)
		ValidateStruct(&Outer{P: nil}, validateFunc(calls))
		if calls["r1"] != 0 {
			t.Errorf("r1: want 0, got %d", calls["r1"])
		}
	})

	t.Run("deeply nested", func(t *testing.T) {
		type A struct {
			V string `validate:"r1"`
		}
		type B struct {
			A A
		}
		type C struct {
			P *B
		}
		calls := make(map[string]int)
		ValidateStruct(&C{P: &B{A: A{V: "v"}}}, validateFunc(calls))
		if calls["r1"] != 1 {
			t.Errorf("r1: want 1, got %d", calls["r1"])
		}
	})
}

func TestCacheDoesNotAffectDifferentTypes(t *testing.T) {
	type S1 struct {
		A string `validate:"r1"`
	}
	type S2 struct {
		A string `validate:"r2"`
	}
	s1 := &S1{A: ""}
	s2 := &S2{A: "ok"}

	var calls1 []string
	var calls2 []string

	ValidateStruct(s1, func(_ reflect.StructField, v reflect.Value, rule string) error {
		calls1 = append(calls1, fmt.Sprintf("%s=%v", rule, v.IsZero()))
		if v.IsZero() {
			return errors.New("required")
		}
		return nil
	})
	ValidateStruct(s2, func(_ reflect.StructField, v reflect.Value, rule string) error {
		calls2 = append(calls2, fmt.Sprintf("%s=%v", rule, v.IsZero()))
		return nil
	})

	if len(calls1) != 1 || calls1[0] != "r1=true" {
		t.Errorf("S1: unexpected calls: %v", calls1)
	}
	if len(calls2) != 1 || calls2[0] != "r2=false" {
		t.Errorf("S2: unexpected calls: %v", calls2)
	}
}
