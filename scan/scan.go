package scan

import (
	"github.com/oshjma/lang/token"
	"github.com/oshjma/lang/util"
)

func Scan(src string) []*token.Token {
	s := &scanner{src: src, pos: -1}
	s.next()
	return s.readTokens()
}

type scanner struct {
	src string          // input source code
	pos int             // current position
	ch byte             // current character
	lastTk *token.Token // last token scanner read
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

func (s *scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *scanner) readTokens() []*token.Token {
	var tokens []*token.Token
	var tk *token.Token
	for {
		s.skipWhitespace()
		tk = s.readToken()
		tokens = append(tokens, tk)
		if (tk.Type == token.EOF) {
			break
		}
	}
	return tokens
}

func (s *scanner) readToken() *token.Token {
	var tk *token.Token
	switch s.ch {
	case '+':
		tk = s.readPunct(token.PLUS)
	case '-':
		if isDigit(s.peekChar()) {
			var last string
			if s.lastTk != nil {
				last = s.lastTk.Type
			}
			if last == token.INT || last == token.RPAREN {
				tk = s.readPunct(token.MINUS)
			} else {
				tk = s.readInt()
			}
		} else {
			tk = s.readPunct(token.MINUS)
		}
	case '*':
		tk = s.readPunct(token.ASTERISK)
	case '/':
		tk = s.readPunct(token.SLASH)
	case '(':
		tk = s.readPunct(token.LPAREN)
	case ')':
		tk = s.readPunct(token.RPAREN)
	case ';':
		tk = s.readPunct(token.SEMICOLON)
	case 0:
		tk = s.readEOF()
	default:
		if isDigit(s.ch) {
			tk = s.readInt()
		} else {
			util.Error("Invalid character %c", s.ch)
		}
	}
	s.lastTk = tk
	return tk
}

func (s *scanner) readPunct(ty string) *token.Token {
	tk := &token.Token{Type: ty, Literal: string(s.ch)}
	s.next()
	return tk
}

func (s *scanner) readEOF() *token.Token {
	tk := &token.Token{Type: token.EOF, Literal: "<EOF>"}
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
	return &token.Token{Type: token.INT, Literal: s.src[pos:s.pos]}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
