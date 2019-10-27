package gen

import (
	"fmt"

	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/token"
	"github.com/oshima/lang/types"
)

// emitter emits the target assembly code.
type emitter struct {
	gvars map[ast.Decl]*gvar
	grans map[ast.Expr]*gran
	garrs map[ast.Expr]*garr
	lvars map[ast.Decl]*lvar
	lrans map[ast.Expr]*lran
	larrs map[ast.Expr]*larr
	strs  map[ast.Expr]*str
	fns   map[ast.Node]*fn
	brs   map[ast.Node]*br
}

func (e *emitter) emit(format string, a ...interface{}) {
	fmt.Printf("\t"+format+"\n", a...)
}

func (e *emitter) emitLabel(label string) {
	fmt.Println(label + ":")
}

// ----------------------------------------------------------------
// Program

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

	for _, gran := range e.grans {
		e.emit(".comm %s,%d,%d", gran.label, 16, 8)
	}

	for _, garr := range e.garrs {
		e.emit(".comm %s,%d,%d", garr.label, garr.len*garr.elemSize, garr.elemSize)
	}

	for node := range e.fns {
		e.emitFunc(node)
	}

	e.emit(".global main")
	e.emitLabel("main")
	e.emit("push rbp")
	e.emit("mov rbp, rsp")

	for _, stmt := range prog.Stmts {
		e.emitStmt(stmt)
	}

	e.emit("leave")
	e.emit("ret")
}

func (e *emitter) emitFunc(node ast.Node) {
	fn := e.fns[node]
	br := e.brs[node]
	endLabel := br.labels[0]

	var params []*ast.VarDecl
	var body *ast.BlockStmt

	switch v := node.(type) {
	case *ast.FuncDecl:
		params = v.Params
		body = v.Body
	case *ast.FuncLit:
		params = v.Params
		body = v.Body
	}

	e.emitLabel(fn.label)
	e.emit("push rbp")
	e.emit("mov rbp, rsp")
	if fn.localArea > 0 {
		e.emit("sub rsp, %d", fn.localArea)
	}

	for i, param := range params {
		lvar := e.lvars[param]
		switch lvar.size {
		case 1:
			e.emit("mov byte ptr [rbp-%d], %s", lvar.offset, paramRegs[1][i])
		case 8:
			e.emit("mov qword ptr [rbp-%d], %s", lvar.offset, paramRegs[8][i])
		}
	}

	e.emitBlockStmt(body)

	e.emitLabel(endLabel)
	e.emit("leave")
	e.emit("ret")
}

// ----------------------------------------------------------------
// Stmt

func (e *emitter) emitStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		e.emitBlockStmt(v)
	case *ast.VarStmt:
		e.emitVarStmt(v)
	case *ast.IfStmt:
		e.emitIfStmt(v)
	case *ast.WhileStmt:
		e.emitWhileStmt(v)
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
	for _, stmt := range stmt.Stmts {
		e.emitStmt(stmt)
	}
}

func (e *emitter) emitVarStmt(stmt *ast.VarStmt) {
	for _, v := range stmt.Vars {
		e.emitExpr(v.Value)

		if lvar, ok := e.lvars[v]; ok {
			switch lvar.size {
			case 1:
				e.emit("mov byte ptr [rbp-%d], al", lvar.offset)
			case 8:
				e.emit("mov qword ptr [rbp-%d], rax", lvar.offset)
			}
		} else if gvar, ok := e.gvars[v]; ok {
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
	br := e.brs[stmt]

	e.emitExpr(stmt.Cond)
	e.emit("cmp rax, 0")

	if stmt.Else == nil {
		endLabel := br.labels[0]
		e.emit("je %s", endLabel)
		e.emitBlockStmt(stmt.Body)
		e.emitLabel(endLabel)
	} else {
		elseLabel := br.labels[0]
		endLabel := br.labels[1]
		e.emit("je %s", elseLabel)
		e.emitBlockStmt(stmt.Body)
		e.emit("jmp %s", endLabel)
		e.emitLabel(elseLabel)
		e.emitStmt(stmt.Else)
		e.emitLabel(endLabel)
	}
}

func (e *emitter) emitWhileStmt(stmt *ast.WhileStmt) {
	br := e.brs[stmt]
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

func (e *emitter) emitForStmt(stmt *ast.ForStmt) {
	br := e.brs[stmt]
	beginLabel := br.labels[0]
	continueLabel := br.labels[1]
	endLabel := br.labels[2]

	switch typ := stmt.Iter.VarType.(type) {
	case *types.Range:
		// init
		e.emitExpr(stmt.Iter.Value) // rax: address of range
		if lvar, ok := e.lvars[stmt.Iter]; ok {
			e.emit("mov qword ptr [rbp-%d], rax", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Iter]; ok {
			e.emit("mov qword ptr %s[rip], rax", gvar.label)
		}
		e.emit("mov rcx, qword ptr [rax]") // rcx: lower limit
		if lvar, ok := e.lvars[stmt.Elem]; ok {
			e.emit("mov qword ptr [rbp-%d], rcx", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Elem]; ok {
			e.emit("mov qword ptr %s[rip], rcx", gvar.label)
		}
		if lvar, ok := e.lvars[stmt.Index]; ok {
			e.emit("mov qword ptr [rbp-%d], 0", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Index]; ok {
			e.emit("mov qword ptr %s[rip], 0", gvar.label)
		}

		// cond
		e.emitLabel(beginLabel)
		e.emit("cmp rcx, qword ptr [rax+8]")
		e.emit("jge %s", endLabel)

		// body
		e.emitBlockStmt(stmt.Body)

		// post
		e.emitLabel(continueLabel)
		if lvar, ok := e.lvars[stmt.Iter]; ok {
			e.emit("mov rax, qword ptr [rbp-%d]", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Iter]; ok {
			e.emit("mov rax, qword ptr %s[rip]", gvar.label)
		}
		if lvar, ok := e.lvars[stmt.Elem]; ok {
			e.emit("inc qword ptr [rbp-%d]", lvar.offset)
			e.emit("mov rcx, qword ptr [rbp-%d]", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Elem]; ok {
			e.emit("inc qword ptr %s[rip]", gvar.label)
			e.emit("mov rcx, qword ptr %s[rip]", gvar.label)
		}
		if lvar, ok := e.lvars[stmt.Index]; ok {
			e.emit("inc qword ptr [rbp-%d]", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Index]; ok {
			e.emit("inc qword ptr %s[rip]", gvar.label)
		}
		e.emit("jmp %s", beginLabel)
		e.emitLabel(endLabel)
	case *types.Array:
		// init
		e.emitExpr(stmt.Iter.Value)
		if lvar, ok := e.lvars[stmt.Iter]; ok {
			e.emit("mov qword ptr [rbp-%d], rax", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Iter]; ok {
			e.emit("mov qword ptr %s[rip], rax", gvar.label)
		}
		e.emit("mov rcx, 0")
		if lvar, ok := e.lvars[stmt.Index]; ok {
			e.emit("mov qword ptr [rbp-%d], rcx", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Index]; ok {
			e.emit("mov qword ptr %s[rip], rcx", gvar.label)
		}

		// cond
		e.emitLabel(beginLabel)
		e.emit("cmp rcx, %d", typ.Len)
		e.emit("jge %s", endLabel)

		// pre
		if lvar, ok := e.lvars[stmt.Elem]; ok {
			switch lvar.size {
			case 1:
				e.emit("mov al, byte ptr [rax+rcx]")
				e.emit("mov byte ptr [rbp-%d], al", lvar.offset)
			case 8:
				e.emit("mov rax, qword ptr [rax+rcx*8]")
				e.emit("mov qword ptr [rbp-%d], rax", lvar.offset)
			}
		} else if gvar, ok := e.gvars[stmt.Elem]; ok {
			switch gvar.size {
			case 1:
				e.emit("mov al, byte ptr [rax+rcx]")
				e.emit("mov byte ptr %s[rip], al", gvar.label)
			case 8:
				e.emit("mov rax, qword ptr [rax+rcx*8]")
				e.emit("mov qword ptr %s[rip], rax", gvar.label)
			}
		}

		// body
		e.emitBlockStmt(stmt.Body)

		// post
		e.emitLabel(continueLabel)
		if lvar, ok := e.lvars[stmt.Iter]; ok {
			e.emit("mov rax, qword ptr [rbp-%d]", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Iter]; ok {
			e.emit("mov rax, qword ptr %s[rip]", gvar.label)
		}
		if lvar, ok := e.lvars[stmt.Index]; ok {
			e.emit("inc qword ptr [rbp-%d]", lvar.offset)
			e.emit("mov rcx, qword ptr [rbp-%d]", lvar.offset)
		} else if gvar, ok := e.gvars[stmt.Index]; ok {
			e.emit("inc qword ptr %s[rip]", gvar.label)
			e.emit("mov rcx, qword ptr %s[rip]", gvar.label)
		}
		e.emit("jmp %s", beginLabel)
		e.emitLabel(endLabel)
	}
}

func (e *emitter) emitContinueStmt(stmt *ast.ContinueStmt) {
	br := e.brs[stmt.Ref]

	switch stmt.Ref.(type) {
	case *ast.WhileStmt:
		beginLabel := br.labels[0]
		e.emit("jmp %s", beginLabel)
	case *ast.ForStmt:
		continueLabel := br.labels[1]
		e.emit("jmp %s", continueLabel)
	}
}

func (e *emitter) emitBreakStmt(stmt *ast.BreakStmt) {
	br := e.brs[stmt.Ref]

	switch stmt.Ref.(type) {
	case *ast.WhileStmt:
		endLabel := br.labels[1]
		e.emit("jmp %s", endLabel)
	case *ast.ForStmt:
		endLabel := br.labels[2]
		e.emit("jmp %s", endLabel)
	}
}

func (e *emitter) emitReturnStmt(stmt *ast.ReturnStmt) {
	br := e.brs[stmt.Ref]
	endLabel := br.labels[0]

	if stmt.Value != nil {
		e.emitExpr(stmt.Value)
	}
	e.emit("jmp %s", endLabel)
}

func (e *emitter) emitAssignStmt(stmt *ast.AssignStmt) {
	switch stmt.Op {
	case token.ASSIGN:
		e.emitExpr(stmt.Value)
	default:
		value := &ast.InfixExpr{Left: stmt.Target, Right: stmt.Value}
		switch stmt.Op {
		case token.ADDASSIGN:
			value.Op = token.PLUS
		case token.SUBASSIGN:
			value.Op = token.MINUS
		case token.MULASSIGN:
			value.Op = token.ASTERISK
		case token.DIVASSIGN:
			value.Op = token.SLASH
		case token.MODASSIGN:
			value.Op = token.PERCENT
		}
		e.emitExpr(value)
	}

	switch v := stmt.Target.(type) {
	case *ast.Ident:
		if lvar, ok := e.lvars[v.Ref.(*ast.VarDecl)]; ok {
			switch lvar.size {
			case 1:
				e.emit("mov byte ptr [rbp-%d], al", lvar.offset)
			case 8:
				e.emit("mov qword ptr [rbp-%d], rax", lvar.offset)
			}
		} else if gvar, ok := e.gvars[v.Ref.(*ast.VarDecl)]; ok {
			switch gvar.size {
			case 1:
				e.emit("mov byte ptr %s[rip], al", gvar.label)
			case 8:
				e.emit("mov qword ptr %s[rip], rax", gvar.label)
			}
		}
	case *ast.IndexExpr:
		e.emit("push rax")
		e.emitExpr(v.Index)
		e.emit("push rax")
		e.emitExpr(v.Left) // rax: address of array head
		e.emit("pop rcx")  // rcx: index
		e.emit("pop rdx")  // rdx: value

		switch sizeOf(stmt.Value.Type()) {
		case 1:
			e.emit("mov byte ptr [rax+rcx], dl")
		case 8:
			e.emit("mov qword ptr [rax+rcx*8], rdx")
		}
	}
}

func (e *emitter) emitExprStmt(stmt *ast.ExprStmt) {
	e.emitExpr(stmt.Expr)
}

// ----------------------------------------------------------------
// Expr

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
	case *ast.Ident:
		e.emitIdent(v)
	case *ast.IntLit:
		e.emitIntLit(v)
	case *ast.BoolLit:
		e.emitBoolLit(v)
	case *ast.StringLit:
		e.emitStringLit(v)
	case *ast.RangeLit:
		e.emitRangeLit(v)
	case *ast.ArrayLit:
		e.emitArrayLit(v)
	case *ast.ArrayShortLit:
		e.emitArrayShortLit(v)
	case *ast.FuncLit:
		e.emitFuncLit(v)
	}
}

func (e *emitter) emitPrefixExpr(expr *ast.PrefixExpr) {
	e.emitExpr(expr.Right)

	switch expr.Op {
	case token.BANG:
		e.emit("xor rax, 1")
	case token.MINUS:
		e.emit("neg rax")
	}
}

func (e *emitter) emitInfixExpr(expr *ast.InfixExpr) {
	switch expr.Op {
	case token.AND, token.OR:
		// do nothing
	default:
		// emit both operands in advance
		e.emitExpr(expr.Left)
		e.emit("push rax")
		e.emitExpr(expr.Right)
		e.emit("mov rcx, rax") // rcx: right
		e.emit("pop rax")      // rax: left
	}

	switch expr.Op {
	case token.PLUS:
		e.emit("add rax, rcx")
	case token.MINUS:
		e.emit("sub rax, rcx")
	case token.ASTERISK:
		e.emit("imul rax, rcx")
	case token.SLASH:
		e.emit("cqo")
		e.emit("idiv rcx")
	case token.PERCENT:
		e.emit("cqo")
		e.emit("idiv rcx")
		e.emit("mov rax, rdx")
	case token.EQ, token.NE, token.LT, token.LE, token.GT, token.GE:
		e.emit("cmp rax, rcx")
		e.emit("%s al", setcc[expr.Op])
		e.emit("movzx rax, al")
	case token.AND:
		br := e.brs[expr]
		endLabel := br.labels[0]

		e.emitExpr(expr.Left)
		e.emit("cmp rax, 0")
		e.emit("je %s", endLabel)
		e.emitExpr(expr.Right)
		e.emitLabel(endLabel)
	case token.OR:
		br := e.brs[expr]
		endLabel := br.labels[0]

		e.emitExpr(expr.Left)
		e.emit("cmp rax, 1")
		e.emit("je %s", endLabel)
		e.emitExpr(expr.Right)
		e.emitLabel(endLabel)
	case token.IN:
		switch v := expr.Right.Type().(type) {
		case *types.Range:
			br := e.brs[expr]
			falseLabel := br.labels[0]
			endLabel := br.labels[1]

			e.emit("cmp rax, qword ptr [rcx]")
			e.emit("jl %s", falseLabel)
			e.emit("cmp rax, qword ptr [rcx+8]")
			e.emit("jge %s", falseLabel)
			e.emit("mov rax, 1")
			e.emit("jmp %s", endLabel)
			e.emitLabel(falseLabel)
			e.emit("mov rax, 0")
			e.emitLabel(endLabel)
		case *types.Array:
			br := e.brs[expr]
			beginLabel := br.labels[0]
			falseLabel := br.labels[1]
			endLabel := br.labels[2]

			len := v.Len
			elemSize := sizeOf(v.ElemType)

			e.emit("mov rdx, rcx")
			e.emit("add rdx, %d", len*elemSize)
			e.emitLabel(beginLabel)
			e.emit("cmp rcx, rdx")
			e.emit("jge %s", falseLabel)
			switch elemSize {
			case 1:
				e.emit("cmp al, byte ptr [rcx]")
			case 8:
				e.emit("cmp rax, qword ptr [rcx]")
			}
			e.emit("lea rcx, [rcx+%d]", elemSize)
			e.emit("jne %s", beginLabel)
			e.emit("mov rax, 1")
			e.emit("jmp %s", endLabel)
			e.emitLabel(falseLabel)
			e.emit("mov rax, 0")
			e.emitLabel(endLabel)
		}
	}
}

func (e *emitter) emitIndexExpr(expr *ast.IndexExpr) {
	e.emitExpr(expr.Index)
	e.emit("push rax")
	e.emitExpr(expr.Left)
	e.emit("pop rcx")

	switch sizeOf(expr.Type()) {
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
	if v, ok := expr.Left.(*ast.Ident); ok {
		if v, ok := v.Ref.(*ast.FuncDecl); ok {
			fn := e.fns[v]
			e.emit("call %s", fn.label)
			return
		}
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
	e.emit("call %s", expr.Name)
}

func (e *emitter) emitIdent(expr *ast.Ident) {
	switch v := expr.Ref.(type) {
	case *ast.VarDecl:
		if lvar, ok := e.lvars[v]; ok {
			switch lvar.size {
			case 1:
				e.emit("movzx rax, byte ptr [rbp-%d]", lvar.offset)
			case 8:
				e.emit("mov rax, qword ptr [rbp-%d]", lvar.offset)
			}
		} else if gvar, ok := e.gvars[v]; ok {
			switch gvar.size {
			case 1:
				e.emit("movzx rax, byte ptr %s[rip]", gvar.label)
			case 8:
				e.emit("mov rax, qword ptr %s[rip]", gvar.label)
			}
		}
	case *ast.FuncDecl:
		fn := e.fns[v]
		e.emit("mov rax, offset flat:%s", fn.label)
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

func (e *emitter) emitRangeLit(expr *ast.RangeLit) {
	if lran, ok := e.lrans[expr]; ok {
		e.emitExpr(expr.Lower)
		e.emit("mov qword ptr [rbp-%d], rax", lran.offset)
		e.emitExpr(expr.Upper)
		e.emit("mov qword ptr [rbp-%d], rax", lran.offset-8)
		e.emit("lea rax, [rbp-%d]", lran.offset)
	} else if gran, ok := e.grans[expr]; ok {
		e.emitExpr(expr.Lower)
		e.emit("mov qword ptr %s[rip], rax", gran.label)
		e.emitExpr(expr.Upper)
		e.emit("mov qword ptr %s[rip+8], rax", gran.label)
		e.emit("mov rax, offset flat:%s", gran.label)
	}
}

func (e *emitter) emitArrayLit(expr *ast.ArrayLit) {
	if larr, ok := e.larrs[expr]; ok {
		for i, elem := range expr.Elems {
			e.emitExpr(elem)
			offset := larr.offset - i*larr.elemSize
			switch larr.elemSize {
			case 1:
				e.emit("mov byte ptr [rbp-%d], al", offset)
			case 8:
				e.emit("mov qword ptr [rbp-%d], rax", offset)
			}
		}
		e.emit("lea rax, [rbp-%d]", larr.offset)
	} else if garr, ok := e.garrs[expr]; ok {
		for i, elem := range expr.Elems {
			e.emitExpr(elem)
			offset := i * garr.elemSize
			switch garr.elemSize {
			case 1:
				e.emit("mov byte ptr %s[rip+%d], al", garr.label, offset)
			case 8:
				e.emit("mov qword ptr %s[rip+%d], rax", garr.label, offset)
			}
		}
		e.emit("mov rax, offset flat:%s", garr.label)
	}
}

func (e *emitter) emitArrayShortLit(expr *ast.ArrayShortLit) {
	if expr.Value != nil {
		e.emitExpr(expr.Value)

		if larr, ok := e.larrs[expr]; ok {
			e.emit("lea rdi, [rbp-%d]", larr.offset)
			e.emit("mov rcx, %d", larr.len)
			switch larr.elemSize {
			case 1:
				e.emit("rep stosb")
			case 8:
				e.emit("rep stosq")
			}
		} else if garr, ok := e.garrs[expr]; ok {
			e.emit("mov rdi, offset flat:%s", garr.label)
			e.emit("mov rcx, %d", garr.len)
			switch garr.elemSize {
			case 1:
				e.emit("rep stosb")
			case 8:
				e.emit("rep stosq")
			}
		}
	}

	if larr, ok := e.larrs[expr]; ok {
		e.emit("lea rax, [rbp-%d]", larr.offset)
	} else if garr, ok := e.garrs[expr]; ok {
		e.emit("mov rax, offset flat:%s", garr.label)
	}
}

func (e *emitter) emitFuncLit(expr *ast.FuncLit) {
	fn := e.fns[expr]
	e.emit("mov rax, offset flat:%s", fn.label)
}
