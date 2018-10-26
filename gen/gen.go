package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
)

func Generate(node *ast.Program) {
	emitProgram(node)
}

func emitProgram(node *ast.Program) {
	emit(".intel_syntax noprefix")
	emit(".text")
	emit(".globl main")
	emit(".type main, @function")
	p("main:")
	emit("push rbp")
	emit("mov rbp, rsp")
	for _, stmt := range node.Statements {
		emitStmt(stmt)
	}
	emit("leave")
	emit("ret")
}

func emitStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.ExprStmt:
		emitExpr(v.Expr)
	}
}

func emitExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		emitPrefixExpr(v)
	case *ast.InfixExpr:
		emitInfixExpr(v)
	case *ast.IntLit:
		emitIntLit(v)
	case *ast.BoolLit:
		emitBoolLit(v)
	}
}

func emitPrefixExpr(expr *ast.PrefixExpr) {
	emitExpr(expr.Right)
	switch expr.Operator {
	case "!":
		emit("xor rax, 1")
	case "-":
		emit("neg rax")
	}
}

func emitInfixExpr(expr *ast.InfixExpr) {
	emitExpr(expr.Right)
	emit("push rax")
	emitExpr(expr.Left)
	emit("pop rcx")
	switch expr.Operator {
	case "+":
		emit("add rax, rcx")
	case "-":
		emit("sub rax, rcx")
	case "*":
		emit("imul rax, rcx")
	case "/":
		emit("cqo")
		emit("idiv rcx")
	case "&&":
		emit("and rax, rcx")
	case "||":
		emit("or rax, rcx")
	case "==", "!=", "<", "<=", ">", ">=":
		emitCmp(expr.Operator)
	}
}

func emitCmp(operator string) {
	emit("cmp rax, rcx")
	switch operator {
	case "==":
		emit("sete al")
	case "!=":
		emit("setne al")
	case "<":
		emit("setl al")
	case "<=":
		emit("setle al")
	case ">":
		emit("setg al")
	case ">=":
		emit("setge al")
	}
	emit("movzx rax, al")
}

func emitIntLit(expr *ast.IntLit) {
	emit("mov rax, %d", expr.Value)
}

func emitBoolLit(expr *ast.BoolLit) {
	if expr.Value {
		emit("mov rax, 1")
	} else {
		emit("mov rax, 0")
	}
}

func p(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

func emit(format string, a ...interface{}) {
	fmt.Print("\t")
	fmt.Printf(format, a...)
	fmt.Print("\n")
}
