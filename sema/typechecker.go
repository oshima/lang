package sema

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/util"
)

/*
 Typechecker - do type checking
*/

type typechecker struct {
	refs  map[ast.Node]ast.Node
	types map[ast.Expr]string
}

func (t *typechecker) typecheckProgram(prog *ast.Program) {
	for _, stmt := range prog.Stmts {
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
	case *ast.ForStmt:
		t.typecheckForStmt(v)
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
	t.typecheckExpr(stmt.Value)

	if stmt.Type == "" {
		if t.types[stmt.Value] == "void" {
			util.Error("Unexpected void value for %s", stmt.Ident)
		}
		stmt.Type = t.types[stmt.Value] // type inference (write on AST node)
	} else {
		if t.types[stmt.Value] != stmt.Type {
			f := "Expected %s value for %s, but got %s"
			util.Error(f, stmt.Type, stmt.Ident, t.types[stmt.Value])
		}
	}
}

func (t *typechecker) typecheckBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.Stmts {
		t.typecheckStmt(stmt_)
	}
}

func (t *typechecker) typecheckIfStmt(stmt *ast.IfStmt) {
	t.typecheckExpr(stmt.Cond)

	if t.types[stmt.Cond] != "bool" {
		f := "Expected bool value for if condition, but got %s"
		util.Error(f, t.types[stmt.Cond])
	}

	t.typecheckBlockStmt(stmt.Conseq)

	if stmt.Altern != nil {
		t.typecheckStmt(stmt.Altern)
	}
}

func (t *typechecker) typecheckForStmt(stmt *ast.ForStmt) {
	t.typecheckExpr(stmt.Cond)

	if t.types[stmt.Cond] != "bool" {
		f := "Expected bool value for while condition, but got %s"
		util.Error(f, t.types[stmt.Cond])
	}

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckReturnStmt(stmt *ast.ReturnStmt) {
	ref := t.refs[stmt].(*ast.FuncDecl)

	if stmt.Value == nil {
		if ref.ReturnType != "void" {
			f := "Expected %s return for %s, but got void"
			util.Error(f, ref.ReturnType, ref.Ident)
		}
		return
	}

	t.typecheckExpr(stmt.Value)

	if t.types[stmt.Value] != ref.ReturnType {
		f := "Expected %s return for %s, but got %s"
		util.Error(f, ref.ReturnType, ref.Ident, t.types[stmt.Value])
	}
}

func (t *typechecker) typecheckAssignStmt(stmt *ast.AssignStmt) {
	ref := t.refs[stmt].(*ast.VarDecl)

	t.typecheckExpr(stmt.Value)

	if t.types[stmt.Value] != ref.Type {
		f := "Expected %s value for %s, but got %s"
		util.Error(f, ref.Type, stmt.Ident, t.types[stmt.Value])
	}
}

func (t *typechecker) typecheckExprStmt(stmt *ast.ExprStmt) {
	t.typecheckExpr(stmt.Expr)
}

func (t *typechecker) typecheckExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		t.typecheckPrefixExpr(v)
	case *ast.InfixExpr:
		t.typecheckInfixExpr(v)
	case *ast.FuncCall:
		t.typecheckFuncCall(v)
	case *ast.VarRef:
		t.typecheckVarRef(v)
	case *ast.IntLit:
		t.types[v] = "int"
	case *ast.BoolLit:
		t.types[v] = "bool"
	case *ast.StringLit:
		t.types[v] = "string"
	}
}

func (t *typechecker) typecheckPrefixExpr(expr *ast.PrefixExpr) {
	t.typecheckExpr(expr.Right)

	switch expr.Op {
	case "!":
		if t.types[expr.Right] != "bool" {
			f := "Expected bool operand for !, but got %s"
			util.Error(f, t.types[expr.Right])
		}
		t.types[expr] = "bool"
	case "-":
		if t.types[expr.Right] != "int" {
			f := "Expected int operand for -, but got %s"
			util.Error(f, t.types[expr.Right])
		}
		t.types[expr] = "int"
	}
}

func (t *typechecker) typecheckInfixExpr(expr *ast.InfixExpr) {
	t.typecheckExpr(expr.Left)
	t.typecheckExpr(expr.Right)

	switch expr.Op {
	case "+", "-", "*", "/", "%":
		if t.types[expr.Left] != "int" || t.types[expr.Right] != "int" {
			f := "Expected int operands for %s, but got %s, %s"
			util.Error(f, expr.Op, t.types[expr.Left], t.types[expr.Right])
		}
		t.types[expr] = "int"
	case "==", "!=":
		if t.types[expr.Left] == "void" || t.types[expr.Right] == "void" {
			util.Error("Unexpected void operand for %s", expr.Op)
		}
		if t.types[expr.Left] != t.types[expr.Right] {
			f := "Expected same type operands for %s, but got %s, %s"
			util.Error(f, expr.Op, t.types[expr.Left], t.types[expr.Right])
		}
		t.types[expr] = "bool"
	case "<", "<=", ">", ">=":
		if t.types[expr.Left] != "int" || t.types[expr.Right] != "int" {
			f := "Expected int operands for %s, but got %s, %s"
			util.Error(f, expr.Op, t.types[expr.Left], t.types[expr.Right])
		}
		t.types[expr] = "bool"
	case "&&", "||":
		if t.types[expr.Left] != "bool" || t.types[expr.Right] != "bool" {
			f := "Expected bool operands for %s, but got %s, %s"
			util.Error(f, expr.Op, t.types[expr.Left], t.types[expr.Right])
		}
		t.types[expr] = "bool"
	}
}

func (t *typechecker) typecheckFuncCall(expr *ast.FuncCall) {
	if _, ok := t.refs[expr]; !ok {
		// FIXME: currently, library functions are only `puts` and `printf`, so this works
		t.types[expr] = "void"
		return
	}

	ref := t.refs[expr].(*ast.FuncDecl)

	if len(expr.Params) != len(ref.Params) {
		f := "Wrong number of parameters for %s (expected %d, given %d)"
		util.Error(f, expr.Ident, len(ref.Params), len(expr.Params))
	}
	for i, param := range expr.Params {
		t.typecheckExpr(param)
		if t.types[param] != ref.Params[i].Type {
			f := "Expected %s value for %s (#%d parameter of %s), but got %s"
			util.Error(f, ref.Params[i].Type, ref.Params[i].Ident, i+1, expr.Ident, t.types[param])
		}
	}
	t.types[expr] = ref.ReturnType
}

func (t *typechecker) typecheckVarRef(expr *ast.VarRef) {
	ref := t.refs[expr].(*ast.VarDecl)
	t.types[expr] = ref.Type
}
