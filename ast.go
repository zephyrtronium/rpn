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

// Abstract syntax tree.
type AST struct {
	Op       operator
	Val      interface{}
	Children []*AST
	Parent   *AST
}

func getast(e *Evaluator, ops []operator) (int, *AST) {
	switch op := ops[len(ops)-1]; op {
	case oNOP:
		n, nn := getast(e, ops[:len(ops)-1])
		return 1 + n, nn
	case oLOAD:
		nn := &AST{op, e.Names[len(e.Names)-e.N-1], nil, nil}
		e.N++
		return 1, nn
	case oCONST:
		nn := &AST{op, e.Consts[len(e.Consts)-e.C-1], nil, nil}
		e.C++
		return 1, nn
	case oABS, oNEG, oNOT, oDENOM, oINV, oNUM, oTRUNC, oFLOOR, oCEIL:
		n, child := getast(e, ops[:len(ops)-1])
		nn := &AST{Op: op, Children: []*AST{child}}
		child.Parent = nn
		return 1 + n, nn
	case oEXP:
		n1, child1 := getast(e, ops[:len(ops)-1])
		n2, child2 := getast(e, ops[:len(ops)-n1-1])
		n3, child3 := getast(e, ops[:len(ops)-n2-n1-1])
		nn := &AST{Op: op, Children: []*AST{child1, child2, child3}}
		child1.Parent, child2.Parent, child3.Parent = nn, nn, nn
		return 1 + n1 + n2 + n3, nn
	default:
		// binary operator
		n1, child1 := getast(e, ops[:len(ops)-1])
		n2, child2 := getast(e, ops[:len(ops)-n1-1])
		nn := &AST{Op: op, Children: []*AST{child1, child2}}
		child1.Parent, child2.Parent = nn, nn
		return 1 + n1 + n2, nn
	}
}

// Compile an AST back into an evaluable expression.
func (nn *AST) RPN(e *Expr) {
	switch nn.Op {
	case oNOP: // do nothing
	case oLOAD:
		e.names = append(e.names, nn.Val.(string))
	case oCONST:
		e.consts = append(e.consts, nn.Val.(*big.Rat))
	default:
		for _, child := range nn.Children {
			child.RPN(e)
		}
	}
	e.ops = append(e.ops, nn.Op)
}
