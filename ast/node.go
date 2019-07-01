package ast

import (
	"github.com/oshima/lang/token"
	"github.com/oshima/lang/types"
)

// ----------------------------------------------------------------
// Interfaces

// Node is the interface for all AST nodes.
type Node interface {
	Pos() *token.Pos
	SetPos(*token.Pos)
}

type node struct {
	pos *token.Pos
}

func (n *node) Pos() *token.Pos       { return n.pos }
func (n *node) SetPos(pos *token.Pos) { n.pos = pos }

// Stmt is the interface for all statement nodes.
type Stmt interface {
	Node
	aStmt()
}

type stmt struct {
	node
}

func (s *stmt) aStmt() {}

// Expr is the interface for all expression nodes.
type Expr interface {
	Node
	Type() types.Type
	SetType(types.Type)
}

type expr struct {
	node
	typ types.Type
}

func (e *expr) Type() types.Type       { return e.typ }
func (e *expr) SetType(typ types.Type) { e.typ = typ }

// Decl is the interface for all declaration nodes.
type Decl interface {
	Node
	aDecl()
}

type decl struct {
	node
}

func (d *decl) aDecl() {}

// ----------------------------------------------------------------
// Program

// Program is the root node of AST.
type Program struct {
	Stmts []Stmt
	node
}

// ----------------------------------------------------------------
// Statement nodes

// BlockStmt represents a block of statements.
type BlockStmt struct {
	Stmts []Stmt
	stmt
}

// VarStmt represents a statement containing a couple of variable declarations.
type VarStmt struct {
	Vars []*VarDecl
	stmt
}

// FuncStmt represents a statement containing a function declaration.
type FuncStmt struct {
	Func *FuncDecl
	stmt
}

// IfStmt represents an if statement.
type IfStmt struct {
	Cond Expr
	Body *BlockStmt
	Else Stmt // *BlockStmt or *IfStmt
	stmt
}

// WhileStmt represents a while statement.
type WhileStmt struct {
	Cond Expr
	Body *BlockStmt
	stmt
}

// ForStmt represents a for statement.
type ForStmt struct {
	Elem  *VarDecl
	Index *VarDecl
	Iter  *VarDecl // implicit variable
	Body  *BlockStmt
	stmt
}

// ContinueStmt represents a continue statement.
type ContinueStmt struct {
	Ref Node // WhileStmt or ForStmt
	stmt
}

// BreakStmt represents a break statement.
type BreakStmt struct {
	Ref Node // WhileStmt or ForStmt
	stmt
}

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	Value Expr
	Ref   Node // FuncLit or FuncDecl
	stmt
}

// AssignStmt represents an assignment.
type AssignStmt struct {
	Op     string
	Target Expr
	Value  Expr
	stmt
}

// ExprStmt represents a statement of stand-alone expression.
type ExprStmt struct {
	Expr Expr
	stmt
}

// ----------------------------------------------------------------
// Expression nodes

// PrefixExpr represents an expression of prefix operator.
type PrefixExpr struct {
	Op    string
	Right Expr
	expr
}

// InfixExpr represents an expression of infix operator.
type InfixExpr struct {
	Op    string
	Left  Expr
	Right Expr
	expr
}

// IndexExpr represents an expression to access an array element.
type IndexExpr struct {
	Left  Expr
	Index Expr
	expr
}

// CallExpr represents an expression to call a function.
type CallExpr struct {
	Left   Expr
	Params []Expr
	expr
}

// LibCallExpr represents an expression to call a library function.
type LibCallExpr struct {
	Name   string
	Params []Expr
	expr
}

// Ident represents an identifier.
type Ident struct {
	Name string
	Ref  Node // VarDecl or FuncDecl
	expr
}

// IntLit represents a literal of integer type.
type IntLit struct {
	Value int
	expr
}

// BoolLit represents a literal of boolean type.
type BoolLit struct {
	Value bool
	expr
}

// StringLit represents a literal of string type.
type StringLit struct {
	Value string
	expr
}

// RangeLit represents a literal of range type.
type RangeLit struct {
	Lower Expr
	Upper Expr
	expr
}

// ArrayLit represents a literal of array type.
type ArrayLit struct {
	Elems []Expr
	expr
}

// ArrayShortLit represents a short form literal of array type.
type ArrayShortLit struct {
	Len      int
	ElemType types.Type
	Value    Expr // initial value for all elements
	expr
}

// FuncLit represents a literal of function type.
type FuncLit struct {
	Params     []*VarDecl
	ReturnType types.Type
	Body       *BlockStmt
	expr
}

// ----------------------------------------------------------------
// Declaration nodes

// VarDecl represents a variable declaration.
type VarDecl struct {
	Name    string
	VarType types.Type
	Value   Expr
	decl
}

// FuncDecl represents a function declaration.
type FuncDecl struct {
	Name       string
	Params     []*VarDecl
	ReturnType types.Type
	Body       *BlockStmt
	decl
}
