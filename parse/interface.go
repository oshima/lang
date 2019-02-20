package parse

import (
	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/token"
)

func Parse(tokens []*token.Token) *ast.Program {
	p := &parser{tokens: tokens, pos: -1}
	p.next()
	return p.parseProgram()
}
