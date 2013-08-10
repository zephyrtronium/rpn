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
