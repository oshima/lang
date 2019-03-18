package scan

import "github.com/oshima/lang/token"

// Scan scans the source code and returns the lexical tokens
func Scan(runes []rune) []*token.Token {
	s := &scanner{runes: runes, pos: -1}
	s.next()
	return s.readTokens()
}
