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

type (
	// A variable in the expression is not in those used for evaluation.
	MissingVar struct {
		Name string
	}

	// A variable of incorrect type is on the stack, or a literal of
	// unsupported type has been parsed.
	TypeError struct {
		Needed string
	}

	// A shift, range multiplication, or binomial coefficent calculation has
	// an operand which is too large.
	OverflowError struct{}

	// Division or modulus by zero.
	DivByZero struct{}

	// An unknown token was parsed by the Go parser.
	BadToken struct{}
)

func (m MissingVar) Error() string  { return "missing var " + m.Name }
func (t TypeError) Error() string   { return "incorrect type; needed " + t.Needed }
func (OverflowError) Error() string { return "overflow" }
func (DivByZero) Error() string     { return "division by zero" }
func (BadToken) Error() string      { return "unrecognized token" }
