package goapp

import (
	"reflect"
)

// InSlice reports whether v is in vs.
func InSlice(v interface{}, vs []interface{}) bool {
	for _, _v := range vs {
		if reflect.DeepEqual(v, _v) {
			return true
		}
	}
	return false
}

// InStrings reports whether s is in ss.
func InStrings(s string, ss []string) bool {
	for _, _s := range ss {
		if _s == s {
			return true
		}
	}
	return false
}

// InInts reports whether v is in vs.
func InInts(v int, vs []int) bool {
	for _, i := range vs {
		if i == v {
			return true
		}
	}
	return false
}

// InUints reports whether v is in vs.
func InUints(v uint, vs []uint) bool {
	for _, i := range vs {
		if i == v {
			return true
		}
	}
	return false
}
