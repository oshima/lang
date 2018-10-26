package token

const (
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	BANG      = "BANG"
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	ASTERISK  = "ASTERISK"
	SLASH     = "SLASH"
	EQ        = "EQ"
	NE        = "NE"
	LT        = "LT"
	LE        = "LE"
	GT        = "GT"
	GE        = "GE"
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
