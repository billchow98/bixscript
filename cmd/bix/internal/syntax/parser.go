// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package syntax

import (
	"errors"
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"github.com/billchow98/bixscript/cmd/bix/internal/ast"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
	"log"
	"math"
	"strconv"
)

type Parser struct {
	lexer    *lexer
	stmts    []ast.Stmt
	previous *token.Token
	current  *token.Token
	// It's possible for us to accumulate multiple errors per statement
	// This is when one statement contains multiple lexer errors
	errors []error
}

type FunctionType int

const (
	FunctionDecl FunctionType = iota
	Method
	FunctionExpr
)

func newParser(s string, file string) *Parser {
	p := &Parser{}
	p.lexer = newLexer(s, file, func(e error) { p.errors = append(p.errors, e) })
	return p
}

func (p *Parser) addUnwrappedError(e error, t *token.Token) {
	if p.errors == nil {
		p.errors = append(p.errors, newError(e.Error(), t))
	}
}

func (p *Parser) newError(s string, t *token.Token) {
	p.addUnwrappedError(errors.New(s), t)
}

func (p *Parser) nextToken() *token.Token {
	p.previous = p.current
	p.current = p.lexer.NextToken()
	return p.current
}

func (p *Parser) peekMatchesOne(types ...token.Type) bool {
	for _, t := range types {
		if p.current.Type == t {
			return true
		}
	}

	return false
}

// Note: Moves to next token if match with p.current is successful
func (p *Parser) matchesOne(types ...token.Type) bool {
	for _, t := range types {
		if p.current.Type == t {
			_ = p.nextToken()
			return true
		}
	}

	return false
}

// Note: Moves to next token if match with p.current is successful
func (p *Parser) expect(msg string, types ...token.Type) {
	for _, t := range types {
		if p.matchesOne(t) {
			return
		}
	}

	p.newError(msg, p.current)
}

func (p *Parser) parseNumber() value.Number {
	n, err := strconv.ParseFloat(p.previous.Value, 64)
	if err != nil {
		p.addUnwrappedError(err, p.previous)
		return 0
	}
	return value.Number(n)
}

func (p *Parser) parseFunction(t FunctionType) ast.Node {
	funcToken := p.previous

	var typeStr string

	switch t {
	case FunctionDecl, FunctionExpr:
		typeStr = "function"
	case Method:
		typeStr = "method"
	default:
		assert.Assert(false) // Unreachable
	}

	var name *token.Token

	if t == FunctionExpr {
		p.expect("expected '('", token.LeftParen)
	} else {
		p.expect(fmt.Sprintf("expected %s name", typeStr), token.Identifier)

		name = p.previous

		p.expect(fmt.Sprintf("expected '(' after %s name", typeStr), token.LeftParen)
	}

	var params []*token.Token

	for p.matchesOne(token.Identifier) {
		params = append(params, p.previous)

		// 255 because 0 is also a possible parameter count
		if len(params) > math.MaxUint8 {
			log.Fatalf("%s has more than %d parameters. sorry, unimplemented", typeStr, math.MaxUint8)
		}

		if !p.matchesOne(token.Comma) {
			break
		}
	}

	p.expect(fmt.Sprintf("expected ')' after %s parameters", typeStr), token.RightParen)

	p.expect(fmt.Sprintf("expected '{' before %s body", typeStr), token.LeftBrace)

	body, ok := p.parseBlockStmt().(*ast.BlockStmt)
	if !ok {
		assert.Assert(body == nil)
		assert.Assert(p.errors != nil)
		return nil
	}

	if t == FunctionExpr {
		return &ast.FuncExpr{Function: funcToken, Params: params, Body: body}
	}
	return &ast.FuncDecl{Function: funcToken, NameToken: name, Params: params, Body: body}
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	if p.matchesOne(token.LeftParen) {
		lParen := p.previous
		e := p.parseExpr()
		p.expect("')' expected", token.RightParen)
		return &ast.GroupExpr{LParen: lParen, Expr: e}
	}
	if p.matchesOne(token.Number) {
		n := p.parseNumber()
		return &ast.LiteralExpr{Value: value.FromNumber(n), ErrorToken: p.previous}
	}
	if p.matchesOne(token.True, token.False) {
		assert.Assert(p.previous.Value == "true" || p.previous.Value == "false")
		b := p.previous.Value == "true"

		return &ast.LiteralExpr{Value: value.FromBoolean(value.Boolean(b)), ErrorToken: p.previous}
	}
	if p.matchesOne(token.String) {
		s := p.previous.Value
		s, err := strconv.Unquote(s)
		if err != nil {
			p.addUnwrappedError(err, p.previous)
		}

		return &ast.LiteralExpr{Value: value.FromString(value.String(s)), ErrorToken: p.previous}
	}
	if p.matchesOne(token.Identifier) {
		return &ast.VariableExpr{Name: p.previous}
	}
	if p.matchesOne(token.Function) {
		e, ok := p.parseFunction(FunctionExpr).(*ast.FuncExpr)
		if !ok {
			assert.Assert(e == nil)
			assert.Assert(p.errors != nil)
			return nil
		}

		return e
	}

	p.newError("expression expected", p.current)
	return nil
}

func (p *Parser) parseCallExpr() ast.Expr {
	e := p.parsePrimaryExpr()

	for {
		if p.matchesOne(token.LeftParen) {
			lParen := p.previous

			var args []ast.Expr

			if p.current.Type != token.RightParen {
				args = append(args, p.parseExpr())

				for p.matchesOne(token.Comma) {
					args = append(args, p.parseExpr())

					if len(args) > math.MaxUint8 {
						log.Fatalf("call expression has more than %d arguments. sorry, unimplemented", math.MaxUint8)
					}
				}
			}

			p.expect("expected ')' after call arguments", token.RightParen)

			e = &ast.CallExpr{Function: e, LParen: lParen, Args: args}
		} else {
			break
		}
	}

	return e
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	if p.matchesOne(token.Minus, token.Bang) {
		op := p.previous
		r := p.parseUnaryExpr()
		return &ast.UnaryExpr{Op: op, Right: r}
	}

	return p.parseCallExpr()
}

func (p *Parser) parseExponentiationExpr() ast.Expr {
	l := p.parseUnaryExpr()

	for p.matchesOne(token.StarStar) {
		op := p.previous
		r := p.parseExponentiationExpr()
		l = &ast.BinaryExpr{Left: l, Op: op, Right: r}
	}

	return l
}

func (p *Parser) parseMultiplicativeExpr() ast.Expr {
	l := p.parseExponentiationExpr()

	for p.matchesOne(token.Star, token.Slash) {
		op := p.previous
		r := p.parseExponentiationExpr()
		l = &ast.BinaryExpr{Left: l, Op: op, Right: r}
	}

	return l
}

func (p *Parser) parseAdditiveExpr() ast.Expr {
	l := p.parseMultiplicativeExpr()

	for p.matchesOne(token.Plus, token.Minus) {
		op := p.previous
		r := p.parseMultiplicativeExpr()
		l = &ast.BinaryExpr{Left: l, Op: op, Right: r}
	}

	return l
}

func (p *Parser) parseComparisonExpr() ast.Expr {
	l := p.parseAdditiveExpr()

	for p.matchesOne(token.Less, token.LessEqual, token.Greater, token.GreaterEqual) {
		op := p.previous
		r := p.parseAdditiveExpr()
		l = &ast.BinaryExpr{Left: l, Op: op, Right: r}
	}

	return l
}

func (p *Parser) parseEqualityExpr() ast.Expr {
	l := p.parseComparisonExpr()

	for p.matchesOne(token.EqualEqual, token.BangEqual) {
		op := p.previous
		r := p.parseComparisonExpr()
		l = &ast.BinaryExpr{Left: l, Op: op, Right: r}
	}

	return l
}

func (p *Parser) parseLogicalAndExpr() ast.Expr {
	l := p.parseEqualityExpr()

	for p.matchesOne(token.And) {
		op := p.previous
		r := p.parseEqualityExpr()
		l = &ast.LogicalExpr{Left: l, Op: op, Right: r}
	}

	return l
}

func (p *Parser) parseLogicalOrExpr() ast.Expr {
	l := p.parseLogicalAndExpr()

	for p.matchesOne(token.Or) {
		op := p.previous
		r := p.parseLogicalAndExpr()
		l = &ast.LogicalExpr{Left: l, Op: op, Right: r}
	}

	return l
}

func (p *Parser) parseExpr() ast.Expr {
	return p.parseLogicalOrExpr()
}

func (p *Parser) parseBlockStmt() ast.Stmt {
	var stmts []ast.Stmt

	lBrace := p.previous

	// Single-line block statement
	if !p.matchesOne(token.Newline) {
		stmts = append(stmts, p.parseStmt(false))
		p.expect("missing matching '}'", token.RightBrace)
		return &ast.BlockStmt{LBrace: lBrace, Stmts: stmts}
	}

	for !p.matchesOne(token.Eof) {
		if p.matchesOne(token.RightBrace) {
			return &ast.BlockStmt{LBrace: lBrace, Stmts: stmts}
		}

		stmts = append(stmts, p.parseStmt(true))
		if p.errors != nil {
			return nil
		}
	}

	p.newError("missing matching '}'", p.current)
	return nil
}

func (p *Parser) parseExprStmt() ast.Stmt {
	e := p.parseExpr()

	if p.matchesOne(token.Equal, token.PlusEqual, token.MinusEqual, token.StarEqual, token.SlashEqual, token.StarStarEqual) {
		equals := p.previous

		if lValue, ok := e.(*ast.VariableExpr); ok {
			var valueExpr ast.Expr

			switch equals.Type {
			case token.Equal:
				valueExpr = p.parseExpr()
			case token.PlusEqual, token.MinusEqual, token.StarEqual, token.SlashEqual, token.StarStarEqual:
				valueExpr = &ast.BinaryExpr{Left: lValue, Op: equals, Right: p.parseExpr()}
			default:
				assert.Assert(false) // Unreachable
			}

			return &ast.AssignStmt{Name: lValue.Name, Equals: equals, Value: valueExpr}
		}

		p.newError("invalid assignment target", equals)
		return nil
	}

	return &ast.ExprStmt{Expr: e}
}

func (p *Parser) parseForStmt() ast.Stmt {
	var init ast.Stmt
	var cond ast.Expr
	var post ast.Stmt

	forToken := p.previous

	if p.matchesOne(token.Semicolon) {
		init = nil
		goto parseCond
	}

	if p.matchesOne(token.Let) {
		init = p.parseLetDecl()
	} else {
		init = p.parseExprStmt()
	}

	p.expect("';' expected", token.Semicolon)

parseCond:
	if p.matchesOne(token.Semicolon) {
		cond = nil
		goto parsePost
	}

	cond = p.parseExpr()

	p.expect("';' expected", token.Semicolon)

parsePost:
	if p.matchesOne(token.LeftBrace) {
		post = nil
		goto parseBody
	}

	post = p.parseExprStmt()

	p.expect("'{' expected", token.LeftBrace)

parseBody:
	body, ok := p.parseBlockStmt().(*ast.BlockStmt)
	if !ok {
		// TODO: DRY
		assert.Assert(body == nil)
		assert.Assert(p.errors != nil)
		return nil
	}

	return &ast.ForStmt{For: forToken, Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseFunctionStmt(t FunctionType) ast.Stmt {
	stmt, ok := p.parseFunction(t).(*ast.FuncDecl)
	if !ok {
		assert.Assert(stmt == nil)
		assert.Assert(p.errors != nil)
		return nil
	}

	return stmt
}

func (p *Parser) parseIfStmt() ast.Stmt {
	ifToken := p.previous

	cond := p.parseExpr()

	p.expect("missing '{' after if condition", token.LeftBrace)

	s := p.parseBlockStmt()
	// An error occurred
	if s == nil {
		return nil
	}

	body, ok := s.(*ast.BlockStmt)
	if !ok {
		assert.Assert(body == nil)
		assert.Assert(p.errors != nil)
		return nil
	}

	var elseStmt ast.Stmt

	if p.matchesOne(token.Else) {
		if p.matchesOne(token.LeftBrace) {
			elseStmt = p.parseBlockStmt()
		} else if p.matchesOne(token.If) {
			elseStmt = p.parseIfStmt()
		} else {
			p.newError("expected if or '{' after else", p.current)
			return nil
		}
	}

	return &ast.IfStmt{If: ifToken, Cond: cond, Body: body, Else: elseStmt}
}

func (p *Parser) parseLetDecl() ast.Stmt {
	let := p.previous

	p.expect("expected variable name", token.Identifier)

	name := p.previous

	var initialiser ast.Expr
	if p.matchesOne(token.Equal) {
		initialiser = p.parseExpr()
	}

	return &ast.LetDecl{Let: let, NameToken: name, Init: initialiser}
}

func (p *Parser) parseReturnStmt(checkLineBreak bool) ast.Stmt {
	returnToken := p.previous

	var expr ast.Expr

	if checkLineBreak {
		if !p.peekMatchesOne(token.Newline, token.Eof) {
			expr = p.parseExpr()
		}
	} else {
		if !p.peekMatchesOne(token.RightBrace) {
			expr = p.parseExpr()
		}
	}

	return &ast.ReturnStmt{Return: returnToken, Expr: expr}
}

func (p *Parser) parseWhileStmt() ast.Stmt {
	while := p.previous

	cond := p.parseExpr()

	p.expect("missing '{' after while condition", token.LeftBrace)

	s := p.parseBlockStmt()
	// An error occurred
	if s == nil {
		return nil
	}

	body, ok := s.(*ast.BlockStmt)
	assert.Assert(ok)

	return &ast.WhileStmt{While: while, Cond: cond, Body: body}
}

func (p *Parser) parseStmt(checkLineBreak bool) ast.Stmt {
	switch {
	case p.matchesOne(token.Error), p.matchesOne(token.Newline):
		return nil
	case p.matchesOne(token.LeftBrace):
		return p.parseBlockStmt()
	}

	var s ast.Stmt

	switch {
	case p.matchesOne(token.For):
		s = p.parseForStmt()
	case p.matchesOne(token.Function):
		s = p.parseFunctionStmt(FunctionDecl)
	case p.matchesOne(token.If):
		s = p.parseIfStmt()
	case p.matchesOne(token.Let):
		s = p.parseLetDecl()
	case p.matchesOne(token.Return):
		s = p.parseReturnStmt(checkLineBreak)
	case p.matchesOne(token.While):
		s = p.parseWhileStmt()
	default:
		s = p.parseExprStmt()
	}

	if checkLineBreak {
		p.expect("newline or eof expected", token.Newline, token.Eof)
	}

	return s
}

// Recovers from panic mode, extending the passed slice errs
// We recover once a newline is reached (newlines after backslashes are ignored)
// or when we reach a keyword starting a new statement
// We need to return []error as we append to errs which may not change the old errs
func (p *Parser) recover() {
	for !p.matchesOne(token.Eof) {
		p.nextToken()

		for !p.matchesOne(token.Eof) && p.matchesOne(token.Error) {
			p.nextToken()
		}

		if p.peekMatchesOne(token.Eof, token.Newline, token.LeftBrace, token.For, token.Function, token.If, token.Let, token.Return, token.While) {
			break
		}
	}
}

func (p *Parser) parse() {
	p.nextToken()

	var errs []error

	for !p.matchesOne(token.Eof) {
		s := p.parseStmt(true)

		if p.errors != nil {
			p.recover()
			errs = append(errs, p.errors...)
			p.errors = nil
		} else {
			p.stmts = append(p.stmts, s)
		}
	}

	p.errors = errs
}

// Parse
// Parsing can have multiple errors but during runtime,
// one error is enough to stop the program
func Parse(s string, file string) ([]ast.Stmt, []error) {
	p := newParser(s, file)
	p.parse()
	return p.stmts, p.errors
}
