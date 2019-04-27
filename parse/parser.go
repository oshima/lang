package parse

import (
	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/token"
	"github.com/oshima/lang/types"
	"github.com/oshima/lang/util"
	"strconv"
)

type parser struct {
	tokens []*token.Token // input tokens
	pos    int            // current position
	tk     *token.Token   // current token
}

func (p *parser) next() {
	p.pos++
	p.tk = p.tokens[p.pos]
}

func (p *parser) peekTk() *token.Token {
	if p.tk.Type == token.EOF {
		util.Error("Unexpected EOF")
	}
	return p.tokens[p.pos+1]
}

func (p *parser) lookPrec() int {
	if prec, ok := precOf[p.tk.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *parser) expect(ty token.Type) {
	if p.tk.Type != ty {
		util.Error("Expected %s but got %s", ty, p.tk.Type)
	}
}

func (p *parser) consume(ty token.Type) {
	if p.tk.Type != ty {
		util.Error("Expected %s but got %s", ty, p.tk.Type)
	}
	p.next()
}

func (p *parser) consumeComma(terminator token.Type) {
	switch p.tk.Type {
	case token.COMMA:
		p.next()
	case terminator:
		// ok
	default:
		util.Error("Expected , or %s but got %s", terminator, p.tk.Type)
	}
}

// ----------------------------------------------------------------
// Type

func (p *parser) parseType() types.Type {
	switch p.tk.Type {
	case token.INT:
		p.next()
		return &types.Int{}
	case token.BOOL:
		p.next()
		return &types.Bool{}
	case token.STRING:
		p.next()
		return &types.String{}
	case token.RANGE:
		p.next()
		return &types.Range{}
	case token.LBRACK:
		return p.parseArray()
	case token.LPAREN:
		return p.parseFunc()
	default:
		util.Error("Unexpected %s", p.tk.Type)
		return nil // unreachable
	}
}

func (p *parser) parseArray() *types.Array {
	p.next()
	p.expect(token.NUMBER)
	len, err := strconv.Atoi(p.tk.Literal)
	if err != nil {
		util.Error("Could not parse %s as integer", p.tk.Literal)
	}
	if len < 0 {
		util.Error("Array length must be non-negative number")
	}
	p.next()
	p.consume(token.RBRACK)
	elemType := p.parseType()
	return &types.Array{Len: len, ElemType: elemType}
}

func (p *parser) parseFunc() *types.Func {
	p.next()
	paramTypes := make([]types.Type, 0, 4)
	for p.tk.Type != token.RPAREN {
		paramTypes = append(paramTypes, p.parseType())
		p.consumeComma(token.RPAREN)
	}
	p.next()
	p.consume(token.ARROW)
	var returnType types.Type
	if p.tk.Type == token.LBRACE {
		p.next()
		p.consume(token.RBRACE)
	} else {
		returnType = p.parseType()
	}
	return &types.Func{ParamTypes: paramTypes, ReturnType: returnType}
}

// ----------------------------------------------------------------
// Program

func (p *parser) parseProgram() *ast.Program {
	stmts := make([]ast.Stmt, 0, 8)
	for p.tk.Type != token.EOF {
		stmts = append(stmts, p.parseStmt())
	}
	return &ast.Program{Stmts: stmts}
}

// ----------------------------------------------------------------
// Stmt

func (p *parser) parseStmt() ast.Stmt {
	switch p.tk.Type {
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.VAR:
		return p.parseVarStmt()
	case token.FUNC:
		return p.parseFuncStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.FOR:
		return p.parseForStmt()
	case token.CONTINUE:
		return p.parseContinueStmt()
	case token.BREAK:
		return p.parseBreakStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	default:
		return p.parseAssignStmtOrExprStmt()
	}
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

func (p *parser) parseVarStmt() *ast.VarStmt {
	p.next()
	vars := make([]*ast.VarDecl, 0, 2)
	for p.tk.Type != token.SEMICOLON {
		v := p.parseVarDecl()
		if v.Value == nil {
			util.Error("%s has no initial value", v.Name)
		}
		vars = append(vars, v)
		p.consumeComma(token.SEMICOLON)
	}
	p.next()
	return &ast.VarStmt{Vars: vars}
}

func (p *parser) parseFuncStmt() *ast.FuncStmt {
	p.next()
	fn := p.parseFuncDecl()
	return &ast.FuncStmt{Func: fn}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	p.next()
	cond := p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	body := p.parseBlockStmt()
	if p.tk.Type != token.ELSE {
		return &ast.IfStmt{Cond: cond, Body: body}
	}
	p.next()
	var els ast.Stmt
	switch p.tk.Type {
	case token.LBRACE:
		els = p.parseBlockStmt()
	case token.IF:
		els = p.parseIfStmt()
	default:
		util.Error("Expected { or if but got %s", p.tk.Type)
	}
	return &ast.IfStmt{Cond: cond, Body: body, Else: els}
}

func (p *parser) parseWhileStmt() *ast.WhileStmt {
	p.next()
	cond := p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	body := p.parseBlockStmt()
	return &ast.WhileStmt{Cond: cond, Body: body}
}

func (p *parser) parseForStmt() *ast.ForStmt {
	p.next()
	elem := &ast.VarDecl{}
	index := &ast.VarDecl{}
	iter := &ast.VarDecl{}
	p.expect(token.IDENT)
	elem.Name = p.tk.Literal
	p.next()
	if p.tk.Type == token.COMMA {
		p.next()
		p.expect(token.IDENT)
		index.Name = p.tk.Literal
		p.next()
	}
	p.consume(token.IN)
	iter.Value = p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	body := p.parseBlockStmt()
	return &ast.ForStmt{Elem: elem, Index: index, Iter: iter, Body: body}
}

func (p *parser) parseContinueStmt() *ast.ContinueStmt {
	p.next()
	p.consume(token.SEMICOLON)
	return &ast.ContinueStmt{}
}

func (p *parser) parseBreakStmt() *ast.BreakStmt {
	p.next()
	p.consume(token.SEMICOLON)
	return &ast.BreakStmt{}
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	p.next()
	if p.tk.Type == token.SEMICOLON {
		p.next()
		return &ast.ReturnStmt{}
	}
	value := p.parseExpr(LOWEST)
	p.consume(token.SEMICOLON)
	return &ast.ReturnStmt{Value: value}
}

func (p *parser) parseAssignStmtOrExprStmt() ast.Stmt {
	expr := p.parseExpr(LOWEST)
	// AssignStmt
	if _, ok := assignOps[p.tk.Type]; ok {
		switch expr.(type) {
		case *ast.Ident, *ast.IndexExpr:
			// ok
		default:
			util.Error("Invalid target in assignment")
		}
		op := p.tk.Literal
		p.next()
		value := p.parseExpr(LOWEST)
		p.consume(token.SEMICOLON)
		return &ast.AssignStmt{Op: op, Target: expr, Value: value}
	}
	// ExprStmt
	p.consume(token.SEMICOLON)
	return &ast.ExprStmt{Expr: expr}
}

// ----------------------------------------------------------------
// Expr

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
		expr = p.parseArrayLitOrArrayShortLit()
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
			expr = p.parseCallExprOrLibCallExpr(expr)
		case token.BETWEEN:
			expr = p.parseRangeLit(expr)
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
	return &ast.InfixExpr{Op: op, Left: left, Right: right}
}

func (p *parser) parseIndexExpr(left ast.Expr) *ast.IndexExpr {
	p.next()
	index := p.parseExpr(LOWEST)
	p.consume(token.RBRACK)
	return &ast.IndexExpr{Left: left, Index: index}
}

func (p *parser) parseCallExprOrLibCallExpr(left ast.Expr) ast.Expr {
	p.next()
	params := make([]ast.Expr, 0, 4)
	for p.tk.Type != token.RPAREN {
		params = append(params, p.parseExpr(LOWEST))
		p.consumeComma(token.RPAREN)
	}
	p.next()
	if v, ok := left.(*ast.Ident); ok {
		if _, ok := libFuncs[v.Name]; ok {
			return &ast.LibCallExpr{Name: v.Name, Params: params}
		}
	}
	return &ast.CallExpr{Left: left, Params: params}
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
			if raw, ok := unescape[ch]; ok {
				value += string(raw)
				escaped = false
			} else {
				util.Error("Unknown escape sequence \\%c", ch)
			}
		} else {
			if ch == '"' {
				continue
			} else if ch == '\\' {
				escaped = true
			} else {
				value += string(ch)
			}
		}
	}
	p.next()
	return &ast.StringLit{Value: value}
}

func (p *parser) parseRangeLit(lower ast.Expr) *ast.RangeLit {
	p.next()
	upper := p.parseExpr(BETWEEN)
	return &ast.RangeLit{Lower: lower, Upper: upper}
}

func (p *parser) parseArrayLitOrArrayShortLit() ast.Expr {
	p.next()
	expr := p.parseExpr(LOWEST)
	// ArrayShortLit
	if p.tk.Type == token.RBRACK {
		if _, ok := typeStart[p.peekTk().Type]; ok {
			i, ok := expr.(*ast.IntLit)
			if !ok || i.Value < 0 {
				util.Error("Array length must be non-negative number")
			}
			p.next()
			elemType := p.parseType()
			p.consume(token.LPAREN)
			var value ast.Expr
			if p.tk.Type != token.RPAREN {
				value = p.parseExpr(LOWEST)
				p.expect(token.RPAREN)
			}
			p.next()
			return &ast.ArrayShortLit{Len: i.Value, ElemType: elemType, Value: value}
		}
	}
	// ArrayLit
	elems := append(make([]ast.Expr, 0, 8), expr)
	p.consumeComma(token.RBRACK)
	for p.tk.Type != token.RBRACK {
		elems = append(elems, p.parseExpr(LOWEST))
		p.consumeComma(token.RBRACK)
	}
	p.next()
	return &ast.ArrayLit{Elems: elems}
}

func (p *parser) parseFuncLitOrGroupedExpr() ast.Expr {
	p.next()
	// grouped expression
	if p.tk.Type != token.RPAREN && p.peekTk().Type != token.COLON {
		expr := p.parseExpr(LOWEST)
		p.consume(token.RPAREN)
		return expr
	}
	// FuncLit
	params := make([]*ast.VarDecl, 0, 4)
	for p.tk.Type != token.RPAREN {
		param := p.parseVarDecl()
		if param.VarType == nil {
			util.Error("Type of parameter must be annotated")
		}
		if param.Value != nil {
			util.Error("Parameter cannot have initial value")
		}
		params = append(params, param)
		p.consumeComma(token.RPAREN)
	}
	p.next()
	p.consume(token.ARROW)
	var returnType types.Type
	if p.tk.Type != token.LBRACE {
		returnType = p.parseType()
		p.expect(token.LBRACE)
	}
	body := p.parseBlockStmt()
	return &ast.FuncLit{Params: params, ReturnType: returnType, Body: body}
}

// ----------------------------------------------------------------
// Decl

func (p *parser) parseVarDecl() *ast.VarDecl {
	p.expect(token.IDENT)
	name := p.tk.Literal
	p.next()
	var varType types.Type
	if p.tk.Type == token.COLON {
		p.next()
		varType = p.parseType()
	}
	var value ast.Expr
	if p.tk.Type == token.ASSIGN {
		p.next()
		value = p.parseExpr(LOWEST)
	}
	return &ast.VarDecl{Name: name, VarType: varType, Value: value}
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	p.expect(token.IDENT)
	name := p.tk.Literal
	p.next()
	p.consume(token.LPAREN)
	params := make([]*ast.VarDecl, 0, 4)
	for p.tk.Type != token.RPAREN {
		param := p.parseVarDecl()
		if param.VarType == nil {
			util.Error("Type of parameter must be annotated")
		}
		if param.Value != nil {
			util.Error("Parameter cannot have initial value")
		}
		params = append(params, param)
		p.consumeComma(token.RPAREN)
	}
	p.next()
	p.consume(token.ARROW)
	var returnType types.Type
	if p.tk.Type != token.LBRACE {
		returnType = p.parseType()
		p.expect(token.LBRACE)
	}
	body := p.parseBlockStmt()
	return &ast.FuncDecl{Name: name, Params: params, ReturnType: returnType, Body: body}
}
