package ast

import "github.com/oshima/lang/types"

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

// All expression nodes implement Expr
type Expr interface {
	Node
	exprNode()
}

// All declaration nodes implement Decl
type Decl interface {
	Node
	declNode()
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

type VarStmt struct {
	Vars []*VarDecl
}

type FuncStmt struct {
	Func *FuncDecl
}

type IfStmt struct {
	Cond Expr
	Body *BlockStmt
	Else Stmt // *BlockStmt or *IfStmt
}

type WhileStmt struct {
	Cond Expr
	Body *BlockStmt
}

type ForStmt struct {
	Elem  *VarDecl
	Index *VarDecl
	Iter  *VarDecl // implicit variable
	Body  *BlockStmt
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
	Op     string
	Target Expr
	Value  Expr
}

type ExprStmt struct {
	Expr Expr
}

func (stmt *BlockStmt) astNode()     {}
func (stmt *BlockStmt) stmtNode()    {}
func (stmt *VarStmt) astNode()       {}
func (stmt *VarStmt) stmtNode()      {}
func (stmt *FuncStmt) astNode()      {}
func (stmt *FuncStmt) stmtNode()     {}
func (stmt *IfStmt) astNode()        {}
func (stmt *IfStmt) stmtNode()       {}
func (stmt *WhileStmt) astNode()     {}
func (stmt *WhileStmt) stmtNode()    {}
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
	Name   string
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

type RangeLit struct {
	Lower Expr
	Upper Expr
}

type ArrayLit struct {
	Elems []Expr
}

type ArrayShortLit struct {
	Len      int
	ElemType types.Type
	Value    Expr // initial value for all elements
}

type FuncLit struct {
	Params     []*VarDecl
	ReturnType types.Type
	Body       *BlockStmt
}

func (expr *PrefixExpr) astNode()     {}
func (expr *PrefixExpr) exprNode()    {}
func (expr *InfixExpr) astNode()      {}
func (expr *InfixExpr) exprNode()     {}
func (expr *IndexExpr) astNode()      {}
func (expr *IndexExpr) exprNode()     {}
func (expr *CallExpr) astNode()       {}
func (expr *CallExpr) exprNode()      {}
func (expr *LibCallExpr) astNode()    {}
func (expr *LibCallExpr) exprNode()   {}
func (expr *Ident) astNode()          {}
func (expr *Ident) exprNode()         {}
func (expr *IntLit) astNode()         {}
func (expr *IntLit) exprNode()        {}
func (expr *BoolLit) astNode()        {}
func (expr *BoolLit) exprNode()       {}
func (expr *StringLit) astNode()      {}
func (expr *StringLit) exprNode()     {}
func (expr *RangeLit) astNode()       {}
func (expr *RangeLit) exprNode()      {}
func (expr *ArrayLit) astNode()       {}
func (expr *ArrayLit) exprNode()      {}
func (expr *ArrayShortLit) astNode()  {}
func (expr *ArrayShortLit) exprNode() {}
func (expr *FuncLit) astNode()        {}
func (expr *FuncLit) exprNode()       {}

/*
 Declaration nodes
*/

type VarDecl struct {
	Name    string
	VarType types.Type
	Value   Expr
}

type FuncDecl struct {
	Name       string
	Params     []*VarDecl
	ReturnType types.Type
	Body       *BlockStmt
}

func (decl *VarDecl) astNode()   {}
func (decl *VarDecl) declNode()  {}
func (decl *FuncDecl) astNode()  {}
func (decl *FuncDecl) declNode() {}
