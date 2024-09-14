// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package runtime

import (
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"github.com/billchow98/bixscript/cmd/bix/internal/compiler"
	"github.com/billchow98/bixscript/cmd/bix/internal/lib"
	"github.com/billchow98/bixscript/cmd/bix/internal/opcode"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
	"log"
	"math"
	"sort"
	"unsafe"
)

type Vm struct {
	acc       value.Value
	stack     []value.Value
	globals   map[*value.Identifier]value.Value
	fp        value.Integer
	constants *value.Value
	function  *value.Function
}

type tokenMapElement struct {
	Pc         int
	ErrorToken *token.Token
}

// New
// Initialise map here. Will last across multiple REPL inputs.
func New() *Vm {
	v := &Vm{globals: make(map[*value.Identifier]value.Value)}
	lib.PopulateNatives(v)
	return v
}

func (v *Vm) RegisterNativeFunction(name *value.Identifier, function value.Value) {
	_, ok := v.globals[name]
	assert.Assert(!ok)

	v.globals[name] = function
}

func (v *Vm) SetLogger(l *log.Logger) {
	lib.SetLogger(l)
}

// Before every run, so we can't initialise globals table here
// as they will be reset after each REPL input
func (v *Vm) prepareRun() {
	v.acc = value.Undefined()
	v.fp = 0
	v.constants = nil
	v.function = nil
}

// Ensure runtime stack has enough space for all local and temporary variables
// We use append here because this might be done on a call stack
// where we still need to preserve existing data
func (v *Vm) growStack() {
	maxStackUsage := v.function.MaxStackUsage
	newSize := int(v.fp) + maxStackUsage
	n := len(v.stack)

	if newSize > n {
		newStack := make([]value.Value, 2*newSize)
		copy(newStack, v.stack)
		v.stack = newStack
		newStack = v.stack[n:]
		newStack[0] = value.Undefined()
		// https://gist.github.com/seeker815/118bb152242cfe476545a5ade43da0b8
		for i := 1; ; i *= 2 {
			copy(newStack[i:], newStack[:i])

			// Proof this covers everything:
			// len(newStack) <= 2 * i
			// len(newStack)-i <= i
			if 2*i >= len(newStack) {
				copy(newStack[len(newStack)-i:], newStack[:i])
				break
			}
		}
	}

	assert.Assert(int(v.fp)+maxStackUsage <= len(v.stack))
}

func (v *Vm) stackAt(i int) *value.Value {
	return &v.stack[int(v.fp)+i]
}

func (v *Vm) setFunction(f *value.Function) {
	v.function = f
	v.constants = &v.function.Constants[:1][0]
}

// No discard! Always update pc with the value here!
func (v *Vm) doReturn() *byte {
	v.setFunction(v.stackAt(compiler.FuncOffset).AsFunction())
	pc := *(**byte)(unsafe.Pointer(v.stackAt(compiler.PcOffset)))
	// This has to come last, or it will affect previous v.stackAt calls
	v.fp = v.stackAt(compiler.FpOffset).AsInteger()
	return pc
}

func (v *Vm) getTokens() []tokenMapElement {
	var tokens []tokenMapElement

	for k, v := range v.function.Tokens {
		tokens = append(tokens, tokenMapElement{k, v})
	}

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Pc < tokens[j].Pc
	})

	return tokens
}

func (v *Vm) initialPc() *byte {
	return &v.function.Bytes[0]
}

func pcAdd(oldPc *byte, i int) *byte {
	return (*byte)(unsafe.Add(unsafe.Pointer(oldPc), i))
}

func (v *Vm) newError(s string, pc *byte) []error {
	tr := newTrace(s)

	found := false
	last := false // true when we want to append one last trace

	// Unwind stack
	for {
		tokens := v.getTokens()

		last = v.fp == 0

		for i := len(tokens) - 1; i >= 0; i-- {
			// Strict inequality because by the time the error occurs,
			// pc already points to the next instruction
			if uintptr(unsafe.Pointer(pcAdd(v.initialPc(), tokens[i].Pc))) < uintptr(unsafe.Pointer(pc)) {
				t := tokens[i].ErrorToken
				assert.Assert(t != nil)
				tr.append(v.function.Name, t)
				found = true
				break
			}
		}

		if last {
			break
		}

		pc = v.doReturn()
	}

	assert.Assert(found) // Impossible not to have a corresponding error token

	return []error{tr.error()}
}

func (v *Vm) Run(s string, file string) []error {
	v.prepareRun()

	c, errs := compiler.Compile(s, file, "<module>")
	if errs != nil {
		return errs
	}

	v.setFunction(c)
	pc := v.initialPc()

	v.growStack()

	newError := func(s string) []error {
		return v.newError(s, pc)
	}

	nextByte := func() byte {
		b := *pc
		pc = pcAdd(pc, 1)
		return b
	}

	nextConstant := func() value.Value {
		return *(*value.Value)(unsafe.Add(unsafe.Pointer(v.constants), uintptr(nextByte())*unsafe.Sizeof(value.Value{})))
	}

loop:
	for {
		switch opcode.Code(nextByte()) {
		case opcode.LdaConstant:
			v.acc = nextConstant()
		case opcode.Add:
			i := int(nextByte())

			switch {
			case v.stackAt(i).IsNumber() && v.acc.IsNumber():
				v.acc = value.FromNumber(v.stackAt(i).AsNumber() + v.acc.AsNumber())
			case v.stackAt(i).IsString() && v.acc.IsString():
				// v.acc must be on the right because it always stores the RHS of a binary expression
				v.acc = value.FromString(v.stackAt(i).AsString() + v.acc.AsString())
			default:
				return newError("addition arguments must both be Numbers or both be Strings")
			}
		case opcode.Sub:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both subtraction arguments must be Numbers")
			}

			v.acc = value.FromNumber(v.stackAt(i).AsNumber() - v.acc.AsNumber())
		case opcode.Mul:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both multiplication arguments must be Numbers")
			}

			v.acc = value.FromNumber(v.stackAt(i).AsNumber() * v.acc.AsNumber())
		case opcode.Div:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both division arguments must be Numbers")
			}

			v.acc = value.FromNumber(v.stackAt(i).AsNumber() / v.acc.AsNumber())
		case opcode.Pow:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both exponentiation arguments must be Numbers")
			}

			v.acc = value.FromNumber(value.Number(math.Pow(float64(v.acc.AsNumber()), float64(v.stackAt(i).AsNumber()))))
		case opcode.Neg:
			if !v.acc.IsNumber() {
				return newError("negation argument must be a Number")
			}

			v.acc = value.FromNumber(-v.acc.AsNumber())
		case opcode.Not:
			if !v.acc.IsBoolean() {
				return newError("logical not argument must be a Boolean")
			}

			v.acc = value.FromBoolean(!v.acc.AsBoolean())
		case opcode.Eq:
			i := int(nextByte())

			v.acc = value.FromBoolean(v.stackAt(i).Equals(v.acc))
		case opcode.Neq:
			i := int(nextByte())

			v.acc = value.FromBoolean(!v.stackAt(i).Equals(v.acc))
		case opcode.CmpLt:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both comparison arguments must be Numbers")
			}

			v.acc = value.FromBoolean(v.stackAt(i).AsNumber() < v.acc.AsNumber())
		case opcode.CmpLe:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both comparison arguments must be Numbers")
			}

			v.acc = value.FromBoolean(v.stackAt(i).AsNumber() <= v.acc.AsNumber())
		case opcode.CmpGt:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both comparison arguments must be Numbers")
			}

			v.acc = value.FromBoolean(v.stackAt(i).AsNumber() > v.acc.AsNumber())
		case opcode.CmpGe:
			i := int(nextByte())

			if !v.stackAt(i).IsNumber() || !v.acc.IsNumber() {
				return newError("both comparison arguments must be Numbers")
			}

			v.acc = value.FromBoolean(v.stackAt(i).AsNumber() >= v.acc.AsNumber())
		case opcode.GlobDecl:
			v.globals[nextConstant().AsIdentifier()] = value.Undefined()
		case opcode.StaGlobal:
			name := nextConstant().AsIdentifier()

			val, found := v.globals[name]
			if !found {
				return newError("assignment to undeclared variable")
			}

			if v.acc.IsUndefined() {
				return newError("cannot set variable to undefined")
			}

			if !val.IsUndefined() && !val.SameTypeAs(v.acc) {
				return newError("variable and assigned value have different types")
			}

			v.globals[name] = v.acc
		case opcode.LdaGlobal:
			name := nextConstant().AsIdentifier()

			_, found := v.globals[name]
			if !found {
				return newError("use of undeclared variable")
			}

			v.acc = v.globals[name]
		case opcode.Star:
			i := int(nextByte()) - value.SignedOffsetBias

			*v.stackAt(i) = v.acc
		case opcode.Ldar:
			i := int(nextByte()) - value.SignedOffsetBias

			v.acc = *v.stackAt(i)
		case opcode.Jump:
			offset := int(nextByte()) - value.SignedOffsetBias
			pc = pcAdd(pc, offset)
		case opcode.JumpIfTrue:
			offset := int(nextByte()) - value.SignedOffsetBias

			if !v.acc.IsBoolean() {
				return newError("condition expression is not a Boolean")
			}

			if v.acc.AsBoolean() {
				pc = pcAdd(pc, offset)
			}
		case opcode.JumpIfFalse:
			offset := int(nextByte()) - value.SignedOffsetBias

			if !v.acc.IsBoolean() {
				return newError("condition expression is not a Boolean")
			}

			if !v.acc.AsBoolean() {
				pc = pcAdd(pc, offset)
			}
		case opcode.Call:
			// Offset of new fp after call
			offset := int(nextByte())

			// Get argc
			argc := int(nextByte())

			// Store pc
			*v.stackAt(offset + compiler.PcOffset) = *(*value.Value)(value.Pointer(&pc))

			// Store fp
			*v.stackAt(offset + compiler.FpOffset) = value.FromInteger(v.fp)

			// Store old function pointer
			*v.stackAt(offset + compiler.FuncOffset) = value.FromFunction(v.function)

			// Do actual call
			v.fp += value.Integer(offset)

			// Assumes pc is the first item after the arguments list
			f := v.stackAt(compiler.PcOffset - argc - 1)
			switch {
			case f.IsFunction():
				v.setFunction(f.AsFunction())
				pc = v.initialPc()

				// Check number of arguments passed
				funcArgc := v.function.Argc
				if funcArgc != lib.VarArgc && argc != funcArgc {
					pc = v.doReturn()
					return newError(fmt.Sprintf("expected %d arguments, got %d", funcArgc, argc))
				}

				v.growStack()
			case f.IsNativeFunction():
				native := f.AsNativeFunction()

				// Check number of arguments passed
				funcArgc := native.Argc
				if funcArgc != lib.VarArgc && argc != funcArgc {
					return newError(fmt.Sprintf("expected %d arguments, got %d", funcArgc, argc))
				}

				v.acc = native.Function(argc, v.stackAt(compiler.PcOffset-argc))

				pc = v.doReturn()
			default:
				pc = v.doReturn()
				return newError("expression is not callable")
			}
		case opcode.Return:
			if v.fp == 0 {
				break loop
			}

			pc = v.doReturn()
		case opcode.LdaUndefined:
			v.acc = value.Undefined()
		default:
			assert.Assert(false) // Unreachable
		}
	}

	return nil
}
