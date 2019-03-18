package sema

import "github.com/oshima/lang/ast"

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
	for _, stmt := range stmt.Stmts {
		if returnableStmt(stmt) {
			return true
		}
	}
	return false
}

func returnableIfStmt(stmt *ast.IfStmt) bool {
	if stmt.Else == nil {
		return false
	}
	return returnableBlockStmt(stmt.Body) && returnableStmt(stmt.Else)
}
