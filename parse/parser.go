package parse

import (
	"fmt"
	"os"
	"strconv"

	"github.com/oshima/lang/ast"
	"github.com/oshima/lang/token"
	"github.com/oshima/lang/types"
)

type parser struct {
	tokens []*token.Token // input tokens
	idx    int            // current index
	tok    *token.Token   // current token (tokens[idx])
}

func (p *parser) next() {
	p.idx++
	p.tok = p.tokens[p.idx]
}

func (p *parser) peek() *token.Token {
	if p.tok.Type == token.EOF {
		return p.tok
	}
	return p.tokens[p.idx+1]
}

func (p *parser) expect(typ token.Type) {
	if p.tok.Type != typ {
		p.error("%s: expected %s, but got %s", p.tok.Pos, typ, p.tok.Type)
	}
}

func (p *parser) consume(typ token.Type) {
	if p.tok.Type != typ {
		p.error("%s: expected %s, but got %s", p.tok.Pos, typ, p.tok.Type)
	}
	p.next()
}

func (p *parser) consumeComma(end token.Type) {
	switch p.tok.Type {
	case token.COMMA:
		p.next()
	case end:
		// ok
	default:
		p.error("%s: expected , or %s, but got %s", p.tok.Pos, end, p.tok.Type)
	}
}

func (p *parser) prec() int {
	if prec, ok := precOf[p.tok.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *parser) error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// ----------------------------------------------------------------
// Program

func (p *parser) parseProgram() *ast.Program {
	prog := new(ast.Program)
	for p.tok.Type != token.EOF {
		prog.Stmts = append(prog.Stmts, p.parseStmt())
	}
	return prog
}

// ----------------------------------------------------------------
// Stmt

func (p *parser) parseStmt() ast.Stmt {
	switch p.tok.Type {
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
	stmt := new(ast.BlockStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	for p.tok.Type != token.RBRACE {
		stmt.Stmts = append(stmt.Stmts, p.parseStmt())
	}
	p.next()
	return stmt
}

func (p *parser) parseVarStmt() *ast.VarStmt {
	stmt := new(ast.VarStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	for p.tok.Type != token.SEMICOLON {
		v := p.parseVarDecl()
		if v.Value == nil {
			p.error("%s: %s has no initial value", v.Pos(), v.Name)
		}
		stmt.Vars = append(stmt.Vars, v)
		p.consumeComma(token.SEMICOLON)
	}
	p.next()
	return stmt
}

func (p *parser) parseFuncStmt() *ast.FuncStmt {
	stmt := new(ast.FuncStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	stmt.Func = p.parseFuncDecl()
	return stmt
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	stmt := new(ast.IfStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	stmt.Cond = p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	stmt.Body = p.parseBlockStmt()
	if p.tok.Type != token.ELSE {
		return stmt
	}
	p.next()
	switch p.tok.Type {
	case token.LBRACE:
		stmt.Else = p.parseBlockStmt()
	case token.IF:
		stmt.Else = p.parseIfStmt()
	default:
		p.error("%s: expected { or if, but got %s", p.tok.Pos, p.tok.Type)
	}
	return stmt
}

func (p *parser) parseWhileStmt() *ast.WhileStmt {
	stmt := new(ast.WhileStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	stmt.Cond = p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	stmt.Body = p.parseBlockStmt()
	return stmt
}

func (p *parser) parseForStmt() *ast.ForStmt {
	stmt := new(ast.ForStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	p.expect(token.IDENT)
	stmt.Elem = &ast.VarDecl{Name: p.tok.Literal}
	p.next()
	if p.tok.Type == token.COMMA {
		p.next()
		p.expect(token.IDENT)
		stmt.Index = &ast.VarDecl{Name: p.tok.Literal}
		p.next()
	} else {
		stmt.Index = &ast.VarDecl{}
	}
	p.consume(token.IN)
	stmt.Iter = &ast.VarDecl{Value: p.parseExpr(LOWEST)}
	p.expect(token.LBRACE)
	stmt.Body = p.parseBlockStmt()
	return stmt
}

func (p *parser) parseContinueStmt() *ast.ContinueStmt {
	stmt := new(ast.ContinueStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	p.consume(token.SEMICOLON)
	return stmt
}

func (p *parser) parseBreakStmt() *ast.BreakStmt {
	stmt := new(ast.BreakStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	p.consume(token.SEMICOLON)
	return stmt
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	stmt := new(ast.ReturnStmt)
	stmt.SetPos(p.tok.Pos)
	p.next()
	if p.tok.Type == token.SEMICOLON {
		p.next()
		return stmt
	}
	stmt.Value = p.parseExpr(LOWEST)
	p.consume(token.SEMICOLON)
	return stmt
}

func (p *parser) parseAssignStmtOrExprStmt() ast.Stmt {
	pos := p.tok.Pos
	expr := p.parseExpr(LOWEST)
	// AssignStmt
	if _, ok := assignOps[p.tok.Type]; ok {
		stmt := new(ast.AssignStmt)
		switch expr.(type) {
		case *ast.Ident, *ast.IndexExpr:
			stmt.Target = expr
		default:
			p.error("%s: invalid target in assignment", expr.Pos())
		}
		stmt.SetPos(p.tok.Pos)
		stmt.Op = p.tok.Type
		p.next()
		stmt.Value = p.parseExpr(LOWEST)
		p.consume(token.SEMICOLON)
		return stmt
	}
	// ExprStmt
	stmt := new(ast.ExprStmt)
	stmt.SetPos(pos)
	stmt.Expr = expr
	p.consume(token.SEMICOLON)
	return stmt
}

// ----------------------------------------------------------------
// Expr

func (p *parser) parseExpr(prec int) ast.Expr {
	var expr ast.Expr

	switch p.tok.Type {
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
		p.error("%s: unexpected %s", p.tok.Pos, p.tok.Type)
	}

	for p.prec() > prec {
		switch p.tok.Type {
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
	expr := new(ast.PrefixExpr)
	expr.SetPos(p.tok.Pos)
	expr.Op = p.tok.Type
	p.next()
	expr.Right = p.parseExpr(PREFIX)
	return expr
}

func (p *parser) parseInfixExpr(left ast.Expr) *ast.InfixExpr {
	expr := new(ast.InfixExpr)
	expr.Left = left
	expr.SetPos(p.tok.Pos)
	expr.Op = p.tok.Type
	prec := p.prec()
	p.next()
	expr.Right = p.parseExpr(prec)
	return expr
}

func (p *parser) parseIndexExpr(left ast.Expr) *ast.IndexExpr {
	expr := new(ast.IndexExpr)
	expr.Left = left
	expr.SetPos(p.tok.Pos)
	p.next()
	expr.Index = p.parseExpr(LOWEST)
	p.consume(token.RBRACK)
	return expr
}

func (p *parser) parseCallExprOrLibCallExpr(left ast.Expr) ast.Expr {
	pos := p.tok.Pos
	p.next()
	params := make([]ast.Expr, 0, 4)
	for p.tok.Type != token.RPAREN {
		params = append(params, p.parseExpr(LOWEST))
		p.consumeComma(token.RPAREN)
	}
	p.next()
	// LibCallExpr
	if v, ok := left.(*ast.Ident); ok {
		if _, ok := libFuncs[v.Name]; ok {
			expr := new(ast.LibCallExpr)
			expr.Name = v.Name
			expr.SetPos(pos)
			expr.Params = params
			return expr
		}
	}
	// CallExpr
	expr := new(ast.CallExpr)
	expr.Left = left
	expr.SetPos(pos)
	expr.Params = params
	return expr
}

func (p *parser) parseIdent() *ast.Ident {
	expr := new(ast.Ident)
	expr.SetPos(p.tok.Pos)
	expr.Name = p.tok.Literal
	p.next()
	return expr
}

func (p *parser) parseIntLit() *ast.IntLit {
	expr := new(ast.IntLit)
	expr.SetPos(p.tok.Pos)
	value, err := strconv.Atoi(p.tok.Literal)
	if err != nil {
		p.error("%s: cannot parse %s as integer", p.tok.Pos, p.tok.Literal)
	}
	expr.Value = value
	p.next()
	return expr
}

func (p *parser) parseBoolLit() *ast.BoolLit {
	expr := new(ast.BoolLit)
	expr.SetPos(p.tok.Pos)
	expr.Value = p.tok.Type == token.TRUE
	p.next()
	return expr
}

func (p *parser) parseStringLit() *ast.StringLit {
	expr := new(ast.StringLit)
	expr.SetPos(p.tok.Pos)
	escaped := false
	for _, ch := range p.tok.Literal {
		if escaped {
			if unescaped, ok := unescape[ch]; ok {
				expr.Value += string(unescaped)
				escaped = false
			} else {
				p.error("%s: unknown escape sequence \\%c", p.tok.Pos, ch)
			}
		} else {
			if ch == '"' {
				continue
			} else if ch == '\\' {
				escaped = true
			} else {
				expr.Value += string(ch)
			}
		}
	}
	p.next()
	return expr
}

func (p *parser) parseRangeLit(lower ast.Expr) *ast.RangeLit {
	expr := new(ast.RangeLit)
	expr.Lower = lower
	expr.SetPos(p.tok.Pos)
	p.next()
	expr.Upper = p.parseExpr(BETWEEN)
	return expr
}

func (p *parser) parseArrayLitOrArrayShortLit() ast.Expr {
	pos := p.tok.Pos
	p.next()
	pick := p.parseExpr(LOWEST)
	// ArrayShortLit
	if p.tok.Type == token.RBRACK {
		if _, ok := typeBegin[p.peek().Type]; ok {
			expr := new(ast.ArrayShortLit)
			expr.SetPos(pos)
			i, ok := pick.(*ast.IntLit)
			if !ok || i.Value < 0 {
				p.error("%s: array length must be non-negative number", i.Pos())
			}
			expr.Len = i.Value
			p.next()
			expr.ElemType = p.parseType()
			p.consume(token.LPAREN)
			if p.tok.Type != token.RPAREN {
				expr.Value = p.parseExpr(LOWEST)
				p.expect(token.RPAREN)
			}
			p.next()
			return expr
		}
	}
	// ArrayLit
	expr := new(ast.ArrayLit)
	expr.SetPos(pos)
	expr.Elems = append(expr.Elems, pick)
	p.consumeComma(token.RBRACK)
	for p.tok.Type != token.RBRACK {
		expr.Elems = append(expr.Elems, p.parseExpr(LOWEST))
		p.consumeComma(token.RBRACK)
	}
	p.next()
	return expr
}

func (p *parser) parseFuncLitOrGroupedExpr() ast.Expr {
	pos := p.tok.Pos
	p.next()
	// FuncLit
	if p.tok.Type == token.RPAREN || p.peek().Type == token.COLON {
		expr := new(ast.FuncLit)
		expr.SetPos(pos)
		for p.tok.Type != token.RPAREN {
			param := p.parseVarDecl()
			if param.VarType == nil {
				p.error("%s: type of %s must be annotated", param.Pos(), param.Name)
			}
			if param.Value != nil {
				p.error("%s: %s cannot have initial value", param.Pos(), param.Name)
			}
			expr.Params = append(expr.Params, param)
			p.consumeComma(token.RPAREN)
		}
		p.next()
		p.consume(token.ARROW)
		if p.tok.Type != token.LBRACE {
			expr.ReturnType = p.parseType()
			p.expect(token.LBRACE)
		}
		expr.Body = p.parseBlockStmt()
		return expr
	}
	// grouped expression
	expr := p.parseExpr(LOWEST)
	p.consume(token.RPAREN)
	return expr
}

// ----------------------------------------------------------------
// Decl

func (p *parser) parseVarDecl() *ast.VarDecl {
	p.expect(token.IDENT)
	decl := new(ast.VarDecl)
	decl.SetPos(p.tok.Pos)
	decl.Name = p.tok.Literal
	p.next()
	if p.tok.Type != token.COLON && p.tok.Type != token.ASSIGN {
		p.error("%s: unexpected %s", p.tok.Pos, p.tok.Type)
	}
	if p.tok.Type == token.COLON {
		p.next()
		decl.VarType = p.parseType()
	}
	if p.tok.Type == token.ASSIGN {
		p.next()
		decl.Value = p.parseExpr(LOWEST)
	}
	return decl
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	p.expect(token.IDENT)
	decl := new(ast.FuncDecl)
	decl.SetPos(p.tok.Pos)
	decl.Name = p.tok.Literal
	p.next()
	p.consume(token.LPAREN)
	for p.tok.Type != token.RPAREN {
		param := p.parseVarDecl()
		if param.VarType == nil {
			p.error("%s: type of %s must be annotated", param.Pos(), param.Name)
		}
		if param.Value != nil {
			p.error("%s: %s cannot have initial value", param.Pos(), param.Name)
		}
		decl.Params = append(decl.Params, param)
		p.consumeComma(token.RPAREN)
	}
	p.next()
	if p.tok.Type != token.LBRACE {
		p.consume(token.ARROW)
		decl.ReturnType = p.parseType()
		p.expect(token.LBRACE)
	}
	decl.Body = p.parseBlockStmt()
	return decl
}

// ----------------------------------------------------------------
// Type

func (p *parser) parseType() types.Type {
	switch p.tok.Type {
	case token.INT:
		p.next()
		return new(types.Int)
	case token.BOOL:
		p.next()
		return new(types.Bool)
	case token.STRING:
		p.next()
		return new(types.String)
	case token.RANGE:
		p.next()
		return new(types.Range)
	case token.LBRACK:
		return p.parseArray()
	case token.LPAREN:
		return p.parseFunc()
	default:
		p.error("%s: unexpected %s", p.tok.Pos, p.tok.Type)
		return nil // unreachable
	}
}

func (p *parser) parseArray() *types.Array {
	typ := new(types.Array)
	p.next()
	p.expect(token.NUMBER)
	len, err := strconv.Atoi(p.tok.Literal)
	if err != nil {
		p.error("%s: cannot parse %s as integer", p.tok.Pos, p.tok.Literal)
	}
	if len < 0 {
		p.error("%s: array length must be non-negative number", p.tok.Pos)
	}
	typ.Len = len
	p.next()
	p.consume(token.RBRACK)
	typ.ElemType = p.parseType()
	return typ
}

func (p *parser) parseFunc() *types.Func {
	typ := new(types.Func)
	p.next()
	for p.tok.Type != token.RPAREN {
		typ.ParamTypes = append(typ.ParamTypes, p.parseType())
		p.consumeComma(token.RPAREN)
	}
	p.next()
	p.consume(token.ARROW)
	if p.tok.Type == token.VOID {
		p.next()
	} else {
		typ.ReturnType = p.parseType()
	}
	return typ
}
