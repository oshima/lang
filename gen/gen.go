package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
)

func Generate(node *ast.Program) {
	emitProgram(node)
}

var uniqueLabel = func() func() string {
	i := 0
	return func() string {
		label := fmt.Sprintf(".L%d", i)
		i += 1
		return label
	}
}()

var setcc = map[string]string{
    "==": "sete",
	"!=": "setne",
	"<":  "setl",
	"<=": "setle",
	">":  "setg",
	">=": "setge",
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
	case *ast.BlockStmt:
		emitBlockStmt(v)
	case *ast.IfStmt:
		emitIfStmt(v)
	case *ast.ExprStmt:
		emitExprStmt(v)
	}
}

func emitBlockStmt(stmt *ast.BlockStmt) {
	for _, s := range stmt.Statements {
		emitStmt(s)
	}
}

func emitIfStmt(stmt *ast.IfStmt) {
	altLabel := uniqueLabel()
	endLabel := uniqueLabel()
	emitExpr(stmt.Cond)
	emit("cmp rax, 0")
	emit("je %s", altLabel)
	emitStmt(stmt.Conseq)
	emit("jmp %s", endLabel)
	p("%s:", altLabel)
	emitStmt(stmt.Altern)
	p("%s:", endLabel)
}

func emitExprStmt(stmt *ast.ExprStmt) {
	emitExpr(stmt.Expr)
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
	emit("%s al", setcc[operator])
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
