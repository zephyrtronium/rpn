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
