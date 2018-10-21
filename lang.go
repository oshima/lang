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
	TK_PLUS = "TK_PLUS"
	TK_MINUS = "TK_MINUS"
	TK_SEMICOLON = "TK_SEMICOLON"
	TK_EOF = "TK_EOF"
	TK_INT = "TK_INT"
)

type Token struct {
	Type string
	Source string
}

// Tokenize

func Tokenize(src string) []*Token {
	l := &Lexer{src: src, pos: -1}
	l.next()

	var tokens []*Token
	var tk *Token
	for {
		l.skipWs()
		tk = l.readToken()
		tokens = append(tokens, tk)
		if (tk.Type == TK_EOF) {
			break
		}
	}
	return tokens
}

type Lexer struct {
	src string // input source code
	pos int    // current position
	ch byte    // current character
}

func (l *Lexer) next() {
	l.pos += 1
	if l.pos < len(l.src) {
		l.ch = l.src[l.pos]
	} else {
		l.ch = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.pos + 1 < len(l.src) {
		return l.src[l.pos + 1]
	}
	return 0
}

func (l *Lexer) skipWs() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.next()
	}
}

func (l *Lexer) readToken() *Token {
	var tk *Token
	switch l.ch {
	case '+':
		tk = l.readPunct(TK_PLUS)
	case '-':
		if isDigit(l.peekChar()) {
			tk = l.readInt()
		} else {
			tk = l.readPunct(TK_MINUS)
		}
	case ';':
		tk = l.readPunct(TK_SEMICOLON)
	case 0:
		tk = l.readPunct(TK_EOF)
	default:
		if isDigit(l.ch) {
			tk = l.readInt()
		} else {
			error("Unexpected %q", string(l.ch))
		}
	}
	return tk
}

func (l *Lexer) readPunct(ty string) *Token {
	tk := &Token{Type: ty, Source: string(l.ch)}
	l.next()
	return tk
}

func (l *Lexer) readInt() *Token {
	pos := l.pos
	if l.ch == '-' {
		l.next()
	}
	l.next()
	for isDigit(l.ch) {
		l.next()
	}
	return &Token{Type: TK_INT, Source: l.src[pos:l.pos]}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// AST

type Node interface {
	AstNode()
}

type Stmt interface {
	Node
	StmtNode()
}

type Expr interface {
	Node
	ExprNode()
}

type Program struct {
	Statements []Stmt
}
func (node *Program) AstNode() {}

type ExprStmt struct {
	Expr Expr
}
func (stmt *ExprStmt) AstNode() {}
func (stmt *ExprStmt) StmtNode() {}

type InfixExpr struct {
	Operator string
	Left Expr
	Right Expr
}
func (expr *InfixExpr) AstNode() {}
func (expr *InfixExpr) ExprNode() {}

type IntLit struct {
	Value int64
}
func (expr *IntLit) AstNode() {}
func (expr *IntLit) ExprNode() {}

// Parse

func Parse(tokens []*Token) *Program {
	p := &Parser{tokens: tokens, pos: -1}
	p.next()

	var statements []Stmt
	var stmt Stmt
	for p.tk.Type != TK_EOF {
		stmt = p.parseStmt()
		statements = append(statements, stmt)
	}
	return &Program{Statements: statements}
}

const (
	PR_LOWEST int = iota
	PR_SUM
)

var precedences = map[string]int{
	TK_PLUS: PR_SUM,
	TK_MINUS: PR_SUM,
}

type Parser struct {
	tokens []*Token // input tokens
	pos int         // current position
	tk *Token       // current token
}

func (p *Parser) next() {
	p.pos += 1
	p.tk = p.tokens[p.pos]
}

func (p *Parser) lookPrecedence() int {
	if pr, ok := precedences[p.tk.Type]; ok {
		return pr
	}
	return PR_LOWEST
}

func (p *Parser) parseStmt() Stmt {
	return p.parseExprStmt()
}

func (p *Parser) parseExprStmt() *ExprStmt {
	expr := p.parseExpr(PR_LOWEST)
	if p.tk.Type != TK_SEMICOLON {
		if p.tk.Type == TK_EOF {
			error("Expected %q but got <EOF>", ";")
		} else {
			error("Expected %q but got %q", ";", p.tk.Source)
		}
	}
	p.next()
	return &ExprStmt{Expr: expr}
}

func (p *Parser) parseExpr(precedence int) Expr {
	var left Expr
	switch p.tk.Type {
	case TK_INT:
		left = p.parseIntLit()
	case TK_EOF:
		error("Unexpected <EOF>")
	default:
		error("Unexpected %q", p.tk.Source)
	}

	for precedence < p.lookPrecedence() {
		switch p.tk.Type {
		case TK_PLUS:
			left = p.parseInfixExpr(left)
		case TK_MINUS:
			left = p.parseInfixExpr(left)
		}
	}
	return left
}

func (p *Parser) parseInfixExpr(left Expr) *InfixExpr {
	operator := p.tk.Source
	precedence := p.lookPrecedence()
	p.next()
	right := p.parseExpr(precedence)
	return &InfixExpr{Operator: operator, Left: left, Right: right}
}

func (p *Parser) parseIntLit() *IntLit {
	value, err := strconv.ParseInt(p.tk.Source, 0, 64)
	if err != nil {
		error("Could not parse %q as integer", p.tk.Source)
	}
	p.next()
	return &IntLit{Value: value}
}

// Generate

func Generate(program *Program) {
	emit(".text")
	emit(".globl main")
	emit(".type main, @function")
	p("main:")
	emit("pushq %%rbp")
	emit("movq %%rsp, %%rbp")
	for _, stmt := range program.Statements {
		emitStmt(stmt)
	}
	emit("leave")
	emit("ret")
}

func emitStmt(stmt Stmt) {
	switch v := stmt.(type) {
	case *ExprStmt:
		emitExpr(v.Expr)
	}
}

func emitExpr(expr Expr) {
	switch v := expr.(type) {
	case *InfixExpr:
		emitInfixExpr(v)
	case *IntLit:
		emitIntLit(v)
	}
}

func emitInfixExpr(expr *InfixExpr) {
	emitExpr(expr.Right)
	emit("pushq %%rax")
	emitExpr(expr.Left)
	emit("popq %%rdx")
	switch expr.Operator {
	case "+":
		emit("addq %%rdx, %%rax")
	case "-":
		emit("subq %%rdx, %%rax")
	}
}

func emitIntLit(expr *IntLit) {
	emit("movq $%d, %%rax", expr.Value)
}

func p(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

func emit(format string, a ...interface{}) {
	fmt.Print("\t")
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// Utils

func error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}

// Main

func main() {
	debug := flag.Bool("d", false, "print tokens and AST for debug")
	flag.Parse()

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		error("Failed to read source code from stdin")
	}

	tokens := Tokenize(string(bytes))
	if *debug {
		pp.Fprintln(os.Stderr, tokens)
	}

	program := Parse(tokens)
	if *debug {
		pp.Fprintln(os.Stderr, program)
	}

	if !*debug {
		Generate(program)
	}
}
