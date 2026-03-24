package go_ecs

import (
	"testing"
)

var ss sparseSet[string]

var i0_0 = "i0_0"
var i0_1 = "i0_1"

var i1_0 = "i1_0"

var i1000_0 = "i1000_0"
var i1000_1 = "i1000_1"

var i300_0 = "i300_0"

// TODO improve tests
func TestBasic(t *testing.T) {
	// Create a new sparse set
	ss = newSparseSet[string]()

	// set new values
	ss.set(0, i0_0)
	ss.set(1, i1_0)
	ss.set(1000, i1000_0)

	// get the value 0
	value, ok := ss.get(0)
	if !ok {
		t.Error("Value not found")
	}
	if value != i0_0 {
		t.Errorf("Wrong second value: %v, not %v", value, i0_0)
	}

	// get the value 1
	value, ok = ss.get(1)
	if !ok {
		t.Error("Value not found")
	}
	if value != i1_0 {
		t.Errorf("Wrong value on index 1: %v, not %v", value, i1_0)
	}

	// get the value 1000
	value, ok = ss.get(1000)
	if !ok {
		t.Error("Value not found")
	}
	if value != i1000_0 {
		t.Errorf("Wrong value on index 1000: %v, not %v", value, i1000_0)
	}

	// Rewrite value 0
	ss.set(0, i0_1)
	value, ok = ss.get(0)
	if !ok {
		t.Error("Value not found")
	}
	if value != i0_1 {
		t.Errorf("Wrong second value: %v, not %v", value, i0_1)
	}

	// Rewrite value 1000
	ss.set(1000, i1000_1)
	value, ok = ss.get(1000)
	if !ok {
		t.Error("Value not found")
	}
	if value != i1000_1 {
		t.Errorf("Wrong second value: %v, not %v", value, i1000_1)
	}

	// Second page
	ss.set(300, i300_0)
	value, ok = ss.get(300)
	if !ok {
		t.Error("Value not found")
	}
	if value != i300_0 {
		t.Errorf("Wrong second value: %v, not %v", value, i300_0)
	}

	// Not existing index
	value, ok = ss.get(123)
	if ok {
		t.Errorf("Found value on non-existing index: %v", value)
	}
	if value != "" {
		t.Errorf("Wrong value on non-existing index: %v, not 0", value)
	}

	err := ss.delete(1000)
	if err != nil {
		t.Error(err)
	}
	value, ok = ss.get(1000)
	if ok {
		t.Errorf("Found value on deleted index: %v", value)
	}
	if value != "" {
		t.Errorf("Wrong value on deleted index: %v, not 0", value)
	}
}
