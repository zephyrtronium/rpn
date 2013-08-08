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

import (
	"bytes"
	"fmt"
	"math/big"
)

// A compiled expression.
type Expr struct {
	ops    []operator
	names  []string
	consts []*big.Rat
}

// Evaluate an expression with variables given in vars.
func (e *Expr) Eval(vars map[string]interface{}) (result *big.Rat, err error) {
	stack := make([]interface{}, 0, len(e.ops))
	names, consts := e.names, e.consts
	for _, op := range e.ops {
		switch op {
		case oNOP: // do nothing
		case oLOAD:
			v := vars[names[0]]
			switch i := v.(type) {
			case *big.Int:
				stack = append(stack, new(big.Int).Set(i))
			case *big.Rat:
				stack = append(stack, new(big.Rat).Set(i))
			default:
				return nil, MissingVar{names[0]}
			}
			names = names[1:]
		case oCONST:
			v := consts[0]
			if v == nil {
				stack = append(stack, nil)
			} else if v.IsInt() {
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
					r := new(big.Rat).SetFrac(b, a)
					if r.IsInt() {
						b.Set(r.Num())
					} else {
						stack[len(stack)-1] = r
					}
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
			if toobig64(a) || toobig64(b) {
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
			if toobig64(a) || toobig64(b) {
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
		case oDENOM:
			x := top(stack)
			switch a := x.(type) {
			case *big.Rat:
				stack[len(stack)-1] = a.Denom()
			case *big.Int:
				a.SetUint64(1)
			default:
				panic("denom: wrong type on stack!")
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
			switch a := x.(type) {
			case *big.Rat:
				stack[len(stack)-1] = a.Num()
			case *big.Int:
				// do nothing
			default:
				panic("num: wrong type on stack!")
			}
		case oTRUNC:
			switch a := top(stack).(type) {
			case *big.Int: // do nothing
			case *big.Rat:
				stack[len(stack)-1] = a.Num().Quo(a.Num(), a.Denom())
			default:
				panic("trunc: unknown type on stack!")
			}
		case oFLOOR:
			switch a := top(stack).(type) {
			case *big.Int: // do nothing
			case *big.Rat:
				stack[len(stack)-1] = a.Num().Div(a.Num(), a.Denom())
			default:
				panic("floor: unknown type on stack!")
			}
		case oCEIL:
			switch a := top(stack).(type) {
			case *big.Int: // do nothing
			case *big.Rat:
				q, r := a.Num().QuoRem(a.Num(), a.Denom(), new(big.Int))
				if r.Sign() > 0 {
					stack[len(stack)-1] = q.Add(q, big.NewInt(1))
				} else {
					stack[len(stack)-1] = q
				}
			}
		default:
			panic("unknown op!")
		}
	}
	switch x := top(stack).(type) {
	case *big.Int:
		return new(big.Rat).SetFrac(x, big.NewInt(1)), nil
	case *big.Rat:
		return new(big.Rat).Set(x), nil
	default:
		panic("wrong type on stack! (return)")
	}
}

// Compute a list of names of variable names in the expression.
func (e *Expr) Vars() []string {
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

// Show the compiled RPN expression.
func (e *Expr) String() string {
	names, consts := e.names, e.consts
	var buf bytes.Buffer
	first := true
	for _, op := range e.ops {
		var s string
		switch op {
		case oNOP:
			s = "NOP"
		case oLOAD:
			s, names = fmt.Sprintf("(%s)", names[0]), names[1:]
		case oCONST:
			if c := consts[0]; c == nil {
				s = "<nil>"
			} else {
				s = consts[0].RatString()
			}
			consts = consts[1:]
		case oABS:
			s = "ABS"
		case oADD:
			s = "+"
		case oMUL:
			s = "*"
		case oNEG:
			s = "NEG"
		case oQUO:
			s = "/"
		case oSUB:
			s = "-"
		case oAND:
			s = "&"
		case oANDNOT:
			s = "&^"
		case oBINOMIAL:
			s = "BINOMIAL"
		case oDIV:
			s = "DIV"
		case oEXP:
			s = "EXP"
		case oGCD:
			s = "GCD"
		case oLSH:
			s = "<<"
		case oMOD:
			s = "MOD"
		case oMODINVERSE:
			s = "MODINV"
		case oMULRANGE:
			s = "MULRANGE"
		case oNOT:
			s = "NOT"
		case oOR:
			s = "|"
		case oREM:
			s = "%"
		case oRSH:
			s = ">>"
		case oXOR:
			s = "^"
		case oDENOM:
			s = "DENOM"
		case oINV:
			s = "INV"
		case oNUM:
			s = "NUM"
		case oTRUNC:
			s = "TRUNC"
		case oFLOOR:
			s = "FLOOR"
		case oCEIL:
			s = "CEIL"
		default:
			panic("unknown op!")
		}
		if !first {
			buf.WriteByte(' ')
		}
		buf.WriteString(s)
		first = false
	}
	return buf.String()
}

func top(stack []interface{}) interface{} {
	return stack[len(stack)-1]
}

func pop(stack *[]interface{}) interface{} {
	v := top(*stack)
	*stack = (*stack)[:len(*stack)-1]
	return v
}

func toobig64(x *big.Int) bool {
	return x.Cmp(two63) >= 0
}

func toobiguint(x *big.Int) bool {
	return x.Cmp(uintmax) >= 0
}

var two63 = new(big.Int).SetUint64(1 << 63)
var uintmax = new(big.Int).Add(new(big.Int).SetUint64(uint64(^uint(0))), big.NewInt(1))
