package token

const (
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	ASTERISK  = "ASTERISK"
	SLASH     = "SLASH"
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	SEMICOLON = "SEMICOLON"
	EOF       = "EOF"
	INT       = "INT"
)

type Token struct {
	Type string
	Literal string
}
