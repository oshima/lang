package scan

import (
	"github.com/oshjma/lang/token"
	"github.com/oshjma/lang/util"
)

func Scan(src string) []*token.Token {
	s := &scanner{src: src, pos: -1}
	s.next()
	s.lastTk = &token.Token{Type: "__DUMMY__"}
	return s.readTokens()
}

type scanner struct {
	src string          // input source code
	pos int             // current position
	ch byte             // current character
	lastTk *token.Token // last token scanner has read
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
	s.skipWhitespace()
	for s.ch != 0 {
		tokens = append(tokens, s.readToken())
		s.skipWhitespace()
	}
	eof := &token.Token{Type: token.EOF, Literal: "<EOF>"}
	return append(tokens, eof)
}

func (s *scanner) readToken() *token.Token {
	var tk *token.Token
	switch {
	case s.ch == '+':
		tk = s.readPunct(token.PLUS)
	case s.ch == '-':
		tk = s.readMinusOrNegativeNumber()
	case s.ch == '*':
		tk = s.readPunct(token.ASTERISK)
	case s.ch == '/':
		tk = s.readPunct(token.SLASH)
	case s.ch == '(':
		tk = s.readPunct(token.LPAREN)
	case s.ch == ')':
		tk = s.readPunct(token.RPAREN)
	case s.ch == ';':
		tk = s.readPunct(token.SEMICOLON)
	case isDigit(s.ch):
		tk = s.readNumber()
	case isAlpha(s.ch):
		tk = s.readKeyword()
	default:
		util.Error("Invalid character %c", s.ch)
	}
	s.lastTk = tk
	return tk
}

func (s *scanner) readPunct(ty string) *token.Token {
	tk := &token.Token{Type: ty, Literal: string(s.ch)}
	s.next()
	return tk
}

func (s *scanner) readNumber() *token.Token {
	pos := s.pos
	if s.ch == '-' {
		s.next()
	}
	s.next()
	for isDigit(s.ch) {
		s.next()
	}
	return &token.Token{Type: token.NUMBER, Literal: s.src[pos:s.pos]}
}

func (s *scanner) readMinusOrNegativeNumber() *token.Token {
	if !isDigit(s.peekChar()) {
		return s.readPunct(token.MINUS)
	}
	ty := s.lastTk.Type
	if ty == token.RPAREN || ty == token.NUMBER || ty == token.TRUE || ty == token.FALSE {
		return s.readPunct(token.MINUS)
	} else {
		return s.readNumber()
	}
}

func (s *scanner) readKeyword() *token.Token {
	pos := s.pos
	s.next()
	for isAlpha(s.ch) || isDigit(s.ch) {
		s.next()
	}
	literal := s.src[pos:s.pos]
	var tk *token.Token
	switch literal {
	case "true":
		tk = &token.Token{Type: token.TRUE, Literal: literal}
	case "false":
		tk = &token.Token{Type: token.FALSE, Literal: literal}
	default:
		util.Error("Unexpected %s", literal)
	}
	return tk
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlpha(ch byte) bool {
	return 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || ch == '_'
}
