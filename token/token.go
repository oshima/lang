package token

const (
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	SEMICOLON = "SEMICOLON"
	EOF       = "EOF"
	INT       = "INT"
)

type Token struct {
	Type string
	Source string
}
