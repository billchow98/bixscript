// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package ast

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
)

type Stmt interface {
	Node
}

type Decl interface {
	Stmt
	Name() *token.Token
}

type AssignStmt struct {
	Name   *token.Token
	Equals *token.Token // For error reporting
	Value  Expr
}

type BlockStmt struct {
	LBrace *token.Token // For error reporting
	Stmts  []Stmt
}

type ExprStmt struct {
	Expr Expr
}

type ForStmt struct {
	For  *token.Token // For error reporting
	Init Stmt         // Could be LetDecl or AssignStmt or ExprStmt
	Cond Expr
	Post Stmt // Could be AssignStmt or ExprStmt
	Body *BlockStmt
}

type FuncDecl struct {
	Function  *token.Token // For error reporting
	NameToken *token.Token
	Params    []*token.Token
	Body      *BlockStmt
}

type IfStmt struct {
	If   *token.Token // For error reporting
	Cond Expr
	Body *BlockStmt
	Else Stmt // Could be IfStmt or BlockStmt
}

type LetDecl struct {
	Let       *token.Token // For error reporting
	NameToken *token.Token
	Init      Expr
}

type ReturnStmt struct {
	Return *token.Token
	Expr   Expr
}

type WhileStmt struct {
	While *token.Token // For error reporting
	Cond  Expr
	Body  *BlockStmt
}

func (s *AssignStmt) Token() *token.Token {
	return s.Equals
}

func (s *BlockStmt) Token() *token.Token {
	return s.LBrace
}

func (s *ExprStmt) Token() *token.Token {
	return s.Expr.Token()
}

func (s *ForStmt) Token() *token.Token {
	return s.For
}

func (s *FuncDecl) Token() *token.Token {
	return s.Function
}

func (s *IfStmt) Token() *token.Token {
	return s.If
}

func (s *LetDecl) Token() *token.Token {
	return s.Let
}

func (s *ReturnStmt) Token() *token.Token {
	return s.Return
}

func (s *WhileStmt) Token() *token.Token {
	return s.While
}

func (s *FuncDecl) Name() *token.Token {
	return s.NameToken
}

func (s *LetDecl) Name() *token.Token {
	return s.NameToken
}
