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
	if p.tk.Type == token.EOF {
		util.Error("Unexpected EOF")
	}
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

/* Type */

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
		util.Error("Array length must be non-negative")
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
	if p.tk.Type == token.BANG {
		p.next()
	} else {
		returnType = p.parseType()
	}
	return &types.Func{ParamTypes: paramTypes, ReturnType: returnType}
}

/* Program */

func (p *parser) parseProgram() *ast.Program {
	stmts := make([]ast.Stmt, 0, 8)
	for p.tk.Type != token.EOF {
		stmts = append(stmts, p.parseStmt())
	}
	return &ast.Program{Stmts: stmts}
}

/* Stmt */

func (p *parser) parseStmt() ast.Stmt {
	switch p.tk.Type {
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.LET:
		return p.parseLetStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.FOR:
		return p.parseForStmtOrForInStmt()
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

func (p *parser) parseLetStmt() *ast.LetStmt {
	p.next()
	vars := make([]*ast.VarDecl, 0, 2)
	for p.tk.Type != token.ASSIGN {
		vars = append(vars, p.parseVarDecl())
		p.consumeComma(token.ASSIGN)
	}
	if len(vars) == 0 {
		util.Error("Unexpected =")
	}
	p.next()
	values := make([]ast.Expr, 0, 2)
	for p.tk.Type != token.SEMICOLON {
		values = append(values, p.parseExpr(LOWEST))
		p.consumeComma(token.SEMICOLON)
	}
	if len(values) == 0 {
		util.Error("Unexpected ;")
	}
	if len(values) != len(vars) {
		util.Error("Wrong number of initializers")
	}
	p.next()
	return &ast.LetStmt{Vars: vars, Values: values}
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

func (p *parser) parseForStmtOrForInStmt() ast.Stmt {
	p.next()
	ty := p.peekTk().Type
	// ForStmt
	if ty != token.COLON && ty != token.COMMA && ty != token.IN {
		cond := p.parseExpr(LOWEST)
		p.expect(token.LBRACE)
		body := p.parseBlockStmt()
		return &ast.ForStmt{Cond: cond, Body: body}
	}
	// ForInStmt
	elem := p.parseVarDecl()
	index := &ast.VarDecl{}
	array := &ast.VarDecl{}
	if p.tk.Type == token.COMMA {
		p.next()
		index = p.parseVarDecl()
	}
	p.consume(token.IN)
	expr := p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	body := p.parseBlockStmt()
	return &ast.ForInStmt{Elem: elem, Index: index, Array: array, Expr: expr, Body: body}
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
	// ExprStmt
	if p.tk.Type == token.SEMICOLON {
		p.next()
		return &ast.ExprStmt{Expr: expr}
	}
	// AssignStmt
	targets := []ast.Expr{expr}
	p.consumeComma(token.ASSIGN)
	for p.tk.Type != token.ASSIGN {
		targets = append(targets, p.parseExpr(LOWEST))
		p.consumeComma(token.ASSIGN)
	}
	for _, target := range targets {
		switch target.(type) {
		case *ast.VarRef, *ast.IndexExpr:
			// ok
		default:
			util.Error("Invalid target in assignment")
		}
	}
	p.next()
	values := make([]ast.Expr, 0, 2)
	for p.tk.Type != token.SEMICOLON {
		values = append(values, p.parseExpr(LOWEST))
		p.consumeComma(token.SEMICOLON)
	}
	if len(values) == 0 {
		util.Error("Unexpected ;")
	}
	if len(values) != len(targets) {
		util.Error("Wrong number of values in assignment")
	}
	p.next()
	return &ast.AssignStmt{Targets: targets, Values: values}
}

/* Decl */

func (p *parser) parseVarDecl() *ast.VarDecl {
	p.expect(token.IDENT)
	ident := p.tk.Literal
	p.next()
	if p.tk.Type != token.COLON {
		return &ast.VarDecl{Ident: ident}
	}
	p.next()
	varType := p.parseType()
	return &ast.VarDecl{Ident: ident, VarType: varType}
}

/* Expr */

func (p *parser) parseExpr(prec int) ast.Expr {
	var expr ast.Expr

	switch p.tk.Type {
	case token.BANG, token.MINUS:
		expr = p.parsePrefixExpr()
	case token.IDENT:
		expr = p.parseVarRef()
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
			expr = p.parseCallExprOrLibCallExpr(expr)
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
	if v, ok := left.(*ast.VarRef); ok {
		if _, ok := libFuncs[v.Ident]; ok {
			return &ast.LibCallExpr{Ident: v.Ident, Params: params}
		}
	}
	return &ast.CallExpr{Left: left, Params: params}
}

func (p *parser) parseVarRef() *ast.VarRef {
	ident := p.tk.Literal
	p.next()
	return &ast.VarRef{Ident: ident}
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
	len_, err := strconv.Atoi(p.tk.Literal)
	if err != nil {
		util.Error("Could not parse %s as integer", p.tk.Literal)
	}
	if len_ < 0 {
		util.Error("Array length must be non-negative")
	}
	p.next()
	p.consume(token.RBRACK)
	elemType := p.parseType()
	p.consume(token.LBRACE)
	elems := make([]ast.Expr, 0, 8)
	for p.tk.Type != token.RBRACE {
		elems = append(elems, p.parseExpr(LOWEST))
		p.consumeComma(token.RBRACE)
	}
	if len(elems) > len_ {
		util.Error("Too many elements in array")
	}
	p.next()
	return &ast.ArrayLit{Len: len_, ElemType: elemType, Elems: elems}
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
		params = append(params, p.parseVarDecl())
		p.consumeComma(token.RPAREN)
	}
	for _, param := range params {
		if param.VarType == nil {
			util.Error("Parameter type must be annotated")
		}
	}
	p.next()
	p.consume(token.ARROW)
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
