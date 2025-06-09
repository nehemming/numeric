package numeric

import (
	"testing"
)

func TestFlag(t *testing.T) {
	var fv fVal

	// Initially flag should be false
	if fv.flag() {
		t.Errorf("Expected flag to be false, got true")
	}

	// Set the flag to true
	fv.setFlag(true)
	if !fv.flag() {
		t.Errorf("Expected flag to be true after setting it")
	}

	// Set the flag to false again
	fv.setFlag(false)
	if fv.flag() {
		t.Errorf("Expected flag to be false after unsetting it")
	}
}

func TestSetAndGetVal(t *testing.T) {
	var fv fVal

	// Test setting and getting values
	vals := []uint32{0, 1, 123456, 0x7FFFFFFF, 0xFFFFFFFF}
	for _, v := range vals {
		fv.setVal(v)
		expected := v & uint32(maskBits)
		actual := fv.val()
		if actual != expected {
			t.Errorf("setVal(%#x): expected val() = %#x, got %#x", v, expected, actual)
		}
	}
}

func TestSetValPreservesFlag(t *testing.T) {
	var fv fVal
	fv.setFlag(true)
	fv.setVal(12345)

	if !fv.flag() {
		t.Errorf("Expected flag to be preserved after setting value")
	}

	if fv.val() != 12345 {
		t.Errorf("Expected value to be 12345, got %d", fv.val())
	}
}

func TestF24_isNegAndSetNeg(t *testing.T) {
	var f f24

	if f.isNeg() {
		t.Errorf("Expected isNeg() to be false initially")
	}

	f.setNeg(true)
	if !f.isNeg() {
		t.Errorf("Expected isNeg() to be true after setNeg(true)")
	}

	f.setNeg(false)
	if f.isNeg() {
		t.Errorf("Expected isNeg() to be false after setNeg(false)")
	}
}

func TestF24_isNaNAndSetNaN(t *testing.T) {
	var f f24

	if f.isNaN() {
		t.Errorf("Expected isNaN() to be false initially")
	}

	f.setNaN(true)
	if !f.isNaN() {
		t.Errorf("Expected isNaN() to be true after setNaN(true)")
	}

	f.setNaN(false)
	if f.isNaN() {
		t.Errorf("Expected isNaN() to be false after setNaN(false)")
	}
}

func TestF24_isOverflowAndSetOverflow(t *testing.T) {
	var f f24

	if f.isOverflow() {
		t.Errorf("Expected isOverflow() to be false initially")
	}

	f.setOverflow(true)
	if !f.isOverflow() {
		t.Errorf("Expected isOverflow() to be true after setOverflow(true)")
	}

	f.setOverflow(false)
	if f.isOverflow() {
		t.Errorf("Expected isOverflow() to be false after setOverflow(false)")
	}
}

func TestF24_isUnderflowAndSetUnderflow(t *testing.T) {
	var f f24

	if f.isUnderflow() {
		t.Errorf("Expected isUnderflow() to be false initially")
	}

	f.setUnderflow(true)
	if !f.isUnderflow() {
		t.Errorf("Expected isUnderflow() to be true after setUnderflow(true)")
	}

	f.setUnderflow(false)
	if f.isUnderflow() {
		t.Errorf("Expected isUnderflow() to be false after setUnderflow(false)")
	}
}
