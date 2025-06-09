package numeric

import (
	"fmt"
)

func ExampleFromFloat64() {
	n := FromFloat64(3.14159)
	fmt.Println(n.Round(3, RoundHalfUp).String())
	// Output: 3.142
}

func ExampleFromInt() {
	n := FromInt(42)
	fmt.Println(n.String())
	// Output: 42
}

func ExampleFromString() {
	n, err := FromString("123.456")
	if err != nil {
		panic(err)
	}
	fmt.Println(n.String())
	// Output: 123.456
}

func ExampleNumeric_Add() {
	x, _ := FromString("1.5")
	y, _ := FromString("2.25")
	fmt.Println(x.Add(y).String())
	// Output: 3.75
}

func ExampleNumeric_Sub() {
	x, _ := FromString("5")
	y, _ := FromString("3.2")
	fmt.Println(x.Sub(y).String())
	// Output: 1.8
}

func ExampleNumeric_Mul() {
	x, _ := FromString("2")
	y, _ := FromString("3.5")
	fmt.Println(x.Mul(y).String())
	// Output: 7
}

func ExampleNumeric_Div() {
	x, _ := FromString("7")
	y, _ := FromString("2")
	fmt.Println(x.Div(y).Round(3, RoundHalfUp).String())
	// Output: 3.5
}

func ExampleNumeric_DivRem() {
	x, _ := FromString("10")
	y, _ := FromString("3")
	q, r := x.DivRem(y)
	fmt.Printf("q=%v r=%v\n", q.String(), r.String())
	// Output: q=3 r=1
}

func ExampleNumeric_Neg() {
	n, _ := FromString("42")
	fmt.Println(n.Neg().String())
	// Output: -42
}

func ExampleNumeric_Abs() {
	n, _ := FromString("-3.14")
	fmt.Println(n.Abs().String())
	// Output: 3.14
}

func ExampleNumeric_IsNaN() {
	n, _ := FromString("NaN")
	fmt.Println(n.IsNaN())
	// Output: true
}

func ExampleNumeric_Sign() {
	n1, _ := FromString("-1")
	n2, _ := FromString("0")
	n3, _ := FromString("3.14")
	fmt.Println(n1.Sign(), n2.Sign(), n3.Sign())
	// Output: -1 1 1
}

func ExampleNumeric_HasOverflow() {
	n, _ := FromString("1999999999999999999.999999999999999999999999999999999999")
	fmt.Println(n.HasOverflow())
	// Output: true
}

func ExampleNumeric_HasUnderflow() {
	n, _ := FromString("0.0000000000000000000000000000000000001")
	fmt.Println(n.HasUnderflow())
	// Output: true
}

func ExampleNumeric_IsZero() {
	n1, _ := FromString("0")
	n2, _ := FromString("0.00000000000000000000000000000000001")
	fmt.Println(n1.IsZero(), n2.IsZero())
	// Output: true false
}

func ExampleNumeric_IsEqual() {
	x, _ := FromString("3.00")
	y, _ := FromString("3")
	fmt.Println(x.IsEqual(y))
	// Output: true
}

func ExampleNumeric_IsLessThan() {
	x, _ := FromString("2.5")
	y, _ := FromString("3")
	fmt.Println(x.IsLessThan(y))
	// Output: true
}

func ExampleNumeric_IsLessThanEqual() {
	x, _ := FromString("2.5")
	y, _ := FromString("2.5")
	fmt.Println(x.IsLessThanEqual(y))
	// Output: true
}

func ExampleNumeric_IsGreaterThan() {
	x, _ := FromString("5")
	y, _ := FromString("2")
	fmt.Println(x.IsGreaterThan(y))
	// Output: true
}

func ExampleNumeric_IsGreaterEqual() {
	x, _ := FromString("5")
	y, _ := FromString("5")
	fmt.Println(x.IsGreaterEqual(y))
	// Output: true
}

func ExampleNumeric_Cmp() {
	x, _ := FromString("2")
	y, _ := FromString("3")
	fmt.Println(x.Cmp(y)) // -1 if x < y
	// Output: -1
}

func ExampleNumeric_MarshalText() {
	n, _ := FromString("1.23")
	text, _ := n.MarshalText()
	fmt.Println(string(text))
	// Output: 1.23
}

func ExampleNumeric_UnmarshalText() {
	var n Numeric
	err := n.UnmarshalText([]byte("456.78"))
	if err != nil {
		panic(err)
	}
	fmt.Println(n.String())
	// Output: 456.78
}

func ExampleNumeric_MarshalJSON() {
	n, _ := FromString("42.42")
	json, _ := n.MarshalJSON()
	fmt.Println(string(json))
	// Output: "42.42"
}

func ExampleNumeric_UnmarshalJSON() {
	var n Numeric
	err := n.UnmarshalJSON([]byte(`"789.01"`))
	if err != nil {
		panic(err)
	}
	fmt.Println(n.String())
	// Output: 789.01
}
