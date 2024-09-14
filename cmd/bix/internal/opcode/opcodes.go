// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package opcode

type Code byte

const (
	LdaConstant Code = iota // Load accumulator from constant
	Add
	Sub
	Mul
	Div
	Pow
	Neg
	Not
	Eq
	Neq
	CmpLt
	CmpLe
	CmpGt
	CmpGe
	GlobDecl
	StaGlobal
	LdaGlobal
	Star // Store accumulator in register
	Ldar // Load accumulator from register
	Jump
	JumpIfTrue
	JumpIfFalse
	Call
	Return
	LdaUndefined
)
