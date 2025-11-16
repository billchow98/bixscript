package repl

import (
	"log"
	"strings"
	"unicode"

	"github.com/billchow98/bixscript/cmd/bix/internal/compiler"
	"github.com/billchow98/bixscript/cmd/bix/internal/runtime"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
)

// extractAllBytecode recursively extracts bytecode from a function and all nested functions
func extractAllBytecode(f *value.Function) string {
	var result strings.Builder

	// Extract nested functions first (they appear in the constant pool)
	for i := 0; i < len(f.Constants); i++ {
		if f.Constants[i].IsFunction() {
			nestedFunc := f.Constants[i].AsFunction()
			result.WriteString(extractAllBytecode(nestedFunc))
		}
	}

	// Then add this function's bytecode
	result.WriteString(f.DebugString())

	return result.String()
}

// Repl represents a stateful REPL session
type Repl struct {
	vm                *runtime.Vm
	inputBuffer       string
	leftBraceCount    int
	lastBytecode      string
	showBytecode      bool
}

// Result represents the result of executing code
type Result struct {
	Output       string   // Print output
	Bytecode     string   // Bytecode disassembly
	Errors       []string // Error messages
	NeedsContinuation bool // True if waiting for more input (multiline)
	PromptSymbol string   // ">>> " or "... "
}

// New creates a new Repl instance
func New() *Repl {
	return &Repl{
		vm:             runtime.New(),
		inputBuffer:    "",
		leftBraceCount: 0,
		lastBytecode:   "",
		showBytecode:   true, // On by default
	}
}

// AddLine adds a line of input and returns the result
// This is the main entry point for both terminal and WASM
func (r *Repl) AddLine(line string) Result {
	// Add line to buffer
	if r.inputBuffer != "" {
		r.inputBuffer += "\n"
	}
	r.inputBuffer += line

	// Track braces for multiline
	if str := strings.TrimLeftFunc(line, unicode.IsSpace); len(str) > 0 && str[0] == '}' {
		r.leftBraceCount--
	}

	needsContinuation := false
	if len(line) > 0 {
		switch line[len(line)-1] {
		case '{':
			r.leftBraceCount++
			needsContinuation = true
		case '\\':
			needsContinuation = true
		default:
			if r.leftBraceCount > 0 {
				needsContinuation = true
			}
		}
	}

	if needsContinuation {
		return Result{
			NeedsContinuation: true,
			PromptSymbol:      "... ",
		}
	}

	// Execute the complete input
	input := r.inputBuffer
	r.inputBuffer = ""
	r.leftBraceCount = 0

	return r.execute(input)
}

// execute runs the input and returns the result
func (r *Repl) execute(input string) Result {
	input = strings.TrimSpace(input)

	if input == "" {
		return Result{
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	}

	// Handle commands
	switch input {
	case "clear":
		return Result{
			Output:            "\x1b[2J\x1b[H", // ANSI clear screen
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	case "reset":
		r.vm = runtime.New()
		return Result{
			Output:            "\x1b[33mVM reset\x1b[0m\n",
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	case "bytecode on":
		r.showBytecode = true
		return Result{
			Output:            "\x1b[32mBytecode display enabled\x1b[0m\n",
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	case "bytecode off":
		r.showBytecode = false
		return Result{
			Output:            "\x1b[33mBytecode display disabled\x1b[0m\n",
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	case "syntax":
		syntax := "\x1b[1mBixScript Syntax Examples:\x1b[0m\n\n" +
			"\x1b[1mVariables:\x1b[0m\n" +
			"  let x = 42\n" +
			"  let name = \"BixScript\"\n" +
			"  x += 10\n\n" +
			"\x1b[1mControl Flow:\x1b[0m\n" +
			"  if x > 0 {\n" +
			"      print(\"positive\")\n" +
			"  } else {\n" +
			"      print(\"non-positive\")\n" +
			"  }\n\n" +
			"\x1b[1mLoops:\x1b[0m\n" +
			"  for let i = 0; i < 10; i += 1 {\n" +
			"      print(i)\n" +
			"  }\n" +
			"  while x > 0 {\n" +
			"      x -= 1\n" +
			"  }\n\n" +
			"\x1b[1mFunctions:\x1b[0m\n" +
			"  function add(a, b) {\n" +
			"      return a + b\n" +
			"  }\n" +
			"  print(add(5, 3))\n\n" +
			"\x1b[1mOperators:\x1b[0m\n" +
			"  Arithmetic: +, -, *, /, **\n" +
			"  Comparison: <, <=, ==, !=, >, >=\n" +
			"  Logical: and, or\n\n" +
			"\x1b[1mTry this (multiline):\x1b[0m\n" +
			"  function fib(n) {\n" +
			"      if n <= 1 {\n" +
			"          return n\n" +
			"      }\n" +
			"      return fib(n-1) + fib(n-2)\n" +
			"  }\n" +
			"  print(fib(10))\n"
		return Result{
			Output:            syntax,
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	case "help":
		help := "\x1b[1mBixScript REPL Commands:\x1b[0m\n" +
			"  clear        - Clear the terminal\n" +
			"  reset        - Reset VM state (clear all globals)\n" +
			"  syntax       - Show syntax examples\n" +
			"  bytecode on  - Enable automatic bytecode display\n" +
			"  bytecode off - Disable automatic bytecode display\n" +
			"  help         - Show this help message\n"
		return Result{
			Output:            help,
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	}

	// Compile without stderr capture to avoid duplicate output
	compiledFunc, compileErrs := compiler.Compile(input, "<stdin>", "<module>")

	var bytecode string
	if compiledFunc != nil {
		bytecode = extractAllBytecode(compiledFunc)
		r.lastBytecode = bytecode
	}

	return r.handleCompileResult(compiledFunc, compileErrs, bytecode, input)
}

// handleCompileResult processes compilation results and executes if successful
func (r *Repl) handleCompileResult(compiledFunc interface{}, compileErrs []error, bytecode string, input string) Result {
	var errStrs []string
	if compileErrs != nil {
		for _, err := range compileErrs {
			errStrs = append(errStrs, err.Error())
		}
		return Result{
			Bytecode:          bytecode,
			Errors:            errStrs,
			NeedsContinuation: false,
			PromptSymbol:      ">>> ",
		}
	}

	// Execute (captures output via vm.SetLogger)
	var outputBuf strings.Builder
	logger := log.New(&outputBuf, "", 0)
	r.vm.SetLogger(logger)

	execErrs := r.vm.Run(input, "<stdin>")

	if execErrs != nil {
		for _, err := range execErrs {
			errStrs = append(errStrs, err.Error())
		}
	}

	// Combine bytecode (from stderr during compilation) with execution output if enabled
	output := outputBuf.String()
	if r.showBytecode && bytecode != "" {
		// Bytecode was generated (function definition), show it in gray
		output = "\x1b[90m" + bytecode + "\x1b[0m" + output
	}

	return Result{
		Output:            output,
		Bytecode:          bytecode,
		Errors:            errStrs,
		NeedsContinuation: false,
		PromptSymbol:      ">>> ",
	}
}

// GetPrompt returns the current prompt symbol
func (r *Repl) GetPrompt() string {
	if r.inputBuffer != "" || r.leftBraceCount > 0 {
		return "... "
	}
	return ">>> "
}
