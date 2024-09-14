// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package runtime

import (
	"errors"
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
)

type traceInfo struct {
	FunctionName *value.Identifier
	File         string
	Line         int
	Col          int
}

type trace struct {
	s     string
	infos []*traceInfo
}

func newTrace(s string) *trace {
	return &trace{s: s}
}

func (t *trace) append(functionName *value.Identifier, token *token.Token) {
	t.infos = append(t.infos, &traceInfo{functionName, token.File, token.Line, token.Col})
}

func (t *trace) error() error {
	pre := "runtime error: "

	suf := ""
	for _, t := range t.infos {
		suf += fmt.Sprintf("\n\tin %s (%s:%d:%d)", t.FunctionName, t.File, t.Line, t.Col)
	}

	return fmt.Errorf("%s%w%s", pre, errors.New(t.s), suf)
}
