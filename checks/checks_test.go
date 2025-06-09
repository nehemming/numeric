package checks

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/nehemming/numeric"
)

func TestStringReps(t *testing.T) {
	tests := []struct {
		val string
	}{
		{"0"},
		{"123"},
		{"-123"},
		{"9999999"},
		{"-9999999"},
		{"1234567.123456789"},
		{"-1234567.123456789"},
		{"0.0000000001"},
		{"-0.0000000001"},
		{"9999999.9999999999"},
		{"-9999999.9999999999"},
		{"1000000.0000000001"},
		{"-1000000.0000000001"},
		{"1"},
		{"-1"},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			dec, err := decimal.NewFromString(tt.val)
			if err != nil {
				t.Fatalf("failed to parse string to decimal: %v", err)
			}

			// lets convert the numbers into numeric format too
			num, err := numeric.FromString(tt.val)
			if err != nil {
				t.Fatalf("failed to parse string to numeric: %v", err)
			}

			str := dec.String()
			if str != tt.val {
				t.Errorf("expected %s, got %s", tt.val, str)
			}

			str = num.String()
			if str != tt.val {
				t.Errorf("expected %s, got %s", tt.val, str)
			}
		})
	}
}

func TestArithmeticOperations(t *testing.T) {
	type pair struct {
		a string
		b string
	}

	tests := []pair{
		{"1", "2"},
		{"9999999", "0.0000000001"},
		{"-1234567.123456789", "7654321.987654321"},
		{"1000000.0000000001", "-1000000.0000000001"},
		{"1.0000000000", "1.0000000000"},
		{"0.0000000001", "-0.0000000001"},
		{"9999999.9999999999", "1"},
	}

	ops := []struct {
		name       string
		do         func(d1, d2 decimal.Decimal) decimal.Decimal
		nDo        func(n1, n2 numeric.Numeric) numeric.Numeric
		skipOnZero bool
	}{
		{
			"add", func(d1, d2 decimal.Decimal) decimal.Decimal { return d1.Add(d2) },
			func(n1, n2 numeric.Numeric) numeric.Numeric { return n1.Add(n2) }, false,
		},
		{
			"sub", func(d1, d2 decimal.Decimal) decimal.Decimal { return d1.Sub(d2) },
			func(n1, n2 numeric.Numeric) numeric.Numeric { return n1.Sub(n2) }, false,
		},
		{
			"mul", func(d1, d2 decimal.Decimal) decimal.Decimal { return d1.Mul(d2) },
			func(n1, n2 numeric.Numeric) numeric.Numeric { return n1.Mul(n2) }, false,
		},
		{
			"div", func(d1, d2 decimal.Decimal) decimal.Decimal { return d1.Div(d2) },
			func(n1, n2 numeric.Numeric) numeric.Numeric { return n1.Div(n2) }, true,
		},
	}

	for _, tt := range tests {
		for _, op := range ops {
			t.Run(tt.a+"_"+op.name+"_"+tt.b, func(t *testing.T) {
				dec1, err := decimal.NewFromString(tt.a)
				if err != nil {
					t.Fatalf("decimal parse error for a: %v", err)
				}
				dec2, err := decimal.NewFromString(tt.b)
				if err != nil {
					t.Fatalf("decimal parse error for b: %v", err)
				}

				num1, err := numeric.FromString(tt.a)
				if err != nil {
					t.Fatalf("numeric parse error for a: %v", err)
				}
				num2, err := numeric.FromString(tt.b)
				if err != nil {
					t.Fatalf("numeric parse error for b: %v", err)
				}

				// Skip div by zero
				if op.skipOnZero && dec2.IsZero() {
					t.Skip("skipping div by zero")
				}

				// Perform operation
				decResult := op.do(dec1, dec2).String()
				numResult := op.nDo(num1, num2).String()

				// Strip leading ~ from numeric string
				numResult = strings.TrimPrefix(numResult, "~")

				// Truncate both to min shared decimal length
				decInt, decFrac := splitParts(decResult)
				numInt, numFrac := splitParts(numResult)

				if decInt != numInt {
					t.Errorf("mismatched integer part in %s: %s vs %s", op.name, decInt, numInt)
					return
				}

				// Compare fractional part up to shortest length
				minFracLen := min(len(decFrac), len(numFrac))
				if decFrac[:minFracLen] != numFrac[:minFracLen] {
					t.Errorf("mismatch in %s fractional part: %s vs %s (up to %d digits)", op.name, decFrac, numFrac, minFracLen)
				}
			})
		}
	}
}

func splitParts(s string) (string, string) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
