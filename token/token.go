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
	PERCENT   = "PERCENT"
	COMMA     = "COMMA"
	SEMICOLON = "SEMICOLON"

	EQ  = "EQ"
	NE  = "NE"
	LT  = "LT"
	LE  = "LE"
	GT  = "GT"
	GE  = "GE"
	AND = "AND"
	OR  = "OR"

	FUNC     = "FUNC"
	VAR      = "VAR"
	IF       = "IF"
	ELSE     = "ELSE"
	FOR      = "FOR"
	RETURN   = "RETURN"
	CONTINUE = "CONTINUE"
	BREAK    = "BREAK"

	INT    = "INT"
	BOOL   = "BOOL"
	STRING = "STRING"

	IDENT  = "IDENT"
	NUMBER = "NUMBER"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	QUOTED = "QUOTED"

	EOF = "EOF"
)

type Token struct {
	Type    string
	Literal string
}
