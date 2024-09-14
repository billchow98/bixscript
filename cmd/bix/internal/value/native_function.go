// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package value

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"unsafe"
)

type NativeFunctionPtr func(argc int, argv *Value) Value
type NativeFunction struct {
	tag      tag
	Name     *Identifier
	Argc     int
	Function NativeFunctionPtr
}

func init() {
	assert.Assert(unsafe.Offsetof(NativeFunction{}.tag) == 0)
}

func NewNativeFunction(name *Identifier, argc int, function NativeFunctionPtr) *NativeFunction {
	return &NativeFunction{BixNativeFunction, name, argc, function}
}
