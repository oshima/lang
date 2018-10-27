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

var punctuations = map[byte]string{
	'(': token.LPAREN,
	')': token.RPAREN,
	'{': token.LBRACE,
	'}': token.RBRACE,
	'+': token.PLUS,
	'*': token.ASTERISK,
	'/': token.SLASH,
	';': token.SEMICOLON,
}

var keywords = map[string]string{
	"true":  token.TRUE,
	"false": token.FALSE,
	"if":    token.IF,
	"else":  token.ELSE,
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

func (s *scanner) expect(ch byte) {
	if s.ch != ch {
		util.Error("Expected %c but got %c", ch, s.ch)
	}
	s.next()
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
	switch s.ch {
	case '(', ')', '{', '}', '+', '*', '/', ';':
		tk = s.readPunct()
	case '!':
		tk = s.readBangOrNotEqual()
	case '-':
		tk = s.readMinusOrNegativeNumber()
	case '=':
		tk = s.readEqual()
	case '<':
		tk = s.readLessOrLessEqual()
	case '>':
		tk = s.readGreaterOrGreaterEqual()
	case '&':
		tk = s.readAnd()
	case '|':
		tk = s.readOr()
	default:
		switch {
		case isDigit(s.ch):
			tk = s.readNumber()
		case isAlpha(s.ch):
			tk = s.readKeyword()
		default:
			util.Error("Invalid character %c", s.ch)
		}
	}
	s.lastTk = tk
	return tk
}

func (s *scanner) readPunct() *token.Token {
	ty := punctuations[s.ch]
	literal := string(s.ch)
	s.next()
	return &token.Token{Type: ty, Literal: literal}
}

func (s *scanner) readEqual() *token.Token {
	s.next()
	s.expect('=')
	return &token.Token{Type: token.EQ, Literal: "=="}
}

func (s *scanner) readAnd() *token.Token {
	s.next()
	s.expect('&')
	return &token.Token{Type: token.AND, Literal: "&&"}
}

func (s *scanner) readOr() *token.Token {
	s.next()
	s.expect('|')
	return &token.Token{Type: token.OR, Literal: "||"}
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

func (s *scanner) readKeyword() *token.Token {
	pos := s.pos
	s.next()
	for isAlpha(s.ch) || isDigit(s.ch) {
		s.next()
	}
	literal := s.src[pos:s.pos]
	ty, ok := keywords[literal]
	if !ok {
		util.Error("Unexpected %s", literal)
	}
	return &token.Token{Type: ty, Literal: literal}
}

func (s *scanner) readBangOrNotEqual() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.NE, Literal: "!="}
	}
	return &token.Token{Type: token.BANG, Literal: "!"}
}

func (s *scanner) readMinusOrNegativeNumber() *token.Token {
	if !isDigit(s.peekChar()) {
		s.next()
		return &token.Token{Type: token.MINUS, Literal: "-"}
	}
	ty := s.lastTk.Type
	if ty == token.RPAREN || ty == token.NUMBER || ty == token.TRUE || ty == token.FALSE {
		s.next()
		return &token.Token{Type: token.MINUS, Literal: "-"}
	}
	return s.readNumber()
}

func (s *scanner) readLessOrLessEqual() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.LE, Literal: "<="}
	}
	return &token.Token{Type: token.LT, Literal: "<"}
}

func (s *scanner) readGreaterOrGreaterEqual() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.GE, Literal: ">="}
	}
	return &token.Token{Type: token.GT, Literal: ">"}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlpha(ch byte) bool {
	return 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || ch == '_'
}
