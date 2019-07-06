package sema

import (
	"fmt"
	"os"

	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/token"
	"github.com/oshima/lang/types"
)

// typechecker performs type checking.
type typechecker struct{}

func (t *typechecker) error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// ----------------------------------------------------------------
// Program

func (t *typechecker) typecheckProgram(prog *ast.Program) {
	for _, stmt := range prog.Stmts {
		t.typecheckStmt(stmt)
	}
}

// ----------------------------------------------------------------
// Stmt

func (t *typechecker) typecheckStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		t.typecheckBlockStmt(v)
	case *ast.VarStmt:
		t.typecheckVarStmt(v)
	case *ast.FuncStmt:
		t.typecheckFuncStmt(v)
	case *ast.IfStmt:
		t.typecheckIfStmt(v)
	case *ast.WhileStmt:
		t.typecheckWhileStmt(v)
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

func (t *typechecker) typecheckBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt := range stmt.Stmts {
		t.typecheckStmt(stmt)
	}
}

func (t *typechecker) typecheckVarStmt(stmt *ast.VarStmt) {
	for _, v := range stmt.Vars {
		t.typecheckVarDecl(v)
	}
}

func (t *typechecker) typecheckFuncStmt(stmt *ast.FuncStmt) {
	t.typecheckFuncDecl(stmt.Func)
}

func (t *typechecker) typecheckIfStmt(stmt *ast.IfStmt) {
	t.typecheckExpr(stmt.Cond)

	if _, ok := stmt.Cond.Type().(*types.Bool); !ok {
		t.error("%s: expected bool condition, but got %s", stmt.Cond.Pos(), stmt.Cond.Type())
	}

	t.typecheckBlockStmt(stmt.Body)

	if stmt.Else != nil {
		t.typecheckStmt(stmt.Else)
	}
}

func (t *typechecker) typecheckWhileStmt(stmt *ast.WhileStmt) {
	t.typecheckExpr(stmt.Cond)

	if _, ok := stmt.Cond.Type().(*types.Bool); !ok {
		t.error("%s: expected bool condition, but got %s", stmt.Cond.Pos(), stmt.Cond.Type())
	}

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckForStmt(stmt *ast.ForStmt) {
	t.typecheckVarDecl(stmt.Iter)

	switch v := stmt.Iter.VarType.(type) {
	case *types.Range:
		stmt.Elem.VarType = new(types.Int)
	case *types.Array:
		stmt.Elem.VarType = v.ElemType
	default:
		t.error("%s: expected range or array, but got %s", stmt.Iter.Value.Pos(), stmt.Iter.VarType)
	}
	stmt.Index.VarType = new(types.Int)

	t.typecheckBlockStmt(stmt.Body)
}

func (t *typechecker) typecheckReturnStmt(stmt *ast.ReturnStmt) {
	var returnType types.Type
	switch v := stmt.Ref.(type) {
	case *ast.FuncDecl:
		returnType = v.ReturnType
	case *ast.FuncLit:
		returnType = v.ReturnType
	}

	if stmt.Value == nil {
		if returnType != nil {
			t.error("%s: expected %s return, but got nothing", stmt.Value.Pos(), returnType)
		}
	} else {
		t.typecheckExpr(stmt.Value)

		if returnType == nil {
			t.error("%s: expected no return, but got %s", stmt.Value.Pos(), stmt.Value.Type())
		}
		if !types.Same(stmt.Value.Type(), returnType) {
			t.error("%s: expected %s return, but got %s", stmt.Value.Pos(), returnType, stmt.Value.Type())
		}
	}
}

func (t *typechecker) typecheckAssignStmt(stmt *ast.AssignStmt) {
	t.typecheckExpr(stmt.Target)
	t.typecheckExpr(stmt.Value)

	switch stmt.Op {
	case token.ASSIGN:
		if !types.Same(stmt.Target.Type(), stmt.Value.Type()) {
			t.error("%s: expected %s value, but got %s", stmt.Value.Pos(), stmt.Target.Type(), stmt.Value.Type())
		}
	default: // +=, -=, *=, /=, %=
		if _, ok := stmt.Target.Type().(*types.Int); !ok {
			t.error("%s: expected int target, but got %s", stmt.Target.Pos(), stmt.Target.Type())
		}
		if _, ok := stmt.Value.Type().(*types.Int); !ok {
			t.error("%s: expected int value, but got %s", stmt.Value.Pos(), stmt.Value.Type())
		}
	}
}

func (t *typechecker) typecheckExprStmt(stmt *ast.ExprStmt) {
	t.typecheckExpr(stmt.Expr)
}

// ----------------------------------------------------------------
// Expr

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
	case *ast.Ident:
		t.typecheckIdent(v)
	case *ast.IntLit:
		v.SetType(new(types.Int))
	case *ast.BoolLit:
		v.SetType(new(types.Bool))
	case *ast.StringLit:
		v.SetType(new(types.String))
	case *ast.RangeLit:
		t.typecheckRangeLit(v)
	case *ast.ArrayLit:
		t.typecheckArrayLit(v)
	case *ast.ArrayShortLit:
		t.typecheckArrayShortLit(v)
	case *ast.FuncLit:
		t.typecheckFuncLit(v)
	}
}

func (t *typechecker) typecheckPrefixExpr(expr *ast.PrefixExpr) {
	t.typecheckExpr(expr.Right)

	switch expr.Op {
	case token.BANG:
		if _, ok := expr.Right.Type().(*types.Bool); !ok {
			t.error("%s: expected bool operand, but got %s", expr.Right.Pos(), expr.Right.Type())
		}
		expr.SetType(new(types.Bool))
	case token.MINUS:
		if _, ok := expr.Right.Type().(*types.Int); !ok {
			t.error("%s: expected int operand, but got %s", expr.Right.Pos(), expr.Right.Type())
		}
		expr.SetType(new(types.Int))
	}
}

func (t *typechecker) typecheckInfixExpr(expr *ast.InfixExpr) {
	t.typecheckExpr(expr.Left)
	t.typecheckExpr(expr.Right)

	if expr.Left.Type() == nil {
		t.error("%s: unexpected void value", expr.Left.Pos())
	}
	if expr.Right.Type() == nil {
		t.error("%s: unexpected void value", expr.Right.Pos())
	}

	switch expr.Op {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT:
		if _, ok := expr.Left.Type().(*types.Int); !ok {
			t.error("%s: expected int operand, but got %s", expr.Left.Pos(), expr.Left.Type())
		}
		if _, ok := expr.Right.Type().(*types.Int); !ok {
			t.error("%s: expected int operand, but got %s", expr.Right.Pos(), expr.Right.Type())
		}
		expr.SetType(new(types.Int))
	case token.EQ, token.NE:
		if !types.Same(expr.Left.Type(), expr.Right.Type()) {
			t.error("%s: expected %s operand, but got %s", expr.Right.Pos(), expr.Left.Type(), expr.Right.Type())
		}
		expr.SetType(new(types.Bool))
	case token.LT, token.LE, token.GT, token.GE:
		if _, ok := expr.Left.Type().(*types.Int); !ok {
			t.error("%s: expected int operand, but got %s", expr.Left.Pos(), expr.Left.Type())
		}
		if _, ok := expr.Right.Type().(*types.Int); !ok {
			t.error("%s: expected int operand, but got %s", expr.Right.Pos(), expr.Right.Type())
		}
		expr.SetType(new(types.Bool))
	case token.AND, token.OR:
		if _, ok := expr.Left.Type().(*types.Bool); !ok {
			t.error("%s: expected bool operand, but got %s", expr.Left.Pos(), expr.Left.Type())
		}
		if _, ok := expr.Right.Type().(*types.Bool); !ok {
			t.error("%s: expected bool operand, but got %s", expr.Right.Pos(), expr.Right.Type())
		}
		expr.SetType(new(types.Bool))
	case token.IN:
		switch v := expr.Right.Type().(type) {
		case *types.Range:
			if _, ok := expr.Left.Type().(*types.Int); !ok {
				t.error("%s: expected int operand, but got %s", expr.Left.Pos(), expr.Left.Type())
			}
		case *types.Array:
			if !types.Same(expr.Left.Type(), v.ElemType) {
				t.error("%s: expected %s operand, but got %s", expr.Left.Pos(), v.ElemType, expr.Left.Type())
			}
		default:
			t.error("%s: expected range or array, but got %s", expr.Right.Pos(), expr.Right.Type())
		}
		expr.SetType(new(types.Bool))
	}
}

func (t *typechecker) typecheckIndexExpr(expr *ast.IndexExpr) {
	t.typecheckExpr(expr.Left)

	arr, ok := expr.Left.Type().(*types.Array)
	if !ok {
		t.error("%s: expected array, but got %s", expr.Left.Pos(), expr.Left.Type())
	}

	t.typecheckExpr(expr.Index)

	if _, ok := expr.Index.Type().(*types.Int); !ok {
		t.error("%s: expected int index, but got %s", expr.Index.Pos(), expr.Index.Type())
	}

	expr.SetType(arr.ElemType)
}

func (t *typechecker) typecheckCallExpr(expr *ast.CallExpr) {
	t.typecheckExpr(expr.Left)

	fn, ok := expr.Left.Type().(*types.Func)
	if !ok {
		t.error("%s: expected function, but got %s", expr.Left.Pos(), expr.Left.Type())
	}

	if len(expr.Params) != len(fn.ParamTypes) {
		t.error("%s: wrong number of parameters (expected %d, got %d)", expr.Pos(), len(fn.ParamTypes), len(expr.Params))
	}
	for i, param := range expr.Params {
		t.typecheckExpr(param)

		if !types.Same(param.Type(), fn.ParamTypes[i]) {
			t.error("%s: expected %s parameter, but got %s", param.Pos(), fn.ParamTypes[i], param.Type())
		}
	}

	expr.SetType(fn.ReturnType)
}

func (t *typechecker) typecheckLibCallExpr(expr *ast.LibCallExpr) {
	for _, param := range expr.Params {
		t.typecheckExpr(param)
	}
	expr.SetType(nil) // both printf and puts return void
}

func (t *typechecker) typecheckIdent(expr *ast.Ident) {
	switch v := expr.Ref.(type) {
	case *ast.VarDecl:
		expr.SetType(v.VarType)
	case *ast.FuncDecl:
		fn := new(types.Func)
		for _, param := range v.Params {
			fn.ParamTypes = append(fn.ParamTypes, param.VarType)
		}
		fn.ReturnType = v.ReturnType
		expr.SetType(fn)
	}
}

func (t *typechecker) typecheckRangeLit(expr *ast.RangeLit) {
	t.typecheckExpr(expr.Lower)
	t.typecheckExpr(expr.Upper)

	if _, ok := expr.Lower.Type().(*types.Int); !ok {
		t.error("%s: expected int boundary, but got %s", expr.Lower.Pos(), expr.Lower.Type())
	}
	if _, ok := expr.Upper.Type().(*types.Int); !ok {
		t.error("%s: expected int boundary, but got %s", expr.Upper.Pos(), expr.Upper.Type())
	}

	expr.SetType(new(types.Range))
}

func (t *typechecker) typecheckArrayLit(expr *ast.ArrayLit) {
	t.typecheckExpr(expr.Elems[0])
	elemType := expr.Elems[0].Type()

	for _, elem := range expr.Elems[1:] {
		t.typecheckExpr(elem)

		if !types.Same(elem.Type(), elemType) {
			t.error("%s: array elements have different types", expr.Pos())
		}
	}

	expr.SetType(&types.Array{Len: len(expr.Elems), ElemType: elemType})
}

func (t *typechecker) typecheckArrayShortLit(expr *ast.ArrayShortLit) {
	if expr.Value != nil {
		t.typecheckExpr(expr.Value)

		if !types.Same(expr.Value.Type(), expr.ElemType) {
			t.error("%s: expected %s element, but got %s", expr.Value.Pos(), expr.ElemType, expr.Value.Type())
		}
	}
	expr.SetType(&types.Array{Len: expr.Len, ElemType: expr.ElemType})
}

func (t *typechecker) typecheckFuncLit(expr *ast.FuncLit) {
	t.typecheckBlockStmt(expr.Body)

	fn := new(types.Func)
	for _, param := range expr.Params {
		fn.ParamTypes = append(fn.ParamTypes, param.VarType)
	}
	fn.ReturnType = expr.ReturnType
	expr.SetType(fn)
}

// ----------------------------------------------------------------
// Decl

func (t *typechecker) typecheckVarDecl(decl *ast.VarDecl) {
	switch v := decl.Value.(type) {
	case *ast.FuncLit:
		fn := new(types.Func)
		for _, param := range v.Params {
			fn.ParamTypes = append(fn.ParamTypes, param.VarType)
		}
		fn.ReturnType = v.ReturnType
		v.SetType(fn)

		if decl.VarType == nil {
			decl.VarType = fn // type inference
		} else {
			if !types.Same(fn, decl.VarType) {
				t.error("%s: expected %s value for %s, but got %s", decl.Value.Pos(), decl.VarType, decl.Name, fn)
			}
		}
		t.typecheckBlockStmt(v.Body)
	default:
		t.typecheckExpr(v)

		if decl.VarType == nil {
			if v.Type() == nil {
				t.error("%s: %s has no initial value", decl.Pos(), decl.Name)
			}
			decl.VarType = v.Type() // type inference
		} else {
			if v.Type() == nil {
				t.error("%s: expected %s value for %s, but got nothing", decl.Value.Pos(), decl.VarType, decl.Name)
			}
			if !types.Same(v.Type(), decl.VarType) {
				t.error("%s: expected %s value for %s, but got %s", decl.Value.Pos(), decl.VarType, decl.Name, v.Type())
			}
		}
	}
}

func (t *typechecker) typecheckFuncDecl(decl *ast.FuncDecl) {
	t.typecheckBlockStmt(decl.Body)
}
