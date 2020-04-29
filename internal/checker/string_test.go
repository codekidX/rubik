package checker

import (
	"testing"
)

func TestStrMin(t *testing.T) {
	fn := stringAssertions["min-len"]
	str := "Ashish"
	if _, ok := fn(str, "8"); ok {
		t.Error("Given a string 'Ashish' of len 6 the min len assersion 8 passed")
	}

	if _, ok := fn(str, "4"); !ok {
		t.Error("Given a string 'Ashish' of len 6 the min len assersion 4 failed")
	}

	if _, ok := fn(str, str); ok {
		t.Error("Given non-integer for equating the assertion passed")
	}
}

func TestStrMax(t *testing.T) {
	fn := stringAssertions["max-len"]
	str := "Ashish"
	if _, ok := fn(str, "4"); ok {
		t.Error("Given a string 'Ashish' of len 6 the max len assertion 4 passed")
	}

	if _, ok := fn(str, "10"); !ok {
		t.Error("Given a string 'Ashish' of len 6 the max len assertion 10 failed")
	}

	if _, ok := fn(str, str); ok {
		t.Error("Given non-integer for equating the assertion passed")
	}
}
