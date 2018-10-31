package token

const (
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	LBRACE    = "LBRACE"
	RBRACE    = "RBRACE"
	ASSIGN    = "ASSIGN"
	BANG      = "BANG"
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	ASTERISK  = "ASTERISK"
	SLASH     = "SLASH"
	SEMICOLON = "SEMICOLON"

	EQ  = "EQ"
	NE  = "NE"
	LT  = "LT"
	LE  = "LE"
	GT  = "GT"
	GE  = "GE"
	AND = "AND"
	OR  = "OR"

	LET  = "LET"
	IF   = "IF"
	ELSE = "ELSE"

	INT  = "INT"
	BOOL = "BOOL"

	IDENT  = "IDENT"
	NUMBER = "NUMBER"
	TRUE   = "TRUE"
	FALSE  = "FALSE"

	EOF = "EOF"
)

type Token struct {
	Type    string
	Literal string
}
