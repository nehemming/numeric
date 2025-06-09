package numeric

import (
	"fmt"
	"testing"
)

func TestF24AddAndSub(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		wantAddStr string
		wantSubStr string
	}

	tests := []testCase{
		{"0", "0", "0", "0"},
		{"NaN", "0", "NaN", "NaN"},
		{"0", "NaN", "NaN", "NaN"},
		{"1", "1", "2", "0"},
		{"1000000000", "1", "1000000001", "999999999"},
		{"-1", "1", "0", "-2"},
		{"~-1", "1", "~0", "~-2"},
		{"-1", "~1", "~0", "~-2"},
		{"1", "-1", "0", "2"},
		{"123456789.987654321", "0.012345679", "123456790", "123456789.975308642"},
		{"1e9", "1e9", "2000000000", "0"},
		{"999999999999999999", "1", "<999999999999999999.999999999999999999999999999999999999", "999999999999999998"}, // overflow
		{"1e-9", "1e-9", "0.000000002", "0"},
		{"-500", "-500", "-1000", "0"},
		{"-100", "300", "200", "-400"},
		{"1e17", "9e17", "<999999999999999999.999999999999999999999999999999999999", "-800000000000000000"},
		{"1e36", "300", "<999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999"},
		{"1e36", "1e37", "<999999999999999999.999999999999999999999999999999999999", "-<999999999999999999.999999999999999999999999999999999999"},
		{"1e36", "1e36", "<999999999999999999.999999999999999999999999999999999999", "-<999999999999999999.999999999999999999999999999999999999"},
	}

	for _, tc := range tests {
		t.Run(tc.xStr+"+"+tc.yStr, func(t *testing.T) {
			x, errX := f24String(tc.xStr)
			y, errY := f24String(tc.yStr)
			if errX != nil || errY != nil {
				t.Fatalf("Invalid input: %v or %v", errX, errY)
			}

			// Test addition
			var addResult f24
			arith.add(&addResult, &x, &y)
			d := addResult.Digits()
			addStr := d.String()
			if addStr != tc.wantAddStr {
				t.Errorf("add(%q, %q) = %q, want %q", tc.xStr, tc.yStr, addStr, tc.wantAddStr)
			}

			// Test subtraction
			var subResult f24
			arith.sub(&subResult, &x, &y)
			d = subResult.Digits()
			subStr := d.String()
			if subStr != tc.wantSubStr {
				t.Errorf("sub(%q, %q) = %q, want %q", tc.xStr, tc.yStr, subStr, tc.wantSubStr)
			}
		})
	}
}

func TestF24MulAndDiv(t *testing.T) {
	type testCase struct {
		xStr, yStr         string
		wantMulStr         string
		wantDivStr         string
		expectMulOverflow  bool
		expectDivUnderflow bool
	}

	tests := []testCase{
		// Trivial cases
		{"NaN", "0", "NaN", "NaN", false, false},
		{"0", "NaN", "NaN", "NaN", false, false},
		{"0", "0", "0", "NaN", false, false},
		{"1", "0", "0", "NaN", false, false},
		{"0", "1", "0", "0", false, false},
		{"1", "1", "1", "1", false, false},
		{"-1", "-1", "1", "1", false, false},
		{"-1", "1", "-1", "-1", false, false},

		// Fractions
		{"2", "0.5", "1", "4", false, false},
		{"2", "2", "4", "1", false, false},
		{"20", "2", "40", "10", false, false},
		{"10", "20", "200", "0.5", false, false},
		{"0.5", "0.5", "0.25", "1", false, false},
		{"0.5", "0.05", "0.025", "10", false, false},
		{"0.5", "0.005", "0.0025", "100", false, false},
		{"0.05", "0.05", "0.0025", "1", false, false},
		{"1", "0.333333333", "0.333333333", "~3.000000003000000003000000003000000003", false, true},
		{"1", "3", "3", "~0.333333333333333333333333333333333333", false, true},
		{"1", "30", "30", "~0.033333333333333333333333333333333333", false, true},

		// Overflow from mul
		{"999999999", "999999999", "999999998000000001", "1", false, false},
		{"9999999990", "999999999", "<999999999999999999.999999999999999999999999999999999999", "10", true, false},
		{"999999999999999999", "2", "<999999999999999999.999999999999999999999999999999999999", "499999999999999999.5", true, false},

		// Underflow from div
		{"1e-18", "1e9", "0.000000001", "0.000000000000000000000000001", false, false},
		{"1", "1e9", "1000000000", "0.000000001", false, false},
		{"1e-37", "1", "~0", "~0", false, true},

		{"1234.567895", "0.0023", "2.8395061585", "536768.65", false, false},
		{"1", "2e-9", "0.000000002", "500000000", false, false},
		{"1", "2e-18", "0.000000000000000002", "500000000000000000", false, false},
		{"1", "2e-27", "0.000000000000000000000000002", "<999999999999999999.999999999999999999999999999999999999", false, false},
		{"1", "2e-36", "0.000000000000000000000000000000000002", "<999999999999999999.999999999999999999999999999999999999", false, false},
		{"2e17", "2e-36", "0.0000000000000000004", "<999999999999999999.999999999999999999999999999999999999", false, false},
		{"2e-36", "2e-36", "~0", "1", false, false},
		{"2e-20", "2e-36", "~0", "10000000000000000", false, false},
		{"<999999999999999999.999999999999999999999999999999999999", "0", "0", "NaN", false, false},
		{"<999999999999999999.999999999999999999999999999999999999", "1", "<999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999", true, false},
	}

	for _, tc := range tests {
		t.Run(tc.xStr+"*"+tc.yStr, func(t *testing.T) {
			x, errX := f24String(tc.xStr)
			y, errY := f24String(tc.yStr)
			if errX != nil || errY != nil {
				t.Fatalf("Invalid input: %v or %v", errX, errY)
			}

			// --- Test Multiply ---
			var mulRes f24
			arith.mul(&mulRes, &x, &y)
			d := mulRes.Digits()
			mulStr := d.String()
			if got := mulRes.isOverflow(); got != tc.expectMulOverflow {
				t.Errorf("mul(%q, %q): expected overflow=%v, got %v", tc.xStr, tc.yStr, tc.expectMulOverflow, got)
			}
			if mulStr != tc.wantMulStr {
				t.Errorf("mul(%q, %q) = %q, want %q", tc.xStr, tc.yStr, mulStr, tc.wantMulStr)
			}

			// --- Test Divide ---
			var divRes f24
			arith.div(&divRes, &x, &y)
			d = divRes.Digits()
			divStr := d.String()
			if got := divRes.isUnderflow(); got != tc.expectDivUnderflow {
				t.Errorf("div(%q, %q): expected underflow=%v, got %v", tc.xStr, tc.yStr, tc.expectDivUnderflow, got)
			}
			if divStr != tc.wantDivStr {
				t.Errorf("div(%q, %q) = %q, want %q", tc.xStr, tc.yStr, divStr, tc.wantDivStr)
			}
		})
	}
}

func TestDiv(t *testing.T) {
	type testCase struct {
		xStr, yStr         string
		wantDivStr         string
		expectDivUnderflow bool
	}

	tests := []testCase{
		{"7.25", "2.5", "2.9", false},
		{"7", "2", "3.5", false},
		{"12345.6789", "987.654", "~12.500003948751283344167086854303227648", true},
		{"2e-20", "2e-36", "10000000000000000", false},
		{"NaN", "0", "NaN", false},
		{"0", "NaN", "NaN", false},
		{"0", "0", "NaN", false},
		{"1", "0", "NaN", false},
		{"0", "1", "0", false},
		{"1", "1", "1", false},
		{"-1", "-1", "1", false},
		{"-1", "1", "-1", false},

		// Fractions
		{"2", "0.5", "4", false},
		{"0.5", "2", "0.25", false},
		{"1e10", "0.5", "20000000000", false},
		{"2", "2", "1", false},
		{"20", "2", "10", false},
		{"10", "20", "0.5", false},
		{"0.5", "0.5", "1", false},
		{"0.5", "0.05", "10", false},
		{"0.5", "0.005", "100", false},
		{"0.05", "0.05", "1", false},
		{"1", "0.333333333", "~3.000000003000000003000000003000000003", true},
		{"1", "3", "~0.333333333333333333333333333333333333", true},
		{"1", "30", "~0.033333333333333333333333333333333333", true},

		// Underflow cases
		{"1e-18", "1e9", "0.000000000000000000000000001", false},
		{"1", "1e9", "0.000000001", false},
		{"1e-37", "1", "~0", true},

		// // Decimal scale
		{"1234.567895", "0.0023", "536768.65", false},
		{"1", "2e-9", "500000000", false},
		{"1", "2e-18", "500000000000000000", false},
		{"1", "2e-27", "<999999999999999999.999999999999999999999999999999999999", false},
		{"1", "2e-36", "<999999999999999999.999999999999999999999999999999999999", false},
		{"2e17", "2e-36", "<999999999999999999.999999999999999999999999999999999999", false},
		{"2e-36", "2e-36", "1", false},
	}

	for _, tc := range tests {
		t.Run(tc.xStr+"_div_"+tc.yStr, func(t *testing.T) {
			x, errX := f24String(tc.xStr)
			y, errY := f24String(tc.yStr)
			if errX != nil || errY != nil {
				t.Fatalf("Invalid input: %v / %v", errX, errY)
			}

			var divRes f24
			arith.div(&divRes, &x, &y)

			d := divRes.Digits()
			gotStr := d.String()
			gotUF := divRes.isUnderflow()

			if gotUF != tc.expectDivUnderflow {
				t.Errorf("div(%q, %q): underflow = %v, want %v", tc.xStr, tc.yStr, gotUF, tc.expectDivUnderflow)
			}

			if gotStr != tc.wantDivStr {
				t.Errorf("div(%q, %q) = %q, want %q", tc.xStr, tc.yStr, gotStr, tc.wantDivStr)
			}
		})
	}
}

func TestF24Round(t *testing.T) {
	type testCase struct {
		xStr   string
		places int
		mode   RoundMode
		want   string
	}

	tests := []testCase{
		{"123.456789", 0, RoundTowards, "123"},
		{"123.456789", 0, RoundAway, "124"},
		{"123.456789", 0, RoundHalfUp, "123"},
		{"123.556789", 0, RoundHalfUp, "124"},
		{"123.999999999", 0, RoundHalfUp, "124"},

		{"123.000001", 5, RoundTowards, "123"},
		{"123.000001", 5, RoundAway, "123.00001"},
		{"123.000005", 5, RoundHalfUp, "123.00001"},
		{"123.000005", 5, RoundHalfDown, "123"},
		{"123.0000055", 5, RoundHalfDown, "123.00001"},

		{"999999999.999999999", -1, RoundAway, "NaN"},
		{"0.0000000001", 9, RoundTowards, "0"},
		{"NaN", 0, RoundHalfUp, "NaN"},
		{"<999999999999999999.999999999999999999999999999999999999", 0, RoundHalfUp, "<999999999999999999.999999999999999999999999999999999999"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("round(%s, %d, %v)", tc.xStr, tc.places, tc.mode), func(t *testing.T) {
			x, err := f24String(tc.xStr)
			if err != nil {
				t.Fatalf("Invalid input: %v", err)
			}

			var z f24
			arith.round(&z, &x, tc.places, tc.mode)
			d := z.Digits()
			got := d.String()
			if got != tc.want {
				t.Errorf("round(%q, %d, %v) = %q, want %q", tc.xStr, tc.places, tc.mode, got, tc.want)
			}
		})
	}
}

func TestF24Quanta(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		mode       RoundMode
		want       string
	}

	tests := []testCase{
		{"123.456", "0.01", RoundTowards, "123.45"},
		{"123.456", "0.01", RoundAway, "123.46"},
		{"123.456", "0.01", RoundHalfUp, "123.46"},
		{"123.444", "0.01", RoundHalfUp, "123.44"},
		{"123", "1", RoundTowards, "123"},
		{"123", "1", RoundAway, "123"},
		{"123", "1", RoundHalfUp, "123"},
		{"1.49", "1", RoundHalfUp, "1"},
		{"1.50", "1", RoundHalfUp, "2"},
		{"-1.5", "1", RoundHalfUp, "-2"},
		{"NaN", "1", RoundHalfUp, "NaN"},
		{"1", "0", RoundHalfUp, "NaN"}, // div by zero case
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("quanta(%s, %s, %v)", tc.xStr, tc.yStr, tc.mode), func(t *testing.T) {
			x, err1 := f24String(tc.xStr)
			y, err2 := f24String(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("Invalid input: %v or %v", err1, err2)
			}

			var z f24
			arith.quanta(&z, &x, &y, tc.mode)
			d := z.Digits()
			got := d.String()
			if got != tc.want {
				t.Errorf("quanta(%q, %q, %v) = %q, want %q", tc.xStr, tc.yStr, tc.mode, got, tc.want)
			}
		})
	}
}

func TestF24Compare(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		want       int
	}

	tests := []testCase{
		{"0", "0", 0},
		{"1", "0", 1},
		{"0", "1", -1},
		{"-1", "-1", 0},
		{"-1", "1", -1},
		{"1", "-1", 1},
		{"123.456", "123.456", 0},
		{"123.456", "123.457", -1},
		{"123.457", "123.456", 1},
		{"1e-30", "0", 1},
		{"0", "1e-30", -1},
		{"NaN", "1", -1},
		{"1", "NaN", 1},
		{"NaN", "NaN", -1},
		// Underflow handling
		{"1e-30", "1e-30", 0},
		{"1e-30", "1e-29", -1},
		{"~1", "1", 1},
		{"~1", "~1", 1},
		{"1", "~1", -1},
		{"~-1", "-1", -1},
		{"~-1", "~-1", -1},
		{"-1", "~-1", 1},
		// Overflow handling
		{"<999999999999999999.999999999999999999999999999999999999", "1", -1},
		{"1", "<999999999999999999.999999999999999999999999999999999999", 1},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("compare(%s,%s)", tc.xStr, tc.yStr), func(t *testing.T) {
			x, err1 := f24String(tc.xStr)
			y, err2 := f24String(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("Invalid input: %v or %v", err1, err2)
			}
			got := arith.compare(&x, &y)
			if got != tc.want {
				t.Errorf("compare(%q, %q) = %d, want %d", tc.xStr, tc.yStr, got, tc.want)
			}
		})
	}
}

func TestF24Equal(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		want       bool
	}

	tests := []testCase{
		{"0", "0", true},
		{"1", "1", true},
		{"-1", "-1", true},
		{"1", "-1", false},
		{"123.456", "123.456", true},
		{"123.456", "123.457", false},
		{"NaN", "NaN", false},
		{"NaN", "1", false},
		{"1", "NaN", false},
		// With underflow
		{"1e-30", "1e-30", true},
		{"1e-38", "1e-38", false},
		{"1e-37", "1", false},
		// With overflow
		{"<999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999", false},
		{"<999999999999999999.999999999999999999999999999999999999", "1", false},
		{"1", "<999999999999999999.999999999999999999999999999999999999", false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("equal(%s,%s)", tc.xStr, tc.yStr), func(t *testing.T) {
			x, err1 := f24String(tc.xStr)
			y, err2 := f24String(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("Invalid input: %v or %v", err1, err2)
			}
			got := arith.equal(&x, &y)
			if got != tc.want {
				t.Errorf("equal(%q, %q) = %v, want %v", tc.xStr, tc.yStr, got, tc.want)
			}
		})
	}
}

func TestF24NegateAndAbs(t *testing.T) {
	type testCase struct {
		xStr        string
		wantNegate  string
		wantAbs     string
		expectNeg   bool
		expectAbs   bool
		expectNaN   bool
		expectOflow bool
		expectUflow bool
	}

	tests := []testCase{
		{"0", "0", "0", false, false, false, false, false},
		{"1", "-1", "1", true, false, false, false, false},
		{"-1", "1", "1", false, false, false, false, false},
		{"123.456", "-123.456", "123.456", true, false, false, false, false},
		{"-999.999", "999.999", "999.999", false, false, false, false, false},
		{"NaN", "NaN", "NaN", false, false, true, false, false},
		{"<999999999999999999.999999999999999999999999999999999999", "-<999999999999999999.999999999999999999999999999999999999", "<999999999999999999.999999999999999999999999999999999999", true, false, false, true, false},
		{"1e-38", "~-0", "~0", true, false, false, false, true},
		{"-1e-38", "~0", "~0", false, false, false, false, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("negate/abs(%s)", tc.xStr), func(t *testing.T) {
			x, err := f24String(tc.xStr)
			if err != nil {
				t.Fatalf("Invalid input: %v", err)
			}

			var gotNeg f24
			arith.negate(&gotNeg, &x)
			d := gotNeg.Digits()
			negStr := d.String()

			if negStr != tc.wantNegate {
				t.Errorf("negate(%q) = %q, want %q", tc.xStr, negStr, tc.wantNegate)
			}
			if gotNeg.isNeg() != tc.expectNeg {
				t.Errorf("negate(%q) sign = %v, want %v", tc.xStr, gotNeg.isNeg(), tc.expectNeg)
			}
			if gotNeg.isNaN() != tc.expectNaN {
				t.Errorf("negate(%q) NaN = %v, want %v", tc.xStr, gotNeg.isNaN(), tc.expectNaN)
			}
			if gotNeg.isOverflow() != tc.expectOflow {
				t.Errorf("negate(%q) overflow = %v, want %v", tc.xStr, gotNeg.isOverflow(), tc.expectOflow)
			}
			if gotNeg.isUnderflow() != tc.expectUflow {
				t.Errorf("negate(%q) underflow = %v, want %v", tc.xStr, gotNeg.isUnderflow(), tc.expectUflow)
			}

			var gotAbs f24
			arith.abs(&gotAbs, &x)
			d = gotAbs.Digits()
			absStr := d.String()

			if absStr != tc.wantAbs {
				t.Errorf("abs(%q) = %q, want %q", tc.xStr, absStr, tc.wantAbs)
			}
			if gotAbs.isNeg() != tc.expectAbs {
				t.Errorf("abs(%q) sign = %v, want %v", tc.xStr, gotAbs.isNeg(), tc.expectAbs)
			}
			if gotAbs.isNaN() != tc.expectNaN {
				t.Errorf("abs(%q) NaN = %v, want %v", tc.xStr, gotAbs.isNaN(), tc.expectNaN)
			}
			if gotAbs.isOverflow() != tc.expectOflow {
				t.Errorf("abs(%q) overflow = %v, want %v", tc.xStr, gotAbs.isOverflow(), tc.expectOflow)
			}
			if gotAbs.isUnderflow() != tc.expectUflow {
				t.Errorf("abs(%q) underflow = %v, want %v", tc.xStr, gotAbs.isUnderflow(), tc.expectUflow)
			}
		})
	}
}

func TestF24DivRem(t *testing.T) {
	type testCase struct {
		xStr, yStr string
		wantQ      string
		wantR      string
	}

	tests := []testCase{
		{"10", "3", "3", "1"},
		{"123.456", "1", "123", "0.456"},
		{"123.456", "0.5", "246", "0.456"},
		{"0", "7", "0", "0"},
		{"7", "1", "7", "0"},
		{"1", "1", "1", "0"},
		{"5", "10", "0", "5"},
		{"-5", "10", "0", "-5"},
		{"5", "-10", "0", "5"},
		{"-5", "-10", "0", "-5"},
		{"-15", "-10", "1", "-5"},
		{"1", "0", "NaN", "NaN"},   // division by zero
		{"NaN", "1", "NaN", "NaN"}, // NaN input
		{"1", "NaN", "NaN", "NaN"}, // NaN input
		{"<999999999999999999.999999999999999999999999999999999999", "1", "NaN", "NaN"}, // overflow case
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("divRem(%s / %s)", tc.xStr, tc.yStr), func(t *testing.T) {
			x, err1 := f24String(tc.xStr)
			y, err2 := f24String(tc.yStr)
			if err1 != nil || err2 != nil {
				t.Fatalf("Invalid input: %v or %v", err1, err2)
			}

			var q, r f24
			arith.divRem(&q, &r, &x, &y)

			dq := q.Digits()
			dr := r.Digits()

			gotQ := dq.String()
			gotR := dr.String()

			if gotQ != tc.wantQ {
				t.Errorf("quotient = %q, want %q", gotQ, tc.wantQ)
			}
			if gotR != tc.wantR {
				t.Errorf("remainder = %q, want %q", gotR, tc.wantR)
			}
		})
	}
}

func TestShouldBeNeg(t *testing.T) {
	type testCase struct {
		xStr   string
		isNeg  bool
		expect bool
	}

	tests := []testCase{
		// NaN → always false
		{"NaN", true, false},
		{"NaN", false, false},

		// Zero, not underflow → always false
		{"0", true, false},
		{"0", false, false},

		// Zero, underflow → should follow isNeg
		{"~0", true, true},   // simulated underflowed zero
		{"~0", false, false}, // simulated underflowed zero

		// Non-zero → just return isNeg
		{"1", true, true},
		{"1", false, false},
		{"-1", true, true},
		{"-1", false, false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("shouldBeNeg(%s, %v)", tc.xStr, tc.isNeg), func(t *testing.T) {
			x, err := f24String(tc.xStr)
			if err != nil {
				t.Fatalf("Invalid input: %v", err)
			}
			got := shouldBeNeg(&x, tc.isNeg)
			if got != tc.expect {
				t.Errorf("shouldBeNeg(%q, %v) = %v, want %v", tc.xStr, tc.isNeg, got, tc.expect)
			}
		})
	}
}
