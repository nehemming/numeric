package numeric

import (
	"fmt"
	"testing"
)

/*
go test -bench=BenchmarkMarshalJSON -cpuprofile=cpu.prof -memprofile=mem.prof -benchmem  github.com/nehemming/numeric=
*/

var (
	a, _ = FromString("12345.6789")
	b, _ = FromString("987.654")

	c, _ = FromString("12345.6")
	d, _ = FromString("1.2")
)

func BenchmarkFromString(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_, _ = FromString("12345.6789")
	}
}

func BenchmarkAdd(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Add(b)
	}
}

func BenchmarkSub(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Sub(b)
	}
}

func BenchmarkMul(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Mul(b)
	}
}

func BenchmarkDivSimpler(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = c.Div(d)
	}
}

func BenchmarkDiv(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Div(b)
	}
}

func BenchmarkDivRem(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_, _ = a.DivRem(b)
	}
}

func BenchmarkRound(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Round(4, RoundHalfUp)
	}
}

func BenchmarkAbs(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Abs()
	}
}

func BenchmarkNeg(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Neg()
	}
}

func BenchmarkIsZero(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.IsZero()
	}
}

func BenchmarkIsEqual(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.IsEqual(b)
	}
}

func BenchmarkCmp(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Cmp(b)
	}
}

func BenchmarkFloat64(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Float64()
	}
}

func BenchmarkInt(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.Int()
	}
}

func BenchmarkString(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_ = a.String()
	}
}

func BenchmarkMarshalText(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_, _ = a.MarshalText()
	}
}

func BenchmarkUnmarshalText(bm *testing.B) {
	text, _ := a.MarshalText()
	var x Numeric
	bm.ResetTimer()
	for i := 0; i < bm.N; i++ {
		_ = x.UnmarshalText(text)
	}
}

func BenchmarkMarshalJSON(bm *testing.B) {
	for i := 0; i < bm.N; i++ {
		_, _ = a.MarshalJSON()
	}
}

func BenchmarkUnmarshalJSON(bm *testing.B) {
	jsonBytes, _ := a.MarshalJSON()
	var x Numeric
	bm.ResetTimer()
	for i := 0; i < bm.N; i++ {
		_ = x.UnmarshalJSON(jsonBytes)
	}
}

/*
go test -bench=BenchmarkFormat -cpuprofile=cpu.prof -memprofile=mem.prof -benchmem  github.com/nehemming/numeric
*/
func BenchmarkFormat(bm *testing.B) {
	var buf [128]byte
	for i := 0; i < bm.N; i++ {
		fmt.Appendf(buf[:0], "%s", a)
	}
}
