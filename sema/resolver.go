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

func (e *env) set(name string, node ast.Node) error {
	if _, ok := e.store[name]; ok {
		return errors.New("Duplicate entries")
	}
	e.store[name] = node
	return nil
}

func (e *env) get(name string) (ast.Node, bool) {
	node, ok := e.store[name]
	if !ok && e.outer != nil {
		node, ok = e.outer.get(name)
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
	case *ast.LetStmt:
		if _, ok := v.Value.(*ast.FuncLit); ok {
			r.resolveLetStmtWithFuncLit(v, e)
		} else {
			r.resolveLetStmt(v, e)
		}
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

func (r *resolver) resolveLetStmt(stmt *ast.LetStmt, e *env) {
	if stmt.Value != nil {
		r.resolveExpr(stmt.Value, e)
	}
	err := e.set(stmt.Ident.Name, stmt)
	if err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
	}
}

func (r *resolver) resolveLetStmtWithFuncLit(stmt *ast.LetStmt, e *env) {
	// register identifier name in advance to enable recursive calls in function
	err := e.set(stmt.Ident.Name, stmt)
	if err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
	}
	r.resolveFuncLit(stmt.Value.(*ast.FuncLit), e)
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
	r.resolveExpr(stmt.Target, e)
	r.resolveExpr(stmt.Value, e)
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
	case *ast.IndexExpr:
		r.resolveIndexExpr(v, e)
	case *ast.CallExpr:
		r.resolveCallExpr(v, e)
	case *ast.LibcallExpr:
		r.resolveLibcallExpr(v, e)
	case *ast.Ident:
		r.resolveIdent(v, e)
	case *ast.ArrayLit:
		r.resolveArrayLit(v, e)
	case *ast.FuncLit:
		r.resolveFuncLit(v, e)
	}
}

func (r *resolver) resolvePrefixExpr(expr *ast.PrefixExpr, e *env) {
	r.resolveExpr(expr.Right, e)
}

func (r *resolver) resolveInfixExpr(expr *ast.InfixExpr, e *env) {
	r.resolveExpr(expr.Left, e)
	r.resolveExpr(expr.Right, e)
}

func (r *resolver) resolveIndexExpr(expr *ast.IndexExpr, e *env) {
	r.resolveExpr(expr.Left, e)
	r.resolveExpr(expr.Index, e)
}

func (r *resolver) resolveCallExpr(expr *ast.CallExpr, e *env) {
	r.resolveExpr(expr.Left, e)
	for _, param := range expr.Params {
		r.resolveExpr(param, e)
	}
}

func (r *resolver) resolveLibcallExpr(expr *ast.LibcallExpr, e *env) {
	for _, param := range expr.Params {
		r.resolveExpr(param, e)
	}
}

func (r *resolver) resolveIdent(expr *ast.Ident, e *env) {
	ref, ok := e.get(expr.Name)
	if !ok {
		util.Error("%s is not declared", expr.Name)
	}
	r.refs[expr] = ref
}

func (r *resolver) resolveArrayLit(expr *ast.ArrayLit, e *env) {
	for _, elem := range expr.Elems {
		r.resolveExpr(elem, e)
	}
}

func (r *resolver) resolveFuncLit(expr *ast.FuncLit, e *env) {
	if _, ok := e.get("return"); ok {
		util.Error("Function literals cannot be nested")
	}
	if expr.ReturnType != nil && !returnableBlockStmt(expr.Body) {
		util.Error("Missing return at end of function")
	}

	e_ := newEnv(e)
	e_.set("return", expr)

	for _, param := range expr.Params {
		r.resolveLetStmt(param, e_)
	}
	r.resolveBlockStmt(expr.Body, e_)
}
