package compiler

import (
	"errors"
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
)

// To be used directly by compiler.go
func newError(s string, t *token.Token) error {
	return fmt.Errorf("%s:%d:%d: compiler error: %w", t.File, t.Line, t.Col, errors.New(s))
}
