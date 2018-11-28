package gen

import "github.com/oshjma/lang/ast"

func Generate(node *ast.Program) {
	r := &resolver{
		relations: make(map[ast.Node]ast.Node),
	}
	r.resolveProgram(node, newEnv(nil))

	t := &typechecker{
		relations: r.relations,
	}
	t.typecheckProgram(node)

	x := &explorer{
		fns:      make(map[*ast.FuncDecl]*fn),
		gvars:    make(map[*ast.VarDecl]*gvar),
		lvars:    make(map[*ast.VarDecl]*lvar),
		strs:     make(map[*ast.StringLit]*str),
		branches: make(map[ast.Stmt]*branch),
	}
	x.exploreProgram(node)

	e := &emitter{
		relations: r.relations,
		fns:       x.fns,
		gvars:     x.gvars,
		lvars:     x.lvars,
		strs:      x.strs,
		branches:  x.branches,
	}
	e.emitProgram(node)
}
