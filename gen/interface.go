package gen

import "github.com/oshima/lang/ast"

// Generate travarses the input AST and emits the target assembly code
func Generate(prog *ast.Program, meta *ast.Metadata) {
	x := &explorer{
		types: meta.Types,
		gvars: make(map[ast.Decl]*gvar),
		grngs: make(map[ast.Expr]*grng),
		garrs: make(map[ast.Expr]*garr),
		lvars: make(map[ast.Decl]*lvar),
		lrngs: make(map[ast.Expr]*lrng),
		larrs: make(map[ast.Expr]*larr),
		strs:  make(map[ast.Expr]*str),
		fns:   make(map[ast.Node]*fn),
		brs:   make(map[ast.Node]*br),
	}
	x.exploreProgram(prog)

	e := &emitter{
		refs:  meta.Refs,
		types: meta.Types,
		gvars: x.gvars,
		grngs: x.grngs,
		garrs: x.garrs,
		lvars: x.lvars,
		lrngs: x.lrngs,
		larrs: x.larrs,
		strs:  x.strs,
		fns:   x.fns,
		brs:   x.brs,
	}
	e.emitProgram(prog)
}
