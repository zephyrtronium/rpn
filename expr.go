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
	v := Evaluator{
		Stack:  make([]interface{}, 0, len(e.ops)),
		Vars:   vars,
		Names:  e.names,
		Consts: e.consts,
	}
	if err = v.Eval(e.ops); err != nil {
		return nil, err
	}
	switch x := v.Top().(type) {
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
