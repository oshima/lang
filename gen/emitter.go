package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
)

/*
 Emitter - emit asm code
*/

type emitter struct {
	relations map[ast.Node]ast.Node

	fns      map[*ast.FuncDecl]*fn
	gvars    map[*ast.VarDecl]*gvar
	lvars    map[*ast.VarDecl]*lvar
	strs     map[*ast.StringLit]*str
	branches map[ast.Stmt]*branch
}

func (e *emitter) emitProgram(node *ast.Program) {
	e.emit(".intel_syntax noprefix")

	if len(e.strs) > 0 {
		e.emit(".section .rodata")
	}
	for _, s := range e.strs {
		e.emitLabel(s.label)
		e.emit(".string %q", s.value)
	}

	e.emit(".text")

	for _, v := range e.gvars {
		e.emit(".comm %s,%d,%d", v.label, v.size, v.size)
	}

	for stmt, _ := range e.fns {
		e.emitFuncDecl(stmt)
	}

	e.emit(".globl main")
	e.emitLabel("main")
	e.emit("push rbp")
	e.emit("mov rbp, rsp")

	for _, stmt := range node.Stmts {
		e.emitStmt(stmt)
	}

	e.emit("leave")
	e.emit("ret")
}

func (e *emitter) emitStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.VarDecl:
		e.emitVarDecl(v)
	case *ast.BlockStmt:
		e.emitBlockStmt(v)
	case *ast.IfStmt:
		e.emitIfStmt(v)
	case *ast.WhileStmt:
		e.emitWhileStmt(v)
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

func (e *emitter) emitFuncDecl(stmt *ast.FuncDecl) {
	fn := e.fns[stmt]
	br := e.branches[stmt]
	endLabel := br.labels[0]

	e.emitLabel(fn.label)
	e.emit("push rbp")
	e.emit("mov rbp, rsp")
	if fn.align > 0 {
		e.emit("sub rsp, %d", fn.align)
	}

	for i, param := range stmt.Params {
		v := e.lvars[param]
		switch v.size {
		case 1:
			e.emit("mov byte ptr [rbp-%d], %s", v.offset, paramRegs[1][i])
		case 8:
			e.emit("mov qword ptr [rbp-%d], %s", v.offset, paramRegs[8][i])
		}
	}

	e.emitBlockStmt(stmt.Body)

	e.emitLabel(endLabel)
	if fn.align > 0 {
		e.emit("add rsp, %d", fn.align)
	}
	e.emit("leave")
	e.emit("ret")
}

func (e *emitter) emitVarDecl(stmt *ast.VarDecl) {
	e.emitExpr(stmt.Value)

	if v, ok := e.lvars[stmt]; ok {
		switch v.size {
		case 1:
			e.emit("mov byte ptr [rbp-%d], al", v.offset)
		case 8:
			e.emit("mov qword ptr [rbp-%d], rax", v.offset)
		}
	} else {
		v := e.gvars[stmt]
		switch v.size {
		case 1:
			e.emit("mov byte ptr %s[rip], al", v.label)
		case 8:
			e.emit("mov qword ptr %s[rip], rax", v.label)
		}
	}
}

func (e *emitter) emitBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.Stmts {
		e.emitStmt(stmt_)
	}
}

func (e *emitter) emitIfStmt(stmt *ast.IfStmt) {
	br := e.branches[stmt]

	e.emitExpr(stmt.Cond)
	e.emit("cmp rax, 0")

	if stmt.Altern == nil {
		endLabel := br.labels[0]
		e.emit("je %s", endLabel)
		e.emitBlockStmt(stmt.Conseq)
		e.emitLabel(endLabel)
	} else {
		altLabel := br.labels[0]
		endLabel := br.labels[1]
		e.emit("je %s", altLabel)
		e.emitBlockStmt(stmt.Conseq)
		e.emit("jmp %s", endLabel)
		e.emitLabel(altLabel)
		e.emitStmt(stmt.Altern)
		e.emitLabel(endLabel)
	}
}

func (e *emitter) emitWhileStmt(stmt *ast.WhileStmt) {
	br := e.branches[stmt]
	beginLabel := br.labels[0]
	endLabel := br.labels[1]

	e.emitLabel(beginLabel)
	e.emitExpr(stmt.Cond)
	e.emit("cmp rax, 0")
	e.emit("je %s", endLabel)
	e.emitBlockStmt(stmt.Body)
	e.emit("jmp %s", beginLabel)
	e.emitLabel(endLabel)
}

func (e *emitter) emitReturnStmt(stmt *ast.ReturnStmt) {
	parent := e.relations[stmt]
	br := e.branches[parent.(*ast.FuncDecl)]
	endLabel := br.labels[0]

	if stmt.Value != nil {
		e.emitExpr(stmt.Value)
	}
	e.emit("jmp %s", endLabel)
}

func (e *emitter) emitContinueStmt(stmt *ast.ContinueStmt) {
	parent := e.relations[stmt]
	br := e.branches[parent.(*ast.WhileStmt)]
	beginLabel := br.labels[0]

	e.emit("jmp %s", beginLabel)
}

func (e *emitter) emitBreakStmt(stmt *ast.BreakStmt) {
	parent := e.relations[stmt]
	br := e.branches[parent.(*ast.WhileStmt)]
	endLabel := br.labels[1]

	e.emit("jmp %s", endLabel)
}

func (e *emitter) emitAssignStmt(stmt *ast.AssignStmt) {
	parent := e.relations[stmt]

	e.emitExpr(stmt.Value)

	if v, ok := e.lvars[parent.(*ast.VarDecl)]; ok {
		switch v.size {
		case 1:
			e.emit("mov byte ptr [rbp-%d], al", v.offset)
		case 8:
			e.emit("mov qword ptr [rbp-%d], rax", v.offset)
		}
	} else {
		v := e.gvars[parent.(*ast.VarDecl)]
		switch v.size {
		case 1:
			e.emit("mov byte ptr %s[rip], al", v.label)
		case 8:
			e.emit("mov qword ptr %s[rip], rax", v.label)
		}
	}
}

func (e *emitter) emitExprStmt(stmt *ast.ExprStmt) {
	e.emitExpr(stmt.Expr)
}

func (e *emitter) emitExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		e.emitPrefixExpr(v)
	case *ast.InfixExpr:
		e.emitInfixExpr(v)
	case *ast.FuncCall:
		e.emitFuncCall(v)
	case *ast.Ident:
		e.emitIdent(v)
	case *ast.IntLit:
		e.emitIntLit(v)
	case *ast.BoolLit:
		e.emitBoolLit(v)
	case *ast.StringLit:
		e.emitStringLit(v)
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
		e.emitCmp(expr.Op)
	}
}

func (e *emitter) emitCmp(op string) {
	e.emit("cmp rax, rcx")
	e.emit("%s al", setcc[op])
	e.emit("movzx rax, al")
}

func (e *emitter) emitFuncCall(expr *ast.FuncCall) {
	for _, param := range expr.Params {
		e.emitExpr(param)
		e.emit("push rax")
	}
	for i, _ := range expr.Params {
		j := len(expr.Params) - i - 1
		e.emit("pop %s", paramRegs[8][j])
	}

	if parent, ok := e.relations[expr]; ok {
		fn := e.fns[parent.(*ast.FuncDecl)]
		e.emit("call %s", fn.label)
	} else {
		e.emit("call %s", expr.Ident.Name) // library function
	}
}

func (e *emitter) emitIdent(expr *ast.Ident) {
	parent := e.relations[expr]

	if v, ok := e.lvars[parent.(*ast.VarDecl)]; ok {
		switch v.size {
		case 1:
			e.emit("movzx rax, byte ptr [rbp-%d]", v.offset)
		case 8:
			e.emit("mov rax, qword ptr [rbp-%d]", v.offset)
		}
	} else {
		v := e.gvars[parent.(*ast.VarDecl)]
		switch v.size {
		case 1:
			e.emit("movzx rax, byte ptr %s[rip]", v.label)
		case 8:
			e.emit("mov rax, qword ptr %s[rip]", v.label)
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
	s := e.strs[expr]
	e.emit("mov rax, offset flat:%s", s.label)
}

func (e *emitter) emitLabel(label string) {
	fmt.Println(label + ":")
}

func (e *emitter) emit(format string, a ...interface{}) {
	fmt.Printf("\t"+format+"\n", a...)
}
