module github.com/nehemming/numeric/checks

go 1.24.3

require github.com/nehemming/numeric v0.0.0

replace github.com/nehemming/numeric => ../

// Import shopspring decimal to use a s a reference for the checks
require github.com/shopspring/decimal v1.4.0
