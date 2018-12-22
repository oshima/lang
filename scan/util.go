package scan

import "github.com/oshjma/lang/token"

var punctuations = map[rune]token.Type{
	'(': token.LPAREN,
	')': token.RPAREN,
	'[': token.LBRACK,
	']': token.RBRACK,
	'{': token.LBRACE,
	'}': token.RBRACE,
	'+': token.PLUS,
	'*': token.ASTERISK,
	'%': token.PERCENT,
	'/': token.SLASH,
	',': token.COMMA,
	';': token.SEMICOLON,
}

var keywords = map[string]token.Type{
	"func":     token.FUNC,
	"var":      token.VAR,
	"if":       token.IF,
	"else":     token.ELSE,
	"for":      token.FOR,
	"return":   token.RETURN,
	"continue": token.CONTINUE,
	"break":    token.BREAK,
	"int":      token.INT,
	"bool":     token.BOOL,
	"string":   token.STRING,
	"true":     token.TRUE,
	"false":    token.FALSE,
}

var exprTerminators = map[token.Type]bool{
	token.RPAREN: true,
	token.RBRACK: true,
	token.IDENT:  true,
	token.NUMBER: true,
	token.TRUE:   true,
	token.FALSE:  true,
	token.QUOTED: true,
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isAlpha(ch rune) bool {
	return 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || ch == '_'
}
