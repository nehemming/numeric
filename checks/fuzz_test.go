package checks

import (
	"strings"
	"testing"

	"github.com/nehemming/numeric"
	"github.com/shopspring/decimal"
)

func FuzzArithmeticConsistency(f *testing.F) {
	// Seed corpus with a few sample values
	seed := []string{
		"1", "-1", "0.0000000001", "-0.0000000001", "9999999.9999999999", "-9999999.9999999999",
	}
	for _, a := range seed {
		for _, b := range seed {
			f.Add(a, b)
		}
	}

	f.Fuzz(func(t *testing.T, aStr string, bStr string) {
		aStr = sanitizeInput(aStr)
		bStr = sanitizeInput(bStr)

		dec1, err := decimal.NewFromString(aStr)
		if err != nil {
			t.Skipf("invalid decimal input A: %q", aStr)
		}
		dec2, err := decimal.NewFromString(bStr)
		if err != nil {
			t.Skipf("invalid decimal input B: %q", bStr)
		}

		num1, err := numeric.FromString(aStr)
		if err != nil {
			t.Skipf("invalid numeric input A: %q", aStr)
		}
		num2, err := numeric.FromString(bStr)
		if err != nil {
			t.Skipf("invalid numeric input B: %q", bStr)
		}

		operations := []struct {
			name       string
			decimalOp  func(a, b decimal.Decimal) decimal.Decimal
			numericOp  func(a, b numeric.Numeric) numeric.Numeric
			skipIfZero bool
		}{
			{"add", decimal.Decimal.Add, numeric.Numeric.Add, false},
			{"sub", decimal.Decimal.Sub, numeric.Numeric.Sub, false},
			{"mul", decimal.Decimal.Mul, numeric.Numeric.Mul, false},
			{"div", decimal.Decimal.Div, numeric.Numeric.Div, true},
		}

		for _, op := range operations {
			if op.skipIfZero && dec2.IsZero() {
				continue
			}
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic in %s with inputs %q and %q: %v", op.name, aStr, bStr, r)
				}
			}()

			decRes := op.decimalOp(dec1, dec2).String()
			numRes := op.numericOp(num1, num2).String()
			hadPrefix := strings.HasPrefix(numRes, "~")
			numRes = strings.TrimPrefix(numRes, "~")

			decInt, decFrac := splitParts(decRes)
			numInt, numFrac := splitParts(numRes)

			if decInt != numInt && (!hadPrefix || numInt != "-"+decInt) {
				t.Errorf("op %s: int part mismatch: %s vs %s", op.name, decInt, numInt)
			} else {
				minFrac := min(len(decFrac), len(numFrac))
				if decFrac[:minFrac] != numFrac[:minFrac] {
					t.Errorf("op %s: frac part mismatch: dec %s vs num %s (first %d digits)", op.name, decFrac, numFrac, minFrac)
				}
			}

			if t.Failed() {
				t.Logf("inputs: A=%s, B=%s, operation=%s\ndec=%s\nnum=%s", aStr, bStr, op.name, decRes, numRes)
			}
		}
	})
}
