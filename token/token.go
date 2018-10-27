package token

const (
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	LBRACE    = "LBRACE"
	RBRACE    = "RBRACE"
	BANG      = "BANG"
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	ASTERISK  = "ASTERISK"
	SLASH     = "SLASH"
	SEMICOLON = "SEMICOLON"

	EQ        = "EQ"
	NE        = "NE"
	LT        = "LT"
	LE        = "LE"
	GT        = "GT"
	GE        = "GE"
	AND       = "AND"
	OR        = "OR"

	NUMBER    = "NUMBER"
	TRUE      = "TRUE"
	FALSE     = "FALSE"

	IF        = "IF"
	ELSE      = "ELSE"

	EOF       = "EOF"
)

type Token struct {
	Type string
	Literal string
}
