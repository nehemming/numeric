package checks

import "github.com/shopspring/decimal"

func sanitizeInput(input string) string {
	// Normalize input to range ±1e9 with ≤10 decimal digits
	d, err := decimal.NewFromString(input)
	if err != nil {
		return "0"
	}
	max := decimal.NewFromFloat(1e9)
	min := max.Neg()
	if d.GreaterThan(max) {
		d = max
	} else if d.LessThan(min) {
		d = min
	}
	// Limit to 10 decimal places
	return d.Round(10).String()
}
