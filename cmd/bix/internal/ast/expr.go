// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package ast

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
)

type Expr interface {
	Node
}

type BinaryExpr struct {
	Left  Expr
	Op    *token.Token
	Right Expr
}

type CallExpr struct {
	Function Expr
	LParen   *token.Token // For error reporting
	Args     []Expr
}

type FuncExpr struct {
	Function *token.Token // For error reporting
	Params   []*token.Token
	Body     *BlockStmt
}

type GroupExpr struct {
	LParen *token.Token // For error reporting
	Expr   Expr
}

type LiteralExpr struct {
	Value      value.Value
	ErrorToken *token.Token // For error reporting
}

type LogicalExpr struct {
	Left  Expr
	Op    *token.Token
	Right Expr
}

type UnaryExpr struct {
	Op    *token.Token
	Right Expr
}

type VariableExpr struct {
	Name *token.Token
}

func (e *BinaryExpr) Token() *token.Token {
	return e.Op
}

func (e *CallExpr) Token() *token.Token {
	return e.LParen
}

func (e *FuncExpr) Token() *token.Token {
	return e.Function
}

func (e *GroupExpr) Token() *token.Token {
	return e.LParen
}

func (e *LiteralExpr) Token() *token.Token {
	return e.ErrorToken
}

func (e *LogicalExpr) Token() *token.Token {
	return e.Op
}

func (e *UnaryExpr) Token() *token.Token {
	return e.Op
}

func (e *VariableExpr) Token() *token.Token {
	return e.Name
}
