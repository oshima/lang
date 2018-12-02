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
	fns      map[*ast.FuncDecl]*fn
	gvars    map[*ast.VarDecl]*gvar
	lvars    map[*ast.VarDecl]*lvar
	strs     map[*ast.StringLit]*str
	branches map[ast.Stmt]*branch

	// Counters of labels
	nGvarLabel   int
	nStrLabel    int
	nBranchLabel int

	// Used for finding local variables
	local  bool
	offset int
}

type fn struct {
	label string
	align int
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

type branch struct {
	labels []string
}

func (x *explorer) gvarLabel(name string) string {
	label := fmt.Sprintf(".GV%d_%s", x.nGvarLabel, name)
	x.nGvarLabel += 1
	return label
}

func (x *explorer) strLabel() string {
	label := fmt.Sprintf(".LC%d", x.nStrLabel)
	x.nStrLabel += 1
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
	case *ast.FuncDecl:
		x.exploreFuncDecl(v)
	case *ast.VarDecl:
		x.exploreVarDecl(v)
	case *ast.BlockStmt:
		x.exploreBlockStmt(v)
	case *ast.IfStmt:
		x.exploreIfStmt(v)
	case *ast.WhileStmt:
		x.exploreWhileStmt(v)
	case *ast.ReturnStmt:
		x.exploreReturnStmt(v)
	case *ast.AssignStmt:
		x.exploreAssignStmt(v)
	case *ast.ExprStmt:
		x.exploreExprStmt(v)
	}
}

func (x *explorer) exploreFuncDecl(stmt *ast.FuncDecl) {
	x.local = true
	x.offset = 0

	for _, param := range stmt.Params {
		x.exploreVarDecl(param)
	}
	x.exploreBlockStmt(stmt.Body)
	endLabel := x.branchLabel()

	x.local = false
	x.fns[stmt] = &fn{
		label: stmt.Ident.Name,
		align: align(x.offset, 16),
	}
	x.branches[stmt] = &branch{labels: []string{endLabel}}
}

func (x *explorer) exploreVarDecl(stmt *ast.VarDecl) {
	if stmt.Value != nil {
		x.exploreExpr(stmt.Value)
	}

	size := sizeof[stmt.Type]
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

func (x *explorer) exploreWhileStmt(stmt *ast.WhileStmt) {
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
	case *ast.FuncCall:
		x.exploreFuncCall(v)
	case *ast.StringLit:
		x.exploreStringLit(v)
	}
}

func (x *explorer) explorePrefixExpr(expr *ast.PrefixExpr) {
	x.exploreExpr(expr.Right)
}

func (x *explorer) exploreInfixExpr(expr *ast.InfixExpr) {
	x.exploreExpr(expr.Left)
	x.exploreExpr(expr.Right)
}

func (x *explorer) exploreFuncCall(expr *ast.FuncCall) {
	for _, param := range expr.Params {
		x.exploreExpr(param)
	}
}

func (x *explorer) exploreStringLit(expr *ast.StringLit) {
	x.strs[expr] = &str{label: x.strLabel(), value: expr.Value}
}
