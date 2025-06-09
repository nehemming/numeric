package nsql

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/nehemming/numeric"
)

func TestNumericVal_ScanAndValue(t *testing.T) {
	var nv NumericVal

	tests := []struct {
		input   any
		wantErr error
		wantStr string
		wantNil bool
	}{
		{int64(42), nil, "42", false},
		{int64(1e18), numeric.ErrIntegerOutOfRange, "", true},
		{float64(3.14), nil, "3.14", false},
		{float64(1e18), numeric.ErrFloatOutOfRange, "", true},
		{[]byte("123.456"), nil, "123.456", false},
		{"789.01", nil, "789.01", false},
		{nil, ErrCannotCoerceScannedType, "", true},
		{true, ErrCannotCoerceScannedType, "", true},
		{float64(1e50), numeric.ErrFloatOutOfRange, "", true}, // overflow
		{float64(0), nil, "0", false},
		{"~789.01", nil, "~789.01", true},
		{[]byte("123!456"), numeric.ErrInvalidCharacter, "", true},
		{"123!456", numeric.ErrInvalidCharacter, "", true},
	}

	for i, tt := range tests {
		nv = NumericVal{} // reset
		err := nv.Scan(tt.input)
		if (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) || (tt.wantErr == nil && err != nil) {
			t.Errorf("Test %d (%v) : Scan() error = %v, want %v", i, tt.input, err, tt.wantErr)
			continue
		}
		if tt.wantErr != nil {
			nv.Numeric = numeric.NaN()
		}

		val, valErr := nv.Value()
		if tt.wantErr != nil || tt.wantNil {
			if val != nil {
				t.Errorf("Test %d  (%v): expected nil value, got %v", i, tt.input, val)
			}
		} else {
			if valErr != nil {
				t.Errorf("Test %d: Value() error = %v", i, valErr)
			}
			if valStr, ok := val.(string); !ok || valStr != tt.wantStr {
				t.Errorf("Test %d: Value = %v, want %v", i, val, tt.wantStr)
			}
		}
	}
}

func TestNumericStr_ScanAndValue(t *testing.T) {
	var ns NumericStr

	tests := []struct {
		input   any
		wantStr string
		wantErr error
	}{
		{nil, "NaN", nil},
		{int64(42), "42", nil},
		{int64(1e18), "", numeric.ErrIntegerOutOfRange},
		{float64(3.14), "3.14", nil},
		{float64(1e18), "", numeric.ErrFloatOutOfRange},
		{[]byte("1.618"), "1.618", nil},
		{"2.718", "2.718", nil},
		{true, "", ErrCannotCoerceScannedType},
		{[]byte("123!456"), "", numeric.ErrInvalidCharacter},
		{"123!456", "", numeric.ErrInvalidCharacter},
	}

	for i, tt := range tests {
		ns = NumericStr{}
		err := ns.Scan(tt.input)
		if (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) || (tt.wantErr == nil && err != nil) {
			t.Errorf("Test %d: Scan() error = %v, want %v", i, err, tt.wantErr)
			continue
		}

		if tt.wantErr != nil {
			ns.Numeric = numeric.NaN()
		}

		val, valErr := ns.Value()
		if tt.wantErr != nil {
			if val != "NaN" {
				t.Errorf("Test %d: expected nil, got %v", i, val)
			}
			continue
		}

		if valErr != nil {
			t.Errorf("Test %d: Value() error = %v", i, valErr)
		}
		if valStr, ok := val.(string); !ok || valStr != tt.wantStr {
			t.Errorf("Test %d: Value = %v, want %v", i, val, tt.wantStr)
		}
	}
}

func TestNullNumericVal_ScanAndValue(t *testing.T) {
	var nv NullNumericVal

	tests := []struct {
		input         any
		expectedError error
		wantStr       string
		valid         bool
	}{
		{nil, nil, "null", false},
		{int64(100), nil, "100", true},
		{int64(1e18), numeric.ErrIntegerOutOfRange, "null", false},
		{float64(100.1), nil, "100.1", true},
		{float64(1e50), numeric.ErrFloatOutOfRange, "null", false}, // overflow
		{[]byte("256"), nil, "256", true},
		{"-512", nil, "-512", true},
		{"~512", nil, "null", false},
		{true, ErrCannotCoerceScannedType, "null", false}, // invalid type
		{[]byte("123!456"), numeric.ErrInvalidCharacter, "null", false},
		{"123!456", numeric.ErrInvalidCharacter, "null", false},
	}

	for i, tt := range tests {
		nv = NullNumericVal{}
		err := nv.Scan(tt.input)
		if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
			t.Errorf("Test %d: Scan() error = %v, want %v", i, err, tt.expectedError)
			continue
		} else if tt.expectedError == nil && err != nil {
			t.Errorf("Test %d: unexpected error = %v", i, err)
		}

		if nv.Valid != tt.valid {
			t.Errorf("Test %d: Valid = %v, want %v", i, nv.Valid, tt.valid)
		}

		gotStr := nv.String()
		if gotStr != tt.wantStr {
			t.Errorf("Test %d: String() = %s, want %s", i, gotStr, tt.wantStr)
		}

		val, valErr := nv.Value()
		if tt.valid {
			if valErr != nil {
				t.Errorf("Test %d: unexpected error from Value(): %v", i, valErr)
			}
			if valStr, ok := val.(string); !ok || valStr != tt.wantStr {
				t.Errorf("Test %d: Value = %v, want %v", i, val, tt.wantStr)
			}
		} else {
			if val != nil {
				t.Errorf("Test %d: expected nil value, got %v", i, val)
			}
		}
	}
}

func TestNullNumericStr_ScanAndValue(t *testing.T) {
	var ns NullNumericStr

	tests := []struct {
		input         any
		expectedError error
		wantStr       string
		valid         bool
	}{
		{nil, nil, "null", false},
		{int64(1234), nil, "1234", true},
		{int64(1e18), numeric.ErrIntegerOutOfRange, "null", false},
		{float64(0.001), nil, "0.001", true},
		{float64(1e50), numeric.ErrFloatOutOfRange, "null", false}, // overflow
		{[]byte("1.23"), nil, "1.23", true},
		{"456.789", nil, "456.789", true},
		{true, ErrCannotCoerceScannedType, "null", false},
		{[]byte("123!456"), numeric.ErrInvalidCharacter, "null", false},
		{"123!456", numeric.ErrInvalidCharacter, "null", false},
	}

	for i, tt := range tests {
		ns = NullNumericStr{}
		err := ns.Scan(tt.input)
		if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
			t.Errorf("Test %d: Scan() error = %v, want %v", i, err, tt.expectedError)
			continue
		} else if tt.expectedError == nil && err != nil {
			t.Errorf("Test %d: unexpected error = %v", i, err)
		}

		if ns.Valid != tt.valid {
			t.Errorf("Test %d: Valid = %v, want %v", i, ns.Valid, tt.valid)
		}

		gotStr := ns.String()
		if gotStr != tt.wantStr {
			t.Errorf("Test %d: String() = %s, want %s", i, gotStr, tt.wantStr)
		}

		val, valErr := ns.Value()
		if tt.valid {
			if valErr != nil {
				t.Errorf("Test %d: unexpected error from Value(): %v", i, valErr)
			}
			if valStr, ok := val.(string); !ok || valStr != tt.wantStr {
				t.Errorf("Test %d: Value = %v, want %v", i, val, tt.wantStr)
			}
		} else {
			if val != nil {
				t.Errorf("Test %d: expected nil value, got %v", i, val)
			}
		}
	}
}

func TestNullNumericVal_Format(t *testing.T) {
	tests := []struct {
		name   string
		n      NullNumericVal
		format string
		want   string
	}{
		{
			name:   "invalid value with %v",
			n:      NullNumericVal{Valid: false},
			format: "%v",
			want:   "null",
		},
		{
			name:   "valid value with %v",
			n:      NullNumericVal{numeric.FromInt(42), true},
			format: "%v",
			want:   "42",
		},
		{
			name:   "valid value with %f",
			n:      NullNumericVal{numeric.FromFloat64(3.14159), true},
			format: "%.2f",
			want:   "3.14",
		},
	}

	for _, tt := range tests {
		got := fmt.Sprintf(tt.format, tt.n)
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNullNumericStr_Format(t *testing.T) {
	tests := []struct {
		name   string
		n      NullNumericStr
		format string
		want   string
	}{
		{
			name:   "invalid value with %v",
			n:      NullNumericStr{Valid: false},
			format: "%v",
			want:   "null",
		},
		{
			name:   "valid value with %v",
			n:      NullNumericStr{numeric.FromInt(99), true},
			format: "%v",
			want:   "99",
		},
		{
			name:   "valid value with %e",
			n:      NullNumericStr{numeric.FromFloat64(1.23e4), true},
			format: "%e",
			want:   "1.230000e+04",
		},
		{
			name:   "valid value with %e",
			n:      NullNumericStr{numeric.FromFloat64(1.23e4), true},
			format: "%-.1e",
			want:   "1.2e+04",
		},
	}

	for _, tt := range tests {
		got := fmt.Sprintf(tt.format, tt.n)
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNullNumericVal_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		value    NullNumericVal
		wantJSON string
	}{
		{
			name:     "valid int",
			value:    NullNumericVal{Numeric: numeric.FromInt(42), Valid: true},
			wantJSON: "\"42\"",
		},
		{
			name:     "valid float",
			value:    NullNumericVal{Numeric: numeric.FromFloat64(3.14), Valid: true},
			wantJSON: "\"3.14\"",
		},
		{
			name:     "null value",
			value:    NullNumericVal{Numeric: numeric.NaN(), Valid: false},
			wantJSON: "null",
		},
	}

	for _, tt := range tests {
		// Marshal to JSON
		data, err := json.Marshal(tt.value)
		if err != nil {
			t.Errorf("%s: Marshal failed: %v", tt.name, err)
			continue
		}

		if string(data) != tt.wantJSON {
			t.Errorf("%s: Marshal output = %s, want %s", tt.name, data, tt.wantJSON)
		}

		// Unmarshal back
		var result NullNumericVal
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("%s: Unmarshal failed: %v", tt.name, err)
			continue
		}

		if result.Valid != tt.value.Valid {
			t.Errorf("%s: Valid mismatch: got %v, want %v", tt.name, result.Valid, tt.value.Valid)
		}

		if result.Valid && result.String() != tt.value.String() {
			t.Errorf("%s: Value mismatch: got %s, want %s", tt.name, result.String(), tt.value.String())
		}
	}
}

func TestNullNumericStr_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		value    NullNumericStr
		wantJSON string
	}{
		{
			name:     "valid int",
			value:    NullNumericStr{Numeric: numeric.FromInt(99), Valid: true},
			wantJSON: "\"99\"",
		},
		{
			name:     "valid float",
			value:    NullNumericStr{Numeric: numeric.FromFloat64(2.718), Valid: true},
			wantJSON: "\"2.718\"",
		},
		{
			name:     "null value",
			value:    NullNumericStr{Numeric: numeric.NaN(), Valid: false},
			wantJSON: "null",
		},
	}

	for _, tt := range tests {
		// Marshal to JSON
		data, err := json.Marshal(tt.value)
		if err != nil {
			t.Errorf("%s: Marshal failed: %v", tt.name, err)
			continue
		}

		if string(data) != tt.wantJSON {
			t.Errorf("%s: Marshal output = %s, want %s", tt.name, data, tt.wantJSON)
		}

		// Unmarshal back
		var result NullNumericStr
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("%s: Unmarshal failed: %v", tt.name, err)
			continue
		}

		if result.Valid != tt.value.Valid {
			t.Errorf("%s: Valid mismatch: got %v, want %v", tt.name, result.Valid, tt.value.Valid)
		}

		if result.Valid && result.String() != tt.value.String() {
			t.Errorf("%s: Value mismatch: got %s, want %s", tt.name, result.String(), tt.value.String())
		}
	}
}
