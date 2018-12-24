package gen

import "github.com/oshjma/lang/ast"

func Generate(prog *ast.Program, meta *ast.Metadata) {
	x := &explorer{
		gvars:    make(map[*ast.LetStmt]*gvar),
		lvars:    make(map[*ast.LetStmt]*lvar),
		strs:     make(map[*ast.StringLit]*str),
		garrs:    make(map[*ast.ArrayLit]*garr),
		larrs:    make(map[*ast.ArrayLit]*larr),
		fns:      make(map[*ast.FuncLit]*fn),
		branches: make(map[ast.Node]*branch),
	}
	x.exploreProgram(prog)

	e := &emitter{
		refs:     meta.Refs,
		types:    meta.Types,
		gvars:    x.gvars,
		lvars:    x.lvars,
		strs:     x.strs,
		garrs:    x.garrs,
		larrs:    x.larrs,
		fns:      x.fns,
		branches: x.branches,
	}
	e.emitProgram(prog)
}
