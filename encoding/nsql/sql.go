// Package nsql provides numeric types that implement the database/sql Scanner and Valuer
// interfaces, allowing safe and consistent handling of numeric values with SQL databases.
//
// The package defines four primary types:
//
//   - NumericVal:
//     A non-null numeric value for NUMERIC/DECIMAL SQL columns. NaN, underflow, and overflow
//     are treated as errors and cause Scan/Value to fail.
//
//   - NumericStr:
//     A numeric value represented as a string (for TEXT/CHAR columns). NaN, underflow, and
//     overflow are permitted and encoded using the numeric package’s string format.
//
//   - NullNumericVal:
//     A nullable numeric value for NUMERIC/DECIMAL columns.
// 	   NaN and underflow are interpreted as SQL NULLs.
//     However scanned values  must be within the valid Numeric range or an error is returned.
// 	   Supports JSON null marshalling. The Valid field indicates
//     whether a non-null value is present.
//
//   - NullNumericStr:
//     A nullable numeric-as-string value for TEXT/CHAR columns. Accepts and encodes NaN,
//     underflow, and overflow using string representation.
// 	   NaN and underflow are interpreted as SQL NULLs.
//     However scanned values  must be within the valid Numeric range or an error is returned.
// 	   Supports JSON null handling and uses the Valid field to indicate presence.
//
// All types are based on github.com/nehemming/numeric for accurate and consistent numeric
// operations, formatting, and validation.

package nsql

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"unsafe"

	"github.com/nehemming/numeric"
)

var (
	// ErrCannotCoerceScannedType is returned when a scanned value cannot be coerced into the expected type.
	ErrCannotCoerceScannedType = errors.New("cannot convert scanned value into type")

	// ErrIsUnderOverNaN is returned when a value cannot be converted to a valid storage type.
	ErrIsUnderOverNaN = errors.New("cannot convert value (NaN/Overflow/Underflow) to storage type")
)

type (
	// NumericVal is a numeric value that can be stored in a database Numeric type
	// The value treats NaN, underflows and overflows as errors.
	NumericVal struct {
		numeric.Numeric
	}

	// NumericStr is a numeric value that can be stored in a database Text/Char type.
	// Underflows, Overflows and NaN are encoded using the numeric String format.
	// A Null value read from the database wil be mapped to a NaN value.
	NumericStr struct {
		numeric.Numeric
	}

	// NullNumericVal is a numeric value that can be stored in a database Numeric type
	// The value treats NaN, underflows and overflows as Null.
	NullNumericVal struct {
		numeric.Numeric
		Valid bool
	}

	// NullNumericStr is a numeric value that can be stored in a database nullable Text/Char type.
	// Underflows, Overflows and NaN are encoded using the numeric String format.
	// Null needs to be explicably set.
	NullNumericStr struct {
		numeric.Numeric
		Valid bool
	}
)

func (nv *NumericVal) Scan(value any) error {
	switch v := value.(type) {
	case int64:
		if err := numeric.ValidateIntRange(v); err != nil {
			return fmt.Errorf("%w: %d", err, v)
		}
		nv.Numeric = numeric.FromInt(v)
	case float64:
		if err := numeric.ValidateFloatRange(v); err != nil {
			return fmt.Errorf("%w: %f", err, v)
		}
		nv.Numeric = numeric.FromFloat64(v)
	case []byte:
		s := unsafe.String(unsafe.SliceData(v), len(v))
		num, err := numeric.FromString(s)
		if err != nil {
			return err
		}
		nv.Numeric = num
	case string:
		num, err := numeric.FromString(v)
		if err != nil {
			return err
		}
		nv.Numeric = num
	default:
		return fmt.Errorf("%w: %T into NumericVal", ErrCannotCoerceScannedType, value)
	}

	return nil
}

func (nv NumericVal) Value() (driver.Value, error) {
	if nv.IsUnderOverNaN() {
		return nil, ErrIsUnderOverNaN
	}
	return nv.String(), nil
}

func (ns *NumericStr) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		ns.Numeric = numeric.NaN()
	case int64:
		if err := numeric.ValidateIntRange(v); err != nil {
			return fmt.Errorf("%w: %d", err, v)
		}
		ns.Numeric = numeric.FromInt(v)
	case float64:
		if err := numeric.ValidateFloatRange(v); err != nil {
			return fmt.Errorf("%w: %f", err, v)
		}
		ns.Numeric = numeric.FromFloat64(v)
	case []byte:
		s := unsafe.String(unsafe.SliceData(v), len(v))
		num, err := numeric.FromString(s)
		if err != nil {
			return err
		}
		ns.Numeric = num
	case string:
		num, err := numeric.FromString(v)
		if err != nil {
			return err
		}
		ns.Numeric = num
	default:
		return fmt.Errorf("%w: %T into NumericStr", ErrCannotCoerceScannedType, value)
	}
	return nil
}

func (ns NumericStr) Value() (driver.Value, error) {
	return ns.String(), nil
}

func (nv *NullNumericVal) Scan(value any) error {
	if value == nil {
		nv.Numeric = numeric.NaN()
		nv.Valid = false
		return nil
	}

	var num numeric.Numeric
	switch v := value.(type) {
	case int64:
		if err := numeric.ValidateIntRange(v); err != nil {
			return fmt.Errorf("%w: %d", err, v)
		}
		num = numeric.FromInt(v)
	case float64:
		if err := numeric.ValidateFloatRange(v); err != nil {
			return fmt.Errorf("%w: %f", err, v)
		}
		num = numeric.FromFloat64(v)
	case []byte:
		s := unsafe.String(unsafe.SliceData(v), len(v))
		n, err := numeric.FromString(s)
		if err != nil {
			return err
		}
		num = n
	case string:
		n, err := numeric.FromString(v)
		if err != nil {
			return err
		}
		num = n
	default:
		return fmt.Errorf("%w: %T into NullNumericVal", ErrCannotCoerceScannedType, value)
	}

	if num.IsUnderOverNaN() {
		nv.Numeric = numeric.NaN()
		nv.Valid = false
	} else {
		nv.Numeric = num
		nv.Valid = true
	}
	return nil
}

func (nv NullNumericVal) Value() (driver.Value, error) {
	if !nv.Valid || nv.IsNaN() || nv.HasOverflow() || nv.HasUnderflow() {
		return nil, nil
	}
	return nv.String(), nil
}

func (ns *NullNumericStr) Scan(value any) error {
	if value == nil {
		ns.Numeric = numeric.NaN()
		ns.Valid = false
		return nil
	}

	var num numeric.Numeric
	switch v := value.(type) {
	case int64:
		if err := numeric.ValidateIntRange(v); err != nil {
			return fmt.Errorf("%w: %d", err, v)
		} else {
			num = numeric.FromInt(v)
		}
	case float64:
		if err := numeric.ValidateFloatRange(v); err != nil {
			return fmt.Errorf("%w: %f", err, v)
		} else {
			num = numeric.FromFloat64(v)
		}
	case []byte:
		s := unsafe.String(unsafe.SliceData(v), len(v))
		n, err := numeric.FromString(s)
		if err != nil {
			return err
		}
		num = n
	case string:
		n, err := numeric.FromString(v)
		if err != nil {
			return err
		}
		num = n
	default:
		return fmt.Errorf("%w: %T into NullNumericStr", ErrCannotCoerceScannedType, value)
	}

	ns.Numeric = num
	ns.Valid = true
	return nil
}

func (ns NullNumericStr) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String(), nil
}

func (nv NullNumericVal) String() string {
	if !nv.Valid {
		return "null"
	}
	return nv.Numeric.String()
}

func (ns NullNumericStr) String() string {
	if !ns.Valid {
		return "null"
	}
	return ns.Numeric.String()
}

// Format implements string formatting for NullNumericVal.
func (n NullNumericVal) Format(f fmt.State, verb rune) {
	if !n.Valid {
		fmt.Fprint(f, "null")
		return
	}
	n.Numeric.Format(f, verb)
}

// Format implements string formatting for NullNumericStr.
func (n NullNumericStr) Format(f fmt.State, verb rune) {
	if !n.Valid {
		fmt.Fprint(f, "null")
		return
	}
	n.Numeric.Format(f, verb)
}

// MarshalJSON for NullNumericVal
func (n NullNumericVal) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Numeric.MarshalJSON()
}

// UnmarshalJSON for NullNumericVal
func (n *NullNumericVal) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Numeric = numeric.NaN()
		n.Valid = false
		return nil
	}
	if err := n.Numeric.UnmarshalJSON(data); err != nil {
		return err
	}
	n.Valid = !n.Numeric.IsUnderOverNaN()
	return nil
}

// MarshalJSON for NullNumericStr
func (n NullNumericStr) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Numeric.MarshalJSON()
}

// UnmarshalJSON for NullNumericStr
func (n *NullNumericStr) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Numeric = numeric.NaN()
		n.Valid = false
		return nil
	}
	if err := n.Numeric.UnmarshalJSON(data); err != nil {
		return err
	}
	n.Valid = true
	return nil
}
