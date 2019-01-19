package parse

import "github.com/oshjma/lang/token"

const (
	LOWEST int = iota
	OR
	AND
	EQUAL
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	SUFFIX
)

var precedences = map[token.Type]int{
	token.OR:       OR,
	token.AND:      AND,
	token.EQ:       EQUAL,
	token.NE:       EQUAL,
	token.LT:       LESSGREATER,
	token.LE:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.GE:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.PERCENT:  PRODUCT,
	token.LBRACK:   SUFFIX,
	token.LPAREN:   SUFFIX,
}

var typeStart = map[token.Type]bool{
	token.INT:    true,
	token.BOOL:   true,
	token.STRING: true,
	token.LBRACK: true,
	token.LPAREN: true,
}

var assignOps = map[token.Type]bool{
	token.ASSIGN:     true,
	token.ADD_ASSIGN: true,
	token.SUB_ASSIGN: true,
	token.MUL_ASSIGN: true,
	token.DIV_ASSIGN: true,
	token.MOD_ASSIGN: true,
}

var unescape = map[rune]rune{
	'a':  '\a',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
	'"':  '"',
	'\\': '\\',
}

var libFuncs = map[string]bool{
	"puts":   true,
	"printf": true,
}
