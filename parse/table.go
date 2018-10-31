package parse

import "github.com/oshjma/lang/token"

const (
	LOWEST int = iota
	ASSIGN
	EQUAL
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
)

var precedences = map[string]int{
	token.ASSIGN:   ASSIGN,
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
	token.AND:      PRODUCT,
}
