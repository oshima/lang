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

/* Program */

func (t *typechecker) typecheckProgram(prog *ast.Program) {
	for _, stmt := range prog.Stmts {
		t.typecheckStmt(stmt)
	}
}

/* Stmt */

func (t *typechecker) typecheckStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		t.typecheckBlockStmt(v)
	case *ast.LetStmt:
		t.typecheckLetStmt(v)
	case *ast.IfStmt:
		t.typecheckIfStmt(v)
	case *ast.ForStmt:
		t.typecheckForStmt(v)
	case *ast.ForInStmt:
		t.typecheckForInStmt(v)
	case *ast.ReturnStmt:
		t.typecheckReturnStmt(v)
	case *ast.AssignStmt:
		t.typecheckAssignStmt(v)
	case *ast.ExprStmt:
		t.typecheckExprStmt(v)
	}
}

func (t *typechecker) typecheckBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.Stmts {
		t.typecheckStmt(stmt_)
	}
}

func (t *typechecker) typecheckLetStmt(stmt *ast.LetStmt) {
	for i, var_ := range stmt.Vars {
		value := stmt.Values[i]

		if fn, ok := value.(*ast.FuncLit); ok {
			paramTypes := make([]types.Type, 0, 4)
			for _, param := range fn.Params {
				paramTypes = append(paramTypes, param.VarType)
			}
			ty := &types.Func{ParamTypes: paramTypes, ReturnType: fn.ReturnType}

			t.types[fn] = ty

			if var_.VarType == nil {
				var_.VarType = ty // type inference (write on AST node)
			} else {
				if !types.Same(ty, var_.VarType) {
					f := "Expected %s value for %s, but got %s"
					util.Error(f, var_.VarType, var_.Ident, ty)
				}
			}

			t.typecheckBlockStmt(fn.Body)
		} else {
			t.typecheckExpr(value)
			ty := t.types[value]

			if var_.VarType == nil {
				if ty == nil {
					util.Error("No initial value for %s", var_.Ident)
				}
				var_.VarType = ty // type inference (write on AST node)
			} else {
				if ty == nil {
					f := "Expected %s value for %s, but got nothing"
					util.Error(f, var_.VarType, var_.Ident)
				}
				if !types.Same(ty, var_.VarType) {
					f := "Expected %s value for %s, but got %s"
					util.Error(f, var_.VarType, var_.Ident, ty)
				}
			}
		}
	}
}

func (t *typechecker) typecheckIfStmt(stmt *ast.IfStmt) {
	t.typecheckExpr(stmt.Cond)
	ty := t.types[stmt.Cond]

	if _, ok := ty.(*types.Bool); !ok {
		util.Error("Expected bool value for if condition, but got %s", ty)
	}

	t.typecheckBlockStmt(stmt.Body)

	if stmt.Else != nil {
		t.typecheckStmt(stmt.Else)
	}
}

func (t *typechecker) typecheckForStmt(stmt *ast.ForStmt) {
	t.typecheckExpr(stmt.Cond)
	ty := t.types[stmt.Cond]

	if _, ok := ty.(*types.Bool); !ok {
		util.Error("Expected bool value for while condition, but got %s", ty)
	}

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckForInStmt(stmt *ast.ForInStmt) {
	t.typecheckExpr(stmt.Expr)
	ty := t.types[stmt.Expr]

	arr, ok := ty.(*types.Array)
	if !ok {
		util.Error("Expected array to iterate, but got %s", ty)
	}

	if stmt.Elem.VarType == nil {
		stmt.Elem.VarType = arr.ElemType
	} else {
		if !types.Same(arr.ElemType, stmt.Elem.VarType) {
			f := "Expected array of %s elements, but got %s"
			util.Error(f, stmt.Elem.VarType, arr.ElemType)
		}
	}
	if stmt.Index.VarType == nil {
		stmt.Index.VarType = &types.Int{}
	} else {
		if _, ok := stmt.Index.VarType.(*types.Int); !ok {
			util.Error("Index variable must be int type")
		}
	}
	stmt.Array.VarType = arr

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckReturnStmt(stmt *ast.ReturnStmt) {
	ref := t.refs[stmt].(*ast.FuncLit)

	if stmt.Value == nil {
		if ref.ReturnType != nil {
			f := "Expected %s return in function, but got nothing"
			util.Error(f, ref.ReturnType)
		}
	} else {
		t.typecheckExpr(stmt.Value)
		ty := t.types[stmt.Value]

		if ref.ReturnType == nil {
			f := "Expected no return in function, but got %s"
			util.Error(f, ty)
		}
		if !types.Same(ty, ref.ReturnType) {
			f := "Expected %s return in function, but got %s"
			util.Error(f, ref.ReturnType, ty)
		}
	}
}

func (t *typechecker) typecheckAssignStmt(stmt *ast.AssignStmt) {
	for i, target := range stmt.Targets {
		value := stmt.Values[i]

		t.typecheckExpr(target)
		t.typecheckExpr(value)
		tty := t.types[target]
		vty := t.types[value]

		if !types.Same(tty, vty) {
			f := "Expected %s value in assignment, but got %s"
			util.Error(f, tty, vty)
		}
	}
}

func (t *typechecker) typecheckExprStmt(stmt *ast.ExprStmt) {
	t.typecheckExpr(stmt.Expr)
}

/* Expr */

func (t *typechecker) typecheckExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		t.typecheckPrefixExpr(v)
	case *ast.InfixExpr:
		t.typecheckInfixExpr(v)
	case *ast.IndexExpr:
		t.typecheckIndexExpr(v)
	case *ast.CallExpr:
		t.typecheckCallExpr(v)
	case *ast.LibCallExpr:
		t.typecheckLibCallExpr(v)
	case *ast.VarRef:
		t.typecheckVarRef(v)
	case *ast.IntLit:
		t.types[v] = &types.Int{}
	case *ast.BoolLit:
		t.types[v] = &types.Bool{}
	case *ast.StringLit:
		t.types[v] = &types.String{}
	case *ast.ArrayLit:
		t.typecheckArrayLit(v)
	case *ast.FuncLit:
		t.typecheckFuncLit(v)
	}
}

func (t *typechecker) typecheckPrefixExpr(expr *ast.PrefixExpr) {
	t.typecheckExpr(expr.Right)
	ty := t.types[expr.Right]

	switch expr.Op {
	case "!":
		if _, ok := ty.(*types.Bool); !ok {
			util.Error("Expected bool operand for !, but got %s", ty)
		}
		t.types[expr] = &types.Bool{}
	case "-":
		if _, ok := ty.(*types.Int); !ok {
			util.Error("Expected int operand for -, but got %s", ty)
		}
		t.types[expr] = &types.Int{}
	}
}

func (t *typechecker) typecheckInfixExpr(expr *ast.InfixExpr) {
	t.typecheckExpr(expr.Left)
	t.typecheckExpr(expr.Right)
	lty := t.types[expr.Left]
	rty := t.types[expr.Right]

	switch expr.Op {
	case "+", "-", "*", "/", "%":
		_, lok := lty.(*types.Int)
		_, rok := rty.(*types.Int)
		if !lok || !rok {
			f := "Expected int operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = &types.Int{}
	case "==", "!=":
		if lty == nil || rty == nil {
			util.Error("Unexpected void operand for %s", expr.Op)
		}
		if !types.Same(lty, rty) {
			f := "Expected same type operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = &types.Bool{}
	case "<", "<=", ">", ">=":
		_, lok := lty.(*types.Int)
		_, rok := rty.(*types.Int)
		if !lok || !rok {
			f := "Expected int operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = &types.Bool{}
	case "&&", "||":
		_, lok := lty.(*types.Bool)
		_, rok := rty.(*types.Bool)
		if !lok || !rok {
			f := "Expected bool operands for %s, but got %s, %s"
			util.Error(f, expr.Op, lty, rty)
		}
		t.types[expr] = &types.Bool{}
	}
}

func (t *typechecker) typecheckIndexExpr(expr *ast.IndexExpr) {
	t.typecheckExpr(expr.Left)
	lty := t.types[expr.Left]

	arr, ok := lty.(*types.Array)
	if !ok {
		util.Error("Expected array to index, but got %s", lty)
	}

	t.typecheckExpr(expr.Index)
	ity := t.types[expr.Index]

	if _, ok := ity.(*types.Int); !ok {
		util.Error("Expected int index, but got %s", ity)
	}

	t.types[expr] = arr.ElemType
}

func (t *typechecker) typecheckCallExpr(expr *ast.CallExpr) {
	t.typecheckExpr(expr.Left)
	ty := t.types[expr.Left]

	fn, ok := ty.(*types.Func)
	if !ok {
		util.Error("Expected function to call, but got %s", ty)
	}

	if len(expr.Params) != len(fn.ParamTypes) {
		f := "Wrong number of parameters (expected %d, given %d)"
		util.Error(f, len(fn.ParamTypes), len(expr.Params))
	}
	for i, param := range expr.Params {
		paramType := fn.ParamTypes[i]

		t.typecheckExpr(param)
		ty := t.types[param]

		if !types.Same(ty, paramType) {
			f := "Expected %s value for #%d parameter, but got %s"
			util.Error(f, paramType, i+1, ty)
		}
	}

	t.types[expr] = fn.ReturnType
}

func (t *typechecker) typecheckLibCallExpr(expr *ast.LibCallExpr) {
	for _, param := range expr.Params {
		t.typecheckExpr(param)
	}
	t.types[expr] = nil // FIXME
}

func (t *typechecker) typecheckVarRef(expr *ast.VarRef) {
	ref := t.refs[expr].(*ast.VarDecl)
	t.types[expr] = ref.VarType
}

func (t *typechecker) typecheckArrayLit(expr *ast.ArrayLit) {
	for _, elem := range expr.Elems {
		t.typecheckExpr(elem)
		ty := t.types[elem]

		if !types.Same(ty, expr.ElemType) {
			f := "Expected %s value for array element, but got %s"
			util.Error(f, expr.ElemType, ty)
		}
	}
	t.types[expr] = &types.Array{Len: expr.Len, ElemType: expr.ElemType}
}

func (t *typechecker) typecheckFuncLit(expr *ast.FuncLit) {
	t.typecheckBlockStmt(expr.Body)

	paramTypes := make([]types.Type, 0, 4)
	for _, param := range expr.Params {
		paramTypes = append(paramTypes, param.VarType)
	}
	t.types[expr] = &types.Func{ParamTypes: paramTypes, ReturnType: expr.ReturnType}
}
