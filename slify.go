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

func foldConsts(nn *AST) {
	for _, child := range nn.Children {
		foldConsts(child)
	}
	switch nn.Op {
	case oNOP, oCONST, oLOAD: // do nothing
	case oABS, oNEG, oNOT, oDENOM, oINV, oNUM, oTRUNC, oFLOOR, oCEIL:
		child := nn.Children[0]
		if child.Op == oCONST {
			v := Evaluator{
				Stack: []interface{}{child.Val},
			}
			if opFuncs[nn.Op](&v) == nil {
				child.Parent = nil
				nn.Children[0] = nil
				nn.Children, nn.Op, nn.Val = nil, oCONST, v.Top()
			}
		}
	case oEXP:
		v := Evaluator{
			Stack: make([]interface{}, 0, 3),
		}
		for _, child := range nn.Children {
			if child.Op == oCONST {
				v.Stack = append(v.Stack, child.Val)
			}
		}
		if len(v.Stack) == 3 && opFuncs[nn.Op](&v) == nil {
			for i, child := range nn.Children {
				child.Parent = nil
				nn.Children[i] = nil
			}
			nn.Children, nn.Op, nn.Val = nil, oCONST, v.Top()
		}
	default:
		// binary operator
		ch1, ch2 := nn.Children[0], nn.Children[1]
		if ch1.Op == oCONST && ch2.Op == oCONST {
			v := Evaluator{
				Stack: []interface{}{ch1.Val, ch2.Val},
			}
			if opFuncs[nn.Op](&v) == nil {
				ch1.Parent, ch2.Parent = nil, nil
				nn.Children[0], nn.Children[1] = nil, nil
				nn.Children, nn.Op, nn.Val = nil, oCONST, v.Top()
			}
		}
	}
}

func redundant(nn *AST) (changed bool) {
	for _, child := range nn.Children {
		if redundant(child) {
			changed = true
		}
	}
	switch nn.Op {
	case oNOP: // do nothing
	case oNEG:
		// NEG NEG is redundant
		if x := nn.Children[0]; x.Op == oNEG {
			linkpast(nn, x.Children[0])
			return true
		} else if x.Op == oSUB {
			// -(x-y) == y-x
			x.Children[0], x.Children[1] = x.Children[1], x.Children[0]
			linkpast(nn, x)
			return true
		}
	case oINV:
		// INV INV is redundant
		if x := nn.Children[0]; x.Op == oINV {
			linkpast(nn, x.Children[0])
			return true
		} else if x.Op == oQUO {
			// 1/(x/y) == y/x
			x.Children[0], x.Children[1] = x.Children[1], x.Children[0]
			linkpast(nn, x)
			return true
		}
	case oMUL:
		x, y := nn.Children[0], nn.Children[1]
		switch {
		case x.Op == oNEG && y.Op == oNEG:
			// -x*-y == x*y
			linkpast(x, x.Children[0])
			linkpast(y, y.Children[0])
			return true
		case x.Op == oINV:
			if y.Op == oINV {
				// 1/x * 1/y == 1/(x*y)
				linkpast(x, x.Children[0])
				linkpast(y, y.Children[0])
				ins := &AST{oINV, nil, []*AST{nn}, nn.Parent}
				if nn.Parent != nil {
					nn.Parent.Children[findme(nn)] = ins
				}
				nn.Parent = ins
				return true
			} else {
				// 1/x * y == y/x
				linkpast(x, x.Children[0])
				nn.Children[0], nn.Children[1] = nn.Children[1], nn.Children[0]
				nn.Op = oQUO
				return true
			}
		case y.Op == oINV:
			// x * 1/y == x/y
			linkpast(y, y.Children[0])
			nn.Op = oQUO
			return true
		case x.Op == oCONST && eqone(x.Val):
			// 1 * y == y
			linkpast(nn, y)
			return true
		case y.Op == oCONST && eqone(y.Val):
			// x * 1 == x
			linkpast(nn, x)
			return true
		}
	case oQUO:
		x, y := nn.Children[0], nn.Children[1]
		switch {
		case x.Op == oNEG && y.Op == oNEG:
			// -x*-y == x*y
			linkpast(x, x.Children[0])
			linkpast(y, y.Children[0])
			return true
		case x.Op == oINV:
			if y.Op == oINV {
				// 1/x / 1/y == y/x
				linkpast(x, x.Children[0])
				linkpast(y, y.Children[0])
				nn.Children[0], nn.Children[1] = nn.Children[1], nn.Children[0]
				return true
			}
			// 1/x / y == 1/(x*y), but that's the same number of operations.
		case y.Op == oINV:
			// x / 1/y == x*y
			linkpast(y, y.Children[0])
			nn.Op = oMUL
			return true
		case x.Op == oCONST && eqone(x.Val):
			// 1 / y == 1/y (nowai)
			nn.Children[0], x.Parent = nil, nil
			nn.Op = oINV
			nn.Children = nn.Children[1:]
			return true
		case y.Op == oCONST && eqone(y.Val):
			// x / 1 == x
			linkpast(nn, x)
			return true
		}
	}
	return changed
}

func linkpast(nn, ch *AST) {
	ch.Parent = nn.Parent
	if nn.Parent != nil {
		nn.Parent.Children[findme(nn)] = ch
		nn.Parent = nil
	}
	for i := range nn.Children {
		nn.Children[i] = nil
	}
}

func findme(me *AST) int {
	if me.Parent == nil {
		return -1
	}
	for i, sibling := range me.Parent.Children {
		if sibling == me { // I'm turning into my brother! D:
			return i
		}
	}
	return -1
}

func eqone(val interface{}) bool {
	switch a := val.(type) {
	case *big.Int:
		if a.Cmp(intOne) == 0 {
			return true
		}
	case *big.Rat:
		if a.Cmp(ratOne) == 0 {
			return true
		}
	}
	return false
}

var intOne = big.NewInt(1)
var ratOne = big.NewRat(1, 1)
