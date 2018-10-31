package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/util"
)

func Generate(prog *ast.Program) {
	g := &generator{}
	g.findGvarsInProgram(prog)
	g.emitProgram(prog)
}

var sizes = map[string]int{
	"int":  8,
	"bool": 1,
}

var setcc = map[string]string{
    "==": "sete",
	"!=": "setne",
	"<":  "setl",
	"<=": "setle",
	">":  "setg",
	">=": "setge",
}

type generator struct {
	nLabel int
	gvars map[*ast.LetStmt]*gvar
}

func (g *generator) nextLabel() string {
	label := fmt.Sprintf(".L%d", g.nLabel)
	g.nLabel += 1
	return label
}

func (g *generator) findGvarsInProgram(node *ast.Program) {
	g.gvars = make(map[*ast.LetStmt]*gvar)
	for _, stmt := range node.Statements {
		switch v := stmt.(type) {
		case *ast.LetStmt:
			label := g.nextLabel() + "_" + v.Ident.Name
			g.gvars[v] = &gvar{label: label, size: sizes[v.Type]}
		case *ast.BlockStmt:
			g.findGvarsInBlockStmt(v)
		case *ast.IfStmt:
			g.findGvarsInIfStmt(v)
		}
	}
}

func (g *generator) findGvarsInBlockStmt(stmt *ast.BlockStmt) {
	for _, _stmt := range stmt.Statements {
		switch v := _stmt.(type) {
		case *ast.LetStmt:
			label := g.nextLabel() + "_" + v.Ident.Name
			g.gvars[v] = &gvar{label: label, size: sizes[v.Type]}
		case *ast.BlockStmt:
			g.findGvarsInBlockStmt(v)
		case *ast.IfStmt:
			g.findGvarsInIfStmt(v)
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

func (g *generator) emitProgram(node *ast.Program) {
	g.emit(".intel_syntax noprefix")

	if len(g.gvars) > 0 {
		g.emit(".bss")
	}
	for _, gvar := range g.gvars {
		if gvar.size > 1 {
			g.emit(".align %d", gvar.size)
		}
		g.emitLabel(gvar.label)
		g.emit(".zero %d", gvar.size)
	}

	g.emit(".text")
	g.emit(".globl main")
	g.emitLabel("main")
	g.emit("push rbp")
	g.emit("mov rbp, rsp")

	e := &env{store: make(map[string]*gvar)}
	for _, stmt := range node.Statements {
		g.emitStmt(e, stmt)
	}

	g.emit("leave")
	g.emit("ret")
}

func (g *generator) emitStmt(e *env, stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		g.emitBlockStmt(e, v)
	case *ast.LetStmt:
		g.emitLetStmt(e, v)
	case *ast.IfStmt:
		g.emitIfStmt(e, v)
	case *ast.ExprStmt:
		g.emitExprStmt(e, v)
	}
}

func (g *generator) emitBlockStmt(e *env, stmt *ast.BlockStmt) {
	newEnv := &env{store: make(map[string]*gvar), outer: e}
	for _, s := range stmt.Statements {
		g.emitStmt(newEnv, s)
	}
}

func (g *generator) emitLetStmt(e *env, stmt *ast.LetStmt) {
	g.emitExpr(e, stmt.Expr)
	gvar := g.gvars[stmt]
	if err := e.set(stmt.Ident.Name, gvar); err != nil {
		util.Error("%s has already been declared", stmt.Ident.Name)
	}
	switch gvar.size {
	case 8:
		g.emit("mov QWORD PTR %s[rip], rax", gvar.label)
	case 1:
		g.emit("mov BYTE PTR %s[rip], al", gvar.label)
	}
}

func (g *generator) emitIfStmt(e *env, stmt *ast.IfStmt) {
	g.emitExpr(e, stmt.Cond)
	g.emit("cmp rax, 0")
	if stmt.Altern == nil {
		endLabel := g.nextLabel()
		g.emit("je %s", endLabel)
		g.emitBlockStmt(e, stmt.Conseq)
		g.emitLabel(endLabel)
	} else {
		altLabel := g.nextLabel()
		endLabel := g.nextLabel()
		g.emit("je %s", altLabel)
		g.emitBlockStmt(e, stmt.Conseq)
		g.emit("jmp %s", endLabel)
		g.emitLabel(altLabel)
		g.emitStmt(e, stmt.Altern)
		g.emitLabel(endLabel)
	}
}

func (g *generator) emitExprStmt(e *env, stmt *ast.ExprStmt) {
	g.emitExpr(e, stmt.Expr)
}

func (g *generator) emitExpr(e *env, expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		g.emitPrefixExpr(e, v)
	case *ast.InfixExpr:
		g.emitInfixExpr(e, v)
	case *ast.Ident:
		g.emitIdent(e, v)
	case *ast.IntLit:
		g.emitIntLit(v)
	case *ast.BoolLit:
		g.emitBoolLit(v)
	}
}

func (g *generator) emitPrefixExpr(e *env, expr *ast.PrefixExpr) {
	g.emitExpr(e, expr.Right)
	switch expr.Operator {
	case "!":
		g.emit("xor rax, 1")
	case "-":
		g.emit("neg rax")
	}
}

func (g *generator) emitInfixExpr(e *env, expr *ast.InfixExpr) {
	g.emitExpr(e, expr.Right)
	g.emit("push rax")
	g.emitExpr(e, expr.Left)
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

func (g *generator) emitIdent(e *env, expr *ast.Ident) {
	gvar, ok := e.get(expr.Name)
	if !ok {
		util.Error("%s is not declared", expr.Name)
	}
	switch gvar.size {
	case 8:
		g.emit("mov rax, QWORD PTR %s[rip]", gvar.label)
	case 1:
		g.emit("movzx rax, BYTE PTR %s[rip]", gvar.label)
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
	fmt.Printf("\t" + format + "\n", a...)
}
