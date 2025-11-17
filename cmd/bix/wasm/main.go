//go:build js && wasm
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
	repls       = make(map[int]*repl.Repl)
	replCounter = 0
	replMutex   sync.Mutex
)

type ReplResult struct {
	Output            string   `json:"output"`
	Bytecode          string   `json:"bytecode"`
	Errors            []string `json:"errors"`
	NeedsContinuation bool     `json:"needsContinuation"`
	PromptSymbol      string   `json:"promptSymbol"`
}

// InitRepl creates a new REPL instance and returns its ID
func InitRepl(this js.Value, args []js.Value) interface{} {
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
	jsonResult := ReplResult{
		Output:            result.Output,
		Bytecode:          result.Bytecode,
		Errors:            result.Errors,
		NeedsContinuation: result.NeedsContinuation,
		PromptSymbol:      result.PromptSymbol,
	}

	jsonBytes, _ := json.Marshal(jsonResult)
	return string(jsonBytes)
}

// ResetVm clears the global state of a REPL (creates new instance)
func ResetVm(this js.Value, args []js.Value) interface{} {
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

// DestroyVm removes a REPL instance
func DestroyVm(this js.Value, args []js.Value) interface{} {
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
	js.Global().Set("bixInitRepl", js.FuncOf(InitRepl))
	js.Global().Set("bixExecuteLine", js.FuncOf(ExecuteLine))
	js.Global().Set("bixResetVm", js.FuncOf(ResetVm))
	js.Global().Set("bixDestroyVm", js.FuncOf(DestroyVm))

	// Keep the program running
	<-c
}
