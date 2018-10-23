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
	NUMBER    = "NUMBER"
	TRUE      = "TRUE"
	FALSE     = "FALSE"
)

type Token struct {
	Type string
	Literal string
}
