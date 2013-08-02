Evaluate yourself some numeric expressions.

Currently, only limited Go syntax is supported, but a more calculator-y syntax is planned. The evaluator understands both arbitrary-precision integers and rationals of two of them, and will automatically promote the latter to the former when possible. Support exists (externally) for all expressions that can be represented using only Go operators (and internally for almost all operations available on math/big's types, but these can't be parsed yet).

For an example usage, see calcule/main.go; this is a program which compiles a supported Go expression and evaluates it. Its usage is of the form `"expression" "var1=1" "var2=2" ...`.
