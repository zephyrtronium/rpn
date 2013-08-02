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

// A compiled expression.
type Expr struct {
	ops    []operator
	names  []string
	consts []big.Rat
}

// Evaluate an expression with variables given in vars.
func (e *Expr) Eval(vars map[string]interface{}) (result *big.Rat, err error) {
	stack := make([]interface{}, 0, len(e.ops))
	names, consts := e.names, e.consts
	for _, op := range e.ops {
		switch op {
		case oNOP: // do nothing
		case oLOAD:
			if names[0] == "" {
				stack = append(stack, nil)
			} else {
				v := vars[names[0]]
				switch i := v.(type) {
				case *big.Int:
					stack = append(stack, new(big.Int).Set(i))
				case *big.Rat:
					stack = append(stack, new(big.Rat).Set(i))
				default:
					return nil, MissingVar{names[0]}
				}
			}
			names = names[1:]
		case oCONST:
			v := &consts[0]
			if v.IsInt() {
				stack = append(stack, new(big.Int).Set(v.Num()))
			} else {
				stack = append(stack, new(big.Rat).Set(v))
			}
			consts = consts[1:]
		case oABS:
			switch i := top(stack).(type) {
			case *big.Int:
				i.Abs(i)
			case *big.Rat:
				i.Abs(i)
			default:
				panic("abs: wrong type on stack!")
			}
		case oADD:
			x := pop(&stack)
			y := top(stack)
			switch a := x.(type) {
			case *big.Int:
				switch b := y.(type) {
				case *big.Int:
					b.Add(b, a)
				case *big.Rat:
					b.Add(b, new(big.Rat).SetFrac(a, big.NewInt(1)))
				default:
					panic("add: wrong type on stack! (int+?)")
				}
			case *big.Rat:
				switch b := y.(type) {
				case *big.Int:
					r := new(big.Rat).SetFrac(b, big.NewInt(1))
					r.Add(r, a)
					stack[len(stack)-1] = r
				case *big.Rat:
					b.Add(b, a)
					if b.IsInt() {
						stack[len(stack)-1] = b.Num()
					}
				default:
					panic("add: wrong type on stack! (rat+?)")
				}
			default:
				panic("add: wrong type on stack! (?+?)")
			}
		case oMUL:
			x := pop(&stack)
			y := top(stack)
			switch a := x.(type) {
			case *big.Int:
				switch b := y.(type) {
				case *big.Int:
					b.Mul(b, a)
				case *big.Rat:
					b.Mul(b, new(big.Rat).SetFrac(a, big.NewInt(1)))
					if b.IsInt() {
						stack[len(stack)-1] = b.Num()
					}
				default:
					panic("mul: wrong type on stack! (int*?)")
				}
			case *big.Rat:
				switch b := y.(type) {
				case *big.Int:
					r := new(big.Rat).SetFrac(b, big.NewInt(1))
					r.Mul(r, a)
					if r.IsInt() {
						stack[len(stack)-1] = r.Num()
					} else {
						stack[len(stack)-1] = r
					}
				case *big.Rat:
					b.Mul(b, a)
					if b.IsInt() {
						stack[len(stack)-1] = b.Num()
					}
				default:
					panic("mul: wrong type on stack! (rat*?)")
				}
			default:
				panic("mul: wrong type on stack! (?*?)")
			}
		case oNEG:
			switch i := top(stack).(type) {
			case *big.Int:
				i.Neg(i)
			case *big.Rat:
				i.Neg(i)
			default:
				panic("neg: wrong type on stack!")
			}
		case oQUO:
			x := pop(&stack)
			y := top(stack)
			switch a := x.(type) {
			case *big.Int:
				switch b := y.(type) {
				case *big.Int:
					b.Quo(b, a)
				case *big.Rat:
					b.Quo(b, new(big.Rat).SetFrac(a, big.NewInt(1)))
					if b.IsInt() {
						stack[len(stack)-1] = b.Num()
					}
				default:
					panic("quo: wrong type on stack! (int/?)")
				}
			case *big.Rat:
				switch b := y.(type) {
				case *big.Int:
					r := new(big.Rat).SetFrac(b, big.NewInt(1))
					r.Quo(r, a)
					if r.IsInt() {
						stack[len(stack)-1] = r.Num()
					} else {
						stack[len(stack)-1] = r
					}
				case *big.Rat:
					b.Quo(b, a)
					if b.IsInt() {
						stack[len(stack)-1] = b.Num()
					}
				default:
					panic("quo: wrong type on stack! (rat/?)")
				}
			default:
				panic("quo: wrong type on stack! (?/?)")
			}
		case oSUB:
			x := pop(&stack)
			y := top(stack)
			switch a := x.(type) {
			case *big.Int:
				switch b := y.(type) {
				case *big.Int:
					b.Sub(b, a)
				case *big.Rat:
					b.Sub(b, new(big.Rat).SetFrac(a, big.NewInt(1)))
				default:
					panic("sub: wrong type on stack! (int-?)")
				}
			case *big.Rat:
				switch b := y.(type) {
				case *big.Int:
					r := new(big.Rat).SetFrac(b, big.NewInt(1))
					r.Sub(r, a)
					stack[len(stack)-1] = r
				case *big.Rat:
					b.Sub(b, a)
					if b.IsInt() {
						stack[len(stack)-1] = b.Num()
					}
				default:
					panic("sub: wrong type on stack! (rat-?)")
				}
			default:
				panic("sub: wrong type on stack! (?-?)")
			}
		case oAND:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat) // TODO: make this error more informative
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.And(b, a)
		case oANDNOT:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.AndNot(b, a)
		case oBINOMIAL:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if toobig(a) || toobig(b) {
				return nil, OverflowError{}
			}
			b.Binomial(b.Int64(), a.Int64())
		case oDIV:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.Div(b, a)
		case oEXP:
			m := pop(&stack)
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			c, cok := m.(*big.Int) // heh
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if m != nil && !cok { // heh
				_ = m.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.Exp(b, a, c)
		case oGCD:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.GCD(nil, nil, b, a)
		case oLSH:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if toobiguint(a) {
				return nil, OverflowError{}
			}
			b.Lsh(b, uint(a.Uint64()))
		case oMOD:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.Mod(b, a)
		case oMODINVERSE:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.ModInverse(b, a)
		case oMULRANGE:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if toobig(a) || toobig(b) {
				return nil, OverflowError{}
			}
			b.MulRange(b.Int64(), a.Int64())
		case oNOT:
			x := top(stack)
			if a, ok := x.(*big.Int); ok {
				a.Not(a)
			} else {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
		case oOR:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.Or(b, a)
		case oRAND:
			panic("not implemented yet")
		case oREM:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.Rem(b, a)
		case oRSH:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if toobiguint(a) {
				return nil, OverflowError{}
			}
			b.Rsh(b, uint(a.Uint64()))
		case oXOR:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			b.Xor(b, a)
		case oFRAC:
			x := pop(&stack)
			y := top(stack)
			a, aok := x.(*big.Int)
			b, bok := y.(*big.Int)
			if !aok {
				_ = x.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if !bok {
				_ = y.(*big.Rat)
				return nil, TypeError{"int"}
			}
			if a.Sign() == 0 {
				return nil, DivByZero{}
			}
			stack[len(stack)-1] = new(big.Rat).SetFrac(b, a)
		case oDENOM:
			x := top(stack)
			if a, ok := x.(*big.Rat); ok {
				stack[len(stack)-1] = a.Denom()
			} else {
				return nil, TypeError{"rat"}
			}
		case oINV:
			// Because x 1 oQUO may be optimized to x INV, we need to handle
			// both int and rat.
			switch i := top(stack).(type) {
			case *big.Int:
				if i.Sign() == 0 {
					return nil, DivByZero{}
				}
				i.Quo(big.NewInt(1), i)
			case *big.Rat:
				if i.Sign() == 0 {
					return nil, DivByZero{}
				}
				i.Inv(i)
			default:
				panic("inv: wrong type on stack!")
			}
		case oNUM:
			x := top(stack)
			if a, ok := x.(*big.Rat); ok {
				stack[len(stack)-1] = a.Num()
			} else {
				return nil, TypeError{"rat"}
			}
		default:
			panic("unknown op!")
		}
	}
	switch x := stack[0].(type) {
	case *big.Int:
		return new(big.Rat).SetFrac(x, big.NewInt(1)), nil
	case *big.Rat:
		return new(big.Rat).Set(x), nil
	default:
		panic("wrong type on stack! (return)")
	}
}

// Compute a list of names of variable names in the expression.
func (e Expr) Vars() []string {
	m := make(map[string]struct{})
	for _, k := range e.names {
		m[k] = struct{}{}
	}
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

// TODO: Expr.String()

func top(stack []interface{}) interface{} {
	return stack[len(stack)-1]
}

func pop(stack *[]interface{}) interface{} {
	v := top(*stack)
	*stack = (*stack)[:len(*stack)-1]
	return v
}

func toobig(x *big.Int) bool {
	return x.Cmp(two63) >= 0
}

func toobiguint(x *big.Int) bool {
	return x.Cmp(uintmax) >= 0
}

var two63 = new(big.Int).SetUint64(1 << 63)
var uintmax = new(big.Int).Add(new(big.Int).SetUint64(uint64(^uint(0))), big.NewInt(1))
