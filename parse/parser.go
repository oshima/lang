package parse

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/token"
	"github.com/oshjma/lang/util"
	"strconv"
)

type parser struct {
	tokens []*token.Token // input tokens
	pos    int            // current position
	tk     *token.Token   // current token
}

func (p *parser) next() {
	p.pos += 1
	p.tk = p.tokens[p.pos]
}

func (p *parser) peekTk() *token.Token {
	return p.tokens[p.pos+1]
}

func (p *parser) lookPrec() int {
	if prec, ok := precedences[p.tk.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *parser) expect(ty string, literal string) {
	if p.tk.Type != ty {
		util.Error("Expected %s but got %s", literal, p.tk.Literal)
	}
}

func (p *parser) expectTypeName() {
	if _, ok := typeNames[p.tk.Type]; !ok {
		util.Error("Expected <type> but got %s", p.tk.Literal)
	}
}

func (p *parser) parseProgram() *ast.Program {
	stmts := make([]ast.Stmt, 0, 8)
	for p.tk.Type != token.EOF {
		stmts = append(stmts, p.parseStmt())
	}
	return &ast.Program{Stmts: stmts}
}

func (p *parser) parseStmt() ast.Stmt {
	switch p.tk.Type {
	case token.FUNC:
		return p.parseFuncDecl()
	case token.VAR:
		return p.parseVarDecl()
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.CONTINUE:
		return p.parseContinueStmt()
	case token.BREAK:
		return p.parseBreakStmt()
	case token.IDENT:
		if p.peekTk().Type == token.ASSIGN {
			return p.parseAssignStmt()
		} else {
			return p.parseExprStmt()
		}
	default:
		return p.parseExprStmt()
	}
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	p.next()
	p.expect(token.IDENT, "<identifier>")
	ident := p.parseIdent()
	p.expect(token.LPAREN, "(")
	p.next()
	params := make([]*ast.VarDecl, 0, 4)
	for p.tk.Type != token.RPAREN {
		p.expect(token.IDENT, "<identifier>")
		ident := p.parseIdent()
		p.expectTypeName()
		ty := p.tk.Literal
		p.next()
		params = append(params, &ast.VarDecl{Ident: ident, Type: ty})
		if p.tk.Type == token.COMMA {
			p.next()
		}
	}
	p.next()
	returnType := "void"
	if _, ok := typeNames[p.tk.Type]; ok {
		returnType = p.tk.Literal
		p.next()
	}
	p.expect(token.LBRACE, "{")
	body := p.parseBlockStmt()
	return &ast.FuncDecl{Ident: ident, Params: params, ReturnType: returnType, Body: body}
}

func (p *parser) parseVarDecl() *ast.VarDecl {
	p.next()
	p.expect(token.IDENT, "<identifier>")
	ident := p.parseIdent()
	var ty string
	if _, ok := typeNames[p.tk.Type]; ok {
		ty = p.tk.Literal
		p.next()
	}
	p.expect(token.ASSIGN, "=")
	p.next()
	value := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON, ";")
	p.next()
	return &ast.VarDecl{Ident: ident, Type: ty, Value: value}
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	p.next()
	stmts := make([]ast.Stmt, 0, 8)
	for p.tk.Type != token.RBRACE {
		stmts = append(stmts, p.parseStmt())
	}
	p.next()
	return &ast.BlockStmt{Stmts: stmts}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	p.next()
	cond := p.parseExpr(LOWEST)
	p.expect(token.LBRACE, "{")
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

func (p *parser) parseWhileStmt() *ast.WhileStmt {
	p.next()
	cond := p.parseExpr(LOWEST)
	p.expect(token.LBRACE, "{")
	body := p.parseBlockStmt()
	return &ast.WhileStmt{Cond: cond, Body: body}
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	p.next()
	if p.tk.Type == token.SEMICOLON {
		p.next()
		return &ast.ReturnStmt{}
	}
	value := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON, ";")
	p.next()
	return &ast.ReturnStmt{Value: value}
}

func (p *parser) parseContinueStmt() *ast.ContinueStmt {
	p.next()
	p.expect(token.SEMICOLON, ";")
	p.next()
	return &ast.ContinueStmt{}
}

func (p *parser) parseBreakStmt() *ast.BreakStmt {
	p.next()
	p.expect(token.SEMICOLON, ";")
	p.next()
	return &ast.BreakStmt{}
}

func (p *parser) parseAssignStmt() *ast.AssignStmt {
	ident := p.parseIdent()
	p.next()
	value := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON, ";")
	p.next()
	return &ast.AssignStmt{Ident: ident, Value: value}
}

func (p *parser) parseExprStmt() *ast.ExprStmt {
	expr := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON, ";")
	p.next()
	return &ast.ExprStmt{Expr: expr}
}

func (p *parser) parseExpr(prec int) ast.Expr {
	var expr ast.Expr

	switch p.tk.Type {
	case token.LPAREN:
		expr = p.parseGroupedExpr()
	case token.BANG, token.MINUS:
		expr = p.parsePrefixExpr()
	case token.IDENT:
		if p.peekTk().Type == token.LPAREN {
			expr = p.parseFuncCall()
		} else {
			expr = p.parseIdent()
		}
	case token.NUMBER:
		expr = p.parseIntLit()
	case token.TRUE, token.FALSE:
		expr = p.parseBoolLit()
	case token.QUOTED:
		expr = p.parseStringLit()
	default:
		util.Error("Unexpected %s", p.tk.Literal)
	}

	for p.lookPrec() > prec {
		expr = p.parseInfixExpr(expr)
	}

	return expr
}

func (p *parser) parseGroupedExpr() ast.Expr {
	p.next()
	expr := p.parseExpr(LOWEST)
	p.expect(token.RPAREN, ")")
	p.next()
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
	prec := p.lookPrec()
	p.next()
	right := p.parseExpr(prec)
	return &ast.InfixExpr{Operator: operator, Left: left, Right: right}
}

func (p *parser) parseFuncCall() *ast.FuncCall {
	ident := p.parseIdent()
	p.next()
	params := make([]ast.Expr, 0, 4)
	for p.tk.Type != token.RPAREN {
		params = append(params, p.parseExpr(LOWEST))
		if p.tk.Type == token.COMMA {
			p.next()
		}
	}
	p.next()
	return &ast.FuncCall{Ident: ident, Params: params}
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

func (p *parser) parseStringLit() *ast.StringLit {
	var value string
	var escaped bool
	for _, ch := range p.tk.Literal {
		if escaped {
			if ch_, ok := unescape[ch]; ok {
				value += string(ch_)
				escaped = false
			} else {
				util.Error("Unknown escape sequence \\%c", ch)
			}
		} else {
			if ch == '\\' {
				escaped = true
			} else if ch != '"' {
				value += string(ch)
			}
		}
	}
	p.next()
	return &ast.StringLit{Value: value}
}
