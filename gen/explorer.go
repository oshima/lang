package gen

import (
	"fmt"
	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/types"
)

// explorer gathers the objects necessary for emitting target assembly code.
type explorer struct {
	types map[ast.Expr]types.Type

	// objects
	gvars map[ast.Decl]*gvar
	grngs map[ast.Expr]*grng
	garrs map[ast.Expr]*garr
	lvars map[ast.Decl]*lvar
	lrngs map[ast.Expr]*lrng
	larrs map[ast.Expr]*larr
	strs  map[ast.Expr]*str
	fns   map[ast.Node]*fn
	brs   map[ast.Node]*br

	// counters of labels
	nGvarLabel int
	nGrngLabel int
	nGarrLabel int
	nStrLabel  int
	nFnLabel   int
	nBrLabel   int

	// used for collecting local objects
	local  bool
	offset int
}

func (x *explorer) gvarLabel() string {
	label := fmt.Sprintf("gvar%d", x.nGvarLabel)
	x.nGvarLabel++
	return label
}

func (x *explorer) grngLabel() string {
	label := fmt.Sprintf("grng%d", x.nGrngLabel)
	x.nGrngLabel++
	return label
}

func (x *explorer) garrLabel() string {
	label := fmt.Sprintf("garr%d", x.nGarrLabel)
	x.nGarrLabel++
	return label
}

func (x *explorer) strLabel() string {
	label := fmt.Sprintf("str%d", x.nStrLabel)
	x.nStrLabel++
	return label
}

func (x *explorer) fnLabel() string {
	label := fmt.Sprintf("fn%d", x.nFnLabel)
	x.nFnLabel++
	return label
}

func (x *explorer) brLabel() string {
	label := fmt.Sprintf(".L%d", x.nBrLabel)
	x.nBrLabel++
	return label
}

// ----------------------------------------------------------------
// Program

func (x *explorer) exploreProgram(node *ast.Program) {
	for _, stmt := range node.Stmts {
		x.exploreStmt(stmt)
	}
}

// ----------------------------------------------------------------
// Stmt

func (x *explorer) exploreStmt(stmt ast.Stmt) {
	switch v := stmt.(type) {
	case *ast.BlockStmt:
		x.exploreBlockStmt(v)
	case *ast.VarStmt:
		x.exploreVarStmt(v)
	case *ast.FuncStmt:
		x.exploreFuncStmt(v)
	case *ast.IfStmt:
		x.exploreIfStmt(v)
	case *ast.WhileStmt:
		x.exploreWhileStmt(v)
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

func (x *explorer) exploreBlockStmt(stmt *ast.BlockStmt) {
	for _, stmt := range stmt.Stmts {
		x.exploreStmt(stmt)
	}
}

func (x *explorer) exploreVarStmt(stmt *ast.VarStmt) {
	for _, v := range stmt.Vars {
		x.exploreVarDecl(v)
	}
}

func (x *explorer) exploreFuncStmt(stmt *ast.FuncStmt) {
	x.exploreFuncDecl(stmt.Func)
}

func (x *explorer) exploreIfStmt(stmt *ast.IfStmt) {
	x.exploreExpr(stmt.Cond)
	x.exploreBlockStmt(stmt.Body)

	if stmt.Else == nil {
		endLabel := x.brLabel()
		x.brs[stmt] = &br{labels: []string{endLabel}}
	} else {
		elseLabel := x.brLabel()
		x.exploreStmt(stmt.Else)
		endLabel := x.brLabel()
		x.brs[stmt] = &br{labels: []string{elseLabel, endLabel}}
	}
}

func (x *explorer) exploreWhileStmt(stmt *ast.WhileStmt) {
	beginLabel := x.brLabel()
	x.exploreExpr(stmt.Cond)
	x.exploreBlockStmt(stmt.Body)
	endLabel := x.brLabel()
	x.brs[stmt] = &br{labels: []string{beginLabel, endLabel}}
}

func (x *explorer) exploreForStmt(stmt *ast.ForStmt) {
	x.exploreVarDecl(stmt.Elem)
	x.exploreVarDecl(stmt.Index)
	x.exploreVarDecl(stmt.Iter)
	beginLabel := x.brLabel()
	x.exploreBlockStmt(stmt.Body)
	continueLabel := x.brLabel()
	endLabel := x.brLabel()
	x.brs[stmt] = &br{labels: []string{beginLabel, continueLabel, endLabel}}
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

// ----------------------------------------------------------------
// Expr

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
	case *ast.LibCallExpr:
		x.exploreLibCallExpr(v)
	case *ast.StringLit:
		x.exploreStringLit(v)
	case *ast.RangeLit:
		x.exploreRangeLit(v)
	case *ast.ArrayLit:
		x.exploreArrayLit(v)
	case *ast.ArrayShortLit:
		x.exploreArrayShortLit(v)
	case *ast.FuncLit:
		x.exploreFuncLit(v)
	}
}

func (x *explorer) explorePrefixExpr(expr *ast.PrefixExpr) {
	x.exploreExpr(expr.Right)
}

func (x *explorer) exploreInfixExpr(expr *ast.InfixExpr) {
	switch expr.Op {
	case "&&", "||":
		x.exploreExpr(expr.Left)
		x.exploreExpr(expr.Right)
	default:
		x.exploreExpr(expr.Right)
		x.exploreExpr(expr.Left)
	}

	switch expr.Op {
	case "&&", "||":
		endLabel := x.brLabel()
		x.brs[expr] = &br{labels: []string{endLabel}}
	case "in":
		ty := x.types[expr.Right]

		switch ty.(type) {
		case *types.Range:
			falseLabel := x.brLabel()
			endLabel := x.brLabel()
			x.brs[expr] = &br{labels: []string{falseLabel, endLabel}}
		case *types.Array:
			beginLabel := x.brLabel()
			falseLabel := x.brLabel()
			endLabel := x.brLabel()
			x.brs[expr] = &br{labels: []string{beginLabel, falseLabel, endLabel}}
		}
	}
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

func (x *explorer) exploreLibCallExpr(expr *ast.LibCallExpr) {
	for _, param := range expr.Params {
		x.exploreExpr(param)
	}
}

func (x *explorer) exploreStringLit(expr *ast.StringLit) {
	x.strs[expr] = &str{label: x.strLabel(), value: expr.Value}
}

func (x *explorer) exploreRangeLit(expr *ast.RangeLit) {
	x.exploreExpr(expr.Lower)
	x.exploreExpr(expr.Upper)

	if x.local {
		x.offset = align(x.offset+16, 8)
		x.lrngs[expr] = &lrng{offset: x.offset}
	} else {
		x.grngs[expr] = &grng{label: x.grngLabel()}
	}
}

func (x *explorer) exploreArrayLit(expr *ast.ArrayLit) {
	for _, elem := range expr.Elems {
		x.exploreExpr(elem)
	}

	ty := x.types[expr].(*types.Array)
	len := ty.Len
	elemSize := sizeOf(ty.ElemType)

	if x.local {
		x.offset = align(x.offset+len*elemSize, elemSize)
		x.larrs[expr] = &larr{offset: x.offset, len: len, elemSize: elemSize}
	} else {
		x.garrs[expr] = &garr{label: x.garrLabel(), len: len, elemSize: elemSize}
	}
}

func (x *explorer) exploreArrayShortLit(expr *ast.ArrayShortLit) {
	if expr.Value != nil {
		x.exploreExpr(expr.Value)
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
		x.exploreVarDecl(param)
	}
	x.exploreBlockStmt(expr.Body)
	endLabel := x.brLabel()

	x.local = false
	x.fns[expr] = &fn{label: x.fnLabel(), localArea: align(x.offset, 16)}
	x.brs[expr] = &br{labels: []string{endLabel}}
}

// ----------------------------------------------------------------
// Decl

func (x *explorer) exploreVarDecl(decl *ast.VarDecl) {
	if decl.Value != nil {
		x.exploreExpr(decl.Value)
	}

	size := sizeOf(decl.VarType)
	if x.local {
		x.offset = align(x.offset+size, size)
		x.lvars[decl] = &lvar{offset: x.offset, size: size}
	} else {
		label := x.gvarLabel()
		if decl.Name != "" {
			label += "_" + decl.Name
		}
		x.gvars[decl] = &gvar{label: label, size: size}
	}
}

func (x *explorer) exploreFuncDecl(decl *ast.FuncDecl) {
	x.local = true
	x.offset = 0

	for _, param := range decl.Params {
		x.exploreVarDecl(param)
	}
	x.exploreBlockStmt(decl.Body)
	endLabel := x.brLabel()

	x.local = false
	x.fns[decl] = &fn{
		label:     x.fnLabel() + "_" + decl.Name,
		localArea: align(x.offset, 16),
	}
	x.brs[decl] = &br{labels: []string{endLabel}}
}
