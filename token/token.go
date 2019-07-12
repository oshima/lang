package token

import "fmt"

// Token represents a lexical token.
type Token struct {
	Type    Type
	Pos     *Pos
	Literal string
}

// Type represents the type of token.
type Type uint8

func (t Type) String() string {
	return strings[t]
}

// Pos represents the position of token.
type Pos struct {
	Line int
	Col  int
}

func (p *Pos) String() string {
	return fmt.Sprintf("%d,%d", p.Line, p.Col)
}

// The list of token types.
const (
	EOF Type = iota
	COMMENT

	LPAREN
	RPAREN
	LBRACK
	RBRACK
	LBRACE
	RBRACE
	COMMA
	COLON
	SEMICOLON
	ASSIGN
	BANG
	PLUS
	MINUS
	ASTERISK
	SLASH
	PERCENT

	BETWEEN
	ARROW

	EQ
	NE
	LT
	LE
	GT
	GE
	AND
	OR

	ADDASSIGN
	SUBASSIGN
	MULASSIGN
	DIVASSIGN
	MODASSIGN

	VAR
	FUNC
	IF
	ELSE
	WHILE
	FOR
	IN
	CONTINUE
	BREAK
	RETURN

	VOID
	INT
	BOOL
	STRING
	RANGE

	IDENT
	NUMBER
	TRUE
	FALSE
	QUOTED
)

var strings = map[Type]string{
	EOF:     "eof",
	COMMENT: "comment",

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

	BETWEEN: "..",
	ARROW:   "->",

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

	VOID:   "void",
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
