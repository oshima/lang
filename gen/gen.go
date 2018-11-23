package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/util"
)

func Generate(node *ast.Program) {
	g := &generator{
		fns:       make(map[*ast.FuncDecl]*fn),
		gvars:     make(map[*ast.VarDecl]*gvar),
		lvars:     make(map[*ast.VarDecl]*lvar),
		strs:      make(map[*ast.StringLit]*str),
		branches:  make(map[ast.Stmt]*branch),
		relations: make(map[ast.Node]ast.Node),
	}
	g.traverseProgram(node, newEnv(nil))
	g.emitProgram(node)
}

type generator struct {
	// Information necessary for emitting assembly code
	fns      map[*ast.FuncDecl]*fn
	gvars    map[*ast.VarDecl]*gvar
	lvars    map[*ast.VarDecl]*lvar
	strs     map[*ast.StringLit]*str
	branches map[ast.Stmt]*branch

	// Relationship between AST nodes: child -> parent
	relations map[ast.Node]ast.Node

	// Counters of labels
	nGvarLabel   int
	nStrLabel    int
	nBranchLabel int

	// Used for finding local variables
	local  bool
	offset int
}

type fn struct {
	label string
	align int
}

type gvar struct {
	label string
	size  int
}

type lvar struct {
	offset int
	size   int
}

type str struct {
	label string
	value string
}

type branch struct {
	labels []string
}

func (g *generator) gvarLabel(name string) string {
	label := fmt.Sprintf(".GV%d_%s", g.nGvarLabel, name)
	g.nGvarLabel += 1
	return label
}

func (g *generator) strLabel() string {
	label := fmt.Sprintf(".LC%d", g.nStrLabel)
	g.nStrLabel += 1
	return label
}

func (g *generator) branchLabel() string {
	label := fmt.Sprintf(".L%d", g.nBranchLabel)
	g.nBranchLabel += 1
	return label
}

/*
 Traverse AST to gather the necessary information for emitting assembly code
*/

func (g *generator) traverseProgram(node *ast.Program, e *env) {
	for _, stmt := range node.Stmts {
		g.traverseStmt(stmt, e)
	}
}

func (g *generator) traverseStmt(stmt ast.Stmt, e *env) {
	switch v := stmt.(type) {
	case *ast.FuncDecl:
		g.traverseFuncDecl(v, e)
	case *ast.VarDecl:
		g.traverseVarDecl(v, e)
	case *ast.BlockStmt:
		g.traverseBlockStmt(v, newEnv(e))
	case *ast.IfStmt:
		g.traverseIfStmt(v, e)
	case *ast.WhileStmt:
		g.traverseWhileStmt(v, e)
	case *ast.ReturnStmt:
		g.traverseReturnStmt(v, e)
	case *ast.ContinueStmt:
		g.traverseContinueStmt(v, e)
	case *ast.BreakStmt:
		g.traverseBreakStmt(v, e)
	case *ast.AssignStmt:
		g.traverseAssignStmt(v, e)
	case *ast.ExprStmt:
		g.traverseExprStmt(v, e)
	}
}

func (g *generator) traverseFuncDecl(stmt *ast.FuncDecl, e *env) {
	if g.local {
		util.Error("Function declarations cannot be nested")
	}
	if err := e.set(stmt.Ident.Name, stmt); err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
	}

	g.local = true
	g.offset = 0

	e_ := newEnv(e)
	e_.set("return", stmt)

	for _, param := range stmt.Params {
		g.traverseVarDecl(param, e_)
	}
	g.traverseBlockStmt(stmt.Body, e_)
	endLabel := g.branchLabel()

	g.local = false
	g.fns[stmt] = &fn{
		label: stmt.Ident.Name,
		align: align(g.offset, 16),
	}
	g.branches[stmt] = &branch{labels: []string{endLabel}}
}

func (g *generator) traverseVarDecl(stmt *ast.VarDecl, e *env) {
	if stmt.Value != nil {
		g.traverseExpr(stmt.Value, e)
	}

	if err := e.set(stmt.Ident.Name, stmt); err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
	}

	if g.local {
		size := sizeof[stmt.Type]
		g.offset = align(g.offset+size, size)
		g.lvars[stmt] = &lvar{offset: g.offset, size: size}
	} else {
		g.gvars[stmt] = &gvar{
			label: g.gvarLabel(stmt.Ident.Name),
			size:  sizeof[stmt.Type],
		}
	}
}

func (g *generator) traverseBlockStmt(stmt *ast.BlockStmt, e *env) {
	for _, stmt_ := range stmt.Stmts {
		g.traverseStmt(stmt_, e)
	}
}

func (g *generator) traverseIfStmt(stmt *ast.IfStmt, e *env) {
	g.traverseExpr(stmt.Cond, e)
	g.traverseBlockStmt(stmt.Conseq, newEnv(e))

	if stmt.Altern == nil {
		endLabel := g.branchLabel()
		g.branches[stmt] = &branch{labels: []string{endLabel}}
	} else {
		altLabel := g.branchLabel()
		g.traverseStmt(stmt.Altern, e)
		endLabel := g.branchLabel()
		g.branches[stmt] = &branch{labels: []string{altLabel, endLabel}}
	}
}

func (g *generator) traverseWhileStmt(stmt *ast.WhileStmt, e *env) {
	beginLabel := g.branchLabel()
	g.traverseExpr(stmt.Cond, e)

	e_ := newEnv(e)
	e_.set("continue", stmt)
	e_.set("break", stmt)

	g.traverseBlockStmt(stmt.Body, e_)
	endLabel := g.branchLabel()
	g.branches[stmt] = &branch{labels: []string{beginLabel, endLabel}}
}

func (g *generator) traverseReturnStmt(stmt *ast.ReturnStmt, e *env) {
	if stmt.Value != nil {
		g.traverseExpr(stmt.Value, e)
	}

	parent, ok := e.get("return")
	if !ok {
		util.Error("Illegal use of return")
	}
	g.relations[stmt] = parent
}

func (g *generator) traverseContinueStmt(stmt *ast.ContinueStmt, e *env) {
	parent, ok := e.get("continue")
	if !ok {
		util.Error("Illegal use of continue")
	}
	g.relations[stmt] = parent
}

func (g *generator) traverseBreakStmt(stmt *ast.BreakStmt, e *env) {
	parent, ok := e.get("break")
	if !ok {
		util.Error("Illegal use of break")
	}
	g.relations[stmt] = parent
}

func (g *generator) traverseAssignStmt(stmt *ast.AssignStmt, e *env) {
	g.traverseExpr(stmt.Value, e)

	parent, ok := e.get(stmt.Ident.Name)
	if !ok {
		util.Error("%s is not declared", stmt.Ident.Name)
	}
	if _, ok := parent.(*ast.VarDecl); !ok {
		util.Error("%s is not a variable")
	}
	g.relations[stmt] = parent
}

func (g *generator) traverseExprStmt(stmt *ast.ExprStmt, e *env) {
	g.traverseExpr(stmt.Expr, e)
}

func (g *generator) traverseExpr(expr ast.Expr, e *env) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		g.traversePrefixExpr(v, e)
	case *ast.InfixExpr:
		g.traverseInfixExpr(v, e)
	case *ast.FuncCall:
		g.traverseFuncCall(v, e)
	case *ast.Ident:
		g.traverseIdent(v, e)
	case *ast.StringLit:
		g.traverseStringLit(v)
	}
}

func (g *generator) traversePrefixExpr(expr *ast.PrefixExpr, e *env) {
	g.traverseExpr(expr.Right, e)
}

func (g *generator) traverseInfixExpr(expr *ast.InfixExpr, e *env) {
	g.traverseExpr(expr.Left, e)
	g.traverseExpr(expr.Right, e)
}

func (g *generator) traverseFuncCall(expr *ast.FuncCall, e *env) {
	for _, param := range expr.Params {
		g.traverseExpr(param, e)
	}

	if parent, ok := e.get(expr.Ident.Name); ok {
		if _, ok := parent.(*ast.FuncDecl); !ok {
			util.Error("%s is not a function", expr.Ident.Name)
		}
		g.relations[expr] = parent
	} else {
		if _, ok := libFns[expr.Ident.Name]; !ok {
			util.Error("%s is not declared", expr.Ident.Name)
		}
	}
}

func (g *generator) traverseIdent(expr *ast.Ident, e *env) {
	parent, ok := e.get(expr.Name)
	if !ok {
		util.Error("%s is not declared", expr.Name)
	}
	if _, ok := parent.(*ast.VarDecl); !ok {
		util.Error("%s is not a variable", expr.Name)
	}
	g.relations[expr] = parent
}

func (g *generator) traverseStringLit(expr *ast.StringLit) {
	g.strs[expr] = &str{label: g.strLabel(), value: expr.Value}
}

/*
 Emit assembly code
*/

func (g *generator) emitProgram(node *ast.Program) {
	g.emit(".intel_syntax noprefix")
	g.emit(".section .rodata")

	for _, s := range g.strs {
		g.emitLabel(s.label)
		g.emit(".string %q", s.value)
	}

	g.emit(".text")

	for _, v := range g.gvars {
		g.emit(".comm %s,%d,%d", v.label, v.size, v.size)
	}

	for stmt, _ := range g.fns {
		g.emitFuncDecl(stmt)
	}

	g.emit(".globl main")
	g.emitLabel("main")
	g.emit("push rbp")
	g.emit("mov rbp, rsp")

	for _, stmt := range node.Stmts {
		g.emitStmt(stmt)
	}

	g.emit("leave")
	g.emit("ret")
}

func (g *generator) emitStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.VarDecl:
		g.emitVarDecl(v)
	case *ast.BlockStmt:
		g.emitBlockStmt(v)
	case *ast.IfStmt:
		g.emitIfStmt(v)
	case *ast.WhileStmt:
		g.emitWhileStmt(v)
	case *ast.ReturnStmt:
		g.emitReturnStmt(v)
	case *ast.ContinueStmt:
		g.emitContinueStmt(v)
	case *ast.BreakStmt:
		g.emitBreakStmt(v)
	case *ast.AssignStmt:
		g.emitAssignStmt(v)
	case *ast.ExprStmt:
		g.emitExprStmt(v)
	}
}

func (g *generator) emitFuncDecl(stmt *ast.FuncDecl) {
	fn := g.fns[stmt]
	br := g.branches[stmt]
	endLabel := br.labels[0]

	g.emitLabel(fn.label)
	g.emit("push rbp")
	g.emit("mov rbp, rsp")
	g.emit("sub rsp, %d", fn.align)

	for i, param := range stmt.Params {
		v := g.lvars[param]
		switch v.size {
		case 1:
			g.emit("mov byte ptr [rbp-%d], %s", v.offset, paramRegs[1][i])
		case 8:
			g.emit("mov qword ptr [rbp-%d], %s", v.offset, paramRegs[8][i])
		}
	}

	g.emitBlockStmt(stmt.Body)

	g.emitLabel(endLabel)
	g.emit("add rsp, %d", fn.align)
	g.emit("leave")
	g.emit("ret")
}

func (g *generator) emitVarDecl(stmt *ast.VarDecl) {
	g.emitExpr(stmt.Value)

	if v, ok := g.lvars[stmt]; ok {
		switch v.size {
		case 1:
			g.emit("mov byte ptr [rbp-%d], al", v.offset)
		case 8:
			g.emit("mov qword ptr [rbp-%d], rax", v.offset)
		}
	} else {
		v := g.gvars[stmt]
		switch v.size {
		case 1:
			g.emit("mov byte ptr %s[rip], al", v.label)
		case 8:
			g.emit("mov qword ptr %s[rip], rax", v.label)
		}
	}
}

func (g *generator) emitBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.Stmts {
		g.emitStmt(stmt_)
	}
}

func (g *generator) emitIfStmt(stmt *ast.IfStmt) {
	br := g.branches[stmt]

	g.emitExpr(stmt.Cond)
	g.emit("cmp rax, 0")

	if stmt.Altern == nil {
		endLabel := br.labels[0]
		g.emit("je %s", endLabel)
		g.emitBlockStmt(stmt.Conseq)
		g.emitLabel(endLabel)
	} else {
		altLabel := br.labels[0]
		endLabel := br.labels[1]
		g.emit("je %s", altLabel)
		g.emitBlockStmt(stmt.Conseq)
		g.emit("jmp %s", endLabel)
		g.emitLabel(altLabel)
		g.emitStmt(stmt.Altern)
		g.emitLabel(endLabel)
	}
}

func (g *generator) emitWhileStmt(stmt *ast.WhileStmt) {
	br := g.branches[stmt]
	beginLabel := br.labels[0]
	endLabel := br.labels[1]

	g.emitLabel(beginLabel)
	g.emitExpr(stmt.Cond)
	g.emit("cmp rax, 0")
	g.emit("je %s", endLabel)
	g.emitBlockStmt(stmt.Body)
	g.emit("jmp %s", beginLabel)
	g.emitLabel(endLabel)
}

func (g *generator) emitReturnStmt(stmt *ast.ReturnStmt) {
	parent := g.relations[stmt]
	br := g.branches[parent.(*ast.FuncDecl)]
	endLabel := br.labels[0]

	if stmt.Value != nil {
		g.emitExpr(stmt.Value)
	}
	g.emit("jmp %s", endLabel)
}

func (g *generator) emitContinueStmt(stmt *ast.ContinueStmt) {
	parent := g.relations[stmt]
	br := g.branches[parent.(*ast.WhileStmt)]
	beginLabel := br.labels[0]

	g.emit("jmp %s", beginLabel)
}

func (g *generator) emitBreakStmt(stmt *ast.BreakStmt) {
	parent := g.relations[stmt]
	br := g.branches[parent.(*ast.WhileStmt)]
	endLabel := br.labels[1]

	g.emit("jmp %s", endLabel)
}

func (g *generator) emitAssignStmt(stmt *ast.AssignStmt) {
	parent := g.relations[stmt]

	g.emitExpr(stmt.Value)

	if v, ok := g.lvars[parent.(*ast.VarDecl)]; ok {
		switch v.size {
		case 1:
			g.emit("mov byte ptr [rbp-%d], al", v.offset)
		case 8:
			g.emit("mov qword ptr [rbp-%d], rax", v.offset)
		}
	} else {
		v := g.gvars[parent.(*ast.VarDecl)]
		switch v.size {
		case 1:
			g.emit("mov byte ptr %s[rip], al", v.label)
		case 8:
			g.emit("mov qword ptr %s[rip], rax", v.label)
		}
	}
}

func (g *generator) emitExprStmt(stmt *ast.ExprStmt) {
	g.emitExpr(stmt.Expr)
}

func (g *generator) emitExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		g.emitPrefixExpr(v)
	case *ast.InfixExpr:
		g.emitInfixExpr(v)
	case *ast.FuncCall:
		g.emitFuncCall(v)
	case *ast.Ident:
		g.emitIdent(v)
	case *ast.IntLit:
		g.emitIntLit(v)
	case *ast.BoolLit:
		g.emitBoolLit(v)
	case *ast.StringLit:
		g.emitStringLit(v)
	}
}

func (g *generator) emitPrefixExpr(expr *ast.PrefixExpr) {
	g.emitExpr(expr.Right)

	switch expr.Operator {
	case "!":
		g.emit("xor rax, 1")
	case "-":
		g.emit("neg rax")
	}
}

func (g *generator) emitInfixExpr(expr *ast.InfixExpr) {
	g.emitExpr(expr.Right)
	g.emit("push rax")
	g.emitExpr(expr.Left)
	g.emit("pop rcx")

	switch expr.Operator {
	case "+":
		g.emit("add rax, rcx")
	case "-":
		g.emit("sub rax, rcx")
	case "*":
		g.emit("imul rax, rcx")
	case "/":
		g.emit("cqo")
		g.emit("idiv rcx")
	case "&&":
		g.emit("and rax, rcx")
	case "||":
		g.emit("or rax, rcx")
	case "==", "!=", "<", "<=", ">", ">=":
		g.emitCmp(expr.Operator)
	}
}

func (g *generator) emitCmp(operator string) {
	g.emit("cmp rax, rcx")
	g.emit("%s al", setcc[operator])
	g.emit("movzx rax, al")
}

func (g *generator) emitFuncCall(expr *ast.FuncCall) {
	for i, param := range expr.Params {
		g.emitExpr(param)
		g.emit("mov %s, rax", paramRegs[8][i])
	}

	if parent, ok := g.relations[expr]; ok {
		fn := g.fns[parent.(*ast.FuncDecl)]
		g.emit("call %s", fn.label)
	} else {
		g.emit("call %s", expr.Ident.Name) // library function
	}
}

func (g *generator) emitIdent(expr *ast.Ident) {
	parent := g.relations[expr]

	if v, ok := g.lvars[parent.(*ast.VarDecl)]; ok {
		switch v.size {
		case 1:
			g.emit("movzx rax, byte ptr [rbp-%d]", v.offset)
		case 8:
			g.emit("mov rax, qword ptr [rbp-%d]", v.offset)
		}
	} else {
		v := g.gvars[parent.(*ast.VarDecl)]
		switch v.size {
		case 1:
			g.emit("movzx rax, byte ptr %s[rip]", v.label)
		case 8:
			g.emit("mov rax, qword ptr %s[rip]", v.label)
		}
	}
}

func (g *generator) emitIntLit(expr *ast.IntLit) {
	g.emit("mov rax, %d", expr.Value)
}

func (g *generator) emitBoolLit(expr *ast.BoolLit) {
	if expr.Value {
		g.emit("mov rax, 1")
	} else {
		g.emit("mov rax, 0")
	}
}

func (g *generator) emitStringLit(expr *ast.StringLit) {
	s := g.strs[expr]
	g.emit("mov rax, offset flat:%s", s.label)
}

func (g *generator) emitLabel(label string) {
	fmt.Println(label + ":")
}

func (g *generator) emit(format string, a ...interface{}) {
	fmt.Printf("\t"+format+"\n", a...)
}
