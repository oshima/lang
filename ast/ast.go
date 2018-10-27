package ast

// Interface for all AST nodes
type Node interface {
	AstNode()
}

// Interface for statement nodes
type Stmt interface {
	Node
	StmtNode()
}

// Interface for expression nodes
type Expr interface {
	Node
	ExprNode()
}

// Root node
type Program struct {
	Statements []Stmt
}
func (node *Program) AstNode() {}

// Statement nodes
type ExprStmt struct {
	Expr Expr
}

type BlockStmt struct {
	Statements []Stmt
}

type IfStmt struct {
	Cond Expr
	Conseq Stmt
	Altern Stmt
}

func (stmt *ExprStmt) AstNode() {}
func (stmt *ExprStmt) StmtNode() {}
func (stmt *BlockStmt) AstNode() {}
func (stmt *BlockStmt) StmtNode() {}
func (stmt *IfStmt) AstNode() {}
func (stmt *IfStmt) StmtNode() {}

// Expression nodes
type PrefixExpr struct {
	Operator string
	Right Expr
}

type InfixExpr struct {
	Operator string
	Left Expr
	Right Expr
}

type IntLit struct {
	Value int64
}

type BoolLit struct {
	Value bool
}

func (expr *PrefixExpr) AstNode() {}
func (expr *PrefixExpr) ExprNode() {}
func (expr *InfixExpr) AstNode() {}
func (expr *InfixExpr) ExprNode() {}
func (expr *IntLit) AstNode() {}
func (expr *IntLit) ExprNode() {}
func (expr *BoolLit) AstNode() {}
func (expr *BoolLit) ExprNode() {}
