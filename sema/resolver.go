package sema

import (
	"errors"
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/util"
)

/*
 Environment - create the scope of name bindings
*/

type env struct {
	store map[string]ast.Node
	outer *env
}

func newEnv(outer *env) *env {
	return &env{
		store: make(map[string]ast.Node),
		outer: outer,
	}
}

func (e *env) set(key string, node ast.Node) error {
	if _, ok := e.store[key]; ok {
		return errors.New("Duplicate entries")
	}
	e.store[key] = node
	return nil
}

func (e *env) get(key string) (ast.Node, bool) {
	node, ok := e.store[key]
	if !ok && e.outer != nil {
		node, ok = e.outer.get(key)
	}
	return node, ok
}

/*
 Resolver - resolve references between AST nodes
*/

type resolver struct {
	refs map[ast.Node]ast.Node
}

func (r *resolver) resolveProgram(prog *ast.Program, e *env) {
	for _, stmt := range prog.Stmts {
		r.resolveStmt(stmt, e)
	}
}

func (r *resolver) resolveStmt(stmt ast.Stmt, e *env) {
	switch v := stmt.(type) {
	case *ast.FuncDecl:
		r.resolveFuncDecl(v, e)
	case *ast.VarDecl:
		r.resolveVarDecl(v, e)
	case *ast.BlockStmt:
		r.resolveBlockStmt(v, newEnv(e))
	case *ast.IfStmt:
		r.resolveIfStmt(v, e)
	case *ast.ForStmt:
		r.resolveForStmt(v, e)
	case *ast.ReturnStmt:
		r.resolveReturnStmt(v, e)
	case *ast.ContinueStmt:
		r.resolveContinueStmt(v, e)
	case *ast.BreakStmt:
		r.resolveBreakStmt(v, e)
	case *ast.AssignStmt:
		r.resolveAssignStmt(v, e)
	case *ast.ExprStmt:
		r.resolveExprStmt(v, e)
	}
}

func (r *resolver) resolveFuncDecl(stmt *ast.FuncDecl, e *env) {
	if _, ok := e.get("return"); ok {
		util.Error("Function declarations cannot be nested")
	}
	if err := e.set(stmt.Ident, stmt); err != nil {
		util.Error("%s has already been declared", stmt.Ident)
	}
	if stmt.ReturnType != "void" && !returnableBlockStmt(stmt.Body) {
		util.Error("Missing return at end of %s", stmt.Ident)
	}

	e_ := newEnv(e)
	e_.set("return", stmt)

	for _, param := range stmt.Params {
		r.resolveVarDecl(param, e_)
	}
	r.resolveBlockStmt(stmt.Body, e_)
}

func (r *resolver) resolveVarDecl(stmt *ast.VarDecl, e *env) {
	if stmt.Value != nil {
		r.resolveExpr(stmt.Value, e)
	}

	if err := e.set(stmt.Ident, stmt); err != nil {
		util.Error("%s has already been declared", stmt.Ident)
	}
}

func (r *resolver) resolveBlockStmt(stmt *ast.BlockStmt, e *env) {
	for _, stmt_ := range stmt.Stmts {
		r.resolveStmt(stmt_, e)
	}
}

func (r *resolver) resolveIfStmt(stmt *ast.IfStmt, e *env) {
	r.resolveExpr(stmt.Cond, e)
	r.resolveBlockStmt(stmt.Conseq, newEnv(e))

	if stmt.Altern != nil {
		r.resolveStmt(stmt.Altern, e)
	}
}

func (r *resolver) resolveForStmt(stmt *ast.ForStmt, e *env) {
	r.resolveExpr(stmt.Cond, e)

	e_ := newEnv(e)
	e_.set("continue", stmt)
	e_.set("break", stmt)

	r.resolveBlockStmt(stmt.Body, e_)
}

func (r *resolver) resolveReturnStmt(stmt *ast.ReturnStmt, e *env) {
	if stmt.Value != nil {
		r.resolveExpr(stmt.Value, e)
	}

	ref, ok := e.get("return")
	if !ok {
		util.Error("Illegal use of return")
	}
	r.refs[stmt] = ref
}

func (r *resolver) resolveContinueStmt(stmt *ast.ContinueStmt, e *env) {
	ref, ok := e.get("continue")
	if !ok {
		util.Error("Illegal use of continue")
	}
	r.refs[stmt] = ref
}

func (r *resolver) resolveBreakStmt(stmt *ast.BreakStmt, e *env) {
	ref, ok := e.get("break")
	if !ok {
		util.Error("Illegal use of break")
	}
	r.refs[stmt] = ref
}

func (r *resolver) resolveAssignStmt(stmt *ast.AssignStmt, e *env) {
	r.resolveExpr(stmt.Value, e)

	ref, ok := e.get(stmt.Ident)
	if !ok {
		util.Error("%s is not declared", stmt.Ident)
	}
	if _, ok := ref.(*ast.VarDecl); !ok {
		util.Error("%s is not a variable")
	}
	r.refs[stmt] = ref
}

func (r *resolver) resolveExprStmt(stmt *ast.ExprStmt, e *env) {
	r.resolveExpr(stmt.Expr, e)
}

func (r *resolver) resolveExpr(expr ast.Expr, e *env) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		r.resolvePrefixExpr(v, e)
	case *ast.InfixExpr:
		r.resolveInfixExpr(v, e)
	case *ast.FuncCall:
		r.resolveFuncCall(v, e)
	case *ast.VarRef:
		r.resolveVarRef(v, e)
	}
}

func (r *resolver) resolvePrefixExpr(expr *ast.PrefixExpr, e *env) {
	r.resolveExpr(expr.Right, e)
}

func (r *resolver) resolveInfixExpr(expr *ast.InfixExpr, e *env) {
	r.resolveExpr(expr.Left, e)
	r.resolveExpr(expr.Right, e)
}

func (r *resolver) resolveFuncCall(expr *ast.FuncCall, e *env) {
	for _, param := range expr.Params {
		r.resolveExpr(param, e)
	}

	ref, ok := e.get(expr.Ident)
	if !ok {
		if _, ok := libFns[expr.Ident]; !ok {
			util.Error("%s is not declared", expr.Ident)
		}
		return
	}
	if _, ok := ref.(*ast.FuncDecl); !ok {
		util.Error("%s is not a function", expr.Ident)
	}
	r.refs[expr] = ref
}

func (r *resolver) resolveVarRef(expr *ast.VarRef, e *env) {
	ref, ok := e.get(expr.Ident)
	if !ok {
		util.Error("%s is not declared", expr.Ident)
	}
	if _, ok := ref.(*ast.VarDecl); !ok {
		util.Error("%s is not a variable", expr.Ident)
	}
	r.refs[expr] = ref
}
