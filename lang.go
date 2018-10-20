package main

import (
	"fmt"
	"os"
	"flag"
	"strconv"
	"io/ioutil"
	"github.com/k0kubun/pp"
)

// Tokens

const (
	TK_MINUS = "TK_MINUS"
	TK_SEMICOLON = "TK_SEMICOLON"
	TK_EOF = "TK_EOF"
	TK_INT = "TK_INT"
)

type Token struct {
	Type string
	Source string
}

// Lexer

type Lexer struct {
	source string
	pos int
}

func (l *Lexer) lookChar() byte {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) peekChar() byte {
	if l.pos + 1 >= len(l.source) {
		return 0
	}
	return l.source[l.pos + 1]
}

func (l *Lexer) next() {
	l.pos += 1
}

func (l *Lexer) skipWs() {
	c := l.lookChar()
	for c == ' ' || c == '\t' || c == '\n' || c == '\r' {
		l.next()
		c = l.lookChar()
	}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	var token Token
	for {
		l.skipWs()
		token = l.readToken()
		tokens = append(tokens, token)
		if (token.Type == TK_EOF) {
			break
		}
	}
	return tokens
}

func (l *Lexer) readToken() Token {
	var token Token
	c := l.lookChar()
	switch c {
	case '-':
		if isDigit(l.peekChar()) {
			token = l.readInteger()
		} else {
			token = Token{Type: TK_MINUS, Source: "-"}
			l.next()
		}
	case ';':
		token = Token{Type: TK_SEMICOLON, Source: ";"}
		l.next()
	case 0:
		token = Token{Type: TK_EOF, Source: ""}
		l.next()
	default:
		if isDigit(c) {
			token = l.readInteger()
		} else {
			error("Unexpected %q", string(c))
		}
	}
	return token
}

func (l *Lexer) readInteger() Token {
	pos := l.pos
	if l.lookChar() == '-' {
		l.next()
	}
	l.next()
	for isDigit(l.lookChar()) {
		l.next()
	}
	return Token{Type: TK_INT, Source: l.source[pos:l.pos]}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// AST

type Node interface {
	Node()
}

type Statement interface {
	Node
	Statement()
}

type Expression interface {
	Node
	Expression()
}

type Program struct {
	Statements []Statement
}
func (node *Program) Node() {}

type ExpressionStatement struct {
	Expression Expression
}
func (stmt *ExpressionStatement) Node() {}
func (stmt *ExpressionStatement) Statement() {}

type IntegerLiteral struct {
	Value int64
}
func (expr *IntegerLiteral) Node() {}
func (expr *IntegerLiteral) Expression() {}

// Parser

type Parser struct {
	tokens []Token
	pos int
}

func (p *Parser) lookToken() Token {
	return p.tokens[p.pos]
}

func (p *Parser) next() {
	p.pos += 1
}

func (p *Parser) ParseProgram() *Program {
	var statements []Statement
	var stmt Statement
	for p.lookToken().Type != TK_EOF {
		stmt = p.parseStatement()
		statements = append(statements, stmt)
	}
	return &Program{Statements: statements}
}

func (p *Parser) parseStatement() Statement {
	return p.parseExpressionStatement()
}

func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	expr := p.parseExpression()
	token := p.lookToken()
	if token.Type != TK_SEMICOLON {
		error("Expected %q but got %q", ";", token.Source)
	}
	p.next()
	return &ExpressionStatement{Expression: expr}
}

func (p *Parser) parseExpression() Expression {
	var expr Expression
	token := p.lookToken()
	switch token.Type {
	case TK_INT:
		expr = p.parseIntegerLiteral()
	default:
		error("Unexpected %q", token.Source)
	}
	return expr
}

func (p *Parser) parseIntegerLiteral() *IntegerLiteral {
	token := p.lookToken()
	value, err := strconv.ParseInt(token.Source, 0, 64)
	if err != nil {
		error("Could not parse %q as integer", token.Source)
	}
	p.next()
	return &IntegerLiteral{Value: value}
}

// Generator

type Generator struct {
	program *Program
}

func (g *Generator) EmitProgram() {
	emit(".text")
	emit(".globl main")
	emit(".type main, @function")
	p("main:")
	emit("pushq %%rbp")
	emit("movq %%rsp, %%rbp")
	for _, stmt := range g.program.Statements {
		g.emitStatement(stmt)
	}
	emit("leave")
	emit("ret")
}

func (g *Generator) emitStatement(stmt Statement) {
	switch v := stmt.(type) {
	case *ExpressionStatement:
		g.emitExpression(v.Expression)
	}
}

func (g *Generator) emitExpression(expr Expression) {
	switch v := expr.(type) {
	case *IntegerLiteral:
		g.emitIntegerLiteral(v)
	}
}

func (g *Generator) emitIntegerLiteral(expr *IntegerLiteral) {
	emit("movq $%d, %%rax", expr.Value)
}

// Utils

func p(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

func emit(format string, a ...interface{}) {
	fmt.Print("\t")
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

func error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}

// Main

func main() {
	debug := flag.Bool("debug", false, "print tokens and AST for debug")
	flag.Parse()

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		error("Failed to read source code from stdin")
	}

	lexer := &Lexer{source: string(bytes)}
	tokens := lexer.Tokenize()
	if *debug {
		pp.Fprintln(os.Stderr, tokens)
	}

	parser := &Parser{tokens: tokens}
	program := parser.ParseProgram()
	if *debug {
		pp.Fprintln(os.Stderr, program)
	}

	generator := &Generator{program: program}
	generator.EmitProgram()
}
