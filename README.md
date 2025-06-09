# numeric

`numeric` is a high-performance Go package for fixed-precision decimal math with explicit handling of overflow, underflow, and inexact values. It is designed for scenarios where **precision**, **determinism**, and **performance** are critical—such as financial, scientific, or blockchain applications.

## ✨ Features

- ✅ Fixed **54-digit decimal** representation  
  - 18 digits above the decimal point  
  - 36 digits below the decimal point
- ✅ Zero allocations for maths operations and parsing.
- ✅ Correct **overflow** and **underflow** detection
- ✅ Fast, deterministic arithmetic
- ✅ Human-readable decimal formatting with symbolic extensions:
  - `~`: underflow/inexact
  - `<...`: overflow
  - `NaN`: invalid or undefined
- ✅ JSON and text marshalling/unmarshalling
- ✅ Compatible formatting float64 and int 

---

## 📦 Installation

```bash
go get github.com/nehemming/numeric
```

Import it:

```go
import "github.com/nehemming/numeric"
```

---

## 🚀 Basic Usage

### Creating Numbers

```go
n1 := numeric.FromInt(42)
n2, _ := numeric.FromString("1.2345")
n3 := numeric.FromFloat64(3.14159)
```

### Arithmetic

```go
sum := n1.Add(n2)
diff := n1.Sub(n2)
product := n1.Mul(n2)
quotient := n1.Div(n2)
neg := n2.Neg()
abs := neg.Abs()
```

### Rounding and Truncation

```go
rounded := n2.Round(2, numeric.RoundHalfUp)
truncated := n2.Truncate(numeric.FromInt(0)) // Equivalent to Round(0, RoundTowards)
```

### Conversion

```go
s := n2.String()  // "1.2345"
f := n2.Float64() // Approximate float64
i := n2.Int()     // Truncated integer
```

### Comparison

```go
if n1.IsGreaterThan(n2) {
    fmt.Println("n1 > n2")
}

switch n1.Cmp(n2) {
case -1:
    fmt.Println("n1 < n2")
case 0:
    fmt.Println("n1 == n2")
case 1:
    fmt.Println("n1 > n2")
}
```

### Underflow and Overflow Detection

```go
n, _ := numeric.FromString("1e-37")
fmt.Println(n.HasUnderflow()) // true
fmt.Println(n.String())       // "~0"
```

---

## 🔢 Special String Formatting Rules

The `numeric` package accepts and produces special formats for edge-case numeric states:

| Input                            | Description              | Output                                                                 |
|----------------------------------|---------------------------|------------------------------------------------------------------------|
| `""`                             | Invalid/empty             | `NaN`                                                                  |
| `"~1.23"`                        | Underflow, inexact        | `~1.23`                                                                |
| `"<1.23"`                        | Overflow                  | `<999999999999999999.999999999999999999999999999999999999`            |
| `"NaN"`                          | Not-a-Number              | `NaN`                                                                  |
| `"1e-37"`                        | Too small                 | `~0`                                                                   |
| `"1e18"`                         | Too large                 | `<999999999999999999.999999999999999999999999999999999999`            |
| `"~0"`                           | Inexact zero              | `~0`                                                                   |
| `"~-<1"`                         | Underflow + overflow      | `~-<999999999999999999.999999999999999999999999999999999999`          |
| `"-1e-37"`                       | Negative underflow        | `~-0`                                                                  |
| `"-1e20"`                        | Negative overflow         | `-<999999999999999999.999999999999999999999999999999999999`           |
| `"+1.23"`                        | Explicit positive         | `1.23`                                                                 |
| `"999999999999999999"`          | Max exact value           | `999999999999999999`                                                  |

These representations are used in both input parsing and output formatting, and provide clear visibility into rounding or representational limits.

## Numeric Type Formatting

The `Numeric` type implements Go's `fmt.Formatter` interface to provide custom formatting behavior using format verbs. Below is a summary of the supported verbs and their corresponding output behavior.

| Verb | Description                                                                 | Output Example                   |
|------|-----------------------------------------------------------------------------|----------------------------------|
| `v`  | Default format. If `#` flag is set, prints as `Numeric(value)`             | `123.45`, `Numeric(123.45)`      |
| `f`  | Decimal point format (float64)                                              | `123.45`                         |
| `e`  | Scientific notation (float64, lower-case `e`)                              | `1.234500e+02`                   |
| `E`  | Scientific notation (float64, upper-case `E`)                              | `1.234500E+02`                   |
| `g`  | Compact format (float64)                                                   | `123.45`                         |
| `G`  | Compact format (float64, upper-case exponent if needed)                    | `123.45`                         |
| `d`  | Decimal integer format (uses `Int()` method)                               | `123`                            |
| `s`  | String format (uses `String()` method)                                     | `123.45`                         |
| `q`  | Quoted string format (uses `String()` method)                              | `"123.45"`                       |
| any other | Unsupported verb; prints a Go-style format error                     | `%!x(Numeric=123.45)` (example)  |

### Notes
- The output respects additional formatting flags (e.g., width, precision) passed to `fmt`.
- Verb `v` with the `#` flag gives a Go-style representation helpful for debugging.
- For unsupported verbs, the formatter prints a placeholder with the `%!verb(Type=value)` pattern.

### Examples

```go
n := NumericFromString("123.45")

fmt.Printf("%v\n", n)       // 123.45
fmt.Printf("%#v\n", n)      // Numeric(123.45)
fmt.Printf("%.2f\n", n)     // 123.45
fmt.Printf("%e\n", n)       // 1.234500e+02
fmt.Printf("%d\n", n)       // 123
fmt.Printf("%q\n", n)       // "123.45"
fmt.Printf("%x\n", n)       // %!x(Numeric=123.45)
```

## 📤 JSON and Text Marshalling

The `Numeric` type implements `json.Marshaler` and `json.Unmarshaler` interfaces:

```go
n := numeric.FromInt(123)
jsonBytes, _ := json.Marshal(n)
fmt.Println(string(jsonBytes)) // Output: "123"
```

To parse from JSON:

```go
var parsed numeric.Numeric
_ = json.Unmarshal([]byte("\"1.23\""), &parsed)
fmt.Println(parsed.String()) // "1.23"
```

Special values like `"NaN"` and symbolic strings like `"~1.23"` and `"<1.23"` are supported.

Text marshalling is also supported via `MarshalText` and `UnmarshalText`.

---

## 📜 License

This project is licensed under the [MIT License](LICENSE).

---

## 🙋 Contributing

Contributions are welcome!

Please ensure your changes:

- Maintain the **zero-allocation** design philosophy
- Respect the **fixed-precision arithmetic** model
- Include test coverage for:
  - Overflow and underflow scenarios
  - Symbolic formatting logic
  - Arithmetic and rounding correctness

---
