package lib

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
	"log"
	"os"
	"sync"
	"unsafe"
)

const (
	VarArgc = -1
)

var (
	natives = make(map[*value.Identifier]value.Value)
	mu      sync.Mutex
	logger  *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", 0)
}

func PopulateNatives(vm vmer) {
	for k, v := range natives {
		vm.RegisterNativeFunction(k, v)
	}
}

func SetLogger(l *log.Logger) {
	logger = l
}

func registerNative(name value.String, argc int, function value.NativeFunctionPtr) {
	mu.Lock()
	defer mu.Unlock()

	// Ensures name is interned
	identifier := value.NewIdentifier(name).AsIdentifier()

	_, ok := natives[identifier]
	assert.Assert(!ok)

	natives[identifier] = value.FromNativeFunction(value.NewNativeFunction(identifier, argc, function))
}

func getArg(argv *value.Value, i int) value.Value {
	return unsafe.Slice(argv, i+1)[i]
}
