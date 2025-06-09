# Benchmark and Fitness Checks

This package is used to perform side by side validation the 
[Numeric](github.com/nehemming/numeric) package vs a reference 
version of [Shopspring Decimal](https://github.com/shopspring/decimal).

The purpose of these tests is solely to compare the output and 
confirm that the `numeric` package matches the reference decimal results.  
[Shopspring Decimal](https://github.com/shopspring/decimal) is widely used and is an excellent package.

[Numeric](github.com/nehemming/numeric) has been created to work in environments that need fixed precision 
accuracy and limited heap allocation.   
If a wider numeric range or maturity matter most [Shopspring Decimal](https://github.com/shopspring/decimal)
may offer a possible alternative.    

---

## Test types

There are 3 types of checks present:

 1) Side by side unit tests.
 2) Fuzz tests comparing the results between `numeric` and `decimal`.
 3) Benchmark tests to compare similar operations.


--- 

## ⏱️ Benchmark Comparison: Decimal vs Numeric

The following metric

| Op  | Input A & B                                 | Decimal ns/op | Numeric ns/op | Decimal B/op | Numeric B/op | Decimal allocs/op | Numeric allocs/op |
|-----|---------------------------------------------|----------------|----------------|----------------|----------------|---------------------|---------------------|
| add | 0.0000000001, 0.0000000002                   | 66.70         | 20.91         | 80            | 0             | 2                   | 0                   |
| add | 100, 0                                       | 53.94         | 21.15         | 40            | 0             | 2                   | 0                   |
| add | 123.456, -654.321                            | 56.60         | 22.36         | 40            | 0             | 2                   | 0                   |
| add | 1, 2                                         | 60.52         | 16.07         | 80            | 0             | 2                   | 0                   |
| add | 9999999.9999999999, -9999999.9999999999      | 68.24         | 11.24         | 40            | 0             | 2                   | 0                   |
| sub | 0.0000000001, 0.0000000002                   | 71.36         | 30.64         | 40            | 0             | 2                   | 0                   |
| sub | 100, 0                                       | 56.01         | 25.79         | 40            | 0             | 2                   | 0                   |
| sub | 123.456, -654.321                            | 65.96         | 26.14         | 80            | 0             | 2                   | 0                   |
| sub | 1, 2                                         | 55.87         | 26.28         | 40            | 0             | 2                   | 0                   |
| sub | 9999999.9999999999, -9999999.9999999999      | 78.38         | 25.10         | 80            | 0             | 2                   | 0                   |
| mul | 0.0000000001, 0.0000000002                   | 71.73         | 25.50         | 80            | 0             | 2                   | 0                   |
| mul | 100, 0                                       | 32.64         | 23.23         | 32            | 0             | 1                   | 0                   |
| mul | 123.456, -654.321                            | 61.23         | 33.48         | 80            | 0             | 2                   | 0                   |
| mul | 1, 2                                         | 60.12         | 23.15         | 80            | 0             | 2                   | 0                   |
| mul | 9999999.9999999999, -9999999.9999999999      | 64.98         | 51.36         | 80            | 0             | 2                   | 0                   |
| div | 0.0000000001, 0.0000000002                   | 290.4         | 193.4         | 184           | 0             | 7                   | 0                   |
| div | 123.456, -654.321                            | 426.8         | 406.5         | 328           | 0             | 12                  | 0                   |
| div | 1, 2                                         | 271.0         | 151.6         | 184           | 0             | 7                   | 0                   |
| div | 9999999.9999999999, -9999999.9999999999      | 391.0         | 127.5         | 264           | 0             | 9                   | 0                   |
 
---

## Contributor Request

Extending validation of numeric is of upmost interest and we would welcome 
help adding additional test cases.
