package gen

import "github.com/oshjma/lang/ast"

var libFns = map[string]bool{
	"puts":   true,
	"printf": true,
}

var sizeof = map[string]int{
	"int":    8,
	"bool":   1,
	"string": 8,
}

var setcc = map[string]string{
	"==": "sete",
	"!=": "setne",
	"<":  "setl",
	"<=": "setle",
	">":  "setg",
	">=": "setge",
}

var paramRegs = map[int][6]string{
	1: [6]string{"dil", "sil", "dl", "cl", "r8b", "r9b"},
	8: [6]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"},
}

func returnableStmt(stmt ast.Stmt) bool {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		return returnableBlockStmt(v)
	case *ast.IfStmt:
		return returnableIfStmt(v)
	case *ast.ReturnStmt:
		return true
	default:
		return false
	}
}

func returnableBlockStmt(stmt *ast.BlockStmt) bool {
	for _, stmt_ := range stmt.Stmts {
		if returnableStmt(stmt_) {
			return true
		}
	}
	return false
}

func returnableIfStmt(stmt *ast.IfStmt) bool {
	if stmt.Altern == nil {
		return false
	}
	return returnableBlockStmt(stmt.Conseq) && returnableStmt(stmt.Altern)
}

// https://en.wikipedia.org/wiki/Data_structure_alignment
func align(n int, boundary int) int {
	return (n + boundary - 1) & -boundary
}
