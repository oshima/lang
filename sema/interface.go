package sema

import "github.com/oshjma/lang/ast"
import "github.com/oshjma/lang/types"

func Analyze(prog *ast.Program) *ast.Metadata {
	r := &resolver{
		refs: make(map[ast.Node]ast.Node),
	}
	r.resolveProgram(prog, newEnv(nil))

	t := &typechecker{
		refs:  r.refs,
		types: make(map[ast.Expr]types.Type),
	}
	t.typecheckProgram(prog)

	return &ast.Metadata{Refs: r.refs, Types: t.types}
}
