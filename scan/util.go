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
	'/': token.SLASH,
	'%': token.PERCENT,
	',': token.COMMA,
	':': token.COLON,
	';': token.SEMICOLON,
}

var keywords = map[string]token.Type{
	"let":      token.LET,
	"if":       token.IF,
	"else":     token.ELSE,
	"for":      token.FOR,
	"continue": token.CONTINUE,
	"break":    token.BREAK,
	"return":   token.RETURN,
	"int":      token.INT,
	"bool":     token.BOOL,
	"string":   token.STRING,
	"true":     token.TRUE,
	"false":    token.FALSE,
}

var exprEnd = map[token.Type]bool{
	token.RPAREN: true,
	token.RBRACK: true,
	token.RBRACE: true,
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
