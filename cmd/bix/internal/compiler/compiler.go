package compiler

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"github.com/billchow98/bixscript/cmd/bix/internal/ast"
	"github.com/billchow98/bixscript/cmd/bix/internal/syntax"
	"github.com/billchow98/bixscript/cmd/bix/internal/token"
	"github.com/billchow98/bixscript/cmd/bix/internal/value"
	"log"
	"math"
)

type local struct {
	name  string
	depth int
}

type compiler struct {
	function   *value.Function
	locals     []local
	scopeDepth int // 0: global scope
	tempCount  int
}

const (
	UninitialisedDepth = -1
	PcOffset           = -3 // Multiple users rely on assumption this is right after function call arguments
	FpOffset           = -2
	FuncOffset         = -1
)

func newCompiler(funcName value.String) *compiler {
	f := value.NewFunction(funcName)
	return &compiler{function: f, scopeDepth: 0, tempCount: 0}
}

// The first c.function.Argc values must be function parameters
func (c *compiler) isFuncParam(i int) bool {
	return i < c.function.Argc
}

// Assumes pc is right after the function arguments
func (c *compiler) funcParamOffset(i int) int {
	return PcOffset - c.function.Argc + i
}

// Returns sentinel value for findLocalInAnyScope
func (c *compiler) localNotFound() int {
	return len(c.locals)
}

// We remember to specially treat function parameters' indices
func (c *compiler) findLocalInAnyScope(name *token.Token) (int, error) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if name.Value == c.locals[i].name {
			if c.locals[i].depth == UninitialisedDepth {
				return i, newError("cannot use local variable in own initialiser", name)
			}

			if c.isFuncParam(i) {
				return c.funcParamOffset(i), nil
			}

			return i, nil
		}
	}

	return c.localNotFound(), nil
}

func (c *compiler) findLocalInCurrentScope(name *token.Token) error {
	for i := len(c.locals) - 1; i >= 0 && c.locals[i].depth == c.scopeDepth; i-- {
		if name.Value == c.locals[i].name {
			return newError("variable with same name exists in current scope", name)
		}
	}

	return nil
}

// Remember that len(c.locals) also includes the function parameters
func (c *compiler) localsCount() int {
	return len(c.locals) - c.function.Argc
}

func (c *compiler) updateMaxStackUsage() {
	localsCount := c.localsCount()
	tempCount := c.tempCount

	if localsCount+tempCount > c.function.MaxStackUsage {
		c.function.MaxStackUsage = localsCount + tempCount
	}
}

// Calling a child function, pushing args, pc, fp, ... should also be counted here
func (c *compiler) pushTemp() {
	c.tempCount++
	c.updateMaxStackUsage()
}

func (c *compiler) popTemp() {
	c.tempCount--
}

func (c *compiler) stackTop() int {
	i := c.localsCount() + c.tempCount - 1

	if i > math.MaxUint8 {
		log.Fatalf("Too many local and temporary variables in one bytecode. sorry, unimplemented")
	}

	return i
}

func (c *compiler) visitExpr(node ast.Expr) error {
	switch node := node.(type) {
	case *ast.BinaryExpr:
		if node.Op.Type == token.StarStar || node.Op.Type == token.StarStarEqual {
			err := c.visitExpr(node.Right)
			if err != nil {
				return err
			}

			c.pushTemp()
			c.function.AddStar(node.Token(), c.stackTop())

			err = c.visitExpr(node.Left)
			if err != nil {
				return err
			}

			c.function.AddPow(node.Token(), c.stackTop())

			c.popTemp()
			break
		}

		err := c.visitExpr(node.Left)
		if err != nil {
			return err
		}

		c.pushTemp()
		c.function.AddStar(node.Token(), c.stackTop())

		err = c.visitExpr(node.Right)
		if err != nil {
			return err
		}

		switch node.Op.Type {
		case token.Plus, token.PlusEqual:
			c.function.AddAdd(node.Token(), c.stackTop())
		case token.Minus, token.MinusEqual:
			c.function.AddSub(node.Token(), c.stackTop())
		case token.Star, token.StarEqual:
			c.function.AddMul(node.Token(), c.stackTop())
		case token.Slash, token.SlashEqual:
			c.function.AddDiv(node.Token(), c.stackTop())
		case token.EqualEqual:
			c.function.AddEq(node.Token(), c.stackTop())
		case token.BangEqual:
			c.function.AddNeq(node.Token(), c.stackTop())
		case token.Less:
			c.function.AddCmpLt(node.Token(), c.stackTop())
		case token.LessEqual:
			c.function.AddCmpLe(node.Token(), c.stackTop())
		case token.Greater:
			c.function.AddCmpGt(node.Token(), c.stackTop())
		case token.GreaterEqual:
			c.function.AddCmpGe(node.Token(), c.stackTop())
		default:
			assert.Assert(false) // Unreachable
		}

		c.popTemp()
	case *ast.CallExpr:
		err := c.visitExpr(node.Function)
		if err != nil {
			return err
		}

		// Save call expression onto stack
		c.pushTemp()
		callableIndex := c.stackTop()
		c.function.AddStar(node.Token(), callableIndex)

		for _, arg := range node.Args {
			err := c.visitExpr(arg)
			if err != nil {
				return err
			}

			c.pushTemp()
			c.function.AddStar(node.Token(), c.stackTop())
		}

		c.pushTemp() // pc
		c.pushTemp() // fp
		c.pushTemp() // function pointer

		// We send the position of the new fp after function call
		c.function.AddCall(node.Token(), c.stackTop()+1, len(node.Args))

		c.popTemp() // Call expression

		for i := 0; i < len(node.Args); i++ {
			c.popTemp()
		}

		c.popTemp() // pc
		c.popTemp() // fp
		c.popTemp() // function pointer
	case *ast.FuncExpr:
		// Create new function compiler
		fc := newCompiler("__anonymous__")

		// Compile function into a *Function, convert to a Value and store in accumulator
		f, err := fc.compileFromAst(node)
		if err != nil {
			return err
		}

		litExpr := &ast.LiteralExpr{Value: value.FromFunction(f), ErrorToken: node.Token()}

		err = c.visitExpr(litExpr)
		if err != nil {
			return err
		}
	case *ast.GroupExpr:
		err := c.visitExpr(node.Expr)
		if err != nil {
			return err
		}
	case *ast.LiteralExpr:
		c.function.AddLdaConstant(node.Token(), node.Value)
	case *ast.LogicalExpr:
		err := c.visitExpr(node.Left)
		if err != nil {
			return err
		}

		var index int

		switch node.Op.Type {
		case token.And:
			index = c.function.AddJumpIfFalse(node.Token())
		case token.Or:
			index = c.function.AddJumpIfTrue(node.Token())
		default:
			assert.Assert(false) // Unreachable
		}

		err = c.visitExpr(node.Right)
		if err != nil {
			return err
		}

		c.function.PatchJump(index, c.function.GetJumpIndex())
	case *ast.UnaryExpr:
		err := c.visitExpr(node.Right)
		if err != nil {
			return err
		}

		switch node.Op.Type {
		case token.Minus:
			c.function.AddNeg(node.Token())
		case token.Bang:
			c.function.AddNot(node.Token())
		default:
			assert.Assert(false) // Unreachable
		}
	case *ast.VariableExpr:
		i, err := c.findLocalInAnyScope(node.Name)
		if err != nil {
			return err
		}

		if i == c.localNotFound() {
			c.function.AddLdaGlobal(node.Token(), value.String(node.Name.Value))
		} else {
			c.function.AddLdar(node.Token(), i)
		}
	default:
		assert.Assert(false) // Unreachable
	}

	return nil
}

func (c *compiler) enterScope() {
	c.scopeDepth++
}

func (c *compiler) exitScope() {
	c.scopeDepth--

	for len(c.locals) > 0 && c.locals[len(c.locals)-1].depth > c.scopeDepth {
		c.locals = c.locals[:len(c.locals)-1]
	}
}

func (c *compiler) inGlobalScope() bool {
	return c.scopeDepth == 0
}

func (c *compiler) declareVariable(node ast.Decl) error {
	if c.inGlobalScope() {
		c.function.AddGlobDecl(node.Token(), value.String(node.Name().Value))
	} else {
		err := c.findLocalInCurrentScope(node.Name())
		if err != nil {
			return err
		}

		c.locals = append(c.locals, local{node.Name().Value, UninitialisedDepth})

		if c.stackTop() > math.MaxUint8 {
			log.Fatalf("Too many local variables in one bytecode. sorry, unimplemented")
		}

		c.updateMaxStackUsage()
	}

	return nil
}

func (c *compiler) defineVariable(node ast.Decl) {
	if c.inGlobalScope() {
		c.function.AddStaGlobal(node.Token(), value.String(node.Name().Value))
	} else {
		c.markInitialised()
		c.function.AddStar(node.Token(), c.stackTop())
	}
}

// Marks most recent local variable added as initialised (so that its value is open for use)
func (c *compiler) markInitialised() {
	c.locals[len(c.locals)-1].depth = c.scopeDepth
}

func (c *compiler) visitStmt(node ast.Stmt) error {
	// Empty statement
	if node == nil {
		return nil
	}

	switch node := node.(type) {
	case *ast.AssignStmt:
		err := c.visitExpr(node.Value)
		if err != nil {
			return err
		}

		i, err := c.findLocalInAnyScope(node.Name)
		assert.Assert(err == nil)

		if i == c.localNotFound() {
			c.function.AddStaGlobal(node.Token(), value.String(node.Name.Value))
		} else {
			c.function.AddStar(node.Token(), i)
		}
	case *ast.BlockStmt:
		c.enterScope()

		for _, s := range node.Stmts {
			errs := c.visitStmt(s)
			if errs != nil {
				return errs
			}
		}

		c.exitScope()
	case *ast.ExprStmt:
		err := c.visitExpr(node.Expr)
		if err != nil {
			return err
		}
	case *ast.ForStmt:
		c.enterScope()

		errs := c.visitStmt(node.Init)
		if errs != nil {
			return errs
		}

		condIndex := c.function.GetJumpIndex()

		err := c.visitExpr(node.Cond)
		if err != nil {
			return err
		}

		breakIndex := c.function.AddJumpIfFalse(node.Token())

		errs = c.visitStmt(node.Body)
		if errs != nil {
			return errs
		}

		errs = c.visitStmt(node.Post)
		if errs != nil {
			return errs
		}

		backIndex := c.function.AddJump(node.Token())

		c.function.PatchJump(backIndex, condIndex)
		c.function.PatchJump(breakIndex, c.function.GetJumpIndex())

		c.exitScope()
	case *ast.FuncDecl:
		err := c.declareVariable(node)
		if err != nil {
			return err
		}

		// Enable recursion
		if !c.inGlobalScope() {
			c.markInitialised()
		}

		// Create new function compiler
		fc := newCompiler(value.String(node.Name().Value))

		// Compile function into a *Function, convert to a Value and store in accumulator
		f, errs := fc.compileFromAst(&ast.FuncExpr{Function: node.Function, Params: node.Params, Body: node.Body})
		if errs != nil {
			return errs
		}

		litExpr := &ast.LiteralExpr{Value: value.FromFunction(f), ErrorToken: node.Token()}

		err = c.visitExpr(litExpr)
		if err != nil {
			return err
		}

		c.defineVariable(node)
	case *ast.IfStmt:
		err := c.visitExpr(node.Cond)
		if err != nil {
			return err
		}

		ifIndex := c.function.AddJumpIfFalse(node.Token())

		errs := c.visitStmt(node.Body)
		if errs != nil {
			return errs
		}

		if node.Else == nil {
			c.function.PatchJump(ifIndex, c.function.GetJumpIndex())
			break
		}

		elseIndex := c.function.AddJump(node.Token())

		c.function.PatchJump(ifIndex, c.function.GetJumpIndex())

		errs = c.visitStmt(node.Else)
		if errs != nil {
			return errs
		}

		c.function.PatchJump(elseIndex, c.function.GetJumpIndex())
	case *ast.LetDecl:
		err := c.declareVariable(node)
		if err != nil {
			return err
		}

		if node.Init == nil {
			if !c.inGlobalScope() {
				c.markInitialised()
			}

			break
		}

		err = c.visitExpr(node.Init)
		if err != nil {
			return err
		}

		c.defineVariable(node)
	case *ast.ReturnStmt:
		if node.Expr == nil {
			c.function.AddLdaUndefined(node.Token())
		} else {
			err := c.visitExpr(node.Expr)
			if err != nil {
				return err
			}
		}

		c.function.AddReturn(node.Token())
	case *ast.WhileStmt:
		condIndex := c.function.GetJumpIndex()

		err := c.visitExpr(node.Cond)
		if err != nil {
			return err
		}

		breakIndex := c.function.AddJumpIfFalse(node.Token())

		errs := c.visitStmt(node.Body)
		if errs != nil {
			return errs
		}

		backIndex := c.function.AddJump(node.Token())

		c.function.PatchJump(breakIndex, c.function.GetJumpIndex())
		c.function.PatchJump(backIndex, condIndex)
	default:
		assert.Assert(false) // Unreachable
	}

	return nil
}

func (c *compiler) compileFromAst(node *ast.FuncExpr) (*value.Function, error) {
	// Set argc
	c.function.Argc = len(node.Params)

	c.enterScope()

	// Parse parameters
	for _, param := range node.Params {
		paramDecl := &ast.LetDecl{Let: node.Function, NameToken: param, Init: nil}

		err := c.declareVariable(paramDecl)
		if err != nil {
			return nil, err
		}

		c.markInitialised()
	}

	for _, s := range node.Body.Stmts {
		err := c.visitStmt(s)
		if err != nil {
			return nil, err
		}
	}

	c.exitScope()

	assert.Assert(len(c.locals) == 0)
	assert.Assert(c.scopeDepth == 0)

	c.function.Finalise()

	return c.function, nil
}

func (c *compiler) compile(s string, file string) (*value.Function, []error) {
	stmts, errs := syntax.Parse(s, file)
	if errs != nil {
		return nil, errs
	}

	for _, stmt := range stmts {
		err := c.visitStmt(stmt)
		if err != nil {
			return nil, []error{err}
		}

		assert.Assert(len(c.locals) == 0)
		assert.Assert(c.scopeDepth == 0)
	}

	c.function.Finalise()

	return c.function, nil
}

// Compile
// Current stack frame format:
// callable | arg 1 | arg 2 | pc | fp | func ptr | var 1 | var 2 | tmp 1 | tmp 2
// -6       | -5    | -4    | -3 | -2 | -1       | 0     | 1     | 2     | 3
// tmp 1 and 2 don't have to be in these positions;
// they simply use the space just above currently needed variables
func Compile(s string, file string, funcName string) (*value.Function, []error) {
	// For now, we don't pass any arguments to the top level.
	return newCompiler(value.String(funcName)).compile(s, file)
}
