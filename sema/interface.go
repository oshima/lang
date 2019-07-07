package sema

import "github.com/oshima/lang/ast"

// Analyze checks if the program is correct.
func Analyze(prog *ast.Program) {
	r := &resolver{}
	r.resolveProgram(prog, newEnv(nil))

	t := &typechecker{}
	t.typecheckProgram(prog)
}
