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

package main

import (
	"fmt"
	"github.com/zephyrtronium/rpn"
	"math/big"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	useRPN := false
	if args[0] == "-rpn" {
		useRPN = true
		args = args[1:]
	}
	vars := make(map[string]interface{})
	for _, v := range args[1:] {
		i := strings.Index(v, "=")
		var ok bool
		vars[v[:i]], ok = rpn.ParseConst(v[i+1:])
		if !ok {
			panic(v[i+1:])
		}
	}
	f := rpn.CompileGo
	if useRPN {
		f = rpn.CompileRPN
	}
	expr, err := f(args[0])
	if err != nil {
		if _, ok := err.(rpn.LargeStack); !ok {
			panic(err)
		}
		fmt.Println(err)
	}
	fmt.Println(expr)
	expr.Slify()
	var res *big.Rat
	res, err = expr.Eval(vars)
	if err != nil {
		panic(err)
	}
	fmt.Println(expr)
	fmt.Println(res.RatString())
}
