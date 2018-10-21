package scan

import (
	"github.com/oshjma/lang/token"
	"github.com/oshjma/lang/util"
)

func Scan(src string) []*token.Token {
	s := &scanner{src: src, pos: -1}
	s.next()

	var tokens []*token.Token
	var tk *token.Token
	for {
		s.skipWs()
		tk = s.readToken()
		tokens = append(tokens, tk)
		if (tk.Type == token.EOF) {
			break
		}
	}
	return tokens
}

type scanner struct {
	src string // input source code
	pos int    // current position
	ch byte    // current character
}

func (s *scanner) next() {
	s.pos += 1
	if s.pos < len(s.src) {
		s.ch = s.src[s.pos]
	} else {
		s.ch = 0
	}
}

func (s *scanner) peekChar() byte {
	if s.pos + 1 < len(s.src) {
		return s.src[s.pos + 1]
	}
	return 0
}

func (s *scanner) skipWs() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *scanner) readToken() *token.Token {
	var tk *token.Token
	switch s.ch {
	case '+':
		tk = s.readPunct(token.PLUS)
	case '-':
		if isDigit(s.peekChar()) {
			tk = s.readInt()
		} else {
			tk = s.readPunct(token.MINUS)
		}
	case ';':
		tk = s.readPunct(token.SEMICOLON)
	case 0:
		tk = s.readPunct(token.EOF)
	default:
		if isDigit(s.ch) {
			tk = s.readInt()
		} else {
			util.Error("Unexpected %q", string(s.ch))
		}
	}
	return tk
}

func (s *scanner) readPunct(ty string) *token.Token {
	tk := &token.Token{Type: ty, Source: string(s.ch)}
	s.next()
	return tk
}

func (s *scanner) readInt() *token.Token {
	pos := s.pos
	if s.ch == '-' {
		s.next()
	}
	s.next()
	for isDigit(s.ch) {
		s.next()
	}
	return &token.Token{Type: token.INT, Source: s.src[pos:s.pos]}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
