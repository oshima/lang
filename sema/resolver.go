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
		// register the function names in advance
		if v, ok := stmt.(*ast.FuncStmt); ok {
			if err := e.set(v.Func.Name, v.Func); err != nil {
				util.Error("%s has already been declared", v.Func.Name)
			}
		}
	}
	for _, stmt := range prog.Stmts {
		r.resolveStmt(stmt, e)
	}
}

/* Stmt */

func (r *resolver) resolveStmt(stmt ast.Stmt, e *env) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		r.resolveBlockStmt(v, newEnv(e))
	case *ast.VarStmt:
		r.resolveVarStmt(v, e)
	case *ast.FuncStmt:
		r.resolveFuncStmt(v, e)
	case *ast.IfStmt:
		r.resolveIfStmt(v, e)
	case *ast.WhileStmt:
		r.resolveWhileStmt(v, e)
	case *ast.ForStmt:
		r.resolveForStmt(v, e)
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
	for _, stmt := range stmt.Stmts {
		// register the function names in advance
		if v, ok := stmt.(*ast.FuncStmt); ok {
			if err := e.set(v.Func.Name, v.Func); err != nil {
				util.Error("%s has already been declared", v.Func.Name)
			}
		}
	}
	for _, stmt_ := range stmt.Stmts {
		r.resolveStmt(stmt_, e)
	}
}

func (r *resolver) resolveVarStmt(stmt *ast.VarStmt, e *env) {
	for _, var_ := range stmt.Vars {
		r.resolveVarDecl(var_, e)
	}
}

func (r *resolver) resolveFuncStmt(stmt *ast.FuncStmt, e *env) {
	r.resolveFuncDecl(stmt.Func, e)
}

func (r *resolver) resolveIfStmt(stmt *ast.IfStmt, e *env) {
	r.resolveExpr(stmt.Cond, e)
	r.resolveBlockStmt(stmt.Body, newEnv(e))

	if stmt.Else != nil {
		r.resolveStmt(stmt.Else, e)
	}
}

func (r *resolver) resolveWhileStmt(stmt *ast.WhileStmt, e *env) {
	r.resolveExpr(stmt.Cond, e)

	e_ := newEnv(e)
	e_.set("continue", stmt)
	e_.set("break", stmt)

	r.resolveBlockStmt(stmt.Body, e_)
}

func (r *resolver) resolveForStmt(stmt *ast.ForStmt, e *env) {
	r.resolveExpr(stmt.Iter.Value, e)

	e_ := newEnv(e)
	e_.set("continue", stmt)
	e_.set("break", stmt)

	r.resolveVarDecl(stmt.Elem, e_)
	if stmt.Index.Name != "" {
		r.resolveVarDecl(stmt.Index, e_)
	}

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
	r.resolveExpr(stmt.Target, e)
	if v, ok := stmt.Target.(*ast.Ident); ok {
		if _, ok := r.refs[v].(*ast.FuncDecl); ok {
			util.Error("%s is not a variable", v.Name)
		}
	}
	r.resolveExpr(stmt.Value, e)
}

func (r *resolver) resolveExprStmt(stmt *ast.ExprStmt, e *env) {
	r.resolveExpr(stmt.Expr, e)
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
	case *ast.Ident:
		r.resolveIdent(v, e)
	case *ast.RangeLit:
		r.resolveRangeLit(v, e)
	case *ast.ArrayLit:
		r.resolveArrayLit(v, e)
	case *ast.ArrayShortLit:
		r.resolveArrayShortLit(v, e)
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

func (r *resolver) resolveIdent(expr *ast.Ident, e *env) {
	ref, ok := e.get(expr.Name)
	if !ok {
		util.Error("%s is not declared", expr.Name)
	}
	r.refs[expr] = ref
}

func (r *resolver) resolveRangeLit(expr *ast.RangeLit, e *env) {
	r.resolveExpr(expr.Lower, e)
	r.resolveExpr(expr.Upper, e)
}

func (r *resolver) resolveArrayLit(expr *ast.ArrayLit, e *env) {
	for _, elem := range expr.Elems {
		r.resolveExpr(elem, e)
	}
}

func (r *resolver) resolveArrayShortLit(expr *ast.ArrayShortLit, e *env) {
	r.resolveExpr(expr.Value, e)
}

func (r *resolver) resolveFuncLit(expr *ast.FuncLit, e *env) {
	if _, ok := e.get("return"); ok {
		util.Error("Functions cannot be nested")
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

/* Decl */

func (r *resolver) resolveVarDecl(decl *ast.VarDecl, e *env) {
	switch v := decl.Value.(type) {
	case nil:
		// ok
	case *ast.FuncLit:
		e_ := newEnv(e)
		e_.set(decl.Name, decl)
		r.resolveFuncLit(v, e_)
	default:
		r.resolveExpr(v, e)
	}

	if err := e.set(decl.Name, decl); err != nil {
		util.Error("%s has already been declared", decl.Name)
	}
}

func (r *resolver) resolveFuncDecl(decl *ast.FuncDecl, e *env) {
	if _, ok := e.get("return"); ok {
		util.Error("Functions cannot be nested")
	}
	if decl.ReturnType != nil && !returnableBlockStmt(decl.Body) {
		util.Error("Missing return at end of function")
	}

	e_ := newEnv(e)
	e_.set("return", decl)

	for _, param := range decl.Params {
		r.resolveVarDecl(param, e_)
	}
	r.resolveBlockStmt(decl.Body, e_)
}
