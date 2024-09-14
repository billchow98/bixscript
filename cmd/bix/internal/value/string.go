// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package value

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"unsafe"
)

type String string

type heapString struct {
	tag tag
	s   String
}

func init() {
	assert.Assert(unsafe.Offsetof(heapString{}.tag) == 0)
}

func newHeapString(s String) *heapString {
	return &heapString{tag: BixString, s: s}
}
