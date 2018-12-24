package sema

import "github.com/oshjma/lang/ast"

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
