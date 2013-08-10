/*
Copyright (c) 2013 Branden J Brown

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following restrictions:

   1. The origin of this software must not be misrepresented; you must not
   claim that you wrote the original software. If you use this software
   in a product, an acknowledgment in the product documentation would be
   appreciated but is not required.

   2. Altered source versions must be plainly marked as such, and must not be
   misrepresented as being the original software.

   3. This notice may not be removed or altered from any source
   distribution.
*/

package rpn

import "math/big"

// Evaluation context. This type is exported to allow eventual user-supplied
// operations.
type Evaluator struct {
	Stack  []interface{}
	Vars   map[string]interface{}
	Names  []string
	Consts []*big.Rat
	N, C   int
}

func (e *Evaluator) eval(ops []operator) (err error) {
	for _, op := range ops {
		if err = opFuncs[op](e); err != nil {
			return err
		}
	}
	return nil
}

// Helper to get the top element on the stack.
func (e *Evaluator) Top() interface{} {
	return e.Stack[len(e.Stack)-1]
}

// Helper to get and remove the top element on the stack.
func (e *Evaluator) Pop() interface{} {
	v := e.Top()
	e.Stack = e.Stack[:len(e.Stack)-1]
	return v
}

// Helper to set the top element on the stack.
func (e *Evaluator) SetTop(v interface{}) {
	e.Stack[len(e.Stack)-1] = v
}

type opFunc func(*Evaluator) error

var opFuncs = [...]opFunc{
	oNOP: func(*Evaluator) error { return nil },
	oLOAD: func(e *Evaluator) error {
		v := e.Vars[e.Names[e.N]]
		switch i := v.(type) {
		case *big.Int:
			e.Stack = append(e.Stack, new(big.Int).Set(i))
		case *big.Rat:
			e.Stack = append(e.Stack, new(big.Rat).Set(i))
		default:
			return MissingVar{e.Names[e.N]}
		}
		e.N++
		return nil
	},
	oCONST: func(e *Evaluator) error {
		v := e.Consts[e.C]
		if v == nil {
			e.Stack = append(e.Stack, nil)
		} else if v.IsInt() {
			e.Stack = append(e.Stack, new(big.Int).Set(v.Num()))
		} else {
			e.Stack = append(e.Stack, new(big.Rat).Set(v))
		}
		e.C++
		return nil
	},
	oABS: numericUnary("ABS", (*big.Int).Abs, (*big.Rat).Abs),
	oADD: numericBinary("ADD", (*big.Int).Add, (*big.Rat).Add),
	oMUL: numericBinary("MUL", (*big.Int).Mul, (*big.Rat).Mul),
	oNEG: numericUnary("NEG", (*big.Int).Neg, (*big.Rat).Neg),
	oQUO: func(e *Evaluator) error {
		x := e.Pop()
		y := e.Top()
		switch a := x.(type) {
		case *big.Int:
			if a.Sign() == 0 {
				return DivByZero{}
			}
			switch b := y.(type) {
			case *big.Int:
				r := new(big.Rat).SetFrac(b, a)
				if r.IsInt() {
					b.Set(r.Num())
				} else {
					e.SetTop(r)
				}
			case *big.Rat:
				b.Quo(b, new(big.Rat).SetFrac(a, big.NewInt(1)))
				if b.IsInt() {
					e.SetTop(b.Num())
				}
			default:
				panic("QUO: wrong type on stack! (int/?)")
			}
		case *big.Rat:
			if a.Sign() == 0 {
				return DivByZero{}
			}
			switch b := y.(type) {
			case *big.Int:
				r := new(big.Rat).SetFrac(b, big.NewInt(1))
				r.Quo(r, a)
				if r.IsInt() {
					e.SetTop(r.Num())
				} else {
					e.SetTop(r)
				}
			case *big.Rat:
				b.Quo(b, a)
				if b.IsInt() {
					e.SetTop(b.Num())
				}
			default:
				panic("QUO: wrong type on stack! (rat/?)")
			}
		default:
			panic("QUO: wrong type on stack! (?/?)")
		}
		return nil
	},
	oSUB:      numericBinary("SUB", (*big.Int).Sub, (*big.Rat).Sub),
	oAND:      integerBinary("AND", (*big.Int).And),
	oANDNOT:   integerBinary("ANDNOT", (*big.Int).AndNot),
	oBINOMIAL: integerOverflow("BINOMIAL", (*big.Int).Binomial),
	oDIV:      integerDivision("DIV", (*big.Int).Div),
	oEXP: func(e *Evaluator) error {
		m := e.Pop()
		x := e.Pop()
		y := e.Top()
		a, aok := x.(*big.Int)
		b, bok := y.(*big.Int)
		c, cok := m.(*big.Int) // heh
		if !aok {
			_ = x.(*big.Rat)
			return TypeError{"int"}
		}
		if !bok {
			_ = y.(*big.Rat)
			return TypeError{"int"}
		}
		if m != nil && !cok { // heh
			_ = m.(*big.Rat)
			return TypeError{"int"}
		}
		invert := a.Sign() < 0
		if invert {
			c = nil
		}
		b.Exp(b, a.Abs(a), c)
		if invert {
			e.SetTop(new(big.Rat).SetFrac(big.NewInt(1), b))
		}
		return nil
	},
	oGCD:        integerBinary("GCD", func(r, x, y *big.Int) *big.Int { return r.GCD(nil, nil, x, y) }),
	oLSH:        integerShift("LSH", (*big.Int).Lsh),
	oMOD:        integerDivision("MOD", (*big.Int).Mod),
	oMODINVERSE: integerBinary("MODINV", (*big.Int).ModInverse),
	oMULRANGE:   integerOverflow("MULRANGE", (*big.Int).MulRange),
	oNOT: func(e *Evaluator) error {
		x := e.Top()
		if a, ok := x.(*big.Int); ok {
			a.Not(a)
		} else {
			_ = x.(*big.Rat)
			return TypeError{"int"}
		}
		return nil
	},
	oOR:  integerBinary("OR", (*big.Int).Or),
	oREM: integerDivision("REM", (*big.Int).Rem),
	oRSH: integerShift("RSH", (*big.Int).Rsh),
	oXOR: integerBinary("XOR", (*big.Int).Xor),
	oDENOM: func(e *Evaluator) error {
		switch a := e.Top().(type) {
		case *big.Rat:
			e.SetTop(a.Denom())
		case *big.Int:
			a.SetUint64(1)
		default:
			panic("DENOM: wrong type on stack!")
		}
		return nil
	},
	oINV: func(e *Evaluator) error {
		switch i := e.Top().(type) {
		case *big.Int:
			if i.Sign() == 0 {
				return DivByZero{}
			}
			e.SetTop(new(big.Rat).SetFrac(big.NewInt(1), i))
		case *big.Rat:
			if i.Sign() == 0 {
				return DivByZero{}
			}
			i.Inv(i)
		default:
			panic("INV: wrong type on stack!")
		}
		return nil
	},
	oNUM: func(e *Evaluator) error {
		switch a := e.Top().(type) {
		case *big.Rat:
			e.SetTop(a.Num())
		case *big.Int:
			// do nothing
		default:
			panic("num: wrong type on stack!")
		}
		return nil
	},
	oTRUNC: numericRound("TRUNC", func(e *Evaluator, a *big.Rat) { e.SetTop(a.Num().Quo(a.Num(), a.Denom())) }),
	oFLOOR: numericRound("TRUNC", func(e *Evaluator, a *big.Rat) { e.SetTop(a.Num().Div(a.Num(), a.Denom())) }),
	oCEIL: numericRound("TRUNC", func(e *Evaluator, a *big.Rat) {
		q, r := a.Num().QuoRem(a.Num(), a.Denom(), new(big.Int))
		if r.Sign() > 0 {
			e.SetTop(q.Add(q, big.NewInt(1)))
		} else {
			e.SetTop(q)
		}
	}),
}

func numericUnary(name string, ints func(_, _ *big.Int) *big.Int, rats func(_, _ *big.Rat) *big.Rat) opFunc {
	return func(e *Evaluator) error {
		switch i := e.Top().(type) {
		case *big.Int:
			ints(i, i)
		case *big.Rat:
			rats(i, i)
		default:
			panic(name + ": wrong type on stack!")
		}
		return nil
	}
}

func numericBinary(name string, ints func(_, _, _ *big.Int) *big.Int, rats func(_, _, _ *big.Rat) *big.Rat) opFunc {
	return func(e *Evaluator) error {
		x := e.Pop()
		y := e.Top()
		switch a := x.(type) {
		case *big.Int:
			switch b := y.(type) {
			case *big.Int:
				ints(b, b, a)
			case *big.Rat:
				rats(b, b, new(big.Rat).SetFrac(a, big.NewInt(1)))
				if b.IsInt() {
					e.SetTop(b.Num())
				}
			default:
				panic(name + ": wrong type on stack! (int*?)")
			}
		case *big.Rat:
			switch b := y.(type) {
			case *big.Int:
				r := new(big.Rat).SetFrac(b, big.NewInt(1))
				rats(r, r, a)
				if r.IsInt() {
					e.SetTop(r.Num())
				} else {
					e.SetTop(r)
				}
			case *big.Rat:
				rats(b, b, a)
				if b.IsInt() {
					e.SetTop(b.Num())
				}
			default:
				panic(name + ": wrong type on stack! (rat*?)")
			}
		default:
			panic(name + ": wrong type on stack! (?*?)")
		}
		return nil
	}
}

func numericRound(name string, f func(*Evaluator, *big.Rat)) opFunc {
	return func(e *Evaluator) error {
		switch a := e.Top().(type) {
		case *big.Int: // do nothing
		case *big.Rat:
			f(e, a)
		default:
			panic(name + ": unknown type on stack!")
		}
		return nil
	}
}

func integerBinary(_ string, f func(_, _, _ *big.Int) *big.Int) opFunc {
	return func(e *Evaluator) error {
		x := e.Pop()
		y := e.Top()
		a, aok := x.(*big.Int)
		b, bok := y.(*big.Int)
		if !aok {
			_ = x.(*big.Rat) // TODO: make this error more informative
			return TypeError{"int"}
		}
		if !bok {
			_ = y.(*big.Rat)
			return TypeError{"int"}
		}
		f(b, b, a)
		return nil
	}
}

func integerDivision(_ string, f func(_, _, _ *big.Int) *big.Int) opFunc {
	return func(e *Evaluator) error {
		x := e.Pop()
		y := e.Top()
		a, aok := x.(*big.Int)
		b, bok := y.(*big.Int)
		if !aok {
			_ = x.(*big.Rat) // TODO: make this error more informative
			return TypeError{"int"}
		}
		if !bok {
			_ = y.(*big.Rat)
			return TypeError{"int"}
		}
		if a.Sign() == 0 {
			return DivByZero{}
		}
		f(b, b, a)
		return nil
	}
}

func integerOverflow(_ string, f func(_ *big.Int, _, _ int64) *big.Int) opFunc {
	return func(e *Evaluator) error {
		x := e.Pop()
		y := e.Top()
		a, aok := x.(*big.Int)
		b, bok := y.(*big.Int)
		if !aok {
			_ = x.(*big.Rat) // TODO: make this error more informative
			return TypeError{"int"}
		}
		if !bok {
			_ = y.(*big.Rat)
			return TypeError{"int"}
		}
		if toobig64(a) || toobig64(b) {
			return OverflowError{}
		}
		f(b, b.Int64(), a.Int64())
		return nil
	}
}

func integerShift(_ string, f func(_, _ *big.Int, _ uint) *big.Int) opFunc {
	return func(e *Evaluator) error {
		x := e.Pop()
		y := e.Top()
		a, aok := x.(*big.Int)
		b, bok := y.(*big.Int)
		if !aok {
			_ = x.(*big.Rat) // TODO: make this error more informative
			return TypeError{"int"}
		}
		if !bok {
			_ = y.(*big.Rat)
			return TypeError{"int"}
		}
		if toobiguint(a) || toobiguint(b) {
			return OverflowError{}
		}
		f(b, b, uint(a.Uint64()))
		return nil
	}
}

func toobig64(x *big.Int) bool {
	return x.Cmp(two63) >= 0
}

func toobiguint(x *big.Int) bool {
	return x.Cmp(uintmax) >= 0
}

var two63 = new(big.Int).SetUint64(1 << 63)
var uintmax = new(big.Int).Add(new(big.Int).SetUint64(uint64(^uint(0))), big.NewInt(1))
