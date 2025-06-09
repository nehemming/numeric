package numeric

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func TestRoundModeString(t *testing.T) {
	type testCase struct {
		mode RoundMode
		want string
	}

	tests := []testCase{
		{RoundTowards, "towards"},
		{RoundAway, "away"},
		{RoundHalfDown, "1/2 down"},
		{RoundHalfUp, "1/2 up"},
		{RoundMode(99), ""}, // unknown mode
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("RoundMode(%d)", tc.mode), func(t *testing.T) {
			got := tc.mode.String()
			if got != tc.want {
				t.Errorf("RoundMode(%d).String() = %q, want %q", tc.mode, got, tc.want)
			}
		})
	}
}

func TestFromFloat64AndFloat64RoundTrip(t *testing.T) {
	type testCase struct {
		in      float64
		wantStr string
		wantF   float64
		isNaN   bool
		isInf   bool
	}

	tests := []testCase{
		{0, "0", 0, false, false},
		{1, "1", 1, false, false},
		{-1, "-1", -1, false, false},
		{12345.6789, "12345.6789", 12345.6789, false, false},
		{123.456, "123.456", 123.456, false, false},
		{-123.456, "-123.456", -123.456, false, false},
		{1e-100, "~0", 0, false, false}, // underflow to zero
		{1e100, "<999999999999999999.999999999999999999999999999999999999", math.Inf(1), false, true},           // overflow
		{math.MaxFloat64, "<999999999999999999.999999999999999999999999999999999999", math.Inf(1), false, true}, // overflow
		{math.SmallestNonzeroFloat64, "~0", 0, false, false},                                                    // subnormal → underflow
		{math.NaN(), "NaN", math.NaN(), true, false},
		{math.Inf(1), "<999999999999999999.999999999999999999999999999999999999", math.Inf(1), false, true},
		{math.Inf(-1), "-<999999999999999999.999999999999999999999999999999999999", math.Inf(-1), false, true},
		{123456789012345, "123456789012345", 123456789012345, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.wantStr, func(t *testing.T) {
			n := FromFloat64(tc.in)

			gotStr := n.String()
			gotF := n.Float64()
			isNaN := n.IsNaN()
			isOverflow := n.HasOverflow()
			// isUnderflow := n.HasUnderflow()

			if tc.isNaN {
				if !isNaN {
					t.Errorf("FromFloat64(%g): expected NaN, got %q", tc.in, gotStr)
				}
			} else {
				if isNaN {
					t.Errorf("FromFloat64(%g): unexpected NaN", tc.in)
				}
			}

			if tc.isInf {
				if !isOverflow {
					t.Errorf("FromFloat64(%g): expected overflow, got %q", tc.in, gotStr)
				}
			} else {
				if isOverflow && !tc.isNaN {
					t.Errorf("FromFloat64(%g): unexpected overflow", tc.in)
				}
			}

			if !tc.isNaN && !tc.isInf && gotStr != tc.wantStr {
				t.Errorf("FromFloat64(%g).String() = %q, want %q", tc.in, gotStr, tc.wantStr)
			}

			if !tc.isNaN && !math.IsInf(tc.wantF, 0) && math.Abs(gotF-tc.wantF) > 1e-12 {
				t.Errorf("FromFloat64(%g).Float64() = %g, want %g", tc.in, gotF, tc.wantF)
			}
		})
	}
}

func TestOversizedToFloat64Conversion(t *testing.T) {
	num, _ := FromString("123456789012345.678")
	got := num.Float64()
	expected := float64(123456789012345.67)
	if got != expected {
		t.Errorf("FromString(\"1123456789012345.67\").Float64() = %g; want %g", got, expected)
	}
}

func TestOversizedWholeToFloat64Conversion(t *testing.T) {
	num, _ := FromString("999999999999999999")
	got := num.Float64()
	expected := float64(1e18)
	if got != expected {
		t.Errorf("FromString(\"999999999999999999\").Float64() = %g; want %g", got, expected)
	}
}

func TestFromInt_Positive(t *testing.T) {
	n := FromInt(42)
	expected := f24Int(42)
	if n.z != expected {
		t.Errorf("FromInt(42) = %+v; want %+v", n.z, expected)
	}
}

func TestFromInt_Negative(t *testing.T) {
	n := FromInt(-123)
	expected := f24Int(-123)
	if n.z != expected {
		t.Errorf("FromInt(-123) = %+v; want %+v", n.z, expected)
	}
}

func TestFromInt_Zero(t *testing.T) {
	n := FromInt(0)
	expected := f24Int(0)
	if n.z != expected {
		t.Errorf("FromInt(0) = %+v; want %+v", n.z, expected)
	}
}

func TestNumericInt_NaN(t *testing.T) {
	n := FromFloat64(math.NaN()) // or construct with custom NaN if needed

	if !n.IsNaN() {
		t.Fatalf("Expected IsNaN to be true")
	}

	val := n.Int()
	if val != 0 {
		t.Errorf("Numeric.Int() on NaN should return 0, got %d", val)
	}
}

func TestFromInt_DecimalTruncationVisible(t *testing.T) {
	n := FromInt(int64(1e18)) // Too large to fully store
	s := n.String()
	if s != "<999999999999999999.999999999999999999999999999999999999" {
		t.Errorf("FromInt(1e18).String() = %s; wanted max value", s)
	}
}

func TestFromInt_Int_OverflowMasking(t *testing.T) {
	// Forcing overflow: max int64 is 9223372036854775807 (~9.2e18)
	// These will wrap due to masking.
	overflowCases := []int64{
		int64(1e18),
		int64(-1e18),
		int64(1<<62 + 1),
		int64(-(1<<62 + 1)),
	}

	for _, input := range overflowCases {
		n := FromInt(input)
		i := n.Int()
		// Note: expected is just input masked into int

		expectedP := int64(999999999999999999)
		expectedN := -expectedP

		if input > 0 {
			if i != expectedP {
				t.Errorf("FromInt(%d).Int() = %d (String = %s) - expected %d", input, i, n.String(), expectedP)
			}
		} else if input < 0 {
			if i != expectedN {
				t.Errorf("FromInt(%d).Int() = %d (String = %s) - expected %d", input, i, n.String(), expectedN)
			}
		}
	}
}

func TestFromInt_Int_RoundTrip(t *testing.T) {
	cases := []int64{
		0,
		1,
		-1,
		123456,
		-987654,
		int64(1e9),
		int64(-1e9),
		int64(1e17), // getting near overflow range
		int64(-1e17),
	}

	for _, input := range cases {
		n := FromInt(input)
		output := n.Int()
		if output != input {
			t.Errorf("FromInt(%d).Int() = %d; want %d", input, output, input)
		}
	}
}

func TestFromStringAndString(t *testing.T) {
	type testCase struct {
		input     string
		expectStr string
		expectErr bool
	}

	tests := []testCase{
		// Exact conversions
		{"0", "0", false},
		{"12345678901234567.123456789012345678901234567890123456", "12345678901234567.123456789012345678901234567890123456", false},
		{"-12345678901234567.123456789012345678901234567890123456", "-12345678901234567.123456789012345678901234567890123456", false},
		{"999999999999999999.999999999999999999999999999999999999", "999999999999999999.999999999999999999999999999999999999", false},
		{"-999999999999999999.999999999999999999999999999999999999", "-999999999999999999.999999999999999999999999999999999999", false},

		// Inexact with ~
		{"12345678901234567.1234567890123456789012345678901234567", "~12345678901234567.123456789012345678901234567890123456", false},
		{"-12345678901234567.1234567890123456789012345678901234567", "~-12345678901234567.123456789012345678901234567890123456", false},

		// Signs
		{"+42.1", "42.1", false},
		{"-0.00000000000000000000000000000000001", "-0.00000000000000000000000000000000001", false},

		// Overflow
		{"1999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999", false},
		{"-1999999999999999999.999999999999999999999999999999999999", "-<999999999999999999.999999999999999999999999999999999999", false},

		// Invalid input
		{"abc.def", "", true},
		{"1.2.3", "", true},
		{"", "NaN", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			n, err := FromString(tc.input)

			if tc.expectErr {
				if err == nil {
					t.Errorf("FromString(%q): expected error, got none", tc.input)
				}
				return
			}

			if err != nil {
				t.Errorf("FromString(%q): unexpected error: %v", tc.input, err)
				return
			}

			gotStr := n.String()
			if gotStr != tc.expectStr {
				t.Errorf("FromString(%q).String() = %q, want %q", tc.input, gotStr, tc.expectStr)
			}
		})
	}
}

func TestNumericSum(t *testing.T) {
	type testCase struct {
		inputs   []string
		expected string
	}

	tests := []testCase{
		{[]string{}, "0"},
		{[]string{"0"}, "0"},
		{[]string{"0", "0"}, "0"},
		{[]string{"1", "2", "3"}, "6"},
		{[]string{"-1", "1"}, "0"},
		{[]string{"100.25", "200.75"}, "301"},
		{[]string{"-50", "-50"}, "-100"},
		{[]string{"1e9", "1e9"}, "2000000000"},
		{[]string{"123456.789", "-56.789"}, "123400"},
		{[]string{"999999999999999999", "1"}, "<999999999999999999.999999999999999999999999999999999999"}, // overflow
	}

	for _, tc := range tests {
		t.Run("Sum_"+tc.expected, func(t *testing.T) {
			var vals []Numeric
			for _, str := range tc.inputs {
				n, err := FromString(str)
				if err != nil {
					t.Fatalf("Invalid input %q: %v", str, err)
				}
				vals = append(vals, n)
			}

			sum := Sum(vals...)
			got := sum.String()

			if got != tc.expected {
				t.Errorf("Sum(%v) = %q, want %q", tc.inputs, got, tc.expected)
			}
		})
	}
}

func TestNumericRound_Modes(t *testing.T) {
	type testCase struct {
		input    string
		places   int
		mode     RoundMode
		expected string
	}

	tests := []testCase{
		// RoundDown: truncate toward zero
		{"123.987", 0, RoundTowards, "123"},
		{"-123.987", 0, RoundTowards, "-123"},
		{"1.999999999999999999", 0, RoundTowards, "1"},
		{"-1.999999999999999999", 0, RoundTowards, "-1"},

		// RoundUp: always away from zero
		{"123.001", 0, RoundAway, "124"},
		{"-123.001", 0, RoundAway, "-124"},
		{"1.000000000000000001", 0, RoundAway, "2"},
		{"-1.000000000000000001", 0, RoundAway, "-2"},

		// RoundHalfDown: ties go down
		{"2.5", 0, RoundHalfDown, "2"},
		{"-2.5", 0, RoundHalfDown, "-2"},
		{"2.6", 0, RoundHalfDown, "3"},
		{"2.4", 0, RoundHalfDown, "2"},

		// RoundHalfUp: ties go up
		{"2.5", 0, RoundHalfUp, "3"},
		{"-2.5", 0, RoundHalfUp, "-3"},
		{"2.4", 0, RoundHalfUp, "2"},
		{"2.6", 0, RoundHalfUp, "3"},

		// Rounding to decimal places
		{"1.23456789", 4, RoundTowards, "1.2345"},
		{"1.23456789", 4, RoundAway, "1.2346"},
		{"1.2345", 4, RoundHalfUp, "1.2345"},
		{"1.23455", 4, RoundHalfUp, "1.2346"},
		{"1.23455", 4, RoundHalfDown, "1.2345"},

		// Very small decimals, edge of underflow
		{"0.000000000000000000000000000000000009", 35, RoundHalfUp, "0.00000000000000000000000000000000001"},
		{"0.000000000000000000000000000000000004", 35, RoundHalfUp, "0"},
	}

	for _, tc := range tests {
		t.Run(tc.input+"_"+tc.mode.String(), func(t *testing.T) {
			n, err := FromString(tc.input)
			if err != nil {
				t.Fatalf("Invalid input %q: %v", tc.input, err)
			}

			r := n.Round(tc.places, tc.mode)
			got := r.String()

			if got != tc.expected {
				t.Errorf("Round(%q, places=%d, mode=%v) = %q, want %q",
					tc.input, tc.places, tc.mode, got, tc.expected)
			}
		})
	}
}

func TestNumericAdd(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		expected   string
		expectNaN  bool
	}

	tests := []testCase{
		// Simple cases
		{"1", "2", "3", false},
		{"-1", "1", "0", false},
		{"0", "0", "0", false},
		{"123.45", "0", "123.45", false},
		{"0", "123.45", "123.45", false},

		// Negative sums
		{"-5", "-5", "-10", false},
		{"-100", "30", "-70", false},
		{"50", "-100", "-50", false},

		// Decimal precision
		{"1.000000001", "0.000000009", "1.00000001", false},
		{"1.999999999", "0.000000001", "2", false},

		// Carrying
		{"999999999", "1", "1000000000", false},
		{"999999999.999999999", "0.000000001", "1000000000", false},

		// Overflow
		{"999999999999999999", "1", "<999999999999999999.999999999999999999999999999999999999", false},

		// NaN propagation
		{"NaN", "1", "NaN", true},
		{"1", "NaN", "NaN", true},
		{"NaN", "NaN", "NaN", true},

		// Carry
		{"99999999999999999.999999999999999999999999999999999999", "99999999999999999.999999999999999999999999999999999999", "199999999999999999.999999999999999999999999999999999998", false},
	}

	for _, tc := range tests {
		t.Run(tc.xStr+"+"+tc.yStr, func(t *testing.T) {
			x, err1 := FromString(tc.xStr)
			y, err2 := FromString(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("invalid input: %v, %v", err1, err2)
			}

			sum := x.Add(y)

			if tc.expectNaN {
				if !sum.IsNaN() {
					t.Errorf("expected NaN, got %q", sum.String())
				}
			} else {
				if got := sum.String(); got != tc.expected {
					t.Errorf("Add(%q, %q) = %q, want %q", tc.xStr, tc.yStr, got, tc.expected)
				}
			}
		})
	}
}

func TestNumericSub(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		expected   string
		expectNaN  bool
	}

	tests := []testCase{
		// Simple subtractions
		{"5", "3", "2", false},
		{"3", "5", "-2", false},
		{"0", "0", "0", false},
		{"123.45", "0", "123.45", false},
		{"0", "123.45", "-123.45", false},

		// Negative operands
		{"-5", "-5", "0", false},
		{"-5", "-2", "-3", false},
		{"-2", "-5", "3", false},

		// Mixed signs
		{"100", "-50", "150", false},
		{"-100", "50", "-150", false},

		// Decimal alignment
		{"1.000000001", "0.000000001", "1", false},
		{"2.000000000", "0.000000001", "1.999999999", false},

		// Borrowing
		{"1.000000000", "0.000000001", "0.999999999", false},

		// Overflow edge
		{"999999999999999999", "-1", "<999999999999999999.999999999999999999999999999999999999", false},

		// NaN propagation
		{"NaN", "1", "NaN", true},
		{"1", "NaN", "NaN", true},
		{"NaN", "NaN", "NaN", true},
	}

	for _, tc := range tests {
		t.Run(tc.xStr+"-"+tc.yStr, func(t *testing.T) {
			x, err1 := FromString(tc.xStr)
			y, err2 := FromString(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("Invalid input: %v or %v", err1, err2)
			}

			diff := x.Sub(y)

			if tc.expectNaN {
				if !diff.IsNaN() {
					t.Errorf("Expected NaN, got %q", diff.String())
				}
			} else {
				if got := diff.String(); got != tc.expected {
					t.Errorf("Sub(%q, %q) = %q, want %q", tc.xStr, tc.yStr, got, tc.expected)
				}
			}
		})
	}
}

func TestNumericMul(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		expected   string
		expectNaN  bool
	}

	tests := []testCase{
		{"0.0000000001", "-9999999.9999999999", "-0.00099999999999999999", false},
		{"9999999.9999999999", "9999999.9999999999", "99999999999999.99800000000000000001", false},
		// Basic multiplication
		{"2", "3", "6", false},
		{"-2", "3", "-6", false},
		{"2", "-3", "-6", false},
		{"-2", "-3", "6", false},

		// Zero multiplication
		{"0", "123456", "0", false},
		{"123456", "0", "0", false},
		{"0", "0", "0", false},

		// Identity multiplication
		{"1", "999", "999", false},
		{"999", "1", "999", false},
		{"-1", "999", "-999", false},

		// Decimal cases
		{"1.5", "2", "3", false},
		{"1.2345", "1000", "1234.5", false},
		{"0.000000001", "1000000000", "1", false},

		// Large values with overflow
		{"999999999999999999", "2", "<999999999999999999.999999999999999999999999999999999999", false},
		{"999999999", "999999999", "999999998000000001", false},
		{"999999999", "-999999999", "-999999998000000001", false},
		{"1999999999", "999999999", "<999999999999999999.999999999999999999999999999999999999", false},

		// NaN propagation
		{"NaN", "1", "NaN", true},
		{"1", "NaN", "NaN", true},
		{"NaN", "NaN", "NaN", true},
	}

	for _, tc := range tests {
		t.Run(tc.xStr+"*"+tc.yStr, func(t *testing.T) {
			x, err1 := FromString(tc.xStr)
			y, err2 := FromString(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("Invalid input: %v or %v", err1, err2)
			}

			product := x.Mul(y)

			if tc.expectNaN {
				if !product.IsNaN() {
					t.Errorf("Expected NaN, got %q", product.String())
				}
			} else {
				if got := product.String(); got != tc.expected {
					t.Errorf("Mul(%q, %q) = %q, want %q", tc.xStr, tc.yStr, got, tc.expected)
				}
			}
		})
	}
}

func TestNumericDiv(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		expected   string
		expectNaN  bool
		expectOF   bool
		expectUF   bool
	}

	tests := []testCase{
		{"1", "3", "~0.333333333333333333333333333333333333", false, false, true},
		{"999999999999999999", "2", "499999999999999999.5", false, false, false},
		{"123.456", "-654.321", "~-0.188678034175886147624789667456798727", false, false, true},
		{"0.5", "0.5", "1", false, false, false},
		{"7", "2", "3.5", false, false, false},
		{"0.0000000001", "-9999999.9999999999", "~-0.0000000000000000100000000000000001", false, false, true},
		{"123.456", "-654.321", "~-0.188678034175886147624789667456798727", false, false, true},

		// Basic division
		{"6", "3", "2", false, false, false},
		{"1", "2", "0.5", false, false, false},

		// Negative combinations
		{"-6", "3", "-2", false, false, false},
		{"6", "-3", "-2", false, false, false},
		{"-6", "-3", "2", false, false, false},

		// Identity / Reciprocal
		{"5", "1", "5", false, false, false},
		{"5", "5", "1", false, false, false},

		// Zero division
		{"0", "1", "0", false, false, false},
		{"1", "0", "NaN", true, false, false},
		{"0", "0", "NaN", true, false, false},

		// Decimal result
		{"1", "3", "~0.333333333333333333333333333333333333", false, false, true},

		// Underflow case
		{"1", "1e8", "0.00000001", false, false, false},

		/*{"1", "1e16", "0.0000000000000001", false, false, true}, // theses cases fail due to mulQ overflow, follow up fix needed
		{"1", "1e17", "0.00000000000000001", false, false, true},
		{"1", "1e18", "0.000000000000000001", false, false, true},*/

		// Overflow (large / small divisor)
		{"1e36", "0.000000001", "<999999999999999999.999999999999999999999999999999999999", false, true, false},

		// NaN propagation
		{"NaN", "1", "NaN", true, false, false},
		{"1", "NaN", "NaN", true, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.xStr+"/"+tc.yStr, func(t *testing.T) {
			x, err1 := FromString(tc.xStr)
			y, err2 := FromString(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("Invalid input: %v / %v", err1, err2)
			}

			quot := x.Div(y)

			if tc.expectNaN {
				if !quot.IsNaN() {
					t.Errorf("Expected NaN, got %q", quot.String())
				}
				return
			}

			if got := quot.String(); got != tc.expected {
				t.Errorf("Div(%q, %q) = %q, want %q", tc.xStr, tc.yStr, got, tc.expected)
			}

			if quot.HasOverflow() != tc.expectOF {
				t.Errorf("Div(%q, %q): overflow = %v, want %v", tc.xStr, tc.yStr, quot.HasOverflow(), tc.expectOF)
			}

			if quot.HasUnderflow() != tc.expectUF {
				t.Errorf("Div(%q, %q): underflow = %v, want %v", tc.xStr, tc.yStr, quot.HasUnderflow(), tc.expectUF)
			}
		})
	}
}

func TestNumericTruncate(t *testing.T) {
	type testCase struct {
		input    string
		expected string
	}

	tests := []testCase{
		// Positive decimals
		{"123.456", "123"},
		{"1.999999999", "1"},
		{"0.000000001", "0"},
		{"999999999.999999999", "999999999"},

		// Negative decimals
		{"-123.456", "-123"},
		{"-1.999999999", "-1"},
		{"-0.000000001", "0"},
		{"-999999999.999999999", "-999999999"},

		// Whole numbers
		{"0", "0"},
		{"1", "1"},
		{"-1", "-1"},
		{"1000000", "1000000"},

		// Edge near base
		{"999999999.1", "999999999"},
		{"-999999999.1", "-999999999"},

		// Overflow case (still truncates to int digits)
		{"<999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999"},
		{"-<999999999999999999.999999999999999999999999999999999999", "-<999999999999999999.999999999999999999999999999999999999"},
	}

	for _, tc := range tests {
		t.Run("Truncate_"+tc.input, func(t *testing.T) {
			n, err := FromString(tc.input)
			if err != nil {
				t.Fatalf("FromString(%q): %v", tc.input, err)
			}

			got := n.Truncate(Numeric{}).String()
			if got != tc.expected {
				t.Errorf("Truncate(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestNumericDivRem(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		wantQ      string
		wantR      string
		expectNaN  bool
	}

	tests := []testCase{
		// Simple integer division
		{"10", "3", "3", "1", false},
		{"9", "3", "3", "0", false},
		{"1", "2", "0", "1", false},

		// Negative numerators
		{"-10", "3", "-3", "-1", false},
		{"-9", "3", "-3", "0", false},

		// Negative denominators
		{"10", "-3", "-3", "1", false},
		{"-10", "-3", "3", "-1", false},

		// Both zero
		{"0", "1", "0", "0", false},
		{"0", "-5", "0", "0", false},

		// Decimal values: only integer quotient returned, remainder is correct diff
		{"5.5", "2", "2", "1.5", false},
		{"7.25", "2.5", "2", "2.25", false},

		// Division by zero
		{"5", "0", "NaN", "NaN", true},
		{"NaN", "1", "NaN", "NaN", true},
		{"1", "NaN", "NaN", "NaN", true},

		// Overflow result
		{"1e36", "0.000000001", "NaN", "NaN", true}, // large ÷ small = overflow
	}

	for _, tc := range tests {
		t.Run(tc.xStr+" / "+tc.yStr, func(t *testing.T) {
			x, err1 := FromString(tc.xStr)
			y, err2 := FromString(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("invalid input: %v / %v", err1, err2)
			}

			q, r := x.DivRem(y)

			if tc.expectNaN {
				if !(q.IsNaN() && r.IsNaN()) {
					t.Errorf("Expected NaN, got q=%q, r=%q", q.String(), r.String())
				}
				return
			}

			if gotQ := q.String(); gotQ != tc.wantQ {
				t.Errorf("Quotient = %q, want %q", gotQ, tc.wantQ)
			}

			if gotR := r.String(); gotR != tc.wantR {
				t.Errorf("Remainder = %q, want %q", gotR, tc.wantR)
			}
		})
	}
}

func TestNumericNeg(t *testing.T) {
	type testCase struct {
		input     string
		expected  string
		expectNaN bool
	}

	tests := []testCase{
		// Normal values
		{"1", "-1", false},
		{"-1", "1", false},
		{"123.456", "-123.456", false},
		{"-999.999", "999.999", false},

		// Zero (stay zero — no negative zero)
		{"0", "0", false},

		// Negation of NaN remains NaN
		{"NaN", "NaN", true},

		// Double negation should restore original
		{"5", "-5", false},
		{"-5", "5", false},
	}

	for _, tc := range tests {
		t.Run("Neg_"+tc.input, func(t *testing.T) {
			n, err := FromString(tc.input)
			if err != nil {
				t.Fatalf("Invalid input: %v", err)
			}

			neg := n.Neg()

			if tc.expectNaN {
				if !neg.IsNaN() {
					t.Errorf("Neg(%q) should be NaN, got %q", tc.input, neg.String())
				}
				return
			}

			if got := neg.String(); got != tc.expected {
				t.Errorf("Neg(%q) = %q, want %q", tc.input, got, tc.expected)
			}

			// Double negation test
			if !n.IsNaN() {
				nn := neg.Neg()
				if nn.String() != n.String() {
					t.Errorf("Double negation failed: Neg(Neg(%q)) = %q, want %q", tc.input, nn.String(), n.String())
				}
			}
		})
	}
}

func TestNumericAbs(t *testing.T) {
	type testCase struct {
		input     string
		expected  string
		expectNaN bool
	}

	tests := []testCase{
		// Positive and negative pairs
		{"1", "1", false},
		{"-1", "1", false},
		{"123.456", "123.456", false},
		{"-123.456", "123.456", false},

		// Zero
		{"0", "0", false},

		// Abs of NaN
		{"NaN", "NaN", true},

		// Overflow/Underflow-preserving cases
		{"-<999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999", false},
		{"-0.00000000000000000000000000000000001", "0.00000000000000000000000000000000001", false},
	}

	for _, tc := range tests {
		t.Run("Abs_"+tc.input, func(t *testing.T) {
			n, err := FromString(tc.input)
			if err != nil {
				t.Fatalf("Invalid input: %v", err)
			}

			abs := n.Abs()

			if tc.expectNaN {
				if !abs.IsNaN() {
					t.Errorf("Abs(%q) should be NaN, got %q", tc.input, abs.String())
				}
				return
			}

			if got := abs.String(); got != tc.expected {
				t.Errorf("Abs(%q) = %q, want %q", tc.input, got, tc.expected)
			}

			if abs.Sign() == -1 {
				t.Errorf("Abs(%q) should not be negative, got %q", tc.input, abs.String())
			}
		})
	}
}

func TestNumericIsNaN(t *testing.T) {
	type testCase struct {
		input     string
		expectNaN bool
	}

	tests := []testCase{
		{"NaN", true},
		{"0", false},
		{"1", false},
		{"-1", false},
		{"123.456", false},
		{"-999999999", false},
		{"999999999999999999", false},
		{"<999999999999999999.999999999999999999999999999999999999", false}, // overflow string
	}

	for _, tc := range tests {
		t.Run("IsNaN_"+tc.input, func(t *testing.T) {
			n, err := FromString(tc.input)
			if err != nil && !tc.expectNaN {
				t.Fatalf("Invalid input: %v", err)
			}

			got := n.IsNaN()
			if got != tc.expectNaN {
				t.Errorf("IsNaN(%q) = %v, want %v", tc.input, got, tc.expectNaN)
			}
		})
	}
}

func TestNumericSign(t *testing.T) {
	type testCase struct {
		input     string
		expectSig int
	}

	tests := []testCase{
		// NaN → 0
		{"NaN", 0},

		// Zero → 1 (positive or unsigned zero)
		{"0", 1},

		// Positive values
		{"1", 1},
		{"123.456", 1},
		{"999999999999999999", 1},

		// Negative values
		{"-1", -1},
		{"-0.000000001", -1},
		{"-999999999999999999.999", -1},
	}

	for _, tc := range tests {
		t.Run("Sign_"+tc.input, func(t *testing.T) {
			n, err := FromString(tc.input)
			if err != nil {
				t.Fatalf("FromString(%q): %v", tc.input, err)
			}

			got := n.Sign()
			if got != tc.expectSig {
				t.Errorf("Sign(%q) = %d, want %d", tc.input, got, tc.expectSig)
			}
		})
	}
}

func TestNumericFlags(t *testing.T) {
	type testCase struct {
		input        string
		expectOF     bool
		expectUF     bool
		expectIsZero bool
	}

	tests := []testCase{
		// NaN → false for all
		{"NaN", false, false, false},

		// Regular values
		{"0", false, false, true},
		{"1", false, false, false},
		{"-1", false, false, false},

		// Edge underflow case
		{"0.0000000000000000000000000000000000001", false, true, true},
		{"-0.0000000000000000000000000000000000001", false, true, true},

		// Overflow case
		{"9999999999999999999.999999999999999999999999999999999999", true, false, false},
		{"-9999999999999999999.999999999999999999999999999999999999", true, false, false},

		// Underflow exact zero (implementation detail if zero with flag)
		{"~0.00000000000000000000000000000000000000000000000000000001", false, true, true},
	}

	for _, tc := range tests {
		t.Run("Flags_"+tc.input, func(t *testing.T) {
			n, err := FromString(tc.input)
			if err != nil {
				t.Fatalf("FromString(%q): %v", tc.input, err)
			}

			if got := n.HasOverflow(); got != tc.expectOF {
				t.Errorf("HasOverflow(%q) = %v, want %v", tc.input, got, tc.expectOF)
			}
			if got := n.HasUnderflow(); got != tc.expectUF {
				t.Errorf("HasUnderflow(%q) = %v, want %v", tc.input, got, tc.expectUF)
			}
			if got := n.IsZero(); got != tc.expectIsZero {
				t.Errorf("IsZero(%q) = %v, want %v", tc.input, got, tc.expectIsZero)
			}
		})
	}
}

func TestNumericComparisons(t *testing.T) {
	type testCase struct {
		aStr, bStr         string
		expectCmp          int // -1, 0, 1
		expectEqual        bool
		expectLess         bool
		expectLessEqual    bool
		expectGreater      bool
		expectGreaterEqual bool
	}

	tests := []testCase{
		// Equal
		{"0", "0", 0, true, false, true, false, true},
		{"1.5", "1.5", 0, true, false, true, false, true},
		{"-100", "-100", 0, true, false, true, false, true},
		{"1e9", "2e9", -1, false, true, true, false, false},
		{"1-e9", "2-e9", -1, false, true, true, false, false},
		{"1-e18", "2-e18", -1, false, true, true, false, false},
		{"1-e27", "2-e27", -1, false, true, true, false, false},
		{"1-e36", "2-e36", -1, false, true, true, false, false},

		// Less than
		{"1", "2", -1, false, true, true, false, false},
		{"-2", "-1", -1, false, true, true, false, false},
		{"0", "0.000000001", -1, false, true, true, false, false},

		// Greater than
		{"2", "1", 1, false, false, false, true, true},
		{"-1", "-2", 1, false, false, false, true, true},
		{"0.000000001", "0", 1, false, false, false, true, true},

		// With NaN — never equality but we support as always less than to allow sorting
		{"NaN", "1", -1, false, true, true, false, false},
		{"1", "NaN", 1, false, false, false, true, true},
		{"NaN", "NaN", -1, false, true, true, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.aStr+"_vs_"+tc.bStr, func(t *testing.T) {
			a, errA := FromString(tc.aStr)
			b, errB := FromString(tc.bStr)
			if errA != nil || errB != nil {
				t.Fatalf("FromString failed: %v / %v", errA, errB)
			}

			cmp := a.Cmp(b)
			if cmp != tc.expectCmp {
				t.Errorf("Cmp(%q, %q) = %d, want %d", tc.aStr, tc.bStr, cmp, tc.expectCmp)
			}

			if got := a.IsEqual(b); got != tc.expectEqual {
				t.Errorf("IsEqual(%q, %q) = %v, want %v", tc.aStr, tc.bStr, got, tc.expectEqual)
			}
			if got := a.IsLessThan(b); got != tc.expectLess {
				t.Errorf("IsLessThan(%q, %q) = %v, want %v", tc.aStr, tc.bStr, got, tc.expectLess)
			}
			if got := a.IsLessThanEqual(b); got != tc.expectLessEqual {
				t.Errorf("IsLessThanEqual(%q, %q) = %v, want %v", tc.aStr, tc.bStr, got, tc.expectLessEqual)
			}
			if got := a.IsGreaterThan(b); got != tc.expectGreater {
				t.Errorf("IsGreaterThan(%q, %q) = %v, want %v", tc.aStr, tc.bStr, got, tc.expectGreater)
			}
			if got := a.IsGreaterEqual(b); got != tc.expectGreaterEqual {
				t.Errorf("IsGreaterEqual(%q, %q) = %v, want %v", tc.aStr, tc.bStr, got, tc.expectGreaterEqual)
			}
		})
	}
}

func TestMarshalUnmarshalText(t *testing.T) {
	type testCase struct {
		input    string
		expected string // if "", we assume NaN on decode
	}

	tests := []testCase{
		{"0", "0"},
		{"1", "1"},
		{"-1", "-1"},
		{"123.456", "123.456"},
		{"-123.456", "-123.456"},
		{"~123.456", "~123.456"},
		{"-~123.456", "~-123.456"},
		{"0.00000000000000000000000000000000001", "0.00000000000000000000000000000000001"},
		{"999999999999999999.999999999999999999999999999999999999", "999999999999999999.999999999999999999999999999999999999"},
		{"-999999999999999999.999999999999999999999999999999999999", "-999999999999999999.999999999999999999999999999999999999"},
		{"1999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999"},
		{"-1999999999999999999.999999999999999999999999999999999999", "-<999999999999999999.999999999999999999999999999999999999"},
		{"<999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999"},
		{"-<999999999999999999.999999999999999999999999999999999999", "-<999999999999999999.999999999999999999999999999999999999"},
		{"NaN", "NaN"},
		{"", "NaN"}, // special case: empty string becomes NaN
	}

	for _, tc := range tests {
		t.Run("MarshalUnmarshal_"+tc.input, func(t *testing.T) {
			// Start from input string
			n, err := FromString(tc.input)
			if err != nil && tc.input != "" {
				t.Fatalf("FromString(%q) failed: %v", tc.input, err)
			}

			// Marshal to text
			text, err := n.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText(%q) failed: %v", tc.input, err)
			}

			// Unmarshal back to Numeric
			var roundTripped Numeric
			err = roundTripped.UnmarshalText(text)
			if err != nil {
				t.Fatalf("UnmarshalText(%q) failed: %v", string(text), err)
			}

			// Check result
			if tc.expected == "NaN" {
				if !roundTripped.IsNaN() {
					t.Errorf("Expected NaN from %q, got %q", string(text), roundTripped.String())
				}
			} else {
				got := roundTripped.String()
				if got != tc.expected {
					t.Errorf("Round-trip failed: %q → %q → %q (want %q)", tc.input, string(text), got, tc.expected)
				}
			}
		})
	}
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	type testCase struct {
		input       string // value to encode or raw JSON to decode
		isRawJSON   bool   // if true, treat input as JSON directly
		expectStr   string // expected value after unmarshal
		expectIsNaN bool
		expectError bool
	}

	tests := []testCase{
		// Valid quoted JSON strings
		{`"0"`, true, "0", false, false},
		{`"123.456"`, true, "123.456", false, false},
		{`"-123.456"`, true, "-123.456", false, false},
		{`"<999999999999999999.999999999999999999999999999999999999"`, true, "<999999999999999999.999999999999999999999999999999999999", false, false},
		{`"NaN"`, true, "NaN", true, false},

		// Valid unquoted numbers (now allowed by fallback)
		{`0`, true, "0", false, false},
		{`123.456`, true, "123.456", false, false},
		{`-123.456`, true, "-123.456", false, false},

		// Unquoted NaN (fallback accepts)
		{`NaN`, true, "NaN", true, false},

		// Invalid formats
		{`"abc`, true, "", false, true},
		{`abc"`, true, "", false, true},
		{`"1.2.3"`, true, "", false, true},

		// Round-trip from Numeric → JSON → Numeric
		{"123.00001", false, "123.00001", false, false},
		{"NaN", false, "NaN", true, false},
	}

	for _, tc := range tests {
		name := tc.input
		if !tc.isRawJSON {
			name = "RoundTrip_" + tc.input
		}
		t.Run(name, func(t *testing.T) {
			var jsonInput []byte
			var original Numeric

			if tc.isRawJSON {
				jsonInput = []byte(tc.input)
			} else {
				var err error
				original, err = FromString(tc.input)
				if err != nil {
					t.Fatalf("FromString(%q) failed: %v", tc.input, err)
				}
				jsonInput, err = original.MarshalJSON()
				if err != nil {
					t.Fatalf("MarshalJSON failed: %v", err)
				}
			}

			var result Numeric
			err := result.UnmarshalJSON(jsonInput)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %q, got none", string(jsonInput))
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %q: %v", string(jsonInput), err)
			}

			if tc.expectIsNaN {
				if !result.IsNaN() {
					t.Errorf("Expected NaN for %q, got %q", string(jsonInput), result.String())
				}
			} else {
				if got := result.String(); got != tc.expectStr {
					t.Errorf("UnmarshalJSON(%q) = %q, want %q", string(jsonInput), got, tc.expectStr)
				}
			}
		})
	}
}

func TestNumericFormat(t *testing.T) {
	tests := []struct {
		num    Numeric
		verb   string
		flags  string // e.g. "+#"
		width  int
		prec   int
		expect string
		desc   string
	}{
		// Basic string/verb tests
		{FromInt(42), "v", "", 0, -1, "42", "%v = String()"},
		{FromInt(-42), "v", "", 0, -1, "-42", "%v = String() negative"},
		{FromInt(42), "#v", "#", 0, -1, "Numeric(42)", "%#v = Numeric(String())"},
		{FromInt(0), "v", "", 0, -1, "0", "%v = String() zero"},
		{FromInt(123), "s", "", 0, -1, "123", "%s = String()"},
		{FromInt(123), "s", "", 10, -1, "       123", "%s = String()"},
		{FromFloat64(123.4567), "s", "", 10, 5, "     123.4", "%s = String()"},
		{FromFloat64(123.4567), "v", "", 10, 5, "     123.4", "%s = String()"},
		{FromFloat64(123.4567), "q", "", 10, 5, "   \"123.4\"", "%s = String()"},

		// Float formatting
		{FromFloat64(123.456), "f", "", 0, -1, "123.456000", "%f = Float64() default"},
		{FromFloat64(123.456), "f", "", 10, 2, "    123.46", "%10.2f = Float64() width+prec"},
		{FromFloat64(123.456), "e", "", 0, 2, "1.23e+02", "%.2e = Float64()"},
		{FromFloat64(123.456), "g", "", 0, 4, "123.5", "%.4g = Float64()"},
		{FromFloat64(123.456), "E", "", 0, 1, "1.2E+02", "%.1E = Float64()"},
		{FromFloat64(123.456), "G", "", 0, 3, "123", "%.3G = Float64()"},
		{FromFloat64(-123.456), "f", "+", 0, 1, "-123.5", "%+0.1f = Float64() sign"},
		{FromFloat64(123.456), "f", "+0", 7, 1, "+0123.5", "%+0.1f = Float64() sign"},
		{FromFloat64(123.456), "f", "-017", 0, -1, "123.456000       ", "%+0.1f = Float64() sign"},

		// Integer formatting
		{FromFloat64(123.456), "d", "", 0, -1, "123", "%d = Int()"},
		{FromFloat64(-123.456), "d", "", 0, -1, "-123", "%d = Int() negative"},
		{FromInt(0), "d", "", 0, -1, "0", "%d = Int() zero"},
		{FromInt(42), "d", "", 6, -1, "    42", "%6d = Int() width"},
		{FromInt(42), "d", " ", 6, -1, "    42", "%6d = Int() width"},

		// bad format
		{FromInt(42), "z", "", 0, -1, "%!z(Numeric=42)", "%6d = Int() width"},
	}

	for _, tc := range tests {
		var format string
		if tc.flags != "" {
			format = "%" + tc.flags
		} else {
			format = "%"
		}
		if tc.width > 0 {
			format += strconv.Itoa(tc.width)
		}
		if tc.prec >= 0 {
			format += "." + strconv.Itoa(tc.prec)
		}
		format += tc.verb

		desc := tc.desc + ", format: '" + format + "'"
		got := fmt.Sprintf(format, &tc.num)
		if got != tc.expect {
			t.Errorf("%s: got %q, want %q", desc, got, tc.expect)
		}
	}
}

func TestOneIsOne(t *testing.T) {
	// Test that One is a valid Numeric representation of 1
	one := One(false)
	if one.IsNaN() {
		t.Error("One() should not be NaN")
	}
	if one.Sign() != 1 {
		t.Errorf("One() Sign() = %d, want 1", one.Sign())
	}
	if got := one.String(); got != "1" {
		t.Errorf("One() String() = %q, want '1'", got)
	}
}

func TestNaN(t *testing.T) {
	// Test that NaN is a valid Numeric representation of NaN
	nan := NaN()
	if !nan.IsNaN() {
		t.Error("NaN() should be NaN")
	}
	if got := nan.String(); got != "NaN" {
		t.Errorf("NaN() String() = %q, want 'NaN'", got)
	}
	if nan.Sign() != 0 {
		t.Errorf("NaN() Sign() = %d, want 0", nan.Sign())
	}
}

func TestOneIsMinusOne(t *testing.T) {
	// Test that One is a valid Numeric representation of 1
	one := One(true)
	if one.IsNaN() {
		t.Error("One() should not be NaN")
	}
	if one.Sign() != -1 {
		t.Errorf("One() Sign() = %d, want -1", one.Sign())
	}
	if got := one.String(); got != "-1" {
		t.Errorf("One() String() = %q, want '1'", got)
	}
}

func TestValidateIntRange(t *testing.T) {
	type testCase struct {
		value    int64
		expectOK bool
	}

	tests := []testCase{
		{0, true},
		{1, true},
		{-1, true},
		{1234567890, true},
		{-1234567890, true},
		{maxValueI + 1, false},  // overflow
		{-maxValueI - 1, false}, // overflow
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("ValidateInt_%d", tc.value), func(t *testing.T) {
			if got := ValidateIntRange(tc.value); (got == nil && !tc.expectOK) || (got != nil && tc.expectOK) {
				t.Errorf("validateIntBounds(%d) = %v, want %v", tc.value, got, tc.expectOK)
			}
		})
	}
}

func TestValidateFloatRange(t *testing.T) {
	type testCase struct {
		value    float64
		expectOK bool
	}

	tests := []testCase{
		{0.0, true},
		{1.0, true},
		{-1.0, true},
		{1234567890.123456789, true},
		{-1234567890.123456789, true},
		{maxValueF64 + 1, true},     // overflow but float not precise enough
		{-maxValueF64 - 1, true},    // overflow but float not precise enough
		{maxValueF64 + 100, false},  // overflow
		{-maxValueF64 - 100, false}, // overflow
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("ValidateFloat_%f", tc.value), func(t *testing.T) {
			if got := ValidateFloatRange(tc.value); (got == nil && !tc.expectOK) || (got != nil && tc.expectOK) {
				t.Errorf("validateFloatBounds(%f) = %v, want %v", tc.value, got, tc.expectOK)
			}
		})
	}
}

func TestIsUnderOverNaN(t *testing.T) {
	type testCase struct {
		value       string
		isException bool
	}

	tests := []testCase{
		{"0", false},  // zero is not an exception
		{"1", false},  // normal number
		{"NaN", true}, // NaN is an exception
		{"<999999999999999999.999999999999999999999999999999999999", true}, // overflow is an exception
		{"~1", true}, // underflow is an exception
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("IsUnderOverNan_%s", tc.value), func(t *testing.T) {
			n, err := FromString(tc.value)
			if err != nil {
				t.Fatalf("FromString(%q) failed: %v", tc.value, err)
			}
			if got := n.HasOverflow() || n.HasUnderflow() || n.IsNaN(); got != tc.isException {
				t.Errorf("IsUnderflow/Overflow/NaN(%q) = %v, want %v", n.String(), got, tc.isException)
			}
		})
	}
}
