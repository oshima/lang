package parse

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/token"
)

func Parse(tokens []*token.Token) *ast.Program {
	p := &parser{tokens: tokens, pos: -1}
	p.next()
	return p.parseProgram()
}
