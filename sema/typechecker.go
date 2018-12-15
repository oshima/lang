package sema

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/types"
	"github.com/oshjma/lang/util"
)

/*
 Typechecker - do type checking
*/

type typechecker struct {
	refs  map[ast.Node]ast.Node
	types map[ast.Expr]types.Type
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
	ty := t.types[stmt.Value]

	if stmt.VarType == types.UNKNOWN {
		if ty == types.VOID {
			util.Error("Unexpected void value for %s", stmt.Ident)
		}
		stmt.VarType = ty // type inference (write on AST node)
	} else {
		if ty != stmt.VarType {
			f := "Expected %s value for %s, but got %s"
			util.Error(f, stmt.VarType, stmt.Ident, ty)
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
	ty := t.types[stmt.Cond]

	if ty != types.BOOL {
		util.Error("Expected bool value for if condition, but got %s", ty)
	}

	t.typecheckBlockStmt(stmt.Conseq)

	if stmt.Altern != nil {
		t.typecheckStmt(stmt.Altern)
	}
}

func (t *typechecker) typecheckForStmt(stmt *ast.ForStmt) {
	t.typecheckExpr(stmt.Cond)
	ty := t.types[stmt.Cond]

	if ty != types.BOOL {
		util.Error("Expected bool value for while condition, but got %s", ty)
	}

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckReturnStmt(stmt *ast.ReturnStmt) {
	ref := t.refs[stmt].(*ast.FuncDecl)

	if stmt.Value == nil {
		if ref.ReturnType != types.VOID {
			f := "Expected %s return in %s, but got void"
			util.Error(f, ref.ReturnType, ref.Ident)
		}
		return
	}

	t.typecheckExpr(stmt.Value)
	ty := t.types[stmt.Value]

	if ty != ref.ReturnType {
		f := "Expected %s return in %s, but got %s"
		util.Error(f, ref.ReturnType, ref.Ident, ty)
	}
}

func (t *typechecker) typecheckAssignStmt(stmt *ast.AssignStmt) {
	ref := t.refs[stmt].(*ast.VarDecl)

	t.typecheckExpr(stmt.Value)
	ty := t.types[stmt.Value]

	if ty != ref.VarType {
		f := "Expected %s value for %s, but got %s"
		util.Error(f, ref.VarType, stmt.Ident, ty)
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
		t.types[v] = types.INT
	case *ast.BoolLit:
		t.types[v] = types.BOOL
	case *ast.StringLit:
		t.types[v] = types.STRING
	}
}

func (t *typechecker) typecheckPrefixExpr(expr *ast.PrefixExpr) {
	t.typecheckExpr(expr.Right)
	ty := t.types[expr.Right]

	switch expr.Op {
	case "!":
		if ty != types.BOOL {
			util.Error("Expected bool operand for !, but got %s", ty)
		}
		t.types[expr] = types.BOOL
	case "-":
		if ty != types.INT {
			util.Error("Expected int operand for -, but got %s", ty)
		}
		t.types[expr] = types.INT
	}
}

func (t *typechecker) typecheckInfixExpr(expr *ast.InfixExpr) {
	t.typecheckExpr(expr.Left)
	t.typecheckExpr(expr.Right)
	lty := t.types[expr.Left]
	rty := t.types[expr.Right]

	switch expr.Op {
	case "+", "-", "*", "/", "%":
		if lty != types.INT || rty != types.INT {
			f := "Expected int operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = types.INT
	case "==", "!=":
		if lty == types.VOID || rty == types.VOID {
			util.Error("Unexpected void operand for %s", expr.Op)
		}
		if lty != rty {
			f := "Expected same type operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = types.BOOL
	case "<", "<=", ">", ">=":
		if lty != types.INT || rty != types.INT {
			f := "Expected int operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = types.BOOL
	case "&&", "||":
		if lty != types.BOOL || rty != types.BOOL {
			f := "Expected bool operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = types.BOOL
	}
}

func (t *typechecker) typecheckFuncCall(expr *ast.FuncCall) {
	if _, ok := t.refs[expr]; !ok {
		// FIXME: currently, library functions are only `puts` and `printf`, so this works
		t.types[expr] = types.VOID
		return
	}

	ref := t.refs[expr].(*ast.FuncDecl)

	if len(expr.Params) != len(ref.Params) {
		f := "Wrong number of parameters for %s (expected %d, given %d)"
		util.Error(f, expr.Ident, len(ref.Params), len(expr.Params))
	}
	for i, param := range expr.Params {
		t.typecheckExpr(param)
		ty := t.types[param]

		if ty != ref.Params[i].VarType {
			f := "Expected %s value for %s (#%d parameter of %s), but got %s"
			util.Error(f, ref.Params[i].VarType, ref.Params[i].Ident, i+1, expr.Ident, ty)
		}
	}
	t.types[expr] = ref.ReturnType
}

func (t *typechecker) typecheckVarRef(expr *ast.VarRef) {
	ref := t.refs[expr].(*ast.VarDecl)
	t.types[expr] = ref.VarType
}
