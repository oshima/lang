package parse

import (
	"github.com/oshjma/lang/ast"
	"github.com/oshjma/lang/token"
	"github.com/oshjma/lang/types"
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

func (p *parser) expect(ty token.Type) {
	if p.tk.Type != ty {
		util.Error("Expected %s but got %s", ty, p.tk.Type)
	}
}

func (p *parser) parseType() types.Type {
	var ty types.Type

	switch p.tk.Type {
	case token.INT:
		p.next()
		ty = &types.Int{}
	case token.BOOL:
		p.next()
		ty = &types.Bool{}
	case token.STRING:
		p.next()
		ty = &types.String{}
	case token.LBRACK:
		ty = p.parseArray()
	case token.LPAREN:
		ty = p.parseFunc()
	default:
		util.Error("Unexpected %s", p.tk.Type)
	}

	return ty
}

func (p *parser) parseArray() *types.Array {
	p.next()
	p.expect(token.NUMBER)
	len, err := strconv.Atoi(p.tk.Literal)
	if err != nil {
		util.Error("Could not parse %s as integer", p.tk.Literal)
	}
	if len < 0 {
		util.Error("Array's length must be non-negative")
	}
	p.next()
	p.expect(token.RBRACK)
	p.next()
	elemType := p.parseType()
	return &types.Array{Len: len, ElemType: elemType}
}

func (p *parser) parseFunc() *types.Func {
	p.next()
	paramTypes := make([]types.Type, 0, 4)
	for p.tk.Type != token.RPAREN {
		paramTypes = append(paramTypes, p.parseType())
		switch p.tk.Type {
		case token.COMMA:
			p.next()
		case token.RPAREN:
			// ok
		default:
			util.Error("Expected , or ) but got %s", p.tk.Type)
		}
	}
	p.next()
	p.expect(token.ARROW)
	p.next()
	var returnType types.Type
	if p.tk.Type == token.BANG {
		p.next()
	} else {
		returnType = p.parseType()
	}
	return &types.Func{ParamTypes: paramTypes, ReturnType: returnType}
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
	case token.LET:
		return p.parseLetStmt()
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.FOR:
		return p.parseForStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.CONTINUE:
		return p.parseContinueStmt()
	case token.BREAK:
		return p.parseBreakStmt()
	default:
		expr := p.parseExpr(LOWEST)
		if p.tk.Type == token.ASSIGN {
			return p.parseAssignStmt(expr)
		} else {
			return p.parseExprStmt(expr)
		}
	}
}

func (p *parser) parseLetStmt() *ast.LetStmt {
	p.next()
	p.expect(token.IDENT)
	ident := p.parseIdent()
	var varType types.Type
	if p.tk.Type == token.COLON {
		p.next()
		varType = p.parseType()
	}
	p.expect(token.ASSIGN)
	p.next()
	value := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON)
	p.next()
	return &ast.LetStmt{Ident: ident, VarType: varType, Value: value}
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
	p.expect(token.LBRACE)
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
		util.Error("Expected { or if but got %s", p.tk.Type)
	}
	return &ast.IfStmt{Cond: cond, Conseq: conseq, Altern: altern}
}

func (p *parser) parseForStmt() *ast.ForStmt {
	p.next()
	cond := p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	body := p.parseBlockStmt()
	return &ast.ForStmt{Cond: cond, Body: body}
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	p.next()
	if p.tk.Type == token.SEMICOLON {
		p.next()
		return &ast.ReturnStmt{}
	}
	value := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON)
	p.next()
	return &ast.ReturnStmt{Value: value}
}

func (p *parser) parseContinueStmt() *ast.ContinueStmt {
	p.next()
	p.expect(token.SEMICOLON)
	p.next()
	return &ast.ContinueStmt{}
}

func (p *parser) parseBreakStmt() *ast.BreakStmt {
	p.next()
	p.expect(token.SEMICOLON)
	p.next()
	return &ast.BreakStmt{}
}

func (p *parser) parseAssignStmt(target ast.Expr) *ast.AssignStmt {
	switch target.(type) {
	case *ast.Ident, *ast.IndexExpr:
		// ok
	default:
		util.Error("Invalid target of assignment")
	}
	p.next()
	value := p.parseExpr(LOWEST)
	p.expect(token.SEMICOLON)
	p.next()
	return &ast.AssignStmt{Target: target, Value: value}
}

func (p *parser) parseExprStmt(expr ast.Expr) *ast.ExprStmt {
	p.expect(token.SEMICOLON)
	p.next()
	return &ast.ExprStmt{Expr: expr}
}

func (p *parser) parseExpr(prec int) ast.Expr {
	var expr ast.Expr

	switch p.tk.Type {
	case token.BANG, token.MINUS:
		expr = p.parsePrefixExpr()
	case token.IDENT:
		expr = p.parseIdent()
	case token.NUMBER:
		expr = p.parseIntLit()
	case token.TRUE, token.FALSE:
		expr = p.parseBoolLit()
	case token.QUOTED:
		expr = p.parseStringLit()
	case token.LBRACK:
		expr = p.parseArrayLit()
	case token.LPAREN:
		expr = p.parseFuncLitOrGroupedExpr()
	default:
		util.Error("Unexpected %s", p.tk.Type)
	}

	for p.lookPrec() > prec {
		switch p.tk.Type {
		case token.LBRACK:
			expr = p.parseIndexExpr(expr)
		case token.LPAREN:
			if v, ok := expr.(*ast.Ident); ok {
				if _, ok := libfuncs[v.Name]; ok {
					expr = p.parseLibcallExpr(v)
					continue
				}
			}
			expr = p.parseCallExpr(expr)
		default:
			expr = p.parseInfixExpr(expr)
		}
	}

	return expr
}

func (p *parser) parsePrefixExpr() *ast.PrefixExpr {
	op := p.tk.Literal
	p.next()
	right := p.parseExpr(PREFIX)
	return &ast.PrefixExpr{Op: op, Right: right}
}

func (p *parser) parseInfixExpr(left ast.Expr) *ast.InfixExpr {
	op := p.tk.Literal
	prec := p.lookPrec()
	p.next()
	right := p.parseExpr(prec)
	return &ast.InfixExpr{Left: left, Op: op, Right: right}
}

func (p *parser) parseIndexExpr(left ast.Expr) *ast.IndexExpr {
	p.next()
	index := p.parseExpr(LOWEST)
	p.expect(token.RBRACK)
	p.next()
	return &ast.IndexExpr{Left: left, Index: index}
}

func (p *parser) parseCallExpr(left ast.Expr) *ast.CallExpr {
	p.next()
	params := make([]ast.Expr, 0, 4)
	for p.tk.Type != token.RPAREN {
		params = append(params, p.parseExpr(LOWEST))
		switch p.tk.Type {
		case token.COMMA:
			p.next()
		case token.RPAREN:
			// ok
		default:
			util.Error("Expected , or ) but got %s", p.tk.Type)
		}
	}
	p.next()
	return &ast.CallExpr{Left: left, Params: params}
}

func (p *parser) parseLibcallExpr(ident *ast.Ident) *ast.LibcallExpr {
	p.next()
	params := make([]ast.Expr, 0, 4)
	for p.tk.Type != token.RPAREN {
		params = append(params, p.parseExpr(LOWEST))
		switch p.tk.Type {
		case token.COMMA:
			p.next()
		case token.RPAREN:
			// ok
		default:
			util.Error("Expected , or ) but got %s", p.tk.Type)
		}
	}
	p.next()
	return &ast.LibcallExpr{Ident: ident, Params: params}
}

func (p *parser) parseIdent() *ast.Ident {
	name := p.tk.Literal
	p.next()
	return &ast.Ident{Name: name}
}

func (p *parser) parseIntLit() *ast.IntLit {
	value, err := strconv.Atoi(p.tk.Literal)
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

func (p *parser) parseArrayLit() *ast.ArrayLit {
	p.next()
	p.expect(token.NUMBER)
	len, err := strconv.Atoi(p.tk.Literal)
	if err != nil {
		util.Error("Could not parse %s as integer", p.tk.Literal)
	}
	if len < 0 {
		util.Error("Array's length must be non-negative")
	}
	p.next()
	p.expect(token.RBRACK)
	p.next()
	elemType := p.parseType()
	p.expect(token.LBRACE)
	p.next()
	elems := make([]ast.Expr, 0, 8)
	for p.tk.Type != token.RBRACE {
		elems = append(elems, p.parseExpr(LOWEST))
		switch p.tk.Type {
		case token.COMMA:
			p.next()
		case token.RBRACE:
		default:
			util.Error("Expected , or } but got %s", p.tk.Type)
		}
	}
	p.next()
	return &ast.ArrayLit{Len: len, ElemType: elemType, Elems: elems}
}

func (p *parser) parseFuncLitOrGroupedExpr() ast.Expr {
	p.next()
	// PATTERN: () ...
	// must be a FuncLit with no parameters
	if p.tk.Type == token.RPAREN {
		p.next()
		p.expect(token.ARROW)
		p.next()
		var returnType types.Type
		switch p.tk.Type {
		case token.LBRACE:
			// ok
		case token.BANG:
			p.next()
		default:
			returnType = p.parseType()
		}
		p.expect(token.LBRACE)
		body := p.parseBlockStmt()
		return &ast.FuncLit{ReturnType: returnType, Body: body}
	}
	// PATTERN: (ident: ...
	// must be a FuncLit with parameters
	if p.tk.Type == token.IDENT && p.peekTk().Type == token.COLON {
		params := make([]*ast.LetStmt, 0, 4)
		for p.tk.Type != token.RPAREN {
			p.expect(token.IDENT)
			ident := p.parseIdent()
			p.expect(token.COLON)
			p.next()
			varType := p.parseType()
			params = append(params, &ast.LetStmt{Ident: ident, VarType: varType})
			switch p.tk.Type {
			case token.COMMA:
				p.next()
			case token.RPAREN:
				// ok
			default:
				util.Error("Expected , or ) but got %s", p.tk.Type)
			}
		}
		p.next()
		p.expect(token.ARROW)
		p.next()
		var returnType types.Type
		switch p.tk.Type {
		case token.LBRACE:
			// ok
		case token.BANG:
			p.next()
		default:
			returnType = p.parseType()
		}
		p.expect(token.LBRACE)
		body := p.parseBlockStmt()
		return &ast.FuncLit{Params: params, ReturnType: returnType, Body: body}
	}
	// PATTERN: otherwise
	// must be a grouped expr
	expr := p.parseExpr(LOWEST)
	p.expect(token.RPAREN)
	p.next()
	return expr
}
