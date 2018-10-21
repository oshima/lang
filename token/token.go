package token

const (
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	ASTERISK  = "ASTERISK"
	SLASH     = "SLASH"
	SEMICOLON = "SEMICOLON"
	EOF       = "EOF"
	INT       = "INT"
)

type Token struct {
	Type string
	Source string
}
