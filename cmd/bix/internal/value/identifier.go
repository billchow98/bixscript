package value

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"sync"
	"unsafe"
)

// For now, only Identifiers are interned

type Identifier struct {
	tag tag
	s   String
}

func init() {
	assert.Assert(unsafe.Offsetof(Identifier{}.tag) == 0)
}

var (
	strings = make(map[String]*Identifier)
	mu      sync.Mutex
)

func NewIdentifier(s String) Value {
	mu.Lock()
	defer mu.Unlock()

	if interned, ok := strings[s]; ok {
		return fromObject(Pointer(interned))
	}

	id := Identifier{tag: BixIdentifier, s: s}
	strings[s] = &id
	return fromObject(Pointer(&id))
}

func (i *Identifier) String() string {
	return string(i.s)
}
