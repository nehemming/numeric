// Package numeric provides high-precision, fixed-point decimal arithmetic with
// overflow and underflow tracking. Numbers are represented as 54-digit decimals
// with 18 digits before and 36 digits after the decimal point.
//
// Internally, the representation is based on a 192-bit fixed-length format
// (f24), making it highly predictable and zero-allocation in most operations.
//
// numeric.Numeric can be created from strings, integers, or floats,
// and supports arithmetic, comparisons, rounding, JSON marshalling,
// and underflow/overflow detection.
//
// Special formats:
// - "~": indicates an underflow or inexact small number.
// - "<...": indicates a numeric overflow.
// - "NaN": indicates an undefined or invalid value.

package numeric

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

const (
	// RoundTowards rounds toward zero (truncates).
	RoundTowards RoundMode = iota

	// RoundAway rounds away from zero.
	RoundAway

	// RoundHalfDown rounds to nearest, but halves are rounded down.
	RoundHalfDown

	// RoundHalfUp rounds to nearest, but halves are rounded up.
	RoundHalfUp
)

// RoundMode represents rounding behavior for Numeric.Round.
type RoundMode int

// roundModeString maps RoundMode values to human-readable strings.
var roundModeString = map[RoundMode]string{
	RoundTowards:  "towards",
	RoundAway:     "away",
	RoundHalfDown: "1/2 down",
	RoundHalfUp:   "1/2 up",
}

var Zero = Numeric{} // Zero represents the numeric zero value.

// String returns the string name for the RoundMode.
func (rm RoundMode) String() string {
	v, ok := roundModeString[rm]
	if ok {
		return v
	}
	return ""
}

// Numeric represents a fixed-point arbitrary-precision decimal number.
type Numeric struct {
	z f24
}

// FromFloat64 creates a Numeric from a float64.
// NOTE!!: Precision may be lost depending on internal representation.
func FromFloat64(f float64) Numeric {
	return Numeric{z: f24Float64(f)}
}

// FromInt creates a Numeric from an int.
func FromInt(i int) Numeric {
	return Numeric{z: f24Int(int64(i))}
}

// One returns a positive or negative 1
func One(isNeg bool) Numeric {
	var f f24
	f[1] = 1
	f.setNeg(isNeg) // Set the sign based on isNeg
	return Numeric{z: f}
}

// FromString parses a string into a Numeric. Returns an error on invalid format.
func FromString(s string) (Numeric, error) {
	z, err := f24String(s)
	if err != nil {
		return Numeric{}, err
	}
	return Numeric{z: z}, nil
}

// Sum returns the sum of a variadic slice of Numerics.
func Sum(nums ...Numeric) Numeric {
	var sum f24
	for _, n := range nums {
		var z f24
		arith.add(&z, &sum, &n.z)
		sum = z
	}
	return Numeric{z: sum}
}

// Round returns a new Numeric rounded to the specified number of decimal places.
// 'places' is digits after the decimal point. 0 means integer rounding.
// Underflow is removed.
func (n Numeric) Round(places int, mode RoundMode) Numeric {
	var z f24
	arith.round(&z, &n.z, places, mode)
	return Numeric{z: z}
}

// Float64 converts the Numeric to a float64.
// NOTE!!: Precision loss possible; not safe for financial calculations.
func (n Numeric) Float64() float64 {
	d := n.z.Digits()
	return d.Float64()
}

// Int converts the Numeric to an int, discarding any fractional part.
// NOTE!!: Overflows are masked to int range; no error is returned.
func (n Numeric) Int() int {
	if n.z.isNaN() {
		return 0
	}
	v := (uint64(n.z[0].val()) * radix) + uint64(n.z[1].val())
	i := int(v & 0x7FFFFFFFFFFFFFFF)
	if n.z.isNeg() {
		return -i
	}
	return i
}

// String returns the decimal string representation of the number.
// This function allocates to the heap the return string.
func (n Numeric) String() string {
	d := n.z.Digits()
	return d.String()
}

// Add returns the sum of n and n2.
func (n Numeric) Add(n2 Numeric) Numeric {
	var z f24
	arith.add(&z, &n.z, &n2.z)
	return Numeric{z: z}
}

// Sub returns the result of subtracting n2 from n.
func (n Numeric) Sub(n2 Numeric) Numeric {
	var z f24
	arith.sub(&z, &n.z, &n2.z)
	return Numeric{z: z}
}

// Mul returns the product of n and n2.
func (n Numeric) Mul(n2 Numeric) Numeric {
	var z f24
	arith.mul(&z, &n.z, &n2.z)
	return Numeric{z: z}
}

// Div returns the quotient of n divided by n2.
func (n Numeric) Div(n2 Numeric) Numeric {
	var z f24
	arith.div(&z, &n.z, &n2.z)
	return Numeric{z: z}
}

// TruncateTo returns n rounded down to the nearest integer.
func (n Numeric) Truncate(n2 Numeric) Numeric {
	var z f24
	arith.round(&z, &n.z, 0, RoundTowards)
	return Numeric{z: z}
}

// DivRem returns the integer quotient and remainder of n / n2.
func (n Numeric) DivRem(n2 Numeric) (Numeric, Numeric) {
	var r, q f24
	arith.divRem(&q, &r, &n.z, &n2.z)
	return Numeric{z: q}, Numeric{z: r}
}

// Neg returns the negated value of n.
func (n Numeric) Neg() Numeric {
	var z f24
	arith.negate(&z, &n.z)
	return Numeric{z: z}
}

// Abs returns the absolute value of n.
func (n Numeric) Abs() Numeric {
	var z f24
	arith.abs(&z, &n.z)
	return Numeric{z: z}
}

// IsNaN returns true if the value is Not-a-Number.
func (n Numeric) IsNaN() bool {
	return n.z.isNaN()
}

// Sign returns -1 if negative, 1 if zero or positive, 0 for NaN
func (n Numeric) Sign() int {
	switch {
	case n.z.isNaN():
		return 0
	case n.z.isNeg():
		return -1
	default:
		return 1
	}
}

// HasOverflow returns true if the number has overflowed.
func (n Numeric) HasOverflow() bool {
	if n.z.isNaN() {
		return false
	}
	return n.z.isOverflow()
}

// HasUnderflow returns true if the number has underflowed.
func (n Numeric) HasUnderflow() bool {
	if n.z.isNaN() {
		return false
	}
	return n.z.isUnderflow()
}

// IsZero returns true if the number is exactly zero.
func (n Numeric) IsZero() bool {
	if n.z.isNaN() {
		return false
	}
	return n.z.isZero()
}

// IsEqual returns true if n == n2, considering special flags.
func (n Numeric) IsEqual(n2 Numeric) bool {
	return arith.equal(&n.z, &n2.z)
}

// IsLessThan returns true if n < n2.
func (n Numeric) IsLessThan(n2 Numeric) bool {
	return arith.compare(&n.z, &n2.z) < 0
}

// IsLessThanEqual returns true if n <= n2.
func (n Numeric) IsLessThanEqual(n2 Numeric) bool {
	c := arith.compare(&n.z, &n2.z)
	return c < 0 || arith.equal(&n.z, &n2.z)
}

// IsGreaterThan returns true if n > n2.
func (n Numeric) IsGreaterThan(n2 Numeric) bool {
	return arith.compare(&n.z, &n2.z) > 0
}

// IsGreaterEqual returns true if n >= n2.
func (n Numeric) IsGreaterEqual(n2 Numeric) bool {
	c := arith.compare(&n.z, &n2.z)
	return c > 0 || arith.equal(&n.z, &n2.z)
}

// Cmp compares n to n2 and returns:
// -1 if n < n2,
//
//	0 if n == n2,
//	1 if n > n2.
func (n Numeric) Cmp(n2 Numeric) int {
	return arith.compare(&n.z, &n2.z)
}

// MarshalText implements encoding.TextMarshaler for text formats.
func (n Numeric) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for text formats.
func (n *Numeric) UnmarshalText(text []byte) error {
	var s string
	if l := len(text); l != 0 {
		s = unsafe.String(&text[0], l)
	}

	nn, err := FromString(s)
	if err != nil {
		return err
	}
	*n = nn
	return nil
}

// MarshalJSON implements json.Marshaler.
// NaN is serialized as the string "NaN".
func (n Numeric) MarshalJSON() ([]byte, error) {
	if n.IsNaN() {
		return []byte(`"NaN"`), nil
	}
	return []byte(`"` + n.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for Numeric.
// Parses quoted decimal strings. Returns error on invalid input.
func (n *Numeric) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return n.UnmarshalText(data)
	}
	return n.UnmarshalText(data[1 : len(data)-1])
}

// Format implements the fmt.Formatter interface for the Numeric type.
//
// It supports the following format verbs:
//
//	Verb | Description
//	-----|-------------------------------------------------------------
//	  v  | Default format using String(). With '#' flag: Numeric(value)
//	  f  | Decimal format using Float64() (e.g., 123.45)
//	  e  | Scientific notation with 'e' using Float64() (e.g., 1.23e+02)
//	  E  | Scientific notation with 'E' using Float64() (e.g., 1.23E+02)
//	  g  | Compact float format using Float64()
//	  G  | Compact float format (upper case) using Float64()
//	  d  | Integer format using Int() (e.g., 123)
//	  s  | String format using String()
//	  q  | Quoted string format using String() (e.g., "123.45")
//	other| Unsupported verb; output will be: %!<verb>(Numeric=<value>)
//
// The method respects width and precision flags defined by fmt.State.
// For example, %.2f will format to 2 decimal places.
//
// Example usage:
//
//	n := NumericFromString("123.45")
//	fmt.Printf("%v\n", n)    // 123.45
//	fmt.Printf("%#v\n", n)   // Numeric(123.45)
//	fmt.Printf("%.2f\n", n)  // 123.45
//	fmt.Printf("%d\n", n)    // 123
//	fmt.Printf("%q\n", n)    // "123.45"
//	fmt.Printf("%x\n", n)    // %!x(Numeric=123.45)
func (n Numeric) Format(f fmt.State, verb rune) {
	fmtS := buildFormatString(f, verb)
	switch verb {
	case 'v':
		s := n.String()
		if f.Flag('#') {
			fmt.Fprintf(f, "Numeric(%s)", s)
		} else {
			fmt.Fprintf(f, fmtS, s)
		}
	case 'f', 'e', 'E', 'g', 'G':
		v := n.Float64()
		fmt.Fprintf(f, fmtS, v)
	case 'd':
		v := n.Int()
		fmt.Fprintf(f, fmtS, v)
	case 's', 'q':
		s := n.String()
		fmt.Fprintf(f, fmtS, s)
	default:
		s := n.String()
		fmt.Fprintf(f, "%%!%c(Numeric=%s)", verb, s)
	}
}

func buildFormatString(f fmt.State, verb rune) string {
	buf := strings.Builder{}
	buf.WriteRune('%')

	// Handle flags
	if f.Flag('+') {
		buf.WriteRune('+')
	}
	if f.Flag('-') {
		buf.WriteRune('-')
	}
	if f.Flag(' ') {
		buf.WriteRune(' ')
	}
	if f.Flag('#') {
		buf.WriteRune('#')
	}
	if f.Flag('0') {
		buf.WriteRune('0')
	}

	// Handle width
	if w, ok := f.Width(); ok {
		var tmp [32]byte
		bt := strconv.AppendInt(tmp[:0], int64(w), 10)
		buf.Write(bt)
	}

	// Handle precision
	if p, ok := f.Precision(); ok {
		var tmp [32]byte
		bt := strconv.AppendInt(tmp[:0], int64(p), 10)
		buf.WriteRune('.')
		buf.Write(bt)
	}

	buf.WriteRune(verb)

	return buf.String()
}
