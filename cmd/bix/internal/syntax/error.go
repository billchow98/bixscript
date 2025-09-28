package syntax

import (
	"errors"
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
)

// Not to be used directly. Call l.newError or p.newError instead.
func newError(s string, t *token.Token) error {
	return fmt.Errorf("%s:%d:%d: syntax error: %w", t.File, t.Line, t.Col, errors.New(s))
}
