package ast

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
)

type Node interface {
	// Token
	// Bear in mind that this function is only used for
	// runtime error messages so just think about how
	// you want line numbers to be displayed
	Token() *token.Token
}

type Visitor interface {
	visit(node Node)
}
