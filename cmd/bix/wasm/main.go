// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"syscall/js"

	"github.com/billchow98/bixscript/cmd/bix/internal/repl"
)

var (
	// Version is set at build time via -ldflags
	Version = "dev"
)

var (
	repls     = make(map[int]*repl.Repl)
	replCounter = 0
	replMutex  sync.Mutex
)

type REPLResult struct {
	Output            string   `json:"output"`
	Bytecode          string   `json:"bytecode"`
	Errors            []string `json:"errors"`
	NeedsContinuation bool     `json:"needsContinuation"`
	PromptSymbol      string   `json:"promptSymbol"`
}

// InitREPL creates a new REPL instance and returns its ID
func InitREPL(this js.Value, args []js.Value) interface{} {
	replMutex.Lock()
	defer replMutex.Unlock()

	replID := replCounter
	replCounter++

	r := repl.New()
	repls[replID] = r

	return replID
}

// ExecuteLine adds a line of input and returns the result
func ExecuteLine(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return `{"errors":["Invalid arguments"]}`
	}

	replID := args[0].Int()
	line := args[1].String()

	replMutex.Lock()
	r, exists := repls[replID]
	replMutex.Unlock()

	if !exists {
		return `{"errors":["Invalid REPL ID"]}`
	}

	// Add line to REPL
	result := r.AddLine(line)

	// Convert to JSON response
	jsonResult := REPLResult{
		Output:            result.Output,
		Bytecode:          result.Bytecode,
		Errors:            result.Errors,
		NeedsContinuation: result.NeedsContinuation,
		PromptSymbol:      result.PromptSymbol,
	}

	jsonBytes, _ := json.Marshal(jsonResult)
	return string(jsonBytes)
}

// ResetVM clears the global state of a REPL (creates new instance)
func ResetVM(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return false
	}

	replID := args[0].Int()

	replMutex.Lock()
	defer replMutex.Unlock()

	if _, exists := repls[replID]; exists {
		// Create new REPL to reset state
		repls[replID] = repl.New()
		return true
	}

	return false
}

// DestroyVM removes a REPL instance
func DestroyVM(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return false
	}

	replID := args[0].Int()

	replMutex.Lock()
	defer replMutex.Unlock()

	if _, exists := repls[replID]; exists {
		delete(repls, replID)
		return true
	}

	return false
}

// GetVersion returns the version string
func GetVersion(this js.Value, args []js.Value) interface{} {
	return Version
}

func main() {
	c := make(chan struct{})

	fmt.Println("BixScript WASM initialized")

	// Register JavaScript functions
	js.Global().Set("bixGetVersion", js.FuncOf(GetVersion))
	js.Global().Set("bixInitREPL", js.FuncOf(InitREPL))
	js.Global().Set("bixExecuteLine", js.FuncOf(ExecuteLine))
	js.Global().Set("bixResetVM", js.FuncOf(ResetVM))
	js.Global().Set("bixDestroyVM", js.FuncOf(DestroyVM))

	// Keep the program running
	<-c
}
