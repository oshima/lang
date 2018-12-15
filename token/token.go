package token

type Type string

const (
	LPAREN    Type = "LPAREN"
	RPAREN    Type = "RPAREN"
	LBRACE    Type = "LBRACE"
	RBRACE    Type = "RBRACE"
	ASSIGN    Type = "ASSIGN"
	BANG      Type = "BANG"
	PLUS      Type = "PLUS"
	MINUS     Type = "MINUS"
	ASTERISK  Type = "ASTERISK"
	SLASH     Type = "SLASH"
	PERCENT   Type = "PERCENT"
	COMMA     Type = "COMMA"
	SEMICOLON Type = "SEMICOLON"

	EQ  Type = "EQ"
	NE  Type = "NE"
	LT  Type = "LT"
	LE  Type = "LE"
	GT  Type = "GT"
	GE  Type = "GE"
	AND Type = "AND"
	OR  Type = "OR"

	FUNC     Type = "FUNC"
	VAR      Type = "VAR"
	IF       Type = "IF"
	ELSE     Type = "ELSE"
	FOR      Type = "FOR"
	RETURN   Type = "RETURN"
	CONTINUE Type = "CONTINUE"
	BREAK    Type = "BREAK"

	INT    Type = "INT"
	BOOL   Type = "BOOL"
	STRING Type = "STRING"

	IDENT  Type = "IDENT"
	NUMBER Type = "NUMBER"
	TRUE   Type = "TRUE"
	FALSE  Type = "FALSE"
	QUOTED Type = "QUOTED"

	EOF Type = "EOF"
)

type Token struct {
	Type    Type
	Literal string
}
