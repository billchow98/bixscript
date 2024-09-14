// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package value

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"unsafe"
)

type tag int16

const (
	BixUndefined tag = iota // Default tag
	BixInteger
	BixNumber
	BixBoolean
	BixString
	BixIdentifier // Internal use
	BixFunction
	BixPointer
	BixNativeFunction
)

type simpleObject struct {
	tag tag
}

func init() {
	assert.Assert(unsafe.Offsetof(simpleObject{}.tag) == 0)
}

var (
	undefinedObject = &simpleObject{tag: BixUndefined}
	trueObject      = &simpleObject{tag: BixBoolean}
	falseObject     = &simpleObject{tag: BixBoolean}
)
