package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
)

func Generate(node *ast.Program) {
	emitProgram(node)
}

func emitProgram(node *ast.Program) {
	emit(".text")
	emit(".globl main")
	emit(".type main, @function")
	p("main:")
	emit("pushq %%rbp")
	emit("movq %%rsp, %%rbp")
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
	case *ast.InfixExpr:
		emitInfixExpr(v)
	case *ast.IntLit:
		emitIntLit(v)
	}
}

func emitInfixExpr(expr *ast.InfixExpr) {
	emitExpr(expr.Right)
	emit("pushq %%rax")
	emitExpr(expr.Left)
	emit("popq %%rcx")
	switch expr.Operator {
	case "+":
		emit("addq %%rcx, %%rax")
	case "-":
		emit("subq %%rcx, %%rax")
	case "*":
		emit("imulq %%rcx, %%rax")
	case "/":
		emit("cqo")
		emit("idivq %%rcx")
	}
}

func emitIntLit(expr *ast.IntLit) {
	emit("movq $%d, %%rax", expr.Value)
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
