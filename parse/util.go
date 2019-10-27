package parse

import "github.com/oshima/lang/token"

// The oparator precedence
const (
	LOWEST int = iota
	OR
	AND
	EQUAL
	LESSGREATER
	SUM
	PRODUCT
	IN
	BETWEEN
	PREFIX
	SUFFIX
)

var precOf = map[token.Type]int{
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
	token.IN:       IN,
	token.BETWEEN:  BETWEEN,
	token.LBRACK:   SUFFIX,
	token.LPAREN:   SUFFIX,
}

var assignOps = map[token.Type]bool{
	token.ASSIGN:    true,
	token.ADDASSIGN: true,
	token.SUBASSIGN: true,
	token.MULASSIGN: true,
	token.DIVASSIGN: true,
	token.MODASSIGN: true,
}

var typeBegin = map[token.Type]bool{
	token.INT:    true,
	token.BOOL:   true,
	token.STRING: true,
	token.RANGE:  true,
	token.LBRACK: true,
	token.LPAREN: true,
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
	"sleep":  true,
}
