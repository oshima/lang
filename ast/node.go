package ast

import "github.com/oshjma/lang/types"

/*
 Interfaces
*/

// for all AST nodes
type Node interface {
	astNode()
}

// for statement nodes
type Stmt interface {
	Node
	stmtNode()
}

// for expression nodes
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

func (prog *Program) astNode() {}

/*
 Statement nodes
*/

type FuncDecl struct {
	Ident      *Ident
	Params     []*VarDecl
	ReturnType types.Type
	Body       *BlockStmt
}

type VarDecl struct {
	Ident   *Ident
	VarType types.Type
	Value   Expr
}

type BlockStmt struct {
	Stmts []Stmt
}

type IfStmt struct {
	Cond   Expr
	Conseq *BlockStmt
	Altern Stmt // *BlockStmt or *IfStmt
}

type ForStmt struct {
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
	Target Expr // *Ident or *IndexExpr
	Value  Expr
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
func (stmt *ForStmt) astNode()       {}
func (stmt *ForStmt) stmtNode()      {}
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
	Left  Expr
	Op    string
	Right Expr
}

type IndexExpr struct {
	Left  Expr
	Index Expr
}

type FuncCall struct {
	Ident  *Ident
	Params []Expr
}

type Ident struct {
	Name string
}

type IntLit struct {
	Value int
}

type BoolLit struct {
	Value bool
}

type StringLit struct {
	Value string
}

type ArrayLit struct {
	Len      int
	ElemType types.Type
	Elems    []Expr
}

func (expr *PrefixExpr) astNode()  {}
func (expr *PrefixExpr) exprNode() {}
func (expr *InfixExpr) astNode()   {}
func (expr *InfixExpr) exprNode()  {}
func (expr *IndexExpr) astNode()   {}
func (expr *IndexExpr) exprNode()  {}
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
func (expr *ArrayLit) astNode()    {}
func (expr *ArrayLit) exprNode()   {}
