package value

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"unsafe"
)

type Pointer unsafe.Pointer

type heapPointer struct {
	tag tag
	p   Pointer
}

func init() {
	assert.Assert(unsafe.Offsetof(heapPointer{}.tag) == 0)
}

func newHeapPointer(p Pointer) *heapPointer {
	return &heapPointer{tag: BixPointer, p: p}
}
