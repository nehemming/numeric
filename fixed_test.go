package numeric

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestParseString(t *testing.T) {
	tests := []struct {
		input        string
		expectNaN    bool
		expectNeg    bool
		expectOver   bool
		expectUnder  bool
		expectDigits string
	}{
		{"", true, false, false, false, "NaN"},
		{"1", false, false, false, false, "1"},
		{"-1", false, true, false, false, "-1"},
		{"1.23456", false, false, false, false, "1.23456"},
		{"1.2345600", false, false, false, false, "1.23456"},
		{"1e3", false, false, false, false, "1000"},
		{"1e17", false, false, false, false, "100000000000000000"},
		{"1e18", false, false, true, false, "<999999999999999999.999999999999999999999999999999999999"}, // Triggers overflow
		{"1.23456e-10", false, false, false, false, "0.000000000123456"},
		{"1e-20", false, false, false, false, "0.00000000000000000001"},
		{"1e-36", false, false, false, false, "0.000000000000000000000000000000000001"},
		{"1e-37", false, false, false, true, "~0"},                                                       // Triggers underflow
		{"~1.23", false, false, false, true, "~1.23"},                                                    // Underflow symbol
		{"<1.23", false, false, true, false, "<999999999999999999.999999999999999999999999999999999999"}, // Overflow symbol
		{"NaN", true, false, false, false, "NaN"},                                                        // NaN
		{"+1.23", false, false, false, false, "1.23"},                                                    // Explicit positive
		{"-1e-20", false, true, false, false, "-0.00000000000000000001"},                                 // Negative underflow
		{"-1e-37", false, true, false, true, "~-0"},                                                      // Negative underflow
		{"-1e20", false, true, true, false, "-<999999999999999999.999999999999999999999999999999999999"}, // Negative overflow
		{"0.0000000000000000000000000000000000001", false, false, false, true, "~0"},                     // Deep underflow
		{"-0.0000000000000000000000000000000000001", false, true, false, true, "~-0"},                    // Deep underflow
		{"999999999999999999", false, false, false, false, "999999999999999999"},
		{"-999999999999999999.999999999999999999999999999999999999", false, true, false, false, "-999999999999999999.999999999999999999999999999999999999"},
		{"0." + strings.Repeat("0", 55) + "1", false, false, false, true, "~0"},
		{strings.Repeat("1", 56), false, false, true, false, "<999999999999999999.999999999999999999999999999999999999"},
		{"123e-4", false, false, false, false, "0.0123"},
		{"123e-60", false, false, false, true, "~0"},
		{"~-<1", false, true, true, true, "~-<999999999999999999.999999999999999999999999999999999999"},
		{"12345678901234567890", false, false, true, false, "<999999999999999999.999999999999999999999999999999999999"},
		{"23.45e-1", false, false, false, false, "2.345"},
		{strings.Repeat("0", 58) + "1", false, false, false, false, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := parseString(tt.input)
			if err != nil {
				t.Errorf("parseString(%q) unexpected error: %v", tt.input, err)
				return
			}

			if d.isNaN != tt.expectNaN {
				t.Errorf("parseString(%q) isNaN = %v, expected %v", tt.input, d.isNaN, tt.expectNaN)
			}
			if d.isNeg != tt.expectNeg {
				t.Errorf("parseString(%q) isNeg = %v, expected %v", tt.input, d.isNeg, tt.expectNeg)
			}
			if d.isOverflow != tt.expectOver {
				t.Errorf("parseString(%q) isOverflow = %v, expected %v", tt.input, d.isOverflow, tt.expectOver)
			}
			if d.isUnderflow != tt.expectUnder {
				t.Errorf("parseString(%q) isUnderflow = %v, expected %v", tt.input, d.isUnderflow, tt.expectUnder)
			}

			if out := d.String(); out != tt.expectDigits {
				t.Errorf("parseString(%q).String() = %q, expected %q", tt.input, out, tt.expectDigits)
			}

			f := d.F24()

			if f.isNaN() != tt.expectNaN {
				t.Errorf("F24(%q) isNaN() = %v, expected %v", tt.input, f.isNaN(), tt.expectNaN)
			}
			if f.isNeg() != tt.expectNeg {
				t.Errorf("F24(%q) isNeg() = %v, expected %v", tt.input, f.isNeg(), tt.expectNeg)
			}
			if f.isOverflow() != tt.expectOver {
				t.Errorf("F24(%q) isOverflow() = %v, expected %v", tt.input, f.isOverflow(), tt.expectOver)
			}
			if f.isUnderflow() != tt.expectUnder {
				t.Errorf("F24(%q) isUnderflow() = %v, expected %v", tt.input, f.isUnderflow(), tt.expectUnder)
			}

			dBack := f.Digits()
			if dBack.isNaN != tt.expectNaN {
				t.Errorf("back digits(%q) isNaN = %v, expected %v", tt.input, dBack.isNaN, tt.expectNaN)
			}
			if dBack.isNeg != tt.expectNeg {
				t.Errorf("back digits(%q) isNeg = %v, expected %v", tt.input, dBack.isNeg, tt.expectNeg)
			}
			if dBack.isOverflow != tt.expectOver {
				t.Errorf("back digits(%q) isOverflow = %v, expected %v", tt.input, dBack.isOverflow, tt.expectOver)
			}
			if dBack.isUnderflow != tt.expectUnder {
				t.Errorf("back digits(%q) isUnderflow = %v, expected %v", tt.input, dBack.isUnderflow, tt.expectUnder)
			}

			if out := dBack.String(); out != tt.expectDigits {
				t.Errorf("back digits(%q).String() = %q, expected %q", tt.input, out, tt.expectDigits)
			}

			b := d.output(nil)
			if out := string(b); out != tt.expectDigits {
				t.Errorf("back digits(%q).output() = %q, expected %q", tt.input, out, tt.expectDigits)
			}
			// test wrapped in
			f, err = f24String(tt.input)
			if err != nil {
				t.Errorf("parseString(%q) unexpected error: %v", tt.input, err)
				return
			}

			if f.isNaN() != tt.expectNaN {
				t.Errorf("f24String(%q) isNaN() = %v, expected %v", tt.input, f.isNaN(), tt.expectNaN)
			}
			if f.isNeg() != tt.expectNeg {
				t.Errorf("f24String(%q) isNeg() = %v, expected %v", tt.input, f.isNeg(), tt.expectNeg)
			}
			if f.isOverflow() != tt.expectOver {
				t.Errorf("f24String(%q) isOverflow() = %v, expected %v", tt.input, f.isOverflow(), tt.expectOver)
			}
			if f.isUnderflow() != tt.expectUnder {
				t.Errorf("f24String(%q) isUnderflow() = %v, expected %v", tt.input, f.isUnderflow(), tt.expectUnder)
			}
		})
	}
}

func TestParseStringErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectError error
	}{
		{"~~1.23", ErrMultipleUnderflowSymbols},
		{"<<1.23", ErrMultipleOverflowSymbols},
		{"--1.23", ErrMultipleMinusSigns},
		{"+-1.23", ErrMultipleMinusSigns},
		{"-+1.23", ErrMultiplePlusSigns},
		{"++1.23", ErrMultiplePlusSigns},
	}

	for _, tt := range tests {
		_, err := parseString(tt.input)
		if err == nil {
			t.Errorf("parseString(%q): expected error %v, got nil", tt.input, tt.expectError)
			continue
		}
		if !errors.Is(err, tt.expectError) {
			t.Errorf("parseString(%q): expected error %v, got %v", tt.input, tt.expectError, err)
		}
	}
}

func TestParseString_ErrorCases(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr error
	}{
		{"1..2", ErrInvalidDecimalPoint},
		{"1.2.3", ErrInvalidDecimalPoint},
		{"1e2e3", ErrMultipleExponents},
		{"1e+2+3", ErrMultipleExponentSigns},
		{"1e", ErrNoExponentValue},
		{"e10", ErrNoDigitsInInput},
		{"abc", ErrInvalidCharacter}, // will match "invalid character: 'a'" dynamically
		{"1a2", ErrInvalidCharacter}, // will match "invalid character: 'a'" dynamically
	}

	for _, tt := range tests {
		_, err := parseString(tt.input)
		if tt.expectedErr != nil {
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("parseString(%q): expected error %v, got %v", tt.input, tt.expectedErr, err)
			}
		}
		f, err := f24String(tt.input)
		if tt.expectedErr != nil {
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("parseString(%q): expected error %v, got %v", tt.input, tt.expectedErr, err)
			}
		}
		if !f.isNaN() {
			t.Errorf("f24String(%q): expected isNaN to be true, got false", tt.input)
		}
	}
}

func TestF24Int(t *testing.T) {
	tests := []struct {
		input        int64
		expectedVal0 uint32
		expectedVal1 uint32
		expectNeg    bool
		expectOver   bool
	}{
		{0, 0, 0, false, false},
		{1, 0, 1, false, false},
		{-1, 0, 1, true, false},
		{1e9 - 1, 0, 999999999, false, false}, // max without crossing baseRadix
		{1e9, 1, 0, false, false},
		{1e9 + 1, 1, 1, false, false},
		{-1e9, 1, 0, true, false},
		{int64(maxValue + 1), 0, 0, false, true}, // triggers overflow
		{-int64(maxValue + 1), 0, 0, true, true}, // negative overflow
	}

	for _, tt := range tests {
		f := f24Int(tt.input)

		if tt.expectOver {
			if !f.isOverflow() {
				t.Errorf("f24Int(%d): expected overflow, got none", tt.input)
			}
			if f.isNeg() != tt.expectNeg {
				t.Errorf("f24Int(%d): expected neg=%v, got %v", tt.input, tt.expectNeg, f.isNeg())
			}
			continue
		}

		if f.isOverflow() {
			t.Errorf("f24Int(%d): unexpected overflow", tt.input)
		}
		if f.isNeg() != tt.expectNeg {
			t.Errorf("f24Int(%d): expected neg=%v, got %v", tt.input, tt.expectNeg, f.isNeg())
		}
		if f[0].val() != tt.expectedVal0 {
			t.Errorf("f24Int(%d): expected f[0]=%d, got %d", tt.input, tt.expectedVal0, f[0].val())
		}
		if f[1].val() != tt.expectedVal1 {
			t.Errorf("f24Int(%d): expected f[1]=%d, got %d", tt.input, tt.expectedVal1, f[1].val())
		}

		// round trip
		d := f.Digits()
		out := d.String()

		str := strconv.FormatInt(tt.input, 10)

		if out != str {
			t.Errorf("f24Int(%d): expected string %q, got %q", tt.input, str, out)
		}
	}
}

func TestF24Float64(t *testing.T) {
	tests := []struct {
		input       float64
		expectedStr string
		expectNeg   bool
		expectNaN   bool
		expectOver  bool
		expectUnder bool
	}{
		{0, "0", false, false, false, false},
		{1.0, "1", false, false, false, false},
		{-1.0, "-1", true, false, false, false},
		{123.456, "123.456", false, false, false, false},
		{-123.456, "-123.456", true, false, false, false},
		{1e3, "1000", false, false, false, false},
		{math.NaN(), "NaN", false, true, false, false},
		{math.Inf(1), "<999999999999999999.999999999999999999999999999999999999", false, false, true, false},
		{math.Inf(-1), "-<999999999999999999.999999999999999999999999999999999999", true, false, true, false},
		{1e-20, "0.00000000000000000001", false, false, false, false},
		{-1e-20, "-0.00000000000000000001", true, false, false, false},
		{1e-40, "~0", false, false, false, true},
		{-1e-40, "~-0", true, false, false, true},
		{1e100, "<999999999999999999.999999999999999999999999999999999999", false, false, true, false}, // overflow by magnitude
	}

	for _, tt := range tests {
		f := f24Float64(tt.input)

		if f.isNaN() != tt.expectNaN {
			t.Errorf("f24Float64(%v): expected isNaN=%v, got %v", tt.input, tt.expectNaN, f.isNaN())
		}
		if f.isNeg() != tt.expectNeg {
			t.Errorf("f24Float64(%v): expected isNeg=%v, got %v", tt.input, tt.expectNeg, f.isNeg())
		}
		if f.isOverflow() != tt.expectOver {
			t.Errorf("f24Float64(%v): expected isOverflow=%v, got %v", tt.input, tt.expectOver, f.isOverflow())
		}
		if f.isUnderflow() != tt.expectUnder {
			t.Errorf("f24Float64(%v): expected isUnderflow=%v, got %v", tt.input, tt.expectUnder, f.isUnderflow())
		}

		// Round trip back to string
		d := f.Digits()
		out := d.String()
		if out != tt.expectedStr {
			t.Errorf("f24Float64(%v): round-trip string = %q, expected %q", tt.input, out, tt.expectedStr)
		}
	}
}

func TestMaxF24String(t *testing.T) {
	f := maxF24
	d := f.Digits()
	out := d.String()

	expected := "999999999999999999.999999999999999999999999999999999999"

	if out != expected {
		t.Errorf("maxF24.Digits().String() = %q, want %q", out, expected)
	}
}

var smallestF24 = f24{
	0, 0, 0,
	0, 0, 1,
}

func TestSmallestF24String(t *testing.T) {
	f := smallestF24
	d := f.Digits()
	out := d.String()

	// Only the last slot is 1, representing 1e-9
	expected := "0.000000000000000000000000000000000001"

	if out != expected {
		t.Errorf("smallestF24.Digits().String() = %q, want %q", out, expected)
	}
}
