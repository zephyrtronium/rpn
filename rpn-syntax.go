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
	"math/big"
	"strings"
	"unicode"
	"unicode/utf8"
)

// parser

// Compile an expression represented in reverse Polish notation. This will
// correctly compile the output of an expression's String() method, and allows
// alternate representations of some things. See
// https://github.com/zephyrtronium/rpn/wiki/RPN-Syntax for more information.
func CompileRPN(expr string) (*Expr, error) {
	e := new(Expr)
	l := lexer{strings.TrimSpace(expr), 0}
	stack := 0
	for {
		t, err := l.next()
		switch t.kind {
		case tBAD:
			return nil, err
		case tLIT:
			e.ops = append(e.ops, oCONST)
			v, _ := new(big.Rat).SetString(t.val)
			e.consts = append(e.consts, v)
			stack++
		case tOP:
			op := ops[t.val]
			switch op {
			case oNOP: // do nothing
			case oABS, oNEG, oNOT, oDENOM, oINV, oNUM, oTRUNC, oFLOOR, oCEIL:
				if stack < 1 {
					return nil, StackError{t.val, l.pos}
				}
			case oEXP:
				stack -= 2
				if stack < 1 {
					return nil, StackError{t.val, l.pos}
				}
			default:
				// binary operator
				stack--
				if stack < 1 {
					return nil, StackError{t.val, l.pos}
				}
			}
			e.ops = append(e.ops, op)
		case tIDENT:
			e.ops = append(e.ops, oLOAD)
			e.names = append(e.names, t.val)
			stack++
		case tNIL:
			e.ops = append(e.ops, oCONST)
			e.consts = append(e.consts, nil)
			stack++
		case tEND:
			if stack > 1 {
				return e, LargeStack{}
			}
			return e, nil
		}
	}
}

// lexer

type lexer struct {
	src string
	pos int
}

type tok struct {
	kind int
	val  string
}

const (
	tBAD = iota
	tLIT
	tOP
	tIDENT
	tNIL
	tEND
)

var ops = map[string]operator{
	"NOP":      oNOP,
	"ABS":      oABS,
	"+":        oADD,
	"ADD":      oADD,
	"*":        oMUL,
	"MUL":      oMUL,
	"NEG":      oNEG,
	"/":        oQUO,
	"QUO":      oQUO,
	"-":        oSUB,
	"SUB":      oSUB,
	"&":        oAND,
	"AND":      oAND,
	"&^":       oANDNOT,
	"ANDNOT":   oANDNOT,
	"BINOMIAL": oBINOMIAL,
	"DIV":      oDIV,
	"EXP":      oEXP,
	"GCD":      oGCD,
	"<<":       oLSH,
	"LSH":      oLSH,
	"MOD":      oMOD,
	"MODINV":   oMODINVERSE,
	"MULRANGE": oMULRANGE,
	"NOT":      oNOT,
	"|":        oOR,
	"OR":       oOR,
	"%":        oREM,
	"REM":      oREM,
	">>":       oRSH,
	"RSH":      oRSH,
	"^":        oXOR,
	"XOR":      oXOR,
	"DENOM":    oDENOM,
	"INV":      oINV,
	"NUM":      oNUM,
	"TRUNC":    oTRUNC,
	"FLOOR":    oFLOOR,
	"CEIL":     oCEIL,
}

func (l *lexer) next() (tok, error) {
	if len(l.src) == 0 {
		return tok{tEND, ""}, nil
	}
	off := strings.IndexFunc(l.src, unicode.IsSpace)
	if off < 0 {
		// end of string
		t, err := l.lexWord(l.src)
		l.src = ""
		return t, err
	}
	s := l.src[:off]
	l.src = l.src[off+1:]
	l.pos += off + 1
	off = strings.IndexFunc(l.src, func(r rune) bool { return !unicode.IsSpace(r) })
	if off < 0 {
		l.src = ""
	} else {
		l.src = l.src[off:]
		l.pos += off
	}
	return l.lexWord(s)
}

func (l *lexer) lexWord(s string) (tok, error) {
	if _, ok := ops[strings.ToUpper(s)]; ok {
		return tok{tOP, strings.ToUpper(s)}, nil
	}
	if nam, ok := lexIdent(s); ok {
		return tok{tIDENT, nam}, nil
	}
	if s == "_" || strings.EqualFold(s, "<nil>") {
		return tok{tNIL, s}, nil
	}
	if _, ok := ParseConst(s); ok {
		return tok{tLIT, s}, nil
	}
	return tok{tBAD, s}, BadRPNToken{s, l.pos}
}

func lexIdent(src string) (string, bool) {
	if src == "_" {
		return "", false
	}
	src = strings.TrimSuffix(strings.TrimPrefix(src, "("), ")")
	if len(src) < 1 {
		return "", false
	}
	if strings.IndexFunc(src, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r)) && r != '_'
	}) >= 0 {
		return "", false
	}
	if r, _ := utf8.DecodeRuneInString(src); unicode.IsDigit(r) {
		return "", false
	}
	return src, true
}
