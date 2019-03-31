package scan

import "github.com/oshima/lang/token"

// Scan separetes the source code into lexical tokens.
func Scan(runes []rune) []*token.Token {
	s := &scanner{runes: runes, pos: -1}
	s.next()
	return s.readTokens()
}
