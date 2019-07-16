package gen

import "github.com/oshima/lang/ast"

// Generate emits the target assembly code.
func Generate(prog *ast.Program) {
	x := &explorer{
		gvars: make(map[ast.Decl]*gvar),
		grans: make(map[ast.Expr]*gran),
		garrs: make(map[ast.Expr]*garr),
		lvars: make(map[ast.Decl]*lvar),
		lrans: make(map[ast.Expr]*lran),
		larrs: make(map[ast.Expr]*larr),
		strs:  make(map[ast.Expr]*str),
		fns:   make(map[ast.Node]*fn),
		brs:   make(map[ast.Node]*br),
	}
	x.exploreProgram(prog)

	e := &emitter{
		gvars: x.gvars,
		grans: x.grans,
		garrs: x.garrs,
		lvars: x.lvars,
		lrans: x.lrans,
		larrs: x.larrs,
		strs:  x.strs,
		fns:   x.fns,
		brs:   x.brs,
	}
	e.emitProgram(prog)
}
