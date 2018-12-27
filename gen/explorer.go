package gen

import (
	"fmt"
	"github.com/oshjma/lang/ast"
)

/*
 Explorer - explore program to collect information necessary for emitting asm code
*/

type explorer struct {
	// Information necessary for emitting asm code
	gvars    map[*ast.LetStmt]*gvar
	lvars    map[*ast.LetStmt]*lvar
	strs     map[*ast.StringLit]*str
	garrs    map[*ast.ArrayLit]*garr
	larrs    map[*ast.ArrayLit]*larr
	fns      map[*ast.FuncLit]*fn
	branches map[ast.Node]*branch

	// Counters of labels
	nGvarLabel   int
	nStrLabel    int
	nGarrLabel   int
	nFnLabel     int
	nBranchLabel int

	// Used for finding local objects
	local  bool
	offset int
}

type gvar struct {
	label string
	size  int
}

type lvar struct {
	offset int
	size   int
}

type str struct {
	label string
	value string
}

type garr struct {
	label    string
	len      int
	elemSize int
}

type larr struct {
	offset   int
	len      int
	elemSize int
}

type fn struct {
	label     string
	localArea int
}

type branch struct {
	labels []string
}

func (x *explorer) gvarLabel(name string) string {
	label := fmt.Sprintf("gvar%d_%s", x.nGvarLabel, name)
	x.nGvarLabel += 1
	return label
}

func (x *explorer) strLabel() string {
	label := fmt.Sprintf("str%d", x.nStrLabel)
	x.nStrLabel += 1
	return label
}

func (x *explorer) garrLabel() string {
	label := fmt.Sprintf("garr%d", x.nGarrLabel)
	x.nGarrLabel += 1
	return label
}

func (x *explorer) fnLabel() string {
	label := fmt.Sprintf("fn%d", x.nFnLabel)
	x.nFnLabel += 1
	return label
}

func (x *explorer) branchLabel() string {
	label := fmt.Sprintf(".L%d", x.nBranchLabel)
	x.nBranchLabel += 1
	return label
}

func (x *explorer) exploreProgram(node *ast.Program) {
	for _, stmt := range node.Stmts {
		x.exploreStmt(stmt)
	}
}

func (x *explorer) exploreStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.LetStmt:
		x.exploreLetStmt(v)
	case *ast.BlockStmt:
		x.exploreBlockStmt(v)
	case *ast.IfStmt:
		x.exploreIfStmt(v)
	case *ast.ForStmt:
		x.exploreForStmt(v)
	case *ast.ReturnStmt:
		x.exploreReturnStmt(v)
	case *ast.AssignStmt:
		x.exploreAssignStmt(v)
	case *ast.ExprStmt:
		x.exploreExprStmt(v)
	}
}

func (x *explorer) exploreLetStmt(stmt *ast.LetStmt) {
	if stmt.Value != nil {
		x.exploreExpr(stmt.Value)
	}

	size := sizeOf(stmt.VarType)
	if x.local {
		x.offset = align(x.offset+size, size)
		x.lvars[stmt] = &lvar{offset: x.offset, size: size}
	} else {
		x.gvars[stmt] = &gvar{label: x.gvarLabel(stmt.Ident.Name), size: size}
	}
}

func (x *explorer) exploreBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt_ := range stmt.Stmts {
		x.exploreStmt(stmt_)
	}
}

func (x *explorer) exploreIfStmt(stmt *ast.IfStmt) {
	x.exploreExpr(stmt.Cond)
	x.exploreBlockStmt(stmt.Conseq)

	if stmt.Altern == nil {
		endLabel := x.branchLabel()
		x.branches[stmt] = &branch{labels: []string{endLabel}}
	} else {
		altLabel := x.branchLabel()
		x.exploreStmt(stmt.Altern)
		endLabel := x.branchLabel()
		x.branches[stmt] = &branch{labels: []string{altLabel, endLabel}}
	}
}

func (x *explorer) exploreForStmt(stmt *ast.ForStmt) {
	beginLabel := x.branchLabel()
	x.exploreExpr(stmt.Cond)
	x.exploreBlockStmt(stmt.Body)
	endLabel := x.branchLabel()
	x.branches[stmt] = &branch{labels: []string{beginLabel, endLabel}}
}

func (x *explorer) exploreReturnStmt(stmt *ast.ReturnStmt) {
	if stmt.Value != nil {
		x.exploreExpr(stmt.Value)
	}
}

func (x *explorer) exploreAssignStmt(stmt *ast.AssignStmt) {
	x.exploreExpr(stmt.Target)
	x.exploreExpr(stmt.Value)
}

func (x *explorer) exploreExprStmt(stmt *ast.ExprStmt) {
	x.exploreExpr(stmt.Expr)
}

func (x *explorer) exploreExpr(expr ast.Expr) {
	switch v := expr.(type) {
	case *ast.PrefixExpr:
		x.explorePrefixExpr(v)
	case *ast.InfixExpr:
		x.exploreInfixExpr(v)
	case *ast.IndexExpr:
		x.exploreIndexExpr(v)
	case *ast.CallExpr:
		x.exploreCallExpr(v)
	case *ast.LibcallExpr:
		x.exploreLibcallExpr(v)
	case *ast.StringLit:
		x.exploreStringLit(v)
	case *ast.ArrayLit:
		x.exploreArrayLit(v)
	case *ast.FuncLit:
		x.exploreFuncLit(v)
	}
}

func (x *explorer) explorePrefixExpr(expr *ast.PrefixExpr) {
	x.exploreExpr(expr.Right)
}

func (x *explorer) exploreInfixExpr(expr *ast.InfixExpr) {
	x.exploreExpr(expr.Left)
	x.exploreExpr(expr.Right)
}

func (x *explorer) exploreIndexExpr(expr *ast.IndexExpr) {
	x.exploreExpr(expr.Left)
	x.exploreExpr(expr.Index)
}

func (x *explorer) exploreCallExpr(expr *ast.CallExpr) {
	x.exploreExpr(expr.Left)
	for _, param := range expr.Params {
		x.exploreExpr(param)
	}
}

func (x *explorer) exploreLibcallExpr(expr *ast.LibcallExpr) {
	for _, param := range expr.Params {
		x.exploreExpr(param)
	}
}

func (x *explorer) exploreStringLit(expr *ast.StringLit) {
	x.strs[expr] = &str{label: x.strLabel(), value: expr.Value}
}

func (x *explorer) exploreArrayLit(expr *ast.ArrayLit) {
	for _, elem := range expr.Elems {
		x.exploreExpr(elem)
	}

	len := expr.Len
	elemSize := sizeOf(expr.ElemType)
	if x.local {
		x.offset = align(x.offset+len*elemSize, elemSize)
		x.larrs[expr] = &larr{offset: x.offset, len: len, elemSize: elemSize}
	} else {
		x.garrs[expr] = &garr{label: x.garrLabel(), len: len, elemSize: elemSize}
	}
}

func (x *explorer) exploreFuncLit(expr *ast.FuncLit) {
	x.local = true
	x.offset = 0

	for _, param := range expr.Params {
		x.exploreLetStmt(param)
	}
	x.exploreBlockStmt(expr.Body)
	endLabel := x.branchLabel()

	x.local = false
	x.fns[expr] = &fn{label: x.fnLabel(), localArea: align(x.offset, 16)}
	x.branches[expr] = &branch{labels: []string{endLabel}}
}
