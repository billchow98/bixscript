package value

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"unsafe"
)

type Number float64

type heapNumber struct {
	tag tag
	n   Number
}

func init() {
	assert.Assert(unsafe.Offsetof(heapNumber{}.tag) == 0)
}

func newHeapNumber(n Number) *heapNumber {
	return &heapNumber{tag: BixNumber, n: n}
}
