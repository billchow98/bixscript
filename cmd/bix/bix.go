package main

import (
	"bufio"
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/runtime"
	"log"
	"os"
	"strings"
	"unicode"
)

const (
	VersionMajor = 0
	VersionMinor = 0
	VersionPatch = 1
)

func runRepl() {
	log.Printf("bix v%d.%d.%d\n", VersionMajor, VersionMinor, VersionPatch)

	v := runtime.New()

	for {
		s := bufio.NewScanner(os.Stdin)

		input := ""

		fmt.Print(">>> ")

		leftBraceCount := 0

	loop:
		for s.Scan() {
			input += s.Text() + string('\n')

			if str := strings.TrimLeftFunc(s.Text(), unicode.IsSpace); len(str) > 0 && str[0] == '}' {
				leftBraceCount--
			}

			if len(s.Text()) > 0 {
				switch s.Text()[len(s.Text())-1] {
				case '{':
					leftBraceCount++
					fmt.Print("... ")
				case '\\':
					fmt.Print("... ")
				default:
					if leftBraceCount > 0 {
						fmt.Print("... ")
					} else {
						break loop
					}
				}
			} else {
				break
			}
		}

		errs := v.Run(input, "<stdin>")
		if errs != nil {
			// REPLs are friendly!
			if errs != nil {
				for _, err := range errs {
					log.Println(err)
				}
			}
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
