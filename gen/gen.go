package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/util"
)

func Generate(node *ast.Program) {
	g := &generator{gvars: make(map[*ast.Ident]*gvar)}
	g.findGvarsInProgram(node)
	g.emitProgram(node)
}

type generator struct {
	nLabel int
	gvars  map[*ast.Ident]*gvar
}

func (g *generator) nextLabel() string {
	label := fmt.Sprintf(".L%d", g.nLabel)
	g.nLabel += 1
	return label
}

func (g *generator) findGvarsInProgram(node *ast.Program) {
	g.findGvarsInBlockStmt(node.TopLevel)
}

func (g *generator) findGvarsInBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.List {
		switch v := stmt_.(type) {
		case *ast.BlockStmt:
			g.findGvarsInBlockStmt(v)
		case *ast.IfStmt:
			g.findGvarsInIfStmt(v)
		case *ast.WhileStmt:
			g.findGvarsInWhileStmt(v)
		case *ast.LetStmt:
			label := g.nextLabel() + "_" + v.Ident.Name
			size := sizes[v.Type]
			g.gvars[v.Ident] = &gvar{label: label, size: size}
		}
	}
}

func (g *generator) findGvarsInIfStmt(stmt *ast.IfStmt) {
	g.findGvarsInBlockStmt(stmt.Conseq)
	switch v := stmt.Altern.(type) {
	case *ast.BlockStmt:
		g.findGvarsInBlockStmt(v)
	case *ast.IfStmt:
		g.findGvarsInIfStmt(v)
	}
}

func (g *generator) findGvarsInWhileStmt(stmt *ast.WhileStmt) {
	g.findGvarsInBlockStmt(stmt.Body)
}

func (g *generator) emitProgram(node *ast.Program) {
	g.emit(".intel_syntax noprefix")

	if len(g.gvars) > 0 {
		g.emit(".bss")
	}
	for _, v := range g.gvars {
		if v.size > 1 {
			g.emit(".align %d", v.size)
		}
		g.emitLabel(v.label)
		g.emit(".zero %d", v.size)
	}

	g.emit(".text")
	g.emit(".globl main")
	g.emitLabel("main")
	g.emit("push rbp")
	g.emit("mov rbp, rsp")
	g.emitBlockStmt(node.TopLevel, newEnv(nil))
	g.emit("leave")
	g.emit("ret")
}

func (g *generator) emitStmt(stmt ast.Stmt, e *env) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		g.emitBlockStmt(v, newEnv(e))
	case *ast.IfStmt:
		g.emitIfStmt(v, e)
	case *ast.WhileStmt:
		g.emitWhileStmt(v, e)
	case *ast.ContinueStmt:
		g.emitContinueStmt(v, e)
	case *ast.BreakStmt:
		g.emitBreakStmt(v, e)
	case *ast.LetStmt:
		g.emitLetStmt(v, e)
	case *ast.AssignStmt:
		g.emitAssignStmt(v, e)
	case *ast.ExprStmt:
		g.emitExprStmt(v, e)
	}
}

func (g *generator) emitBlockStmt(stmt *ast.BlockStmt, e *env) {
	for _, stmt_ := range stmt.List {
		g.emitStmt(stmt_, e)
	}
}

func (g *generator) emitIfStmt(stmt *ast.IfStmt, e *env) {
	g.emitExpr(stmt.Cond, e)
	g.emit("cmp rax, 0")
	if stmt.Altern == nil {
		endLabel := g.nextLabel()
		g.emit("je %s", endLabel)
		g.emitBlockStmt(stmt.Conseq, newEnv(e))
		g.emitLabel(endLabel)
	} else {
		altLabel := g.nextLabel()
		endLabel := g.nextLabel()
		g.emit("je %s", altLabel)
		g.emitBlockStmt(stmt.Conseq, newEnv(e))
		g.emit("jmp %s", endLabel)
		g.emitLabel(altLabel)
		g.emitStmt(stmt.Altern, e)
		g.emitLabel(endLabel)
	}
}

func (g *generator) emitWhileStmt(stmt *ast.WhileStmt, e *env) {
	beginLabel := g.nextLabel()
	endLabel := g.nextLabel()

	g.emitLabel(beginLabel)
	g.emitExpr(stmt.Cond, e)
	g.emit("cmp rax, 0")
	g.emit("je %s", endLabel)

	newE := newEnv(e)
	newE.setDest("continue", &dest{label: beginLabel})
	newE.setDest("break", &dest{label: endLabel})
	g.emitBlockStmt(stmt.Body, newE)

	g.emit("jmp %s", beginLabel)
	g.emitLabel(endLabel)
}

func (g *generator) emitContinueStmt(stmt *ast.ContinueStmt, e *env) {
	d, ok := e.getDest("continue")
	if !ok {
		util.Error("Illegal use of continue")
	}
	g.emit("jmp %s", d.label)
}

func (g *generator) emitBreakStmt(stmt *ast.BreakStmt, e *env) {
	d, ok := e.getDest("break")
	if !ok {
		util.Error("Illegal use of break")
	}
	g.emit("jmp %s", d.label)
}

func (g *generator) emitLetStmt(stmt *ast.LetStmt, e *env) {
	g.emitExpr(stmt.Expr, e)
	v := g.gvars[stmt.Ident]
	if err := e.setGvar(stmt.Ident.Name, v); err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
	}
	switch v.size {
	case 8:
		g.emit("mov QWORD PTR %s[rip], rax", v.label)
	case 1:
		g.emit("mov BYTE PTR %s[rip], al", v.label)
	}
}

func (g *generator) emitAssignStmt(stmt *ast.AssignStmt, e *env) {
	g.emitExpr(stmt.Expr, e)
	v, ok := e.getGvar(stmt.Ident.Name)
	if !ok {
		util.Error("%s is not declared", stmt.Ident.Name)
	}
	switch v.size {
	case 8:
		g.emit("mov QWORD PTR %s[rip], rax", v.label)
	case 1:
		g.emit("mov BYTE PTR %s[rip], al", v.label)
	}
}

func (g *generator) emitExprStmt(stmt *ast.ExprStmt, e *env) {
	g.emitExpr(stmt.Expr, e)
}

func (g *generator) emitExpr(expr ast.Expr, e *env) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		g.emitPrefixExpr(v, e)
	case *ast.InfixExpr:
		g.emitInfixExpr(v, e)
	case *ast.Ident:
		g.emitIdent(v, e)
	case *ast.IntLit:
		g.emitIntLit(v)
	case *ast.BoolLit:
		g.emitBoolLit(v)
	}
}

func (g *generator) emitPrefixExpr(expr *ast.PrefixExpr, e *env) {
	g.emitExpr(expr.Right, e)
	switch expr.Operator {
	case "!":
		g.emit("xor rax, 1")
	case "-":
		g.emit("neg rax")
	}
}

func (g *generator) emitInfixExpr(expr *ast.InfixExpr, e *env) {
	g.emitExpr(expr.Right, e)
	g.emit("push rax")
	g.emitExpr(expr.Left, e)
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

func (g *generator) emitIdent(expr *ast.Ident, e *env) {
	v, ok := e.getGvar(expr.Name)
	if !ok {
		util.Error("%s is not declared", expr.Name)
	}
	switch v.size {
	case 8:
		g.emit("mov rax, QWORD PTR %s[rip]", v.label)
	case 1:
		g.emit("movzx rax, BYTE PTR %s[rip]", v.label)
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

func (g *generator) emitLabel(label string) {
	fmt.Println(label + ":")
}

func (g *generator) emit(format string, a ...interface{}) {
	fmt.Printf("\t"+format+"\n", a...)
}
