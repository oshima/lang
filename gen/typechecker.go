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
	ty := t.typecheckExpr(stmt.Value)

	if stmt.Type == "" {
		stmt.Type = ty // type inference (overwrite on AST node)
		return
	}

	if ty != stmt.Type {
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
		util.Error("Expected bool value for while condition, but got %s", ty)
	}

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckReturnStmt(stmt *ast.ReturnStmt) {
	parent := t.relations[stmt].(*ast.FuncDecl)

	if stmt.Value == nil {
		if parent.ReturnType != "void" {
			f := "Expected %s return for %s, but got void"
			util.Error(f, parent.ReturnType, parent.Ident.Name)
		}
	} else {
		if ty := t.typecheckExpr(stmt.Value); ty != parent.ReturnType {
			f := "Expected %s return for %s, but got %s"
			util.Error(f, parent.ReturnType, parent.Ident.Name, ty)
		}
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
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		return t.typecheckPrefixExpr(v)
	case *ast.InfixExpr:
		return t.typecheckInfixExpr(v)
	case *ast.FuncCall:
		return t.typecheckFuncCall(v)
	case *ast.Ident:
		return t.typecheckIdent(v)
	case *ast.IntLit:
		return "int"
	case *ast.BoolLit:
		return "bool"
	case *ast.StringLit:
		return "string"
	default:
		// unreachable here
		return ""
	}
}

func (t *typechecker) typecheckPrefixExpr(expr *ast.PrefixExpr) string {
	rty := t.typecheckExpr(expr.Right)

	switch expr.Op {
	case "!":
		if rty != "bool" {
			util.Error("Expected bool operand for !, but got %s", rty)
		}
		return "bool"
	case "-":
		if rty != "int" {
			util.Error("Expected int operand for -, but got %s", rty)
		}
		return "int"
	default:
		// unreachable here
		return ""
	}
}

func (t *typechecker) typecheckInfixExpr(expr *ast.InfixExpr) string {
	lty := t.typecheckExpr(expr.Left)
	rty := t.typecheckExpr(expr.Right)

	switch op := expr.Op; op {
	case "+", "-", "*", "/", "%":
		if lty != "int" || rty != "int" {
			util.Error("Expected int operands for %s, but got %s, %s", op, lty, rty)
		}
		return "int"
	case "==", "!=":
		if lty == "void" || rty == "void" {
			util.Error("Unexpected void operand for %s", op)
		}
		if lty != rty {
			util.Error("Expected same type operands for %s, but got %s, %s", op, lty, rty)
		}
		return "bool"
	case "<", "<=", ">", ">=":
		if lty != "int" || rty != "int" {
			util.Error("Expected int operands for %s, but got %s, %s", op, lty, rty)
		}
		return "bool"
	case "&&", "||":
		if lty != "bool" || rty != "bool" {
			util.Error("Expected bool operands for %s, but got %s, %s", op, lty, rty)
		}
		return "bool"
	default:
		// unreachable here
		return ""
	}
}

func (t *typechecker) typecheckFuncCall(expr *ast.FuncCall) string {
	if _, ok := t.relations[expr]; !ok {
		// FIXME: currently, library functions are only `puts` and `printf`, so this works
		return "void"
	}

	parent := t.relations[expr].(*ast.FuncDecl)

	if len(expr.Params) != len(parent.Params) {
		f := "Wrong number of parameters for %s (expected %d, given %d)"
		util.Error(f, expr.Ident.Name, len(parent.Params), len(expr.Params))
	}
	for i, param := range expr.Params {
		if ty := t.typecheckExpr(param); ty != parent.Params[i].Type {
			f := "Expected %s value for %s (#%d parameter of %s), but got %s"
			util.Error(f, parent.Params[i].Type, parent.Params[i].Ident.Name, i+1, expr.Ident.Name, ty)
		}
	}
	return parent.ReturnType
}

func (t *typechecker) typecheckIdent(expr *ast.Ident) string {
	parent := t.relations[expr].(*ast.VarDecl)
	return parent.Type
}
