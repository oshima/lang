package token

type Type string

const (
	LPAREN    Type = "LPAREN"
	RPAREN    Type = "RPAREN"
	LBRACK    Type = "LBRACK"
	RBRACK    Type = "RBRACK"
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

	LET      Type = "LET"
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

// for error messages
var strings = map[Type]string{
	LPAREN:    "(",
	RPAREN:    ")",
	LBRACK:    "[",
	RBRACK:    "]",
	LBRACE:    "{",
	RBRACE:    "}",
	ASSIGN:    "=",
	BANG:      "!",
	PLUS:      "+",
	MINUS:     "-",
	ASTERISK:  "*",
	SLASH:     "/",
	PERCENT:   "%",
	COMMA:     ",",
	SEMICOLON: ";",

	EQ:  "==",
	NE:  "!=",
	LT:  "<",
	LE:  "<=",
	GT:  ">",
	GE:  ">=",
	AND: "&&",
	OR:  "||",

	LET:      "let",
	IF:       "if",
	ELSE:     "else",
	FOR:      "for",
	RETURN:   "return",
	CONTINUE: "continue",
	BREAK:    "break",

	INT:    "int",
	BOOL:   "bool",
	STRING: "string",

	IDENT:  "identifier",
	NUMBER: "number",
	TRUE:   "true",
	FALSE:  "false",
	QUOTED: "quoted characters",

	EOF: "EOF",
}

func (t Type) String() string {
	return strings[t]
}

type Token struct {
	Type    Type
	Literal string
}
