package gen

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/util"
)

/*
 Typechecker - do type checking
*/

type typechecker struct {
	relations map[ast.Node]ast.Node
}

func (t *typechecker) typecheckProgram(node *ast.Program) {
	for _, stmt := range node.Stmts {
		t.typecheckStmt(stmt)
	}
}

func (t *typechecker) typecheckStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.FuncDecl:
		t.typecheckFuncDecl(v)
	case *ast.VarDecl:
		t.typecheckVarDecl(v)
	case *ast.BlockStmt:
		t.typecheckBlockStmt(v)
	case *ast.IfStmt:
		t.typecheckIfStmt(v)
	case *ast.WhileStmt:
		t.typecheckWhileStmt(v)
	case *ast.ReturnStmt:
		t.typecheckReturnStmt(v)
	case *ast.AssignStmt:
		t.typecheckAssignStmt(v)
	case *ast.ExprStmt:
		t.typecheckExprStmt(v)
	}
}

func (t *typechecker) typecheckFuncDecl(stmt *ast.FuncDecl) {
	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckVarDecl(stmt *ast.VarDecl) {
	if ty := t.typecheckExpr(stmt.Value); ty != stmt.Type {
		f := "Expected %s value for %s, but got %s"
		util.Error(f, stmt.Type, stmt.Ident.Name, ty)
	}
}

func (t *typechecker) typecheckBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.Stmts {
		t.typecheckStmt(stmt_)
	}
}

func (t *typechecker) typecheckIfStmt(stmt *ast.IfStmt) {
	if ty := t.typecheckExpr(stmt.Cond); ty != "bool" {
		util.Error("Expected bool value for if condition, but got %s", ty)
	}

	t.typecheckBlockStmt(stmt.Conseq)

	if stmt.Altern != nil {
		t.typecheckStmt(stmt.Altern)
	}
}

func (t *typechecker) typecheckWhileStmt(stmt *ast.WhileStmt) {
	if ty := t.typecheckExpr(stmt.Cond); ty != "bool" {
		util.Error("Expected bool value for while condition, but got %s value", ty)
	}

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckReturnStmt(stmt *ast.ReturnStmt) {
	parent := t.relations[stmt].(*ast.FuncDecl)

	if stmt.Value == nil {
		if parent.RetType != "void" {
			f := "Expected %s value for return of %s, but nothing returned"
			util.Error(f, parent.RetType, parent.Ident.Name)
		}
		return
	}

	if ty := t.typecheckExpr(stmt.Value); ty != parent.RetType {
		f := "Expected %s value for return of %s, but got %s"
		util.Error(f, parent.RetType, parent.Ident.Name, ty)
	}
}

func (t *typechecker) typecheckAssignStmt(stmt *ast.AssignStmt) {
	parent := t.relations[stmt].(*ast.VarDecl)

	if ty := t.typecheckExpr(stmt.Value); ty != parent.Type {
		f := "Expected %s value for %s, but got %s"
		util.Error(f, parent.Type, stmt.Ident.Name, ty)
	}
}

func (t *typechecker) typecheckExprStmt(stmt *ast.ExprStmt) {
	t.typecheckExpr(stmt.Expr)
}

func (t *typechecker) typecheckExpr(expr ast.Expr) string {
	var ty string

	switch v := expr.(type) {
	case *ast.PrefixExpr:
		ty = t.typecheckPrefixExpr(v)
	case *ast.InfixExpr:
		ty = t.typecheckInfixExpr(v)
	case *ast.FuncCall:
		ty = t.typecheckFuncCall(v)
	case *ast.Ident:
		ty = t.typecheckIdent(v)
	case *ast.IntLit:
		ty = "int"
	case *ast.BoolLit:
		ty = "bool"
	case *ast.StringLit:
		ty = "string"
	}
	return ty
}

func (t *typechecker) typecheckPrefixExpr(expr *ast.PrefixExpr) string {
	var ty string
	tyr := t.typecheckExpr(expr.Right)

	switch expr.Operator {
	case "!":
		if tyr != "bool" {
			util.Error("Expected bool value for operand of !, but got %s", tyr)
		}
		ty = "bool"
	case "-":
		if tyr != "int" {
			util.Error("Expected int value for operand of -, but got %s", tyr)
		}
		ty = "int"
	}
	return ty
}

func (t *typechecker) typecheckInfixExpr(expr *ast.InfixExpr) string {
	var ty string
	tyl := t.typecheckExpr(expr.Left)
	tyr := t.typecheckExpr(expr.Right)

	switch op := expr.Operator; op {
	case "+", "-", "*", "/", "%":
		if tyl != "int" || tyr != "int" {
			util.Error("Expected int values for operands of %s, but got %s, %s", op, tyl, tyr)
		}
		ty = "int"
	case "==", "!=":
		if tyl != "int" && tyl != "bool" {
			util.Error("Expected int or bool value for left operand of %s, but got %s", op, tyl)
		}
		if tyl == "int" && tyr != "int" {
			util.Error("Expected int value for right operand of %s, but got %s", op, tyr)
		}
		if tyl == "bool" && tyr != "bool" {
			util.Error("Expected bool value for right operand of %s, but got %s", op, tyr)
		}
		ty = "bool"
	case "<", "<=", ">", ">=":
		if tyl != "int" || tyr != "int" {
			util.Error("Expected int values for operands of %s, but got %s, %s", op, tyl, tyr)
		}
		ty = "bool"
	case "&&", "||":
		if tyl != "bool" || tyr != "bool" {
			util.Error("Expected bool values for operands of %s, but got %s, %s", op, tyl, tyr)
		}
		ty = "bool"
	}
	return ty
}

func (t *typechecker) typecheckFuncCall(expr *ast.FuncCall) string {
	if _, ok := t.relations[expr]; !ok {
		// FIXME: currently, library functions are only `puts` and `printf`, so this works
		return "void"
	}

	parent := t.relations[expr].(*ast.FuncDecl)

	if len(expr.Params) != len(parent.Params) {
		f := "Wrong number of parameters for %s: expected %d, given %d"
		util.Error(f, expr.Ident.Name, len(parent.Params), len(expr.Params))
	}
	for i, param := range expr.Params {
		if ty := t.typecheckExpr(param); ty != parent.Params[i].Type {
			f := "Expected %s value for %s (#%d parameter of %s), but got %s"
			util.Error(f, parent.Params[i].Type, parent.Params[i].Ident.Name, i+1, expr.Ident.Name, ty)
		}
	}
	return parent.RetType
}

func (t *typechecker) typecheckIdent(expr *ast.Ident) string {
	parent := t.relations[expr].(*ast.VarDecl)
	return parent.Type
}
