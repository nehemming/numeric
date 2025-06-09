package numeric

// Bit constants for fVal internal representation
// The integer values stored in f24 do not use the full bits so we borrow the high bit
// for logical flags. The lower 31 bits are used for the actual numeric value.
const (
	flagBit  = fVal(1) << 31    // MSB used as a boolean flag (sign, NaN, overflow, etc.)
	maskBits = fVal(0x7FFFFFFF) // Lower 31 bits hold the actual numeric value
)

// Indices within f24 used for logical flags
const (
	signFlag      = iota // Index for negative sign flag
	nanFlag              // Index for NaN (Not-a-Number) flag
	overflowFlag         // Index for overflow flag
	underflowFlag        // Index for underflow flag
)

// fVal represents a 32-bit unit: high bit is a flag, lower 31 bits store value
type fVal uint32

// flag returns true if the high bit is set.
func (fv *fVal) flag() bool {
	return (*fv & flagBit) != 0
}

// setFlag sets or clears the high bit (flag).
func (fv *fVal) setFlag(flag bool) {
	if flag {
		*fv |= flagBit // Set MSB
	} else {
		*fv &^= flagBit // Clear MSB (AND NOT)
	}
}

// val returns the numeric portion (lower 31 bits).
func (fv *fVal) val() uint32 {
	return uint32(*fv & maskBits)
}

// setVal sets the numeric portion (lower 31 bits), preserving the flag bit.
func (fv *fVal) setVal(val uint32) {
	*fv = (*fv & flagBit) | (maskBits & fVal(val))
}

// isNeg checks if the number is marked negative.
func (f *f24) isNeg() bool {
	return f[signFlag].flag()
}

// setNeg sets or clears the negative sign flag.
func (f *f24) setNeg(neg bool) {
	f[signFlag].setFlag(neg)
}

// isNaN checks if the value is NaN.
func (f *f24) isNaN() bool {
	return f[nanFlag].flag()
}

// setNaN sets or clears the NaN flag.
func (f *f24) setNaN(isNan bool) {
	f[nanFlag].setFlag(isNan)
}

// isOverflow checks if the overflow flag is set.
func (f *f24) isOverflow() bool {
	return f[overflowFlag].flag()
}

// setOverflow sets or clears the overflow flag.
func (f *f24) setOverflow(isOverflow bool) {
	f[overflowFlag].setFlag(isOverflow)
}

// isUnderflow checks if the underflow flag is set.
func (f *f24) isUnderflow() bool {
	return f[underflowFlag].flag()
}

// setUnderflow sets or clears the underflow flag.
func (f *f24) setUnderflow(isUnderflow bool) {
	f[underflowFlag].setFlag(isUnderflow)
}
