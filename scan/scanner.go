package scan

import (
	"github.com/oshima/lang/token"
	"github.com/oshima/lang/util"
)

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

func (s *scanner) skipWs() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *scanner) consume(ch rune) {
	if s.ch != ch {
		util.Error("Expected %c but got %c", ch, s.ch)
	}
	s.next()
}

func (s *scanner) readTokens() []*token.Token {
	tokens := make([]*token.Token, 0, 64)
	s.skipWs()
	for s.ch != 0 {
		tokens = append(tokens, s.readToken())
		s.skipWs()
	}
	return append(tokens, &token.Token{Type: token.EOF})
}

func (s *scanner) readToken() *token.Token {
	var tk *token.Token
	switch s.ch {
	case '(', ')', '[', ']', '{', '}', ',', ':', ';':
		tk = s.readPunct()
	case '=':
		tk = s.readAssignOrEqual()
	case '!':
		tk = s.readBangOrNotEqual()
	case '+':
		tk = s.readPlusOrAddAssign()
	case '-':
		tk = s.readMinusOrSubAssignOrArrowOrNumber()
	case '*':
		tk = s.readAsteriskOrMulAssign()
	case '/':
		tk = s.readSlashOrDivAssign()
	case '%':
		tk = s.readPercentOrModAssign()
	case '<':
		tk = s.readLessOrLessEqual()
	case '>':
		tk = s.readGreaterOrGreaterEqual()
	case '&':
		tk = s.readAnd()
	case '|':
		tk = s.readOr()
	case '.':
		tk = s.readBetween()
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
		return &token.Token{Type: token.ADD_ASSIGN, Literal: "+="}
	}
	return &token.Token{Type: token.PLUS, Literal: "+"}
}

func (s *scanner) readMinusOrSubAssignOrArrowOrNumber() *token.Token {
	nextCh := s.peekCh()
	if nextCh == '=' {
		s.next()
		s.next()
		return &token.Token{Type: token.SUB_ASSIGN, Literal: "-="}
	}
	if nextCh == '>' {
		s.next()
		s.next()
		return &token.Token{Type: token.ARROW, Literal: "->"}
	}
	if isDigit(nextCh) {
		if s.lastTk == nil {
			return s.readNumber()
		}
		if _, ok := exprEnd[s.lastTk.Type]; ok {
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
		return &token.Token{Type: token.MUL_ASSIGN, Literal: "*="}
	}
	return &token.Token{Type: token.ASTERISK, Literal: "*"}
}

func (s *scanner) readSlashOrDivAssign() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.DIV_ASSIGN, Literal: "/="}
	}
	return &token.Token{Type: token.SLASH, Literal: "/"}
}

func (s *scanner) readPercentOrModAssign() *token.Token {
	s.next()
	if s.ch == '=' {
		s.next()
		return &token.Token{Type: token.MOD_ASSIGN, Literal: "%="}
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
	pos := s.pos
	s.next()
	for s.ch != '"' {
		if s.ch == '\\' {
			s.next()
		}
		if s.ch == 0 {
			util.Error("Unexpected EOF")
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
