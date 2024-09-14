// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package value

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"unsafe"
)

type Integer int64

type heapInteger struct {
	tag tag
	i   Integer
}

func init() {
	assert.Assert(unsafe.Offsetof(heapInteger{}.tag) == 0)
}

func newHeapInteger(i Integer) *heapInteger {
	return &heapInteger{tag: BixInteger, i: i}
}
