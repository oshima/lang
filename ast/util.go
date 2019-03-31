package ast

// Returnable checks if the input statement can return a value in function
func Returnable(stmt Stmt) bool {
	switch v := stmt.(type) {
	case *BlockStmt:
		for _, stmt := range v.Stmts {
			if Returnable(stmt) {
				return true
			}
		}
		return false
	case *IfStmt:
		if v.Else == nil {
			return false
		}
		return Returnable(v.Body) && Returnable(v.Else)
	case *ReturnStmt:
		return true
	default:
		return false
	}
}
