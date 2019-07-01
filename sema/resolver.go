package sema

import (
	"fmt"
	"os"

	"github.com/oshima/lang/ast"
)

// resolver resolves the references between the AST nodes.
type resolver struct{}

func (r *resolver) error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// ----------------------------------------------------------------
// Program

func (r *resolver) resolveProgram(prog *ast.Program, e *env) {
	// register the function names in advance
	for _, stmt := range prog.Stmts {
		if v, ok := stmt.(*ast.FuncStmt); ok {
			if err := e.set(v.Func.Name, v.Func); err != nil {
				r.error("%s: %s has already been declared", v.Func.Pos(), v.Func.Name)
			}
		}
	}
	for _, stmt := range prog.Stmts {
		r.resolveStmt(stmt, e)
	}
}

// ----------------------------------------------------------------
// Stmt

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
	// register the function names in advance
	for _, stmt := range stmt.Stmts {
		if v, ok := stmt.(*ast.FuncStmt); ok {
			if err := e.set(v.Func.Name, v.Func); err != nil {
				r.error("%s: %s has already been declared", v.Func.Pos(), v.Func.Name)
			}
		}
	}
	for _, stmt := range stmt.Stmts {
		r.resolveStmt(stmt, e)
	}
}

func (r *resolver) resolveVarStmt(stmt *ast.VarStmt, e *env) {
	for _, v := range stmt.Vars {
		r.resolveVarDecl(v, e)
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

	ne := newEnv(e)
	ne.set("continue", stmt)
	ne.set("break", stmt)

	r.resolveBlockStmt(stmt.Body, ne)
}

func (r *resolver) resolveForStmt(stmt *ast.ForStmt, e *env) {
	r.resolveExpr(stmt.Iter.Value, e)

	ne := newEnv(e)
	ne.set("continue", stmt)
	ne.set("break", stmt)

	r.resolveVarDecl(stmt.Elem, ne)
	if stmt.Index.Name != "" {
		r.resolveVarDecl(stmt.Index, ne)
	}

	r.resolveBlockStmt(stmt.Body, ne)
}

func (r *resolver) resolveContinueStmt(stmt *ast.ContinueStmt, e *env) {
	ref, ok := e.get("continue")
	if !ok {
		r.error("%s: illegal use of continue", stmt.Pos())
	}
	stmt.Ref = ref
}

func (r *resolver) resolveBreakStmt(stmt *ast.BreakStmt, e *env) {
	ref, ok := e.get("break")
	if !ok {
		r.error("%s: illegal use of break", stmt.Pos())
	}
	stmt.Ref = ref
}

func (r *resolver) resolveReturnStmt(stmt *ast.ReturnStmt, e *env) {
	if stmt.Value != nil {
		r.resolveExpr(stmt.Value, e)
	}

	ref, ok := e.get("return")
	if !ok {
		r.error("%s: illegal use of return", stmt.Pos())
	}
	stmt.Ref = ref
}

func (r *resolver) resolveAssignStmt(stmt *ast.AssignStmt, e *env) {
	r.resolveExpr(stmt.Target, e)
	if v, ok := stmt.Target.(*ast.Ident); ok {
		if _, ok := v.Ref.(*ast.FuncDecl); ok {
			r.error("%s: %s is not a variable", v.Pos(), v.Name)
		}
	}
	r.resolveExpr(stmt.Value, e)
}

func (r *resolver) resolveExprStmt(stmt *ast.ExprStmt, e *env) {
	r.resolveExpr(stmt.Expr, e)
}

// ----------------------------------------------------------------
// Expr

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
		r.error("%s: %s is not declared", expr.Pos(), expr.Name)
	}
	expr.Ref = ref
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
		r.error("%s: functions cannot be nested", expr.Pos())
	}
	if expr.ReturnType != nil && !ast.Returnable(expr.Body) {
		r.error("%s: missing return at end of function", expr.Body.Pos())
	}

	ne := newEnv(e)
	ne.set("return", expr)

	for _, param := range expr.Params {
		r.resolveVarDecl(param, ne)
	}
	r.resolveBlockStmt(expr.Body, ne)
}

// ----------------------------------------------------------------
// Decl

func (r *resolver) resolveVarDecl(decl *ast.VarDecl, e *env) {
	switch v := decl.Value.(type) {
	case nil:
		// ok
	case *ast.FuncLit:
		ne := newEnv(e)
		ne.set(decl.Name, decl)
		r.resolveFuncLit(v, ne)
	default:
		r.resolveExpr(v, e)
	}

	if err := e.set(decl.Name, decl); err != nil {
		r.error("%s: %s has already been declared", decl.Pos(), decl.Name)
	}
}

func (r *resolver) resolveFuncDecl(decl *ast.FuncDecl, e *env) {
	if _, ok := e.get("return"); ok {
		r.error("%s: functions cannot be nested", decl.Pos())
	}
	if decl.ReturnType != nil && !ast.Returnable(decl.Body) {
		r.error("%s: missing return at end of function", decl.Body.Pos())
	}

	ne := newEnv(e)
	ne.set("return", decl)

	for _, param := range decl.Params {
		r.resolveVarDecl(param, ne)
	}
	r.resolveBlockStmt(decl.Body, ne)
}
