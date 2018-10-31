package parse

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/token"
	"github.com/oshjma/lang/util"
	"strconv"
)

func Parse(tokens []*token.Token) *ast.Program {
	p := &parser{tokens: tokens, pos: -1}
	p.next()
	return p.parseProgram()
}

type parser struct {
	tokens []*token.Token // input tokens
	pos    int            // current position
	tk     *token.Token   // current token
}

func (p *parser) next() {
	p.pos += 1
	p.tk = p.tokens[p.pos]
}

func (p *parser) expect(ty string, literal string) {
	if p.tk.Type != ty {
		util.Error("Expected %s but got %s", literal, p.tk.Literal)
	}
	p.next()
}

func (p *parser) peekToken() *token.Token {
	return p.tokens[p.pos+1]
}

func (p *parser) lookPrecedence() int {
	if pr, ok := precedences[p.tk.Type]; ok {
		return pr
	}
	return LOWEST
}

func (p *parser) parseProgram() *ast.Program {
	var list []ast.Stmt
	for p.tk.Type != token.EOF {
		list = append(list, p.parseStmt())
	}
	return &ast.Program{List: list}
}

func (p *parser) parseStmt() ast.Stmt {
	switch p.tk.Type {
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.LET:
		return p.parseLetStmt()
	case token.IDENT:
		return p.parseAssignStmtOrExprStmt()
	default:
		return p.parseExprStmt()
	}
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	p.next()
	var list []ast.Stmt
	for p.tk.Type != token.RBRACE {
		list = append(list, p.parseStmt())
	}
	p.next()
	return &ast.BlockStmt{List: list}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	p.next()
	cond := p.parseExpr(LOWEST)
	if p.tk.Type != token.LBRACE {
		util.Error("Expected { but got %s", p.tk.Literal)
	}
	conseq := p.parseBlockStmt()
	if p.tk.Type != token.ELSE {
		return &ast.IfStmt{Cond: cond, Conseq: conseq}
	}
	p.next()
	var altern ast.Stmt
	switch p.tk.Type {
	case token.LBRACE:
		altern = p.parseBlockStmt()
	case token.IF:
		altern = p.parseIfStmt()
	default:
		util.Error("Expected { or if but got %s", p.tk.Literal)
	}
	return &ast.IfStmt{Cond: cond, Conseq: conseq, Altern: altern}
}

func (p *parser) parseLetStmt() *ast.LetStmt {
	p.next()
	if p.tk.Type != token.IDENT {
		util.Error("Expected <identifier> but got %s", p.tk.Literal)
	}
	ident := p.parseIdent()
	if p.tk.Type != token.INT && p.tk.Type != token.BOOL {
		util.Error("Expected <type> but got %s", p.tk.Literal)
	}
	ty := p.tk.Literal
	p.next()
	p.expect(token.ASSIGN, "=")
	expr := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON, ";")
	return &ast.LetStmt{Ident: ident, Type: ty, Expr: expr}
}

func (p *parser) parseAssignStmtOrExprStmt() ast.Stmt {
	if p.peekToken().Type == token.ASSIGN {
		return p.parseAssignStmt()
	}
	return p.parseExprStmt()
}

func (p *parser) parseAssignStmt() *ast.AssignStmt {
	ident := p.parseIdent()
	p.next()
	expr := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON, ";")
	return &ast.AssignStmt{Ident: ident, Expr: expr}
}

func (p *parser) parseExprStmt() *ast.ExprStmt {
	expr := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON, ";")
	return &ast.ExprStmt{Expr: expr}
}

func (p *parser) parseExpr(precedence int) ast.Expr {
	var expr ast.Expr
	switch p.tk.Type {
	case token.LPAREN:
		expr = p.parseGroupedExpr()
	case token.BANG, token.MINUS:
		expr = p.parsePrefixExpr()
	case token.IDENT:
		expr = p.parseIdent()
	case token.NUMBER:
		expr = p.parseIntLit()
	case token.TRUE, token.FALSE:
		expr = p.parseBoolLit()
	default:
		util.Error("Unexpected %s", p.tk.Literal)
	}
	for p.lookPrecedence() > precedence {
		expr = p.parseInfixExpr(expr)
	}
	return expr
}

func (p *parser) parseGroupedExpr() ast.Expr {
	p.next()
	expr := p.parseExpr(LOWEST)
	p.expect(token.RPAREN, ")")
	return expr
}

func (p *parser) parsePrefixExpr() *ast.PrefixExpr {
	operator := p.tk.Literal
	p.next()
	right := p.parseExpr(PREFIX)
	return &ast.PrefixExpr{Operator: operator, Right: right}
}

func (p *parser) parseInfixExpr(left ast.Expr) *ast.InfixExpr {
	operator := p.tk.Literal
	precedence := p.lookPrecedence()
	p.next()
	right := p.parseExpr(precedence)
	return &ast.InfixExpr{Operator: operator, Left: left, Right: right}
}

func (p *parser) parseIdent() *ast.Ident {
	name := p.tk.Literal
	p.next()
	return &ast.Ident{Name: name}
}

func (p *parser) parseIntLit() *ast.IntLit {
	value, err := strconv.ParseInt(p.tk.Literal, 0, 64)
	if err != nil {
		util.Error("Could not parse %s as integer", p.tk.Literal)
	}
	p.next()
	return &ast.IntLit{Value: value}
}

func (p *parser) parseBoolLit() *ast.BoolLit {
	value := p.tk.Type == token.TRUE
	p.next()
	return &ast.BoolLit{Value: value}
}
