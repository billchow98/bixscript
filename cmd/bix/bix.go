package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/billchow98/bixscript/cmd/bix/internal/repl"
	"github.com/billchow98/bixscript/cmd/bix/internal/runtime"
)

var (
	// Version is set at build time via -ldflags
	Version = "dev"
)

func runRepl() {
	fmt.Printf("bix %s\n", Version)
	fmt.Println()
	fmt.Println("Type BixScript code and press Enter to execute.")
	fmt.Println("Commands: syntax, clear, reset, bytecode [on|off], help")
	fmt.Println()

	r := repl.New()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("\x1b[1;36m>>> \x1b[0m")

	for scanner.Scan() {
		line := scanner.Text()

		result := r.AddLine(line)

		// Print output if any (contains ANSI color codes)
		if result.Output != "" {
			fmt.Print(result.Output)
		}

		// Print errors if any (in red)
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				fmt.Printf("\x1b[31m%s\x1b[0m\n", err)
			}
		}

		// Print next prompt with color
		if result.PromptSymbol == ">>> " {
			fmt.Print("\x1b[1;36m>>> \x1b[0m")
		} else {
			fmt.Print("\x1b[1;33m... \x1b[0m")
		}
	}
}

func runFile() {
	filename := os.Args[1]

	bytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}

	input := string(bytes)

	v := runtime.New()

	errs := v.Run(input, filename)
	if errs != nil {
		for i, err := range errs {
			switch i {
			case len(errs) - 1:
				log.Fatalln(err)
			default:
				log.Println(err)
			}
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	switch len(os.Args) {
	case 1:
		runRepl()
	case 2:
		runFile()
	default:
		fmt.Println("Usage: bix [FILE]")
		os.Exit(1)
	}
}
