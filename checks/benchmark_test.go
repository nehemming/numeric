package checks

import (
	"testing"

	"github.com/nehemming/numeric"
	"github.com/shopspring/decimal"
)

type (
	opFuncDecimal func(a, b decimal.Decimal) decimal.Decimal
	opFuncNumeric func(a, b numeric.Numeric) numeric.Numeric
)

func benchmarkDecimalOp(b *testing.B, name, aStr, bStr string, op opFuncDecimal) {
	aStr = sanitizeInput(aStr)
	bStr = sanitizeInput(bStr)

	a, err := decimal.NewFromString(aStr)
	if err != nil {
		b.Skip("invalid decimal input:", aStr)
	}
	bb, err := decimal.NewFromString(bStr)
	if err != nil {
		b.Skip("invalid decimal input:", bStr)
	}

	if name == "div" && bb.IsZero() {
		b.Skip("division by zero")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = op(a, bb)
	}
}

func benchmarkNumericOp(b *testing.B, name, aStr, bStr string, op opFuncNumeric) {
	aStr = sanitizeInput(aStr)
	bStr = sanitizeInput(bStr)

	a, err := numeric.FromString(aStr)
	if err != nil {
		b.Skip("invalid numeric input:", aStr)
	}
	bb, err := numeric.FromString(bStr)
	if err != nil {
		b.Skip("invalid numeric input:", bStr)
	}

	if name == "div" && bb.Cmp(numeric.Zero) == 0 {
		b.Skip("division by zero")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = op(a, bb)
	}
}

func BenchmarkArithmeticOps(b *testing.B) {
	inputs := []struct {
		a, b string
	}{
		{"1", "2"},
		{"123.456", "-654.321"},
		{"9999999.9999999999", "-9999999.9999999999"},
		{"0.0000000001", "0.0000000002"},
		{"100", "0"}, // to test div zero handling
	}

	decimalOps := map[string]opFuncDecimal{
		"add": decimal.Decimal.Add,
		"sub": decimal.Decimal.Sub,
		"mul": decimal.Decimal.Mul,
		"div": decimal.Decimal.Div,
	}

	numericOps := map[string]opFuncNumeric{
		"add": numeric.Numeric.Add,
		"sub": numeric.Numeric.Sub,
		"mul": numeric.Numeric.Mul,
		"div": numeric.Numeric.Div,
	}

	for _, in := range inputs {
		for name, op := range decimalOps {
			b.Run("Decimal/"+name+"/"+in.a+"_"+in.b, func(b *testing.B) {
				benchmarkDecimalOp(b, name, in.a, in.b, op)
			})
		}
		for name, op := range numericOps {
			b.Run("Numeric/"+name+"/"+in.a+"_"+in.b, func(b *testing.B) {
				benchmarkNumericOp(b, name, in.a, in.b, op)
			})
		}
	}
}
