package sema

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/util"
)

/*
 Resolver - resolve references between AST nodes
*/

type resolver struct {
	refs map[ast.Node]ast.Node
}

/* Program */

func (r *resolver) resolveProgram(prog *ast.Program, e *env) {
	for _, stmt := range prog.Stmts {
		r.resolveStmt(stmt, e)
	}
}

/* Stmt */

func (r *resolver) resolveStmt(stmt ast.Stmt, e *env) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		r.resolveBlockStmt(v, newEnv(e))
	case *ast.LetStmt:
		r.resolveLetStmt(v, e)
	case *ast.IfStmt:
		r.resolveIfStmt(v, e)
	case *ast.ForStmt:
		r.resolveForStmt(v, e)
	case *ast.ForInStmt:
		r.resolveForInStmt(v, e)
	case *ast.ContinueStmt:
		r.resolveContinueStmt(v, e)
	case *ast.BreakStmt:
		r.resolveBreakStmt(v, e)
	case *ast.ReturnStmt:
		r.resolveReturnStmt(v, e)
	case *ast.AssignStmt:
		r.resolveAssignStmt(v, e)
	case *ast.ExprStmt:
		r.resolveExprStmt(v, e)
	}
}

func (r *resolver) resolveBlockStmt(stmt *ast.BlockStmt, e *env) {
	for _, stmt_ := range stmt.Stmts {
		r.resolveStmt(stmt_, e)
	}
}

func (r *resolver) resolveLetStmt(stmt *ast.LetStmt, e *env) {
	for i, value := range stmt.Values {
		var_ := stmt.Vars[i]

		if fn, ok := value.(*ast.FuncLit); ok {
			e_ := newEnv(e)
			r.resolveVarDecl(var_, e_)
			r.resolveFuncLit(fn, e_)
		} else {
			r.resolveExpr(value, e)
		}
	}
	for _, var_ := range stmt.Vars {
		r.resolveVarDecl(var_, e)
	}
}

func (r *resolver) resolveIfStmt(stmt *ast.IfStmt, e *env) {
	r.resolveExpr(stmt.Cond, e)
	r.resolveBlockStmt(stmt.Body, newEnv(e))

	if stmt.Else != nil {
		r.resolveStmt(stmt.Else, e)
	}
}

func (r *resolver) resolveForStmt(stmt *ast.ForStmt, e *env) {
	r.resolveExpr(stmt.Cond, e)

	e_ := newEnv(e)
	e_.set("continue", stmt)
	e_.set("break", stmt)

	r.resolveBlockStmt(stmt.Body, e_)
}

func (r *resolver) resolveForInStmt(stmt *ast.ForInStmt, e *env) {
	r.resolveExpr(stmt.Expr, e)

	e_ := newEnv(e)
	e_.set("continue", stmt)
	e_.set("break", stmt)

	r.resolveVarDecl(stmt.Elem, e_)
	r.resolveVarDecl(stmt.Index, e_)

	r.resolveBlockStmt(stmt.Body, e_)
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

func (r *resolver) resolveAssignStmt(stmt *ast.AssignStmt, e *env) {
	for _, t := range stmt.Targets {
		r.resolveExpr(t, e)
	}
	for _, v := range stmt.Values {
		r.resolveExpr(v, e)
	}
}

func (r *resolver) resolveExprStmt(stmt *ast.ExprStmt, e *env) {
	r.resolveExpr(stmt.Expr, e)
}

/* Decl */

func (r *resolver) resolveVarDecl(decl *ast.VarDecl, e *env) {
	if decl.Ident == "" {
		return // don't register implicit variable
	}
	err := e.set(decl.Ident, decl)
	if err != nil {
		util.Error("%s has already been declared", decl.Ident)
	}
}

/* Expr */

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
	case *ast.LibCallExpr:
		r.resolveLibCallExpr(v, e)
	case *ast.VarRef:
		r.resolveVarRef(v, e)
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

func (r *resolver) resolveLibCallExpr(expr *ast.LibCallExpr, e *env) {
	for _, param := range expr.Params {
		r.resolveExpr(param, e)
	}
}

func (r *resolver) resolveVarRef(expr *ast.VarRef, e *env) {
	ref, ok := e.get(expr.Ident)
	if !ok {
		util.Error("%s is not declared", expr.Ident)
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
		r.resolveVarDecl(param, e_)
	}
	r.resolveBlockStmt(expr.Body, e_)
}
