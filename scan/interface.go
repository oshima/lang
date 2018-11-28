package scan

import "github.com/oshjma/lang/token"

func Scan(runes []rune) []*token.Token {
	s := &scanner{runes: runes, pos: -1}
	s.next()
	return s.readTokens()
}
