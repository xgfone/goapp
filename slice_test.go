package goapp

import "testing"

func TestInInts(t *testing.T) {
	if !InInts(1, []int{1, 2, 3}) {
		t.Fail()
	}
	if InInts(0, []int{1, 2, 3}) {
		t.Fail()
	}
}

func TestInUints(t *testing.T) {
	if !InUints(1, []uint{1, 2, 3}) {
		t.Fail()
	}
	if InUints(0, []uint{1, 2, 3}) {
		t.Fail()
	}
}

func TestInStrings(t *testing.T) {
	if !InStrings("a", []string{"a", "b", "c"}) {
		t.Fail()
	}
	if InStrings("z", []string{"a", "b", "c"}) {
		t.Fail()
	}
}

func TestInSlice(t *testing.T) {
	if !InSlice(1, []interface{}{1, 2, 3}) {
		t.Fail()
	}
	if InSlice(0, []interface{}{1, 2, 3}) {
		t.Fail()
	}
}
