package ast

import "github.com/oshima/lang/types"

// ----------------------------------------------------------------
// Interfaces

// Node is the interface for all AST nodes
type Node interface {
	astNode()
}

// Stmt is the interface for all statement nodes
type Stmt interface {
	Node
	stmtNode()
}

// Expr is the interface for all expression nodes
type Expr interface {
	Node
	exprNode()
}

// Decl is the interface for all declaration nodes
type Decl interface {
	Node
	declNode()
}

// ----------------------------------------------------------------
// Program

// Program is the root node of AST
type Program struct {
	Stmts []Stmt
}

func (prog *Program) astNode() {}

// ----------------------------------------------------------------
// Statement nodes

// BlockStmt represents a block of statements
type BlockStmt struct {
	Stmts []Stmt
}

// VarStmt represents a statement containing a couple of variable declarations
type VarStmt struct {
	Vars []*VarDecl
}

// FuncStmt represents a statement containing a function declaration
type FuncStmt struct {
	Func *FuncDecl
}

// IfStmt represents an if statement
type IfStmt struct {
	Cond Expr
	Body *BlockStmt
	Else Stmt // *BlockStmt or *IfStmt
}

// WhileStmt represents a while statement
type WhileStmt struct {
	Cond Expr
	Body *BlockStmt
}

// ForStmt represents a for statement
type ForStmt struct {
	Elem  *VarDecl
	Index *VarDecl
	Iter  *VarDecl // implicit variable
	Body  *BlockStmt
}

// ContinueStmt represents a continue statement
type ContinueStmt struct {
	_ byte
}

// BreakStmt represents a break statement
type BreakStmt struct {
	_ byte
}

// ReturnStmt represents a return statement
type ReturnStmt struct {
	Value Expr
}

// AssignStmt represents an assignment
type AssignStmt struct {
	Op     string
	Target Expr
	Value  Expr
}

// ExprStmt represents a statement of stand-alone expression
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

// ----------------------------------------------------------------
// Expression nodes

// PrefixExpr represents an expression of prefix operator
type PrefixExpr struct {
	Op    string
	Right Expr
}

// InfixExpr represents an expression of infix operator
type InfixExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

// IndexExpr represents an expression to access an array element
type IndexExpr struct {
	Left  Expr
	Index Expr
}

// CallExpr represents an expression to call a function
type CallExpr struct {
	Left   Expr
	Params []Expr
}

// LibCallExpr represents an expression to call a library function
type LibCallExpr struct {
	Name   string
	Params []Expr
}

// Ident represents an identifier
type Ident struct {
	Name string
}

// IntLit represents a literal of integer type
type IntLit struct {
	Value int
}

// BoolLit represents a literal of boolean type
type BoolLit struct {
	Value bool
}

// StringLit represents a literal of string type
type StringLit struct {
	Value string
}

// RangeLit represents a literal of range type
type RangeLit struct {
	Lower Expr
	Upper Expr
}

// ArrayLit represents a literal of array type
type ArrayLit struct {
	Elems []Expr
}

// ArrayShortLit represents a short form literal of array type
type ArrayShortLit struct {
	Len      int
	ElemType types.Type
	Value    Expr // initial value for all elements
}

// FuncLit represents a literal of function type
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

// ----------------------------------------------------------------
// Declaration nodes

// VarDecl represents a variable declaration
type VarDecl struct {
	Name    string
	VarType types.Type
	Value   Expr
}

// FuncDecl represents a function declaration
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
