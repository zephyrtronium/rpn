Evaluate yourself some numeric expressions.

Currently, only limited Go syntax is supported, but a more calculator-y syntax is planned. The evaluator understands both arbitrary-precision integers and rationals of two of them, and will automatically promote the latter to the former when possible.

For an example usage, see calcule/main.go; this is a program which compiles a supported Go expression and evaluates it. Its usage is of the form `"expression" "var1=1" "var2=2" ...`.

Supported operations in Go syntax:

 - (x)
 - +x
 - -x
 - ^x (integer x)
 - x+y
 - x-y
 - x*y
 - x/y
 - x%y (integers x and y)
 - x&y (integers x and y)
 - x|y (integers x and y)
 - x^y (integers x and y)
 - x&^y (integers x and y)
 - x<<y (integers x and y)
 - x>>y (integers x and y)
 - abs(x) - absolute value
 - inv(x) - 1/x
 - binomial(x, y) - binomial coefficent of integers x and y
 - div(x, y) - euclidean division of integers x and y
 - mod(x, y) - euclidean modulo of integers x and y
 - gcd(x, y) - greatest common denominator of integers x and y
 - exp(x, y[, m]) - exponentiation, optionally modulo m, of integers x, y, and m
 - modinv(x, p) - modular inverse of integer x in Z/pZ with p assumed prime
 - mulrange(x, y) - product of all integers in the range [x, y], with integers x and y
 - frac(x, y) - convert integers x and y to rational x/y
 - denom(x) - denominator of rational x
 - num(x) - numerator of rational x
