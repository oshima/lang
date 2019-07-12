package scan

import (
	"fmt"
	"os"

	"github.com/oshima/lang/token"
)

type scanner struct {
	runes   []rune       // source code
	idx     int          // current index
	ch      rune         // current character (runes[idx])
	line    int          // current line
	col     int          // current column
	lastTok *token.Token // last token scanner has read
}

func (s *scanner) next() {
	if s.ch == '\n' {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	s.idx++
	if s.idx < len(s.runes) {
		s.ch = s.runes[s.idx]
	} else {
		s.ch = 0
	}
}

func (s *scanner) peek() rune {
	if s.idx+1 < len(s.runes) {
		return s.runes[s.idx+1]
	}
	return 0
}

func (s *scanner) consume(ch rune) {
	switch s.ch {
	case ch:
		s.next()
	case '\n':
		s.error("unexpected newline")
	case 0:
		s.error("unexpected eof")
	default:
		s.error("unexpected %c", s.ch)
	}
}

func (s *scanner) skipWs() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *scanner) pos() *token.Pos {
	return &token.Pos{Line: s.line, Col: s.col}
}

func (s *scanner) error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s.pos().String()+": "+format+"\n", a...)
	os.Exit(1)
}

func (s *scanner) readTokens() []*token.Token {
	tokens := make([]*token.Token, 0, 64)
	s.skipWs()
	for s.ch != 0 {
		pos := s.pos()
		tok := s.readToken()
		tok.Pos = pos
		if tok.Type != token.COMMENT {
			tokens = append(tokens, tok)
		}
		s.lastTok = tok
		s.skipWs()
	}
	eof := &token.Token{Type: token.EOF, Pos: s.pos()}
	return append(tokens, eof)
}

func (s *scanner) readToken() *token.Token {
	switch s.ch {
	case '#':
		return s.readComment()
	case '(', ')', '[', ']', '{', '}', ',', ':', ';':
		return s.readPunct()
	case '=':
		return s.readAssignOrEqual()
	case '!':
		return s.readBangOrNotEqual()
	case '+':
		return s.readPlusOrAddAssign()
	case '-':
		return s.readMinusOrSubAssignOrArrowOrNumber()
	case '*':
		return s.readAsteriskOrMulAssign()
	case '/':
		return s.readSlashOrDivAssign()
	case '%':
		return s.readPercentOrModAssign()
	case '<':
		return s.readLessOrLessEqual()
	case '>':
		return s.readGreaterOrGreaterEqual()
	case '&':
		return s.readAnd()
	case '|':
		return s.readOr()
	case '.':
		return s.readBetween()
	case '"':
		return s.readQuoted()
	default:
		switch {
		case isDigit(s.ch):
			return s.readNumber()
		case isAlpha(s.ch):
			return s.readKeywordOrIdentifier()
		default:
			s.error("invalid character %c", s.ch)
			return nil // unreachable
		}
	}
}

func (s *scanner) readComment() *token.Token {
	idx := s.idx
	s.next()
	for s.ch != '\n' && s.ch != 0 {
		s.next()
	}
	literal := string(s.runes[idx:s.idx])
	if s.ch == '\n' {
		s.next()
	}
	return &token.Token{Type: token.COMMENT, Literal: literal}
}

func (s *scanner) readPunct() *token.Token {
	typ := punctuations[s.ch]
	literal := string(s.ch)
	s.next()
	return &token.Token{Type: typ, Literal: literal}
}

func (s *scanner) readAssignOrEqual() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.EQ, Literal: "=="}
	}
	return &token.Token{Type: token.ASSIGN, Literal: "="}
}

func (s *scanner) readBangOrNotEqual() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.NE, Literal: "!="}
	}
	return &token.Token{Type: token.BANG, Literal: "!"}
}

func (s *scanner) readPlusOrAddAssign() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.ADDASSIGN, Literal: "+="}
	}
	return &token.Token{Type: token.PLUS, Literal: "+"}
}

func (s *scanner) readMinusOrSubAssignOrArrowOrNumber() *token.Token {
	nextCh := s.peek()
	if nextCh == '=' {
		s.next()
		s.next()
		return &token.Token{Type: token.SUBASSIGN, Literal: "-="}
	}
	if nextCh == '>' {
		s.next()
		s.next()
		return &token.Token{Type: token.ARROW, Literal: "->"}
	}
	if isDigit(nextCh) {
		if s.lastTok == nil {
			return s.readNumber()
		}
		if _, ok := exprEnd[s.lastTok.Type]; ok {
			s.next()
			return &token.Token{Type: token.MINUS, Literal: "-"}
		}
		return s.readNumber()
	}
	s.next()
	return &token.Token{Type: token.MINUS, Literal: "-"}
}

func (s *scanner) readAsteriskOrMulAssign() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.MULASSIGN, Literal: "*="}
	}
	return &token.Token{Type: token.ASTERISK, Literal: "*"}
}

func (s *scanner) readSlashOrDivAssign() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.DIVASSIGN, Literal: "/="}
	}
	return &token.Token{Type: token.SLASH, Literal: "/"}
}

func (s *scanner) readPercentOrModAssign() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.MODASSIGN, Literal: "%="}
	}
	return &token.Token{Type: token.PERCENT, Literal: "%"}
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
	s.consume('&')
	return &token.Token{Type: token.AND, Literal: "&&"}
}

func (s *scanner) readOr() *token.Token {
	s.next()
	s.consume('|')
	return &token.Token{Type: token.OR, Literal: "||"}
}

func (s *scanner) readBetween() *token.Token {
	s.next()
	s.consume('.')
	return &token.Token{Type: token.BETWEEN, Literal: ".."}
}

func (s *scanner) readQuoted() *token.Token {
	idx := s.idx
	s.next()
	for s.ch != '"' {
		if s.ch == '\\' {
			s.next()
		}
		if s.ch == 0 {
			s.error("unexpected eof")
		}
		s.next()
	}
	s.next()
	literal := string(s.runes[idx:s.idx])
	return &token.Token{Type: token.QUOTED, Literal: literal}
}

func (s *scanner) readNumber() *token.Token {
	idx := s.idx
	if s.ch == '-' {
		s.next()
	}
	s.next()
	for isDigit(s.ch) {
		s.next()
	}
	literal := string(s.runes[idx:s.idx])
	return &token.Token{Type: token.NUMBER, Literal: literal}
}

func (s *scanner) readKeywordOrIdentifier() *token.Token {
	idx := s.idx
	s.next()
	for isAlpha(s.ch) || isDigit(s.ch) {
		s.next()
	}
	literal := string(s.runes[idx:s.idx])
	if typ, ok := keywords[literal]; ok {
		return &token.Token{Type: typ, Literal: literal}
	}
	return &token.Token{Type: token.IDENT, Literal: literal}
}
