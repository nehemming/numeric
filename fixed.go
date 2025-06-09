package numeric

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

// const below are used to  provide consistent values
// The constants should not be changed.  Many function use unrolled logic and
// the values are essentially bound to the constants.

const (
	radix            = uint64(1e9) // radix is base 1e9
	radixI           = int64(1e9)
	radixHalfI       = int64(5e8)                 // radixHalfI is used for normalization pivoting
	maxDigit         = 1e9 - 1                    // maxDigit largest single digit in the radix base
	maxUnit          = fVal(maxDigit)             // maxUnit in fVal format
	precision        = 54                         // precision is 18.36 => 54.
	maxWholeDigits   = 18                         // maxWholeDigits is the maximum number of whole digits allowed before overflow
	maxDecimals      = precision - maxWholeDigits // maxDecimals is the maximum number of decimal digits allowed
	radixDigits      = 9                          // base 10 digits in 1 radix unit.
	maxDecimalPlaces = 4 * radixDigits            // maxDecimalPlaces number of decimal digits.
	decMul           = uint64(1e8)                // decMul used in decimal conversions from strings.
	maxValue         = uint64(1e18) - 1           // maxValue is the maximum integer value value that can be represented in f24
	lowIndex         = len(f24{}) - 1             // lowIndex is the index of the lowest fVal in f24
	lenF24           = len(f24{})                 // lenF24 is the length of f24, which is 6
	decIndex         = 2                          // decIndex is the index of the first decimal place in f24
)

type (
	f24 [6]fVal

	digits struct {
		v           [precision]uint8 // v is an array type to avoid allocations.
		pointIdx    int              // pointIdx index of decimal point (before).
		count       int              // count is number of digits.
		isNeg       bool             // isNeg indicates if the number is negative.
		isNaN       bool             // isNaN indicates if the number is NaN (not a number).
		isOverflow  bool             // isOverflow indicates if the number is too large to represent.
		isUnderflow bool             // isUnderflow indicates if the number is too small to represent.
	}
)

var (
	// ErrParseFormatNumeric indicates a general numeric format error.
	ErrParseFormatNumeric = errors.New("invalid numeric format")

	// ErrMultipleUnderflowSymbols is returned when multiple '~' underflow markers are present in the input.
	ErrMultipleUnderflowSymbols = errors.New("multiple underflow symbols (~) not allowed")

	// ErrMultipleOverflowSymbols is returned when multiple '<' overflow markers are present in the input.
	ErrMultipleOverflowSymbols = errors.New("multiple overflow symbols (<) not allowed")

	// ErrMultipleMinusSigns is returned when multiple '-' signs are found in the prefix, indicating conflicting negative flags.
	ErrMultipleMinusSigns = errors.New("multiple sign symbols (-) not allowed")

	// ErrMultiplePlusSigns is returned when multiple '+' signs are found in the prefix, indicating conflicting positive flags.
	ErrMultiplePlusSigns = errors.New("multiple sign symbols (+) not allowed")

	// ErrInvalidDecimalPoint is returned when an input contains more than one decimal point or uses it incorrectly with an exponent.
	ErrInvalidDecimalPoint = errors.New("invalid decimal points")

	// ErrMultipleExponents is returned when more than one exponent ('e' or 'E') is encountered in the input string.
	ErrMultipleExponents = errors.New("multiple exponents")

	// ErrMultipleExponentSigns is returned when multiple '+' or '-' signs are used in the exponent portion of the input.
	ErrMultipleExponentSigns = errors.New("multiple exponent signs")

	// ErrNoExponentValue is returned when an exponent character is present but not followed by a valid numeric value.
	ErrNoExponentValue = errors.New("no exponent value")

	// ErrNoDigitsInInput is returned when no numeric digits are found in the input.
	ErrNoDigitsInInput = errors.New("no digits in input")

	// ErrInvalidCharacter is returned when an invalid character is encountered in the input string.
	ErrInvalidCharacter = errors.New("invalid character")
)

var maxF24 = f24{
	maxUnit, maxUnit, maxUnit,
	maxUnit, maxUnit, maxUnit,
}

// f24Int creates a f24 from a integer value.
// not values > 1e18-1 will overflow.
func f24Int(v int64) f24 {
	var f f24
	var u uint64
	isNeg := v < 0
	if isNeg {
		f.setNeg(true)
		u = uint64(-v)
	} else {
		u = uint64(v)
	}

	if u >= maxValue {
		return overflow(isNeg)
	}
	f[0].setVal(uint32(u / radix))
	f[1].setVal(uint32(u % radix))

	return f
}

// f24Float64 creates a f24 from a float64 value.
// It handles special cases like NaN, Infinity, and zero.
// It uses strconv to convert the float to a string and then parses it.
// The function returns a f24 representation of the float64 value.
func f24Float64(v float64) f24 {
	var f f24
	switch {
	case math.IsNaN(v):
		f.setNaN(true)
	case math.IsInf(v, 1):
		f = overflow(false)
	case math.IsInf(v, -1):
		f = overflow(true)
	case v == 0:
	default:
		var buf [24]byte
		b := strconv.AppendFloat(buf[:0], v, 'g', -1, 64)
		s := unsafe.String(&b[0], len(b))
		d, err := parseString(s)
		if err != nil { // this really should not be possible as format is valid.
			f.setNaN(true)
		} else {
			f = d.F24()
		}
	}
	return f
}

// f24String creates a f24 from a string value.
// It parses the string to extract numeric values, handling special cases like NaN and overflow.
// If the string is not a valid numeric format, it returns a f24 with NaN set.
// Both number and exponent formats are parsed, allowing for scientific notation.
func f24String(v string) (f24, error) {
	d, err := parseString(v)
	if err != nil {
		var f f24
		f.setNaN(true)
		return f, err
	}
	return d.F24(), nil
}

// Digits converts a f24 to a digits structure.
// Digits makes it easier to convert and output base 10 values.
func (f *f24) Digits() digits {
	var d digits
	d.isNeg = f.isNeg()
	d.isNaN = f.isNaN()
	d.isOverflow = f.isOverflow()
	d.isUnderflow = f.isUnderflow()

	if d.isNaN {
		return d
	}

	pos := 0
	var lead bool
	for i := 0; i < 2; i++ {
		v := uint64(f[i].val())
		if !lead && v == 0 {
			continue
		}

		for mul := decMul; mul > 0; mul /= 10 {
			u := v / mul
			v = v % mul
			if lead || u != 0 {
				d.v[pos] = uint8(u)
				pos++
				lead = true
			}
		}
	}

	d.pointIdx = pos
	d.count = d.pointIdx

	// now handle decimal places
	for i := decIndex; i < lenF24; i++ {
		v := uint64(f[i].val())
		if v == 0 {
			pos += radixDigits
			continue
		}

		for mul := decMul; mul > 0; mul /= 10 {
			u := v / mul
			v = v % mul
			d.v[pos] = uint8(u)
			pos++
			if u != 0 {
				d.count = pos
			}
		}
	}

	return d
}

func (f *f24) isZero() bool {
	if f[0].val() != 0 {
		return false
	}
	if f[1].val() != 0 {
		return false
	}
	if f[2].val() != 0 {
		return false
	}
	if f[3].val() != 0 {
		return false
	}
	if f[4].val() != 0 {
		return false
	}
	if f[5].val() != 0 {
		return false
	}
	return true
}

// F24 converts digits to a f24 representation.
func (d *digits) F24() f24 {
	var f f24
	if d.isNaN {
		f.setNeg(d.isNeg)
		f.setNaN(true)
		return f
	}

	if d.isOverflow || d.pointIdx > maxWholeDigits {
		f = overflow(d.isNeg)
		f.setUnderflow(d.isUnderflow)
		return f
	} else {
		// handle whole integer part
		var val uint64
		for _, v := range d.v[:d.pointIdx] {
			val = val*10 + uint64(v)
		}
		f[0].setVal(uint32(val / radix))
		f[1].setVal(uint32(val % radix))
	}

	underflow := d.isUnderflow
	dp := d.count - d.pointIdx
	if dp > maxDecimalPlaces {
		underflow = true
		dp = maxDecimalPlaces
	}

	// handle decimals
	start := d.pointIdx
	for p := decIndex; p < lenF24 && dp > 0; p++ {
		end := start + min(dp, radixDigits)
		dp -= radixDigits
		var val uint64
		mul := decMul
		for _, v := range d.v[start:end] {
			val += uint64(v) * mul
			mul /= 10
		}
		f[p].setVal(uint32(val % radix))
		start = end
	}

	if d.isNeg {
		f.setNeg(true)
	}
	if underflow {
		f.setUnderflow(true)
	}

	return f
}

// Float64 converts digits to a float64 representation.
// It handles special cases like NaN and overflow.
// Output may not be an exact representation.
func (d *digits) Float64() float64 {
	switch {
	case d.isNaN:
		return math.NaN()
	case d.isOverflow:
		var sign int
		if d.isNeg {
			sign = -1
		} else {
			sign = 1
		}
		return math.Inf(sign)

	default:
		var buf [56]byte
		var pos, sd int
		if d.isNeg {
			buf[pos] = '-'
			pos++
		}
		if d.pointIdx == 0 {
			buf[pos] = '0'
			pos++
		} else {
			for _, v := range d.v[:d.pointIdx] {
				if sd < 18 {
					buf[pos] = '0' + v
				} else {
					buf[pos] = '0'
				}
				pos++
				sd++
			}
		}

		if d.count-d.pointIdx > 0 && sd < 18 {
			var dotted bool
			var zeros int
			for _, v := range d.v[d.pointIdx:d.count] {
				if sd >= 18 {
					break
				}
				sd++
				if v == 0 {
					zeros++
					continue
				}
				if !dotted {
					dotted = true
					buf[pos] = '.'
					pos++
				}
				for range zeros {
					buf[pos] = '0'
					pos++
				}
				zeros = 0
				buf[pos] = '0' + v
				pos++
			}
		}

		f, err := strconv.ParseFloat(unsafe.String(&buf[0], pos), 64)
		if err != nil {
			return math.NaN()
		}

		return f
	}
}

// output formats the digits into a byte slice.
func (d *digits) output(buf []byte) []byte {
	b := bytes.NewBuffer(buf) // Wrap existing slice; reuse memory

	if d.isUnderflow {
		b.WriteByte('~')
	}
	if d.isNeg {
		b.WriteByte('-')
	}
	if d.isOverflow {
		b.WriteByte('<')
	}
	if d.isNaN {
		b.WriteString("NaN")
		return b.Bytes()
	}
	if d.count == 0 {
		b.WriteByte('0')
		return b.Bytes()
	}
	if d.pointIdx == 0 {
		b.WriteByte('0')
	} else {
		for _, v := range d.v[:d.pointIdx] {
			b.WriteByte('0' + v)
		}
	}

	if d.count-d.pointIdx > 0 {
		var dotted bool
		var dp int
		var zeros int
		for _, v := range d.v[d.pointIdx:d.count] {
			if dp == maxDecimalPlaces {
				break
			}
			dp++
			if v == 0 {
				zeros++
				continue
			}
			if !dotted {
				b.WriteByte('.')
				dotted = true
			}
			for i := 0; i < zeros; i++ {
				b.WriteByte('0')
			}
			zeros = 0
			b.WriteByte('0' + v)
		}
	}
	return b.Bytes()
}

// String formats the digits into a string representation.
// This function allocates the result to the heap.
func (d *digits) String() string {
	var sb strings.Builder

	if d.isUnderflow {
		sb.WriteRune('~')
	}
	if d.isNeg {
		sb.WriteRune('-')
	}
	if d.isOverflow {
		sb.WriteRune('<')
	}
	if d.isNaN {
		return "NaN"
	}
	if d.count == 0 {
		sb.WriteRune('0')
		return sb.String()
	}
	if d.pointIdx == 0 {
		sb.WriteRune('0')
	} else {
		for _, v := range d.v[:d.pointIdx] {
			sb.WriteByte('0' + v)
		}
	}

	if d.count-d.pointIdx > 0 {
		var dotted bool
		var dp int
		var zeros int
		for _, v := range d.v[d.pointIdx:d.count] {
			if dp == maxDecimalPlaces {
				break
			}
			dp++
			if v == 0 {
				zeros++
				continue
			}
			if !dotted {
				sb.WriteByte('.')
				dotted = true
			}
			for range zeros {
				sb.WriteByte('0')
			}
			zeros = 0
			sb.WriteByte('0' + v)
		}
	}

	return sb.String()
}

func (d *digits) parsePrefix(s string) (string, error) {
	var posSeen bool
	if s == "NaN" {
		d.isNaN = true
		return "", nil
	}
	for pos, ch := range s {
		switch ch {
		case '~':
			if d.isUnderflow {
				return "", ErrMultipleUnderflowSymbols
			}
			d.isUnderflow = true
		case '<':
			if d.isOverflow {
				return "", ErrMultipleOverflowSymbols
			}
			d.isOverflow = true
		case '-':
			if d.isNeg || posSeen {
				return "", ErrMultipleMinusSigns
			}
			d.isNeg = true
		case '+':
			if d.isNeg || posSeen {
				return "", ErrMultiplePlusSigns
			}
			d.isNeg = false
			posSeen = true
		default:
			return s[pos:], nil
		}
	}

	return "", nil
}

// parseString is the main string parser to digits.
func (d *digits) parseString(s string) error {
	if len(s) == 0 {
		d.isNaN = true
		return nil
	}

	var count int
	var digitCount int
	var overflows int
	var underflow int
	var pointIdx int
	var sawDigit bool
	var expSeen bool
	var expSign int
	var decimalSeen bool
	var hasExp bool
	var expVal int
	var lead bool
	for _, ch := range s {
		switch {
		case ch >= '0' && ch <= '9':
			if expSeen {
				if expSign == 0 {
					expSign = 1
				}
				expVal = expVal*10 + int(ch-'0')
				hasExp = true
			} else {
				sawDigit = true
				switch {
				case count < precision:
					u := uint8(ch - '0')
					if u == 0 && !lead {
						continue
					}
					lead = true
					d.v[count] = uint8(ch) - '0'
					count++
					if decimalSeen {
						digitCount++
					}
				case decimalSeen && digitCount > maxDecimals:
					underflow++
				default:
					overflows++
				}
			}
		case ch == '.':
			if decimalSeen || expSeen {
				return ErrInvalidDecimalPoint
			}
			decimalSeen = true
			lead = true
			pointIdx = count
		case ch == 'e' || ch == 'E':
			if expSeen {
				return ErrMultipleExponents
			}
			expSeen = true
		case ch == '+' || ch == '-':
			if expSign != 0 {
				return ErrMultipleExponentSigns
			}
			if ch == '-' {
				expSign = -1
			} else {
				expSign = 1
			}
		default:
			return fmt.Errorf("%w: %q", ErrInvalidCharacter, ch)
		}
	}

	if expSeen && !hasExp {
		return ErrNoExponentValue
	}
	if !sawDigit {
		return ErrNoDigitsInInput
	}
	if overflows > 0 {
		d.isOverflow = true
		return nil
	}
	if underflow > 0 {
		d.isUnderflow = true
	}
	if !decimalSeen {
		pointIdx = count
	}

	d.scale(pointIdx, count, expVal, expSign)
	return nil
}

func (d *digits) scale(pointIdx, count, expVal, expSign int) {
	// handle exponent, moves decimal place
	switch expSign {
	case 1:
		dp := pointIdx + expVal
		if dp > maxWholeDigits {
			d.isOverflow = true
			return
		}
		pointIdx = dp
		count = max(count, dp)
	case -1:
		dp := pointIdx - expVal
		if count-dp > maxDecimalPlaces {
			d.isUnderflow = true
		}
		if dp >= 0 {
			pointIdx = dp
		} else {
			var padded [precision]uint8
			pointIdx = 0
			insert := min(precision, -dp)
			if insert < precision {
				copy(padded[insert:], d.v[:count])
				count = insert + count
			} else {
				count = 0
			}
			d.v = padded
		}
	}
	dp := count - pointIdx
	if dp > maxDecimalPlaces {
		d.isUnderflow = true
	}

	whole := count - dp
	if whole > maxWholeDigits {
		d.isOverflow = true
		return
	}

	d.count = count
	d.pointIdx = pointIdx
}

func (d *digits) setOverflow() {
	d.isOverflow = true
	d.pointIdx = maxWholeDigits
	d.count = precision
	for i := range precision {
		d.v[i] = 9
	}
}

// overflow is used by digits during oarsing to set a overflow max value.
// in arith operations it is better to use its overflow function.
func overflow(isNeg bool) f24 {
	f := maxF24
	f.setNeg(isNeg)
	f.setOverflow(true)
	return f
}

// parseString op level parser function.
func parseString(s string) (digits, error) {
	var d digits

	original := s
	s, err := d.parsePrefix(strings.TrimSpace(s))
	if err != nil {
		return digits{}, fmt.Errorf("%w: %w for %s", ErrParseFormatNumeric, err, original)
	}
	if d.isNaN {
		return d, nil
	}

	if err := d.parseString(s); err != nil {
		return digits{}, fmt.Errorf("%w: %w for %s", ErrParseFormatNumeric, err, original)
	}

	if d.isOverflow {
		d.setOverflow()
	}

	return d, nil
}
