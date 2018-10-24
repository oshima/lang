package token

const (
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	BANG      = "BANG"
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	ASTERISK  = "ASTERISK"
	SLASH     = "SLASH"
	AND       = "AND"
	OR        = "OR"
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
