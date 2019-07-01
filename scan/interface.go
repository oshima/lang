package scan

import "github.com/oshima/lang/token"

// Scan separetes the source code into lexical tokens.
func Scan(runes []rune) []*token.Token {
	s := &scanner{runes: runes, idx: -1, line: 1, col: 0}
	s.next()
	return s.readTokens()
}
