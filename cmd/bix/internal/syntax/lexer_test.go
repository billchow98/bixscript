package syntax

import (
	"errors"
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"testing"
)

type errorCatcher struct {
	errors []error
}

type unwrapType int

const (
	doUnwrap unwrapType = iota
	doNotUnwrap
)

func (ec *errorCatcher) errorHandler(err error) {
	ec.errors = append(ec.errors, err)
}

func lexErrorTestRunner(t *testing.T, s string, want string, unwrap unwrapType) {
	ec := &errorCatcher{}
	l := newLexer(s, "lexer_test.go", ec.errorHandler)

	for !l.atEnd() {
		_ = l.NextToken()
	}

	errs := ec.errors

	if errs == nil {
		t.Errorf("No error occurred, want error %q", want)
	} else if len(errs) != 1 {
		t.Errorf("got %d errors, want 1 error", len(errs))

		for i, err := range errs {
			t.Errorf("error #%d: got error %q", i+1, errors.Unwrap(err))
		}
	} else {
		if unwrap == doUnwrap && errors.Unwrap(errs[0]).Error() != want {
			t.Errorf("got error %q, want error %q", errors.Unwrap(errs[0]), want)
		} else if unwrap == doNotUnwrap && errs[0].Error() != want {
			t.Errorf("got error %q, want error %q", errs[0], want)
		}
	}
}

func nextTokenTestRunner(t *testing.T, src string, want string, ty token.Type) {
	ec := &errorCatcher{}

	l := newLexer(src, "lexer_test.go", ec.errorHandler)

	if ec.errors != nil {
		t.Fatalf("%d errors occurred. Showing first error: %v", len(ec.errors), ec.errors[0])
	}

	got := l.NextToken()
	if got.Value != want {
		t.Fatalf("got %q, want %q", got.Value, want)
	}
	if got.Type != ty {
		t.Fatalf("%q: got %v, want %v", src, got.Type, ty)
	}
}

func TestNextToken(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		src, want string
		ty        token.Type
	}{
		{".", ".", token.Dot},
		{"123", "123", token.Number},
		{"+123", "+123", token.Number},
		{"-123", "-123", token.Number},
		{"123.456", "123.456", token.Number},
		{"123.", "123.", token.Number},
		{".456", ".456", token.Number},
		{"123\n", "123", token.Number},
		{`""`, `""`, token.String},
		{"\n", "\n", token.Newline},
		{"\r\n", "\r\n", token.Newline},
		{"\r\t\r\n", "\r\n", token.Newline},
		{`"\a\b\f\n\r\t\v\\\""`, `"\a\b\f\n\r\t\v\\\""`, token.String},
		{`"\x7f\u00A0\U0010ffff"`, `"\x7f\u00A0\U0010ffff"`, token.String},
		{"false", "false", token.False},
		{"for", "for", token.For},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v,%v", tt.src, tt.want, tt.ty)
		t.Run(name, func(t *testing.T) {
			nextTokenTestRunner(t, tt.src, tt.want, tt.ty)
		})
	}
}

func TestLexErrors(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"\"", "unterminated string"},
		{"\"\n", "newline in string"},
		{"\"\\z\"", "unknown escape"},
		{"\"\\xg0\"", "invalid character in hexadecimal escape"},
		{"\"\\x0G\"", "invalid character in hexadecimal escape"},
		{"\\\n\n", "empty line after '\\'"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			lexErrorTestRunner(t, tt.src, tt.want, doUnwrap)
		})
	}
}

func TestLexErrorsLineNumber(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"\\\n\n", "lexer_test.go:2:1: syntax error: empty line after '\\'"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v,%v", tt.src, tt.want)
		t.Run(name, func(t *testing.T) {
			lexErrorTestRunner(t, tt.src, tt.want, doNotUnwrap)
		})
	}
}
