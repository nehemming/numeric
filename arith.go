package numeric

import (
	"math/bits"
)

type (
	arithmetic struct{}
	divUnit    = int64
	divArray   [8]divUnit
)

// arith functions are intended for internal calculation logic only.
// They work with *f24 types in the form z = x op y where op is one of
// add, sub, mul, div, etc.  Z is aloways assumed to be zero value prior to the operation.
// The functions handle NaN, underflow, overflow, and sign as needed.
var arith arithmetic

var (
	mulOffset = [6]int{1, 0, -1, -2, -3, -4}
	powers    = [radixDigits + 1]uint64{1, 10, 100, 1000, 10000, 100_000, 1_000_000, 10_000_000, 100_000_000, 1000_000_000}
)

func (arithmetic) overflow(z *f24) {
	z[0].setVal(maxDigit)
	z[1].setVal(maxDigit)
	z[2].setVal(maxDigit)
	z[3].setVal(maxDigit)
	z[4].setVal(maxDigit)
	z[5].setVal(maxDigit)
	z.setOverflow(true)
}

func (arithmetic) add(z, x, y *f24) {
	if x.isNaN() || y.isNaN() {
		z.setNaN(true)
		return
	}

	if x.isUnderflow() || y.isUnderflow() {
		z.setUnderflow(true)
	}

	// If both have same sign, do digit-wise addition
	if isNeg := x.isNeg(); isNeg == y.isNeg() {
		z.setNeg(isNeg)
		if x.isOverflow() || y.isOverflow() {
			arith.overflow(z)
			return
		}
		arith.unsignedAdd(z, x, y)
	} else {
		if x.isOverflow() || y.isOverflow() {
			z.setNeg(x.isNeg() || y.isOverflow())
			arith.overflow(z)
			return
		}

		// Signs differ, perform subtraction: big - small
		// Determine which operand has greater magnitude
		switch arith.unsignedCompare(x, y) {
		case 0:
			// x == y → result is zero
			return
		case 1:
			// |x| > |y| → result sign = x.sign
			arith.unsignedSub(z, x, y)
			z.setNeg(x.isNeg())
		case -1:
			// |y| > |x| → result sign = y.sign
			arith.unsignedSub(z, y, x)
			z.setNeg(y.isNeg())
		}
	}
}

func (arith arithmetic) sub(z, x, y *f24) {
	var yNeg f24
	arith.negate(&yNeg, y)
	arith.add(z, x, &yNeg)
}

func (arithmetic) mul(z, x, y *f24) {
	if x.isNaN() || y.isNaN() {
		z.setNaN(true)
		return
	}

	isNeg := x.isNeg() != y.isNeg()
	z.setNeg(isNeg)

	if x.isUnderflow() || y.isUnderflow() {
		z.setUnderflow(true)
	}

	if !y.isZero() && (x.isOverflow() || y.isOverflow()) {
		arith.overflow(z)
		return
	}

	var accumulator [12]uint64

	// Multiply 6×6 base-1e9 digits
	for i := lowIndex; i >= 0; i-- {
		xi := uint64(x[i].val())
		if xi == 0 {
			continue
		}
		for j := lowIndex; j >= 0; j-- {
			yj := uint64(y[j].val())
			if yj == 0 {
				continue
			}
			pos := 3 - mulOffset[i] - mulOffset[j]
			v := xi * yj
			accumulator[pos] += v % radix
			if accumulator[pos] >= radix {
				accumulator[pos] -= radix
				accumulator[pos-1]++
			}

			accumulator[pos-1] += v / radix
			if accumulator[pos-1] >= radix {
				accumulator[pos-1] -= radix
				accumulator[pos-2]++
			}
		}
	}

	// check for an overflow.
	if accumulator[0] != 0 || accumulator[1] != 0 {
		arith.overflow(z)
		return
	}
	z[0].setVal(uint32(accumulator[2]))
	z[1].setVal(uint32(accumulator[3]))
	z[2].setVal(uint32(accumulator[4]))
	z[3].setVal(uint32(accumulator[5]))
	z[4].setVal(uint32(accumulator[6]))
	z[5].setVal(uint32(accumulator[7]))
	if accumulator[8] != 0 || accumulator[9] != 0 || accumulator[10] != 0 || accumulator[11] != 0 {
		z.setUnderflow(true)
		return
	}
}

func (arith arithmetic) div(z, x, y *f24) {
	// check for NaN's
	if x.isNaN() || y.isNaN() || y.isZero() {
		z.setNaN(true)
		return
	}
	// get negative sign
	isNeg := x.isNeg() != y.isNeg()
	defer func() {
		// ensure we have a closure here on final z.
		z.setNeg(shouldBeNeg(z, isNeg))
	}()

	// if overflowing result is an overflow
	if x.isOverflow() || y.isOverflow() {
		arith.overflow(z)
		return
	}

	// if either value has an underflow this will be under too.
	z.setUnderflow(x.isUnderflow() || y.isUnderflow())

	// when x = 0 so is z.
	if x.isZero() {
		return
	}

	// arith.divLong(z, x, y)
	arith.divInner(z, x, y)
}

func (arithmetic) divInner(z, x, y *f24) {
	// This is a implementation of the Knuth division algorithm
	// on digits of radix based integers with a fixed point after 2 digits.

	// step 1 we are going to shift the numerator and denominator
	// to the left until the first digit of the denominator is non-zero.
	// we will also shift the numerator to the left
	var num divArray // Holds extended numerator (192 bits extended to 256 for computation)
	var den divArray // Holds extended denominator

	var firstNumDigit, firstDenDigit int // Track positions of first non-zero digit in x and y
	dp := 1                              // Denominator pointer index (starts at 1 for alignment)
	np := 1                              // Numerator pointer index (starts at 1 for alignment)
	// Normalize numerator and denominator into divUnit arrays
	for i := range lenF24 {
		nv := divUnit(x[i].val()) // Get the value from x[i]
		dv := divUnit(y[i].val()) // Get the value from y[i]

		// Skip leading zeroes in numerator
		if firstNumDigit != 0 || nv != 0 {
			if firstNumDigit == 0 {
				firstNumDigit = i + 1
			}
			num[np] = nv
			np++
		}

		// Skip leading zeroes in denominator
		if firstDenDigit != 0 || dv != 0 {
			if firstDenDigit == 0 {
				firstDenDigit = i + 1
			}
			den[dp] = dv
			dp++
		}
	}

	// Shift is used to place digits relative to the decimal point
	shift := 1 + firstNumDigit - firstDenDigit
	if num[1] < den[1] {
		shift++
	}

	// we have the values in num and den non zero at idx 1
	// now if the first digit < radixHalf we need to normilize.
	if num[1] < radixHalfI {
		if normalization := radixHalfI / (num[1] + 1); normalization != 0 {
			_, np = num.mul(normalization)
			_, dp = den.mul(normalization)
		}

		// shift to left to remove zeros
		if np > 0 {
			p := 0
			for i := np; i < len(num); i++ {
				num[p] = num[i]
				num[i] = 0
				p++
			}
		}

		if dp > 0 {
			p := 0
			for i := dp; i < len(den); i++ {
				den[p] = den[i]
				den[i] = 0
				p++
			}
		}
	}
	den.trimZero()
	dEst := den[0]*radixI + den[1]

	// now we have adjusted the values
	// now we can follow the Knuth algorithm
	var res divArray     // Holds the raw division result
	var q, carry divUnit // Holds quotient estimate
	np = 0
	for i := range len(res) {
		if num[np] == 0 && carry == 0 {
			res[i] = 0
			_, isZero := num.trimZero()
			if isZero {
				break // If numerator is zero, we can stop
			}
			continue
		}
		q, carry = num.estimate(np, dEst, carry) // Estimate quotient digit
		// Move to next numerator digit
		if q == 0 {
			trimmed, isZero := num.trimZero()
			if carry == 0 && isZero {
				break // If numerator is zero, we can stop
			}
			if !trimmed {
				np = 1
			}
			continue // If estimate is zero, continue to next digit
		}

		prod := den
		for q != 0 {
			mulCarry, _ := prod.mul(q) // Multiply denominator by quotient estimate
			prod.shiftCarry(mulCarry)
			prod.trimZero()
			rem := num
			// Subtract product from numerator
			if borrow := rem.sub(&prod); borrow != 0 {
				q-- // Retry with smaller quotient if borrow occurred
				continue
			}

			// Attempt to move the digits left, if thereis carry we
			// will handle it here too.
			num = rem
			trimmed, _ := num.trimZero()
			switch {
			case !trimmed: // carry present
				np = 1
				carry = num[0]
			case num[0] == 0: // new shifted value is zero, clear carry
				np = 0
				carry = 0
			case np != 0: // we still have carry after the shift.
				np = 1
				carry = num[0]
			default:
				carry = 0
			}

			break // Successful subtraction, exit loop
		}
		res[i] = q
	}

	// Adjust for a leading zero offset in results.
	if res[0] == 0 {
		shift--
	}

	// correct values for decimal point.
	for i, v := range res {
		rp := shift + i
		if rp >= 0 && rp <= lowIndex {
			z[rp].setVal(uint32(v))
		} else if rp < 0 && v > 0 {
			// overflow.
			arith.overflow(z)
			return
		} else if rp > lowIndex && v != 0 {
			// If we have a digit beyond the lowIndex, it means overflow.
			z.setUnderflow(true)
			return
		}
	}
}

// u128 calculates a hi:lo uint64 from the divArray at index i adding in any carry.
func (d *divArray) u128(i int, carry divUnit) (hi, lo uint64) {
	// carry * radix2 + d[i] radix + d[i+1]
	lo = uint64(d[i])*radix + uint64(d[i+1])
	if carry == 0 {
		return
	}
	chi, clo := bits.Mul64(uint64(carry)*radix, radix) // Multiply carry by radix

	lo, c := bits.Add64(clo, lo, 0) // add low together
	hi = chi + c                    // carry from low addition

	return hi, lo
}

func (d *divArray) estimate(i int, den divUnit, carry divUnit) (divUnit, divUnit) {
	// use math.bits to calc a integer quotient estimate
	denU := uint64(den)

	hi, lo := d.u128(i, carry)
	q, r := bits.Div64(hi, lo, denU)

	if q == 0 {
		// remainder needs to be reduced by 1 radix as we multiplied by 1 radix in u128.
		carry = divUnit(r / radix)
	} else {
		carry = 0
	}

	return divUnit(q), carry
}

func (d *divArray) sub(other *divArray) divUnit {
	// Subtract each element of other from da and handle underflow
	var borrow divUnit
	for i := len(d) - 1; i >= 0; i-- {
		diff := d[i] - other[i] - borrow
		if diff < 0 {
			diff += radixI
			borrow = 1
		} else {
			borrow = 0
		}
		d[i] = diff
	}
	return borrow
}

func (d *divArray) mul(x divUnit) (divUnit, int) {
	// Multiply each element by x and handle overflow
	var carry divUnit
	var firstNonZero int // Track first non-zero digit position
	for i := len(d) - 1; i >= 0; i-- {
		product := d[i]*x + carry
		d[i] = product % radixI  // Store the result in the current position
		carry = product / radixI // Carry over to the next position
		if d[i] != 0 {
			firstNonZero = i // Update first non-zero position
		}
	}
	return carry, firstNonZero
}

func (d *divArray) shiftCarry(carry divUnit) {
	// Shift the carry to the left, adding it to the next element
	if carry == 0 {
		return // No carry to shift
	}
	for i := len(d) - 1; i > 0; i-- {
		d[i] = d[i-1]
	}
	d[0] = carry
}

func (d *divArray) trimZero() (trimmed bool, isZero bool) {
	if d[0] != 0 {
		return false, false // No leading zero to trim
	}
	isZero = true // Assume zero unless we find a non-zero digit
	for i := 1; i < len(d)-1; i++ {
		v := d[i]
		if v != 0 {
			isZero = false // Found a non-zero digit
		}
		d[i-1] = v
	}
	return true, isZero
}

func (arithmetic) negate(z, x *f24) {
	if x.isNaN() {
		z.setNaN(true)
		return
	}
	*z = *x // copy value and flags

	// Make z negative if x was positive and not a real zero (unless it's a zero due to underflow).
	z.setNeg(!x.isNeg() && !(x.isZero() && !x.isUnderflow())) // flip sign
}

func (arithmetic) abs(z, x *f24) {
	if x.isNaN() {
		z.setNaN(true)
		return
	}
	*z = *x // copy value and flags
	z.setNeg(false)
}

// unsignedCompare compares |x| and |y|
// Returns 1 if |x| > |y|, -1 if |x| < |y|, 0 if equal
func (arithmetic) unsignedCompare(x, y *f24) int {
	xv := x[0].val()
	yv := y[0].val()
	if xv > yv {
		return 1
	} else if xv < yv {
		return -1
	}

	xv = x[1].val()
	yv = y[1].val()
	if xv > yv {
		return 1
	} else if xv < yv {
		return -1
	}

	xv = x[2].val()
	yv = y[2].val()
	if xv > yv {
		return 1
	} else if xv < yv {
		return -1
	}

	xv = x[3].val()
	yv = y[3].val()
	if xv > yv {
		return 1
	} else if xv < yv {
		return -1
	}

	xv = x[4].val()
	yv = y[4].val()
	if xv > yv {
		return 1
	} else if xv < yv {
		return -1
	}

	xv = x[5].val()
	yv = y[5].val()
	if xv > yv {
		return 1
	} else if xv < yv {
		return -1
	}

	return 0
}

func (arith arithmetic) compare(x, y *f24) int {
	if x.isNaN() {
		return -1 // NaN's are not equal but for comparison we will treat as less
	}

	if y.isNaN() {
		return 1 // NaN's are not equal but for comparison we will treat as less
	}

	var cmp int

	xs, ys := x.isNeg(), y.isNeg()
	if xs == ys {
		switch {
		case x.isOverflow():
			cmp = -1
		case y.isOverflow():
			cmp = 1
		default:
			cmp = arith.unsignedCompare(x, y)
			if cmp == 0 {
				switch {
				case x.isUnderflow():
					cmp = 1
				case y.isUnderflow():
					cmp = -1
				}
			}
		}
	}

	switch {
	case xs && ys:
		cmp = -cmp
	case xs:
		cmp = -1
	case ys:
		cmp = 1
	}

	return cmp
}

func (arith arithmetic) equal(x, y *f24) bool {
	if arith.hasExceptionalState(x) || arith.hasExceptionalState(y) {
		return false
	}
	return *x == *y
}

func (arith arithmetic) hasExceptionalState(f *f24) bool {
	return f.isNaN() || f.isUnderflow() || f.isOverflow()
}

// unsignedSub performs z = |a| - |b|
// Assumes |a| ≥ |b|
func (arithmetic) unsignedSub(z, a, b *f24) {
	const radix = uint64(1e9)
	var borrow uint64

	// i = 5
	ai := uint64(a[5].val())
	bi := uint64(b[5].val()) + borrow
	if ai < bi {
		ai += radix
		borrow = 1
	} else {
		borrow = 0
	}
	z[5].setVal(uint32(ai - bi))

	// i = 4
	ai = uint64(a[4].val())
	bi = uint64(b[4].val()) + borrow
	if ai < bi {
		ai += radix
		borrow = 1
	} else {
		borrow = 0
	}
	z[4].setVal(uint32(ai - bi))

	// i = 3
	ai = uint64(a[3].val())
	bi = uint64(b[3].val()) + borrow
	if ai < bi {
		ai += radix
		borrow = 1
	} else {
		borrow = 0
	}
	z[3].setVal(uint32(ai - bi))

	// i = 2
	ai = uint64(a[2].val())
	bi = uint64(b[2].val()) + borrow
	if ai < bi {
		ai += radix
		borrow = 1
	} else {
		borrow = 0
	}
	z[2].setVal(uint32(ai - bi))

	// i = 1
	ai = uint64(a[1].val())
	bi = uint64(b[1].val()) + borrow
	if ai < bi {
		ai += radix
		borrow = 1
	} else {
		borrow = 0
	}
	z[1].setVal(uint32(ai - bi))

	// i = 0
	// the borrow is not reliant as a has to be bigger.
	ai = uint64(a[0].val())
	bi = uint64(b[0].val()) + borrow
	if ai < bi {
		ai += radix
	}
	z[0].setVal(uint32(ai - bi))
}

// unsignedAdd performs |z| = |x| + |y|
func (arithmetic) unsignedAdd(z, x, y *f24) {
	var carry uint64

	// i = 5
	sum := uint64(x[5].val()) + uint64(y[5].val()) + carry
	if sum >= radix {
		carry = 1
		sum -= radix
	} else {
		carry = 0
	}
	z[5].setVal(uint32(sum))

	// i = 4
	sum = uint64(x[4].val()) + uint64(y[4].val()) + carry
	if sum >= radix {
		carry = 1
		sum -= radix
	} else {
		carry = 0
	}
	z[4].setVal(uint32(sum))

	// i = 3
	sum = uint64(x[3].val()) + uint64(y[3].val()) + carry
	if sum >= radix {
		carry = 1
		sum -= radix
	} else {
		carry = 0
	}
	z[3].setVal(uint32(sum))

	// i = 2
	sum = uint64(x[2].val()) + uint64(y[2].val()) + carry
	if sum >= radix {
		carry = 1
		sum -= radix
	} else {
		carry = 0
	}
	z[2].setVal(uint32(sum))

	// i = 1
	sum = uint64(x[1].val()) + uint64(y[1].val()) + carry
	if sum >= radix {
		carry = 1
		sum -= radix
	} else {
		carry = 0
	}
	z[1].setVal(uint32(sum))

	// i = 0
	sum = uint64(x[0].val()) + uint64(y[0].val()) + carry
	if sum >= radix {
		carry = 1
		sum -= radix
	} else {
		carry = 0
	}
	z[0].setVal(uint32(sum))

	// Check for overflow
	if carry > 0 {
		*z = overflow(z.isNeg())
	}
}

func (arith arithmetic) round(z, x *f24, y int, mode RoundMode) {
	isNeg := x.isNeg()
	defer func() {
		// ensure we have a closure here on final z.
		z.setNeg(shouldBeNeg(z, isNeg))
	}()
	switch {
	case x.isNaN():
		z.setNaN(true)
	case x.isOverflow():
		arith.overflow(z)
	case x.isZero():
	case y < 0:
		z.setNaN(true)
	default:
		idx := decIndex + y/radixDigits
		v := uint64(x[idx].val())

		pow := radixDigits - y%radixDigits
		p := powers[pow]
		rem := v % p
		v -= rem
		switch mode {
		case RoundAway:
			if rem > 0 {
				v += p
			} else {
				for i := idx + 1; i < lenF24; i++ {
					if x[i].val() != 0 {
						v += p
						break
					}
				}
			}
		case RoundTowards:
		case RoundHalfDown:
			if rem > p/2 {
				v += p
			}
		case RoundHalfUp:
			if (rem + 1) > p/2 {
				v += p
			}
		}
		carry := v / radix
		v %= radix
		for i := idx - 1; i >= 0; i-- {
			xi := uint64(x[i].val()) + carry
			carry = xi / radix
			xi %= radix
			z[i].setVal(uint32(xi))
		}
		z[idx].setVal(uint32(v))
		for i := idx + 1; i < lenF24; i++ {
			z[i].setVal(0)
		}
	}
}

func (arith arithmetic) quanta(z, x, y *f24, mode RoundMode) {
	var w f24
	arith.div(&w, x, y)
	if w.isNaN() || w.isOverflow() {
		z.setNaN(true)
		return
	}

	var u f24
	arith.round(&u, &w, 0, mode)
	arith.mul(z, &u, y)
}

func (arith arithmetic) divRem(q, r, x, y *f24) {
	var w f24
	arith.div(&w, x, y)
	if w.isNaN() || w.isOverflow() {
		q.setNaN(true)
		r.setNaN(true)
		return
	}
	arith.round(q, &w, 0, RoundTowards)

	var u f24
	arith.mul(&u, q, y)
	arith.sub(r, x, &u)
}

func shouldBeNeg(x *f24, isNeg bool) bool {
	if x.isNaN() {
		return false
	}
	if x.isZero() {
		if x.isUnderflow() {
			return isNeg
		}
		return false
	}
	return isNeg
}
