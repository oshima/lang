package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/types"
)

/*
 Emitter - emit asm code
*/

type emitter struct {
	refs  map[ast.Node]ast.Node
	types map[ast.Expr]types.Type

	gvars    map[*ast.VarDecl]*gvar
	lvars    map[*ast.VarDecl]*lvar
	strs     map[*ast.StringLit]*str
	garrs    map[*ast.ArrayLit]*garr
	larrs    map[*ast.ArrayLit]*larr
	fns      map[*ast.FuncLit]*fn
	branches map[ast.Node]*branch
}

func (e *emitter) emit(format string, a ...interface{}) {
	fmt.Printf("\t"+format+"\n", a...)
}

func (e *emitter) emitLabel(label string) {
	fmt.Println(label + ":")
}

/* Program */

func (e *emitter) emitProgram(prog *ast.Program) {
	e.emit(".intel_syntax noprefix")

	if len(e.strs) > 0 {
		e.emit(".section .rodata")
	}
	for _, str := range e.strs {
		e.emitLabel(str.label)
		e.emit(".string %q", str.value)
	}

	e.emit(".text")

	for _, gvar := range e.gvars {
		e.emit(".comm %s,%d,%d", gvar.label, gvar.size, gvar.size)
	}

	for _, garr := range e.garrs {
		e.emit(".comm %s,%d,%d", garr.label, garr.len*garr.elemSize, garr.elemSize)
	}

	for expr, _ := range e.fns {
		e.emitFuncCode(expr)
	}

	e.emit(".globl main")
	e.emitLabel("main")
	e.emit("push rbp")
	e.emit("mov rbp, rsp")

	for _, stmt := range prog.Stmts {
		e.emitStmt(stmt)
	}

	e.emit("leave")
	e.emit("ret")
}

func (e *emitter) emitFuncCode(expr *ast.FuncLit) {
	fn := e.fns[expr]
	branch := e.branches[expr]
	endLabel := branch.labels[0]

	e.emitLabel(fn.label)
	e.emit("push rbp")
	e.emit("mov rbp, rsp")
	if fn.localArea > 0 {
		e.emit("sub rsp, %d", fn.localArea)
	}

	for i, param := range expr.Params {
		lvar := e.lvars[param]
		switch lvar.size {
		case 1:
			e.emit("mov byte ptr [rbp-%d], %s", lvar.offset, paramRegs[1][i])
		case 8:
			e.emit("mov qword ptr [rbp-%d], %s", lvar.offset, paramRegs[8][i])
		}
	}

	e.emitBlockStmt(expr.Body)

	e.emitLabel(endLabel)
	if fn.localArea > 0 {
		e.emit("add rsp, %d", fn.localArea)
	}
	e.emit("leave")
	e.emit("ret")
}

/* Stmt */

func (e *emitter) emitStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		e.emitBlockStmt(v)
	case *ast.LetStmt:
		e.emitLetStmt(v)
	case *ast.IfStmt:
		e.emitIfStmt(v)
	case *ast.ForStmt:
		e.emitForStmt(v)
	case *ast.ReturnStmt:
		e.emitReturnStmt(v)
	case *ast.ContinueStmt:
		e.emitContinueStmt(v)
	case *ast.BreakStmt:
		e.emitBreakStmt(v)
	case *ast.AssignStmt:
		e.emitAssignStmt(v)
	case *ast.ExprStmt:
		e.emitExprStmt(v)
	}
}

func (e *emitter) emitBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.Stmts {
		e.emitStmt(stmt_)
	}
}

func (e *emitter) emitLetStmt(stmt *ast.LetStmt) {
	for i, var_ := range stmt.Vars {
		value := stmt.Values[i]

		e.emitExpr(value)

		if lvar, ok := e.lvars[var_]; ok {
			switch lvar.size {
			case 1:
				e.emit("mov byte ptr [rbp-%d], al", lvar.offset)
			case 8:
				e.emit("mov qword ptr [rbp-%d], rax", lvar.offset)
			}
		} else {
			gvar := e.gvars[var_]
			switch gvar.size {
			case 1:
				e.emit("mov byte ptr %s[rip], al", gvar.label)
			case 8:
				e.emit("mov qword ptr %s[rip], rax", gvar.label)
			}
		}
	}
}

func (e *emitter) emitIfStmt(stmt *ast.IfStmt) {
	branch := e.branches[stmt]

	e.emitExpr(stmt.Cond)
	e.emit("cmp rax, 0")

	if stmt.Else == nil {
		endLabel := branch.labels[0]
		e.emit("je %s", endLabel)
		e.emitBlockStmt(stmt.Body)
		e.emitLabel(endLabel)
	} else {
		altLabel := branch.labels[0]
		endLabel := branch.labels[1]
		e.emit("je %s", altLabel)
		e.emitBlockStmt(stmt.Body)
		e.emit("jmp %s", endLabel)
		e.emitLabel(altLabel)
		e.emitStmt(stmt.Else)
		e.emitLabel(endLabel)
	}
}

func (e *emitter) emitForStmt(stmt *ast.ForStmt) {
	branch := e.branches[stmt]
	beginLabel := branch.labels[0]
	endLabel := branch.labels[1]

	e.emitLabel(beginLabel)
	e.emitExpr(stmt.Cond)
	e.emit("cmp rax, 0")
	e.emit("je %s", endLabel)
	e.emitBlockStmt(stmt.Body)
	e.emit("jmp %s", beginLabel)
	e.emitLabel(endLabel)
}

func (e *emitter) emitContinueStmt(stmt *ast.ContinueStmt) {
	ref := e.refs[stmt].(*ast.ForStmt)
	branch := e.branches[ref]
	beginLabel := branch.labels[0]

	e.emit("jmp %s", beginLabel)
}

func (e *emitter) emitBreakStmt(stmt *ast.BreakStmt) {
	ref := e.refs[stmt].(*ast.ForStmt)
	branch := e.branches[ref]
	endLabel := branch.labels[1]

	e.emit("jmp %s", endLabel)
}

func (e *emitter) emitReturnStmt(stmt *ast.ReturnStmt) {
	ref := e.refs[stmt].(*ast.FuncLit)
	branch := e.branches[ref]
	endLabel := branch.labels[0]

	if stmt.Value != nil {
		e.emitExpr(stmt.Value)
	}
	e.emit("jmp %s", endLabel)
}

func (e *emitter) emitAssignStmt(stmt *ast.AssignStmt) {
	for i, value := range stmt.Values {
		e.emitExpr(value)
		if i < len(stmt.Values)-1 {
			e.emit("push rax")
		}
	}
	for i, _ := range stmt.Targets {
		j := len(stmt.Targets) - 1 - i // reverse order
		target := stmt.Targets[j]
		value := stmt.Values[j]

		switch v := target.(type) {
		case *ast.VarRef:
			ref := e.refs[v].(*ast.VarDecl)

			if i > 0 {
				e.emit("pop rax")
			}
			if lvar, ok := e.lvars[ref]; ok {
				switch lvar.size {
				case 1:
					e.emit("mov byte ptr [rbp-%d], al", lvar.offset)
				case 8:
					e.emit("mov qword ptr [rbp-%d], rax", lvar.offset)
				}
			} else {
				gvar := e.gvars[ref]
				switch gvar.size {
				case 1:
					e.emit("mov byte ptr %s[rip], al", gvar.label)
				case 8:
					e.emit("mov qword ptr %s[rip], rax", gvar.label)
				}
			}
		case *ast.IndexExpr:
			if i == 0 {
				e.emit("push rax")
			}
			e.emitExpr(v.Index)
			e.emit("push rax")
			e.emitExpr(v.Left) // rax: address of array head
			e.emit("pop rcx")  // rcx: index
			e.emit("pop rdx")  // rdx: value

			switch sizeOf(e.types[value]) {
			case 1:
				e.emit("mov byte ptr [rax+rcx], dl")
			case 8:
				e.emit("mov qword ptr [rax+rcx*8], rdx")
			}
		}
	}
}

func (e *emitter) emitExprStmt(stmt *ast.ExprStmt) {
	e.emitExpr(stmt.Expr)
}

/* Expr */

func (e *emitter) emitExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		e.emitPrefixExpr(v)
	case *ast.InfixExpr:
		e.emitInfixExpr(v)
	case *ast.IndexExpr:
		e.emitIndexExpr(v)
	case *ast.CallExpr:
		e.emitCallExpr(v)
	case *ast.LibCallExpr:
		e.emitLibCallExpr(v)
	case *ast.VarRef:
		e.emitVarRef(v)
	case *ast.IntLit:
		e.emitIntLit(v)
	case *ast.BoolLit:
		e.emitBoolLit(v)
	case *ast.StringLit:
		e.emitStringLit(v)
	case *ast.ArrayLit:
		e.emitArrayLit(v)
	case *ast.FuncLit:
		e.emitFuncLit(v)
	}
}

func (e *emitter) emitPrefixExpr(expr *ast.PrefixExpr) {
	e.emitExpr(expr.Right)

	switch expr.Op {
	case "!":
		e.emit("xor rax, 1")
	case "-":
		e.emit("neg rax")
	}
}

func (e *emitter) emitInfixExpr(expr *ast.InfixExpr) {
	e.emitExpr(expr.Right)
	e.emit("push rax")
	e.emitExpr(expr.Left)
	e.emit("pop rcx")

	switch expr.Op {
	case "+":
		e.emit("add rax, rcx")
	case "-":
		e.emit("sub rax, rcx")
	case "*":
		e.emit("imul rax, rcx")
	case "/":
		e.emit("cqo")
		e.emit("idiv rcx")
	case "%":
		e.emit("cqo")
		e.emit("idiv rcx")
		e.emit("mov rax, rdx")
	case "&&":
		e.emit("and rax, rcx")
	case "||":
		e.emit("or rax, rcx")
	case "==", "!=", "<", "<=", ">", ">=":
		e.emit("cmp rax, rcx")
		e.emit("%s al", setcc[expr.Op])
		e.emit("movzx rax, al")
	}
}

func (e *emitter) emitIndexExpr(expr *ast.IndexExpr) {
	e.emitExpr(expr.Index)
	e.emit("push rax")
	e.emitExpr(expr.Left)
	e.emit("pop rcx")

	switch sizeOf(e.types[expr]) {
	case 1:
		e.emit("movzx rax, byte ptr [rax+rcx]")
	case 8:
		e.emit("mov rax, qword ptr [rax+rcx*8]")
	}
}

func (e *emitter) emitCallExpr(expr *ast.CallExpr) {
	for _, param := range expr.Params {
		e.emitExpr(param)
		e.emit("push rax")
	}
	for i := range expr.Params {
		j := len(expr.Params) - 1 - i // reverse order
		e.emit("pop %s", paramRegs[8][j])
	}
	e.emitExpr(expr.Left)
	e.emit("call rax")
}

func (e *emitter) emitLibCallExpr(expr *ast.LibCallExpr) {
	for _, param := range expr.Params {
		e.emitExpr(param)
		e.emit("push rax")
	}
	for i := range expr.Params {
		j := len(expr.Params) - 1 - i // reverse order
		e.emit("pop %s", paramRegs[8][j])
	}
	e.emit("call %s", expr.Ident)
}

func (e *emitter) emitVarRef(expr *ast.VarRef) {
	ref := e.refs[expr].(*ast.VarDecl)

	if lvar, ok := e.lvars[ref]; ok {
		switch lvar.size {
		case 1:
			e.emit("movzx rax, byte ptr [rbp-%d]", lvar.offset)
		case 8:
			e.emit("mov rax, qword ptr [rbp-%d]", lvar.offset)
		}
	} else {
		gvar := e.gvars[ref]
		switch gvar.size {
		case 1:
			e.emit("movzx rax, byte ptr %s[rip]", gvar.label)
		case 8:
			e.emit("mov rax, qword ptr %s[rip]", gvar.label)
		}
	}
}

func (e *emitter) emitIntLit(expr *ast.IntLit) {
	e.emit("mov rax, %d", expr.Value)
}

func (e *emitter) emitBoolLit(expr *ast.BoolLit) {
	if expr.Value {
		e.emit("mov rax, 1")
	} else {
		e.emit("mov rax, 0")
	}
}

func (e *emitter) emitStringLit(expr *ast.StringLit) {
	str := e.strs[expr]
	e.emit("mov rax, offset flat:%s", str.label)
}

func (e *emitter) emitArrayLit(expr *ast.ArrayLit) {
	if larr, ok := e.larrs[expr]; ok {
		for i, elem := range expr.Elems {
			e.emitExpr(elem)
			elemOffset := larr.offset - i*larr.elemSize
			switch larr.elemSize {
			case 1:
				e.emit("mov byte ptr [rbp-%d], al", elemOffset)
			case 8:
				e.emit("mov qword ptr [rbp-%d], rax", elemOffset)
			}
		}
		e.emit("lea rax, [rbp-%d]", larr.offset)
	} else {
		garr := e.garrs[expr]
		for i, elem := range expr.Elems {
			e.emitExpr(elem)
			elemOffset := i * garr.elemSize
			switch garr.elemSize {
			case 1:
				e.emit("mov byte ptr %s[rip+%d], al", garr.label, elemOffset)
			case 8:
				e.emit("mov qword ptr %s[rip+%d], rax", garr.label, elemOffset)
			}
		}
		e.emit("mov rax, offset flat:%s", garr.label)
	}
}

func (e *emitter) emitFuncLit(expr *ast.FuncLit) {
	fn := e.fns[expr]
	e.emit("mov rax, offset flat:%s", fn.label)
}
