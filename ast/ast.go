package ast

/*
 Interfaces
*/

// for all AST nodes
type Node interface {
	Node()
}

// for all statement nodes
type Stmt interface {
	Node
	Stmt()
}

// for all expression nodes
type Expr interface {
	Node
	Expr()
}

/*
 Root node
*/

type Program struct {
	Statements []Stmt
}

func (node *Program) Node() {}

/*
 Statement nodes
*/

type ExprStmt struct {
	Expr Expr
}

type BlockStmt struct {
	Statements []Stmt
}

type IfStmt struct {
	Cond   Expr
	Conseq *BlockStmt
	Altern Stmt // *BlockStmt or *IfStmt
}

type LetStmt struct {
	Ident *Ident
	Type  string
	Expr  Expr
}

func (stmt *ExprStmt) Node()  {}
func (stmt *ExprStmt) Stmt()  {}
func (stmt *BlockStmt) Node() {}
func (stmt *BlockStmt) Stmt() {}
func (stmt *IfStmt) Node()    {}
func (stmt *IfStmt) Stmt()    {}
func (stmt *LetStmt) Node()   {}
func (stmt *LetStmt) Stmt()   {}

/*
 Expression nodes
*/

type PrefixExpr struct {
	Operator string
	Right    Expr
}

type InfixExpr struct {
	Operator string
	Left     Expr
	Right    Expr
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

func (expr *PrefixExpr) Node() {}
func (expr *PrefixExpr) Expr() {}
func (expr *InfixExpr) Node()  {}
func (expr *InfixExpr) Expr()  {}
func (expr *Ident) Node()      {}
func (expr *Ident) Expr()      {}
func (expr *IntLit) Node()     {}
func (expr *IntLit) Expr()     {}
func (expr *BoolLit) Node()    {}
func (expr *BoolLit) Expr()    {}
