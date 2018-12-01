package parse

import "github.com/oshjma/lang/token"

const (
	LOWEST int = iota
	EQUAL
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
)

var precedences = map[string]int{
	token.EQ:       EQUAL,
	token.NE:       EQUAL,
	token.LT:       LESSGREATER,
	token.LE:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.GE:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.OR:       SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.PERCENT:  PRODUCT,
	token.AND:      PRODUCT,
}

var typeNames = map[string]bool{
	token.INT:    true,
	token.BOOL:   true,
	token.STRING: true,
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
