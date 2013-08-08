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
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
)

// Compile an expression represented in Go syntax.
func CompileGo(expr string) (*Expr, error) {
	tree, err := parser.ParseExpr(expr)
	if err != nil {
		return nil, err
	}
	return CompileGoAST(tree)
}

// Compile a Go AST representation of an expression.
func CompileGoAST(node ast.Node) (*Expr, error) {
	exp := new(Expr)
	err := goast(node, exp)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

func goast(node ast.Node, e *Expr) error {
	switch nn := node.(type) {
	case *ast.Ident:
		e.ops = append(e.ops, oLOAD)
		e.names = append(e.names, nn.Name)
	case *ast.BasicLit:
		if nn.Kind == token.INT || nn.Kind == token.FLOAT {
			if x, ok := new(big.Rat).SetString(nn.Value); ok {
				e.ops = append(e.ops, oCONST)
				e.consts = append(e.consts, x)
			} else {
				panic("can this even happen?")
			}
		} else {
			return TypeError{"int or float"}
		}
	case *ast.BinaryExpr:
		if err := goast(nn.X, e); err != nil {
			return err
		}
		if err := goast(nn.Y, e); err != nil {
			return err
		}
		op := oNOP
		switch nn.Op {
		case token.ADD:
			op = oADD
		case token.MUL:
			op = oMUL
		case token.QUO:
			op = oQUO
		case token.SUB:
			op = oSUB
		case token.AND:
			op = oAND
		case token.AND_NOT:
			op = oANDNOT
		case token.SHL:
			op = oLSH
		case token.OR:
			op = oOR
		case token.REM:
			op = oREM
		case token.SHR:
			op = oRSH
		case token.XOR:
			op = oXOR
		default:
			panic("can this even happen?")
		}
		e.ops = append(e.ops, op)
	case *ast.UnaryExpr:
		if err := goast(nn.X, e); err != nil {
			return err
		}
		op := oNOP
		switch nn.Op {
		case token.ADD:
			break
		case token.SUB:
			op = oNEG
		case token.XOR:
			op = oNOT
		default:
			panic("can this even happen?")
		}
		e.ops = append(e.ops, op)
	case *ast.ParenExpr:
		if err := goast(nn.X, e); err != nil {
			return err
		}
	case *ast.CallExpr:
		if ident, ok := nn.Fun.(*ast.Ident); ok {
			op := oNOP
			m, n := -1, -1
			switch ident.Name {
			case "abs":
				op = oABS
				n = 1
			case "binomial":
				op = oBINOMIAL
				n = 2
			case "div":
				op = oDIV
				n = 2
			case "exp":
				op = oEXP
				m, n = 2, 3
			case "gcd":
				op = oGCD
				n = 2
			case "mod":
				op = oMOD
				n = 2
			case "modinv":
				op = oMODINVERSE
				n = 2
			case "mulrange":
				op = oMULRANGE
				n = 2
			case "rand":
				panic("not implemented yet")
			case "frac":
				op = oFRAC
				n = 2
			case "denom":
				op = oDENOM
				n = 1
			case "inv":
				op = oINV
				n = 1
			case "num":
				op = oNUM
				n = 1
			case "trunc":
				op = oTRUNC
				n = 1
			case "floor":
				op = oFLOOR
				n = 1
			case "ceil":
				op = oCEIL
				n = 1
			default:
				return BadGoToken{}
			}
			if n >= 0 {
				if m < 0 {
					if err := chkargs(nn, n, e); err != nil {
						return err
					}
					e.ops = append(e.ops, op)
				} else {
					if err := chkargs2(nn, m, n, e); err != nil {
						return err
					}
					e.ops = append(e.ops, op)
				}
			}
		} else {
			return BadGoToken{}
		}
	default:
		return BadGoToken{}
	}
	return nil
}

func chkargs(nn *ast.CallExpr, n int, e *Expr) error {
	if len(nn.Args) != n {
		return BadCall{n}
	}
	for i := 0; i < n; i++ {
		if err := goast(nn.Args[i], e); err != nil {
			return err
		}
	}
	return nil
}

func chkargs2(nn *ast.CallExpr, m, n int, e *Expr) error {
	if len(nn.Args) < m || len(nn.Args) > n {
		return BadCall{m}
	}
	for i := 0; i < n; i++ {
		if i >= len(nn.Args) {
			for i < n {
				e.ops = append(e.ops, oCONST)
				e.consts = append(e.consts, nil)
				i++
			}
			break
		} else if err := goast(nn.Args[i], e); err != nil {
			return err
		}
	}
	return nil
}
