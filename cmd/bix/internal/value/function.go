// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package value

import (
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"github.com/billchow98/bixscript/cmd/bix/internal/opcode"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"log"
	"math"
	"os"
	"unsafe"
)

// SignedOffsetBias
// Bias added when emitting bytecode, subtracted when executing it
const SignedOffsetBias = (math.MaxUint8 + 1) >> 1

type Function struct {
	tag            tag
	Name           *Identifier
	Argc           int
	Bytes          []byte
	Constants      []Value
	Tokens         map[int]*token.Token
	MaxStackUsage  int
	constantsIndex map[Value]byte
}

func init() {
	assert.Assert(unsafe.Offsetof(Function{}.tag) == 0)
}

// NewFunction
// Constants must have at least 1 element for VM's constants pointer to work
func NewFunction(name String) *Function {
	return &Function{tag: BixFunction, Name: NewIdentifier(name).AsIdentifier(), Argc: 0, Constants: make([]Value, 0, 1), Tokens: make(map[int]*token.Token), MaxStackUsage: 0, constantsIndex: make(map[Value]byte)}
}

func maxIntLength(i int) int {
	if i == 0 {
		return 1
	}
	return int(math.Floor(math.Log10(float64(i))) + 1)
}

func (f *Function) debugSimple(s string) (int, string, []any) {
	return 1, s, nil
}

func (f *Function) debugLdaConstant(i int) (int, string, []any) {
	return 2, "LdaConstant", []any{fmt.Sprintf("[%d]", f.Bytes[i+1])}
}

func (f *Function) debugBinaryOp(name string, i int) (int, string, []any) {
	return 2, name, []any{fmt.Sprintf("r%d", f.Bytes[i+1])}
}

func (f *Function) debugUnaryOp(name string) (int, string, []any) {
	return 1, name, nil
}

func (f *Function) debugGlobOp(name string, i int) (int, string, []any) {
	return 2, name, []any{fmt.Sprintf("[%d]", f.Bytes[i+1])}
}

func (f *Function) debugLocalOp(name string, i int) (int, string, []any) {
	offset := int(f.Bytes[i+1]) - SignedOffsetBias

	offsetStr := fmt.Sprintf("%d", offset)

	if offset < 0 {
		offsetStr = string('(') + offsetStr + string(')')
	}

	return 2, name, []any{fmt.Sprintf("r%s", offsetStr)}
}

func (f *Function) debugJumpOp(name string, i int) (int, string, []any) {
	return 2, name, []any{
		fmt.Sprintf("[%d]", 2+int(f.Bytes[i+1])-SignedOffsetBias),
		fmt.Sprintf("(%d)", i+2+int(f.Bytes[i+1])-SignedOffsetBias),
	}
}

func (f *Function) debugCallOp(name string, i int) (int, string, []any) {
	return 3, name, []any{fmt.Sprintf("[%d]", f.Bytes[i+1]), fmt.Sprintf("[%d]", f.Bytes[i+2])}
}

func (f *Function) DebugString() string {
	const maxN = 3

	// Disassembly of bytecode

	s := fmt.Sprintf("Disassembly (function %s)\n", f.Name)

	for i, n := 0, 0; i < len(f.Bytes); i += n {
		s += fmt.Sprintf("%0*d\t", maxIntLength(len(f.Bytes)), i)

		var code string
		var args []any

		switch opcode.Code(f.Bytes[i]) {
		case opcode.LdaConstant:
			// n: number of bytes used by instruction
			n, code, args = f.debugLdaConstant(i)
		case opcode.Add:
			n, code, args = f.debugBinaryOp("Add", i)
		case opcode.Sub:
			n, code, args = f.debugBinaryOp("Sub", i)
		case opcode.Mul:
			n, code, args = f.debugBinaryOp("Mul", i)
		case opcode.Div:
			n, code, args = f.debugBinaryOp("Div", i)
		case opcode.Pow:
			n, code, args = f.debugBinaryOp("Pow", i)
		case opcode.Neg:
			n, code, args = f.debugUnaryOp("Neg")
		case opcode.Not:
			n, code, args = f.debugUnaryOp("Not")
		case opcode.Eq:
			n, code, args = f.debugBinaryOp("Eq", i)
		case opcode.Neq:
			n, code, args = f.debugBinaryOp("Neq", i)
		case opcode.CmpLt:
			n, code, args = f.debugBinaryOp("CmpLt", i)
		case opcode.CmpLe:
			n, code, args = f.debugBinaryOp("CmpLe", i)
		case opcode.CmpGt:
			n, code, args = f.debugBinaryOp("CmpGt", i)
		case opcode.CmpGe:
			n, code, args = f.debugBinaryOp("CmpGe", i)
		case opcode.GlobDecl:
			n, code, args = f.debugGlobOp("GlobDecl", i)
		case opcode.StaGlobal:
			n, code, args = f.debugGlobOp("StaGlobal", i)
		case opcode.LdaGlobal:
			n, code, args = f.debugGlobOp("LdaGlobal", i)
		case opcode.Star:
			n, code, args = f.debugLocalOp("Star", i)
		case opcode.Ldar:
			n, code, args = f.debugLocalOp("Ldar", i)
		case opcode.Jump:
			n, code, args = f.debugJumpOp("Jump", i)
		case opcode.JumpIfTrue:
			n, code, args = f.debugJumpOp("JumpIfTrue", i)
		case opcode.JumpIfFalse:
			n, code, args = f.debugJumpOp("JumpIfFalse", i)
		case opcode.Call:
			n, code, args = f.debugCallOp("Call", i)
		case opcode.Return:
			n, code, args = f.debugSimple("Return")
		case opcode.LdaUndefined:
			n, code, args = f.debugSimple("LdaUndefined")
		default:
			assert.Assert(false) // Unreachable
		}

		assert.Assert(n <= maxN)

		for j := 0; j < maxN; j++ {
			switch {
			case j == 0:
				s += fmt.Sprintf("%02x", f.Bytes[i+j])
			case j < n:
				s += fmt.Sprintf(" %02x", f.Bytes[i+j])
			default:
				s += fmt.Sprintf("%3s", "")
			}
		}
		s += string('\t')

		s += code
		s += string(' ')

		if len(args) != 0 {
			for j := 0; j < len(args); j++ {
				if j == 0 {
					s += fmt.Sprintf("%v", args[j])
				} else {
					s += fmt.Sprintf(", %v", args[j])
				}
			}
		}

		s += string('\n')
	}

	// Constant pool

	s += fmt.Sprintf("Constant pool (size %d)\n", len(f.Constants))

	for i := 0; i < len(f.Constants); i++ {
		s += fmt.Sprintf("%0*d\t", maxIntLength(len(f.Constants)), i)

		switch {
		case f.Constants[i].IsNumber():
			s += "<Number>     \t"
		case f.Constants[i].IsBoolean():
			s += "<Boolean>    \t"
		case f.Constants[i].IsString():
			s += "<String>     \t"
		case f.Constants[i].IsUndefined():
			s += "<Undefined>  \t"
		case f.Constants[i].IsIdentifier():
			s += "<Identifier> \t"
		case f.Constants[i].IsFunction():
			s += "<Function>   \t"
		default:
			assert.Assert(false) // Unreachable
		}

		s += fmt.Sprintf("%s\n", f.Constants[i].DebugString())
	}

	return s
}

func (f *Function) appendCode(t *token.Token, byte1 byte, bytes ...byte) {
	_, found := f.Tokens[len(f.Bytes)]
	assert.Assert(!found)

	f.Tokens[len(f.Bytes)] = t
	f.Bytes = append(f.Bytes, append([]byte{byte1}, bytes...)...)
}

func (f *Function) addConstant(value Value) byte {
	if i, ok := f.constantsIndex[value]; ok {
		return i
	}

	f.Constants = append(f.Constants, value)
	if len(f.Constants)-1 > math.MaxUint8 {
		log.Fatalf("Too many constants in one bytecode. sorry, unimplemented")
	}

	f.constantsIndex[value] = byte(len(f.Constants) - 1)
	return f.constantsIndex[value]
}

func (f *Function) AddLdaConstant(t *token.Token, value Value) {
	f.appendCode(t, byte(opcode.LdaConstant), f.addConstant(value))
}

func (f *Function) addBinOp(op opcode.Code, t *token.Token, i int) {
	assert.Assert(int(byte(i)) == i)
	f.appendCode(t, byte(op), byte(i))
}

func (f *Function) AddAdd(t *token.Token, i int) {
	f.addBinOp(opcode.Add, t, i)
}

func (f *Function) AddSub(t *token.Token, i int) {
	f.addBinOp(opcode.Sub, t, i)
}

func (f *Function) AddMul(t *token.Token, i int) {
	f.addBinOp(opcode.Mul, t, i)
}

func (f *Function) AddDiv(t *token.Token, i int) {
	f.addBinOp(opcode.Div, t, i)
}

func (f *Function) AddPow(t *token.Token, i int) {
	f.addBinOp(opcode.Pow, t, i)
}

func (f *Function) AddNeg(t *token.Token) {
	f.appendCode(t, byte(opcode.Neg))
}

func (f *Function) AddNot(t *token.Token) {
	f.appendCode(t, byte(opcode.Not))
}

func (f *Function) AddEq(t *token.Token, i int) {
	f.addBinOp(opcode.Eq, t, i)
}

func (f *Function) AddNeq(t *token.Token, i int) {
	f.addBinOp(opcode.Neq, t, i)
}

func (f *Function) AddCmpLt(t *token.Token, i int) {
	f.addBinOp(opcode.CmpLt, t, i)
}

func (f *Function) AddCmpLe(t *token.Token, i int) {
	f.addBinOp(opcode.CmpLe, t, i)
}

func (f *Function) AddCmpGt(t *token.Token, i int) {
	f.addBinOp(opcode.CmpGt, t, i)
}

func (f *Function) AddCmpGe(t *token.Token, i int) {
	f.addBinOp(opcode.CmpGe, t, i)
}

func (f *Function) AddGlobDecl(t *token.Token, name String) {
	nameIdx := f.addConstant(NewIdentifier(name))
	f.appendCode(t, byte(opcode.GlobDecl), nameIdx)
}

func (f *Function) AddStaGlobal(t *token.Token, name String) {
	nameIdx := f.addConstant(NewIdentifier(name))
	f.appendCode(t, byte(opcode.StaGlobal), nameIdx)
}

func (f *Function) AddLdaGlobal(t *token.Token, name String) {
	nameIdx := f.addConstant(NewIdentifier(name))
	f.appendCode(t, byte(opcode.LdaGlobal), nameIdx)
}

func (f *Function) addRegOp(op opcode.Code, t *token.Token, i int) {
	offset := i + SignedOffsetBias

	if offset > math.MaxUint8 {
		log.Fatalf("Register index too large. sorry, unimplemented")
	}

	if offset < 0 {
		log.Fatalf("Register index too small. sorry, unimplemented")
	}

	assert.Assert(int(byte(offset)) == offset)
	f.appendCode(t, byte(op), byte(offset))
}

func (f *Function) AddStar(t *token.Token, i int) {
	f.addRegOp(opcode.Star, t, i)
}

func (f *Function) AddLdar(t *token.Token, i int) {
	f.addRegOp(opcode.Ldar, t, i)
}

// jumpOffset
// from: index of jump source
// to: index of jump destination
func jumpOffset(from int, to int) int {
	// -1 for VM's pc pointing at next byte
	// -1 for current fixed 1-byte length of jump offset to be read
	return to - from + SignedOffsetBias - 2
}

// addJump
// For the group of AddJump* functions, a valid index means a backward jump to the instruction
// beginning with that index
// An invalid index will leave the bytecode open for patching later (forward jump) and
// returns the index of the jump instruction
func (f *Function) addJump(op opcode.Code, t *token.Token) int {
	f.appendCode(t, byte(op), byte(0))
	return len(f.Bytes) - 2
}

func (f *Function) GetJumpIndex() int {
	return len(f.Bytes)
}

// AddJump
// For the group of AddJump* functions, a valid index means a backward jump
// An invalid index will leave the bytecode open for patching later (forward jump)
func (f *Function) AddJump(t *token.Token) int {
	return f.addJump(opcode.Jump, t)
}

// AddJumpIfTrue
// For the group of AddJump* functions, a valid index means a backward jump
// An invalid index will leave the bytecode open for patching later (forward jump)
func (f *Function) AddJumpIfTrue(t *token.Token) int {
	return f.addJump(opcode.JumpIfTrue, t)
}

// AddJumpIfFalse
// For the group of AddJump* functions, a valid index means a backward jump
// An invalid index will leave the bytecode open for patching later (forward jump)
func (f *Function) AddJumpIfFalse(t *token.Token) int {
	return f.addJump(opcode.JumpIfFalse, t)
}

// PatchJump
// from: index of jump instruction
// to: index of target instruction
func (f *Function) PatchJump(from int, to int) {
	offset := jumpOffset(from, to)

	if offset > math.MaxUint8 {
		log.Fatalf("Forward jump too large. sorry, unimplemented")
	}

	if offset < 0 {
		log.Fatalf("Backward jump too large. sorry, unimplemented")
	}

	assert.Assert(int(byte(offset)) == offset)
	// For now, we have a fixed 1-byte offset for all jumps
	f.Bytes[from+1] = byte(offset)
}

func (f *Function) AddCall(t *token.Token, i int, argc int) {
	assert.Assert(int(byte(i)) == i)
	assert.Assert(int(byte(argc)) == argc)
	f.appendCode(t, byte(opcode.Call), byte(i), byte(argc))
}

func (f *Function) AddReturn(t *token.Token) {
	f.appendCode(t, byte(opcode.Return))
}

func (f *Function) AddLdaUndefined(t *token.Token) {
	f.appendCode(t, byte(opcode.LdaUndefined))
}

func (f *Function) Finalise() {
	f.AddLdaUndefined(nil)
	f.AddReturn(nil)
	f.constantsIndex = nil
	_, _ = fmt.Fprintf(os.Stderr, "%s", f.DebugString())
}
