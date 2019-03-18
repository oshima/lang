package sema

import (
	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/types"
)

// Analyze travarses the input AST and check if the program is correct
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
