package gen

import "github.com/oshjma/lang/ast"

func Generate(prog *ast.Program, meta *ast.Metadata) {
	x := &explorer{
		fns:      make(map[*ast.FuncDecl]*fn),
		gvars:    make(map[*ast.VarDecl]*gvar),
		lvars:    make(map[*ast.VarDecl]*lvar),
		strs:     make(map[*ast.StringLit]*str),
		branches: make(map[ast.Stmt]*branch),
	}
	x.exploreProgram(prog)

	e := &emitter{
		refs:     meta.Refs,
		types:    meta.Types,
		fns:      x.fns,
		gvars:    x.gvars,
		lvars:    x.lvars,
		strs:     x.strs,
		branches: x.branches,
	}
	e.emitProgram(prog)
}
