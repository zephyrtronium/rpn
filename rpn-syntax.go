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
	l := lexer{expr, 0}
	for {
		t, err := l.next()
		switch t.kind {
		case tBAD:
			return nil, err
		case tLIT:
			e.ops = append(e.ops, oCONST)
			v, _ := new(big.Rat).SetString(t.val)
			e.consts = append(e.consts, v)
		case tOP:
			e.ops = append(e.ops, ops[t.val])
		case tIDENT:
			e.ops = append(e.ops, oLOAD)
			e.names = append(e.names, t.val)
		case tNIL:
			e.ops = append(e.ops, oCONST)
			e.consts = append(e.consts, nil)
		case tEOE:
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
	tEOE
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
	"RAND":     oRAND,
	"%":        oREM,
	"REM":      oREM,
	">>":       oRSH,
	"RSH":      oRSH,
	"^":        oXOR,
	"XOR":      oXOR,
	"FRAC":     oFRAC,
	"DENOM":    oDENOM,
	"INV":      oINV,
	"NUM":      oNUM,
}

func (l *lexer) next() (tok, error) {
	if len(l.src) == 0 {
		return tok{tEOE, ""}, nil
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
