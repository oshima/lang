package parse

import (
	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/token"
)

// Parse parses the input tokens and constructs the AST.
func Parse(tokens []*token.Token) *ast.Program {
	p := &parser{tokens: tokens, idx: -1}
	p.next()
	return p.parseProgram()
}
