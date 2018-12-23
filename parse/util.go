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
