package ast

import "github.com/oshjma/lang/types"

/*
 Interfaces
*/

// All AST nodes implement Node
type Node interface {
	astNode()
}

// All statement nodes implement Stmt
type Stmt interface {
	Node
	stmtNode()
}

// All declaration nodes implement Decl
type Decl interface {
	Node
	declNode()
}

// All expression nodes implement Expr
type Expr interface {
	Node
	exprNode()
}

/*
 Program (root node)
*/

type Program struct {
	Stmts []Stmt
}

func (prog *Program) astNode() {}

/*
 Statement nodes
*/

type BlockStmt struct {
	Stmts []Stmt
}

type LetStmt struct {
	Vars   []*VarDecl
	Values []Expr
}

type IfStmt struct {
	Cond Expr
	Body *BlockStmt
	Else Stmt // *BlockStmt or *IfStmt
}

type ForStmt struct {
	Cond Expr
	Body *BlockStmt
}

type ContinueStmt struct {
	_ byte
}

type BreakStmt struct {
	_ byte
}

type ReturnStmt struct {
	Value Expr
}

type AssignStmt struct {
	Targets []Expr // consist of *VarRef or *IndexExpr
	Values  []Expr
}

type ExprStmt struct {
	Expr Expr
}

func (stmt *BlockStmt) astNode()     {}
func (stmt *BlockStmt) stmtNode()    {}
func (stmt *LetStmt) astNode()       {}
func (stmt *LetStmt) stmtNode()      {}
func (stmt *IfStmt) astNode()        {}
func (stmt *IfStmt) stmtNode()       {}
func (stmt *ForStmt) astNode()       {}
func (stmt *ForStmt) stmtNode()      {}
func (stmt *ContinueStmt) astNode()  {}
func (stmt *ContinueStmt) stmtNode() {}
func (stmt *BreakStmt) astNode()     {}
func (stmt *BreakStmt) stmtNode()    {}
func (stmt *ReturnStmt) astNode()    {}
func (stmt *ReturnStmt) stmtNode()   {}
func (stmt *AssignStmt) astNode()    {}
func (stmt *AssignStmt) stmtNode()   {}
func (stmt *ExprStmt) astNode()      {}
func (stmt *ExprStmt) stmtNode()     {}

/*
 Declaration nodes
*/

type VarDecl struct {
	Ident   string
	VarType types.Type
}

func (decl *VarDecl) astNode()  {}
func (decl *VarDecl) declNode() {}

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

type IndexExpr struct {
	Left  Expr
	Index Expr
}

type CallExpr struct {
	Left   Expr
	Params []Expr
}

type LibCallExpr struct {
	Ident  string
	Params []Expr
}

type VarRef struct {
	Ident string
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

type FuncLit struct {
	Params     []*VarDecl
	ReturnType types.Type
	Body       *BlockStmt
}

func (expr *PrefixExpr) astNode()   {}
func (expr *PrefixExpr) exprNode()  {}
func (expr *InfixExpr) astNode()    {}
func (expr *InfixExpr) exprNode()   {}
func (expr *IndexExpr) astNode()    {}
func (expr *IndexExpr) exprNode()   {}
func (expr *CallExpr) astNode()     {}
func (expr *CallExpr) exprNode()    {}
func (expr *LibCallExpr) astNode()  {}
func (expr *LibCallExpr) exprNode() {}
func (expr *VarRef) astNode()       {}
func (expr *VarRef) exprNode()      {}
func (expr *IntLit) astNode()       {}
func (expr *IntLit) exprNode()      {}
func (expr *BoolLit) astNode()      {}
func (expr *BoolLit) exprNode()     {}
func (expr *StringLit) astNode()    {}
func (expr *StringLit) exprNode()   {}
func (expr *ArrayLit) astNode()     {}
func (expr *ArrayLit) exprNode()    {}
func (expr *FuncLit) astNode()      {}
func (expr *FuncLit) exprNode()     {}
