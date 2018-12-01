package gen

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
 Resolver - resolve relationship between AST nodes
*/

type resolver struct {
	relations map[ast.Node]ast.Node // child node -> parent node
}

func (r *resolver) resolveProgram(node *ast.Program, e *env) {
	for _, stmt := range node.Stmts {
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
	case *ast.WhileStmt:
		r.resolveWhileStmt(v, e)
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
	if err := e.set(stmt.Ident.Name, stmt); err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
	}
	if stmt.RetType != "void" && !returnableBlockStmt(stmt.Body) {
		util.Error("Missing return at end of %s", stmt.Ident.Name)
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

	if err := e.set(stmt.Ident.Name, stmt); err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
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

func (r *resolver) resolveWhileStmt(stmt *ast.WhileStmt, e *env) {
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

	parent, ok := e.get("return")
	if !ok {
		util.Error("Illegal use of return")
	}
	r.relations[stmt] = parent
}

func (r *resolver) resolveContinueStmt(stmt *ast.ContinueStmt, e *env) {
	parent, ok := e.get("continue")
	if !ok {
		util.Error("Illegal use of continue")
	}
	r.relations[stmt] = parent
}

func (r *resolver) resolveBreakStmt(stmt *ast.BreakStmt, e *env) {
	parent, ok := e.get("break")
	if !ok {
		util.Error("Illegal use of break")
	}
	r.relations[stmt] = parent
}

func (r *resolver) resolveAssignStmt(stmt *ast.AssignStmt, e *env) {
	r.resolveExpr(stmt.Value, e)

	parent, ok := e.get(stmt.Ident.Name)
	if !ok {
		util.Error("%s is not declared", stmt.Ident.Name)
	}
	if _, ok := parent.(*ast.VarDecl); !ok {
		util.Error("%s is not a variable")
	}
	r.relations[stmt] = parent
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
	case *ast.Ident:
		r.resolveIdent(v, e)
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

	if parent, ok := e.get(expr.Ident.Name); ok {
		if _, ok := parent.(*ast.FuncDecl); !ok {
			util.Error("%s is not a function", expr.Ident.Name)
		}
		r.relations[expr] = parent
	} else {
		if _, ok := libFns[expr.Ident.Name]; !ok {
			util.Error("%s is not declared", expr.Ident.Name)
		}
	}
}

func (r *resolver) resolveIdent(expr *ast.Ident, e *env) {
	parent, ok := e.get(expr.Name)
	if !ok {
		util.Error("%s is not declared", expr.Name)
	}
	if _, ok := parent.(*ast.VarDecl); !ok {
		util.Error("%s is not a variable", expr.Name)
	}
	r.relations[expr] = parent
}
