package scan

import (
	"github.com/oshjma/lang/token"
	"github.com/oshjma/lang/util"
)

func Scan(runes []rune) []*token.Token {
	s := &scanner{runes: runes, pos: -1}
	s.next()
	return s.readTokens()
}

type scanner struct {
	runes  []rune       // input source code
	pos    int          // current position
	ch     rune         // current character
	lastTk *token.Token // last token scanner has read
}

func (s *scanner) next() {
	s.pos += 1
	if s.pos < len(s.runes) {
		s.ch = s.runes[s.pos]
	} else {
		s.ch = 0
	}
}

func (s *scanner) peekCh() rune {
	if s.pos+1 < len(s.runes) {
		return s.runes[s.pos+1]
	}
	return 0
}

func (s *scanner) expect(ch rune) {
	if s.ch != ch {
		util.Error("Expected %c but got %c", ch, s.ch)
	}
}

func (s *scanner) skipWs() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *scanner) readTokens() []*token.Token {
	tokens := make([]*token.Token, 0, 64)
	s.skipWs()
	for s.ch != 0 {
		tokens = append(tokens, s.readToken())
		s.skipWs()
	}
	eof := &token.Token{Type: token.EOF, Literal: "<EOF>"}
	return append(tokens, eof)
}

func (s *scanner) readToken() *token.Token {
	var tk *token.Token
	switch s.ch {
	case '(', ')', '{', '}', '+', '*', '/', ',', ';':
		tk = s.readPunct()
	case '!':
		tk = s.readBangOrNotEqual()
	case '-':
		tk = s.readMinusOrNegativeNumber()
	case '=':
		tk = s.readAssignOrEqual()
	case '<':
		tk = s.readLessOrLessEqual()
	case '>':
		tk = s.readGreaterOrGreaterEqual()
	case '&':
		tk = s.readAnd()
	case '|':
		tk = s.readOr()
	case '"':
		tk = s.readQuoted()
	default:
		switch {
		case isDigit(s.ch):
			tk = s.readNumber()
		case isAlpha(s.ch):
			tk = s.readKeywordOrIdentifier()
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

func (s *scanner) readBangOrNotEqual() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.NE, Literal: "!="}
	}
	return &token.Token{Type: token.BANG, Literal: "!"}
}

func (s *scanner) readMinusOrNegativeNumber() *token.Token {
	if !isDigit(s.peekCh()) {
		s.next()
		return &token.Token{Type: token.MINUS, Literal: "-"}
	}
	if s.lastTk == nil {
		return s.readNumber()
	}
	if _, ok := exprTerminators[s.lastTk.Type]; ok {
		s.next()
		return &token.Token{Type: token.MINUS, Literal: "-"}
	}
	return s.readNumber()
}

func (s *scanner) readAssignOrEqual() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.EQ, Literal: "=="}
	}
	return &token.Token{Type: token.ASSIGN, Literal: "="}
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

func (s *scanner) readAnd() *token.Token {
	s.next()
	s.expect('&')
	s.next()
	return &token.Token{Type: token.AND, Literal: "&&"}
}

func (s *scanner) readOr() *token.Token {
	s.next()
	s.expect('|')
	s.next()
	return &token.Token{Type: token.OR, Literal: "||"}
}

func (s *scanner) readQuoted() *token.Token {
	pos := s.pos
	s.next()
	for s.ch != '"' {
		if s.ch == '\\' {
			s.next()
		}
		s.next()
	}
	s.next()
	literal := string(s.runes[pos:s.pos])
	return &token.Token{Type: token.QUOTED, Literal: literal}
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
	literal := string(s.runes[pos:s.pos])
	return &token.Token{Type: token.NUMBER, Literal: literal}
}

func (s *scanner) readKeywordOrIdentifier() *token.Token {
	pos := s.pos
	s.next()
	for isAlpha(s.ch) || isDigit(s.ch) {
		s.next()
	}
	literal := string(s.runes[pos:s.pos])
	if ty, ok := keywords[literal]; ok {
		return &token.Token{Type: ty, Literal: literal}
	}
	return &token.Token{Type: token.IDENT, Literal: literal}
}
