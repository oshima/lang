package ast

/*
 Interfaces
*/

// for all AST nodes
type Node interface {
	astNode()
}

// for all statement nodes
type Stmt interface {
	Node
	stmtNode()
}

// for all expression nodes
type Expr interface {
	Node
	exprNode()
}

/*
 Root node
*/

type Program struct {
	Stmts []Stmt
}

func (node *Program) astNode() {}

/*
 Statement nodes
*/

type VarDecl struct {
	Ident *Ident
	Type  string
	Value Expr
}

type FuncDecl struct {
	Ident      *Ident
	Params     []*VarDecl
	ReturnType string
	Body       *BlockStmt
}

type BlockStmt struct {
	Stmts []Stmt
}

type IfStmt struct {
	Cond   Expr
	Conseq *BlockStmt
	Altern Stmt // *BlockStmt or *IfStmt
}

type WhileStmt struct {
	Cond Expr
	Body *BlockStmt
}

type ReturnStmt struct {
	Value Expr
}

type ContinueStmt struct {
	dummy byte
}

type BreakStmt struct {
	dummy byte
}

type AssignStmt struct {
	Ident *Ident
	Value Expr
}

type ExprStmt struct {
	Expr Expr
}

func (stmt *VarDecl) astNode()       {}
func (stmt *VarDecl) stmtNode()      {}
func (stmt *FuncDecl) astNode()      {}
func (stmt *FuncDecl) stmtNode()     {}
func (stmt *BlockStmt) astNode()     {}
func (stmt *BlockStmt) stmtNode()    {}
func (stmt *IfStmt) astNode()        {}
func (stmt *IfStmt) stmtNode()       {}
func (stmt *WhileStmt) astNode()     {}
func (stmt *WhileStmt) stmtNode()    {}
func (stmt *ReturnStmt) astNode()    {}
func (stmt *ReturnStmt) stmtNode()   {}
func (stmt *ContinueStmt) astNode()  {}
func (stmt *ContinueStmt) stmtNode() {}
func (stmt *BreakStmt) astNode()     {}
func (stmt *BreakStmt) stmtNode()    {}
func (stmt *AssignStmt) astNode()    {}
func (stmt *AssignStmt) stmtNode()   {}
func (stmt *ExprStmt) astNode()      {}
func (stmt *ExprStmt) stmtNode()     {}

/*
 Expression nodes
*/

type PrefixExpr struct {
	Op    string
	Right Expr
}

type InfixExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

type FuncCall struct {
	Ident  *Ident
	Params []Expr
}

type Ident struct {
	Name string
}

type IntLit struct {
	Value int64
}

type BoolLit struct {
	Value bool
}

type StringLit struct {
	Value string
}

func (expr *PrefixExpr) astNode()  {}
func (expr *PrefixExpr) exprNode() {}
func (expr *InfixExpr) astNode()   {}
func (expr *InfixExpr) exprNode()  {}
func (expr *FuncCall) astNode()    {}
func (expr *FuncCall) exprNode()   {}
func (expr *Ident) astNode()       {}
func (expr *Ident) exprNode()      {}
func (expr *IntLit) astNode()      {}
func (expr *IntLit) exprNode()     {}
func (expr *BoolLit) astNode()     {}
func (expr *BoolLit) exprNode()    {}
func (expr *StringLit) astNode()   {}
func (expr *StringLit) exprNode()  {}
