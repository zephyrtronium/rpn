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
)

// Parse a literal into either a *big.Int or *big.Rat. s may be an integer
// literal with case-insensitive prefix 0x for hexadecimal, 0 for octal, 0b
// for binary, and decimal otherwise; or it may be a fraction of two decimal
// integers separated by a /; or it may be a floating-point constant. The
// returned interface{} is the value of appropriate type, and the bool
// indicates success.
func ParseConst(s string) (interface{}, bool) {
	if strings.IndexAny(s, "./") != -1 || strings.IndexAny(s, "eE") != -1 && !(len(s) > 2 && strings.EqualFold(s[:2], "0x")) {
		return new(big.Rat).SetString(s)
	}
	return new(big.Int).SetString(s, 0)
}

// Panic if err is not nil; otherwise return e. Must(CompileGo(expr)) can be
// used to compile an expression when failure is already fatal.
func Must(e *Expr, err error) *Expr {
	if err != nil {
		panic(err)
	}
	return e
}
