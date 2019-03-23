package token

// Token represents the lexical token
type Token struct {
	Type    Type
	Literal string
}

// Type represents the type of token
type Type string

// The list of token types
const (
	COMMENT Type = "COMMENT"
	EOF     Type = "EOF"

	LPAREN    Type = "LPAREN"
	RPAREN    Type = "RPAREN"
	LBRACK    Type = "LBRACK"
	RBRACK    Type = "RBRACK"
	LBRACE    Type = "LBRACE"
	RBRACE    Type = "RBRACE"
	COMMA     Type = "COMMA"
	COLON     Type = "COLON"
	SEMICOLON Type = "SEMICOLON"
	ASSIGN    Type = "ASSIGN"
	BANG      Type = "BANG"
	PLUS      Type = "PLUS"
	MINUS     Type = "MINUS"
	ASTERISK  Type = "ASTERISK"
	SLASH     Type = "SLASH"
	PERCENT   Type = "PERCENT"

	EQ  Type = "EQ"
	NE  Type = "NE"
	LT  Type = "LT"
	LE  Type = "LE"
	GT  Type = "GT"
	GE  Type = "GE"
	AND Type = "AND"
	OR  Type = "OR"

	ADDASSIGN Type = "ADDASSIGN"
	SUBASSIGN Type = "SUBASSIGN"
	MULASSIGN Type = "MULASSIGN"
	DIVASSIGN Type = "DIVASSIGN"
	MODASSIGN Type = "MODASSIGN"

	BETWEEN Type = "BETWEEN"
	ARROW   Type = "ARROW"

	VAR      Type = "VAR"
	FUNC     Type = "FUNC"
	IF       Type = "IF"
	ELSE     Type = "ELSE"
	WHILE    Type = "WHILE"
	FOR      Type = "FOR"
	IN       Type = "IN"
	CONTINUE Type = "CONTINUE"
	BREAK    Type = "BREAK"
	RETURN   Type = "RETURN"

	INT    Type = "INT"
	BOOL   Type = "BOOL"
	STRING Type = "STRING"
	RANGE  Type = "RANGE"

	IDENT  Type = "IDENT"
	NUMBER Type = "NUMBER"
	TRUE   Type = "TRUE"
	FALSE  Type = "FALSE"
	QUOTED Type = "QUOTED"
)

// for error messages
var strings = map[Type]string{
	COMMENT: "comment",
	EOF:     "EOF",

	LPAREN:    "(",
	RPAREN:    ")",
	LBRACK:    "[",
	RBRACK:    "]",
	LBRACE:    "{",
	RBRACE:    "}",
	COMMA:     ",",
	COLON:     ":",
	SEMICOLON: ";",
	ASSIGN:    "=",
	BANG:      "!",
	PLUS:      "+",
	MINUS:     "-",
	ASTERISK:  "*",
	SLASH:     "/",
	PERCENT:   "%",

	EQ:  "==",
	NE:  "!=",
	LT:  "<",
	LE:  "<=",
	GT:  ">",
	GE:  ">=",
	AND: "&&",
	OR:  "||",

	ADDASSIGN: "+=",
	SUBASSIGN: "-=",
	MULASSIGN: "*=",
	DIVASSIGN: "/=",
	MODASSIGN: "%=",

	BETWEEN: "..",
	ARROW:   "->",

	VAR:      "var",
	FUNC:     "func",
	IF:       "if",
	ELSE:     "else",
	WHILE:    "while",
	FOR:      "for",
	IN:       "in",
	CONTINUE: "continue",
	BREAK:    "break",
	RETURN:   "return",

	INT:    "int",
	BOOL:   "bool",
	STRING: "string",
	RANGE:  "range",

	IDENT:  "identifier",
	NUMBER: "number",
	TRUE:   "true",
	FALSE:  "false",
	QUOTED: "quoted characters",
}

func (t Type) String() string {
	return strings[t]
}
