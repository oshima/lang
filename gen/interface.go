package gen

import "github.com/oshima/lang/ast"

// Generate travarses the input AST and emits the target assembly code
func Generate(prog *ast.Program, meta *ast.Metadata) {
	x := &explorer{
		types:    meta.Types,
		gvars:    make(map[*ast.VarDecl]*gvar),
		lvars:    make(map[*ast.VarDecl]*lvar),
		strs:     make(map[*ast.StringLit]*str),
		grngs:    make(map[*ast.RangeLit]*grng),
		lrngs:    make(map[*ast.RangeLit]*lrng),
		garrs:    make(map[ast.Expr]*garr),
		larrs:    make(map[ast.Expr]*larr),
		fns:      make(map[ast.Node]*fn),
		branches: make(map[ast.Node]*branch),
	}
	x.exploreProgram(prog)

	e := &emitter{
		refs:     meta.Refs,
		types:    meta.Types,
		gvars:    x.gvars,
		lvars:    x.lvars,
		strs:     x.strs,
		grngs:    x.grngs,
		lrngs:    x.lrngs,
		garrs:    x.garrs,
		larrs:    x.larrs,
		fns:      x.fns,
		branches: x.branches,
	}
	e.emitProgram(prog)
}
