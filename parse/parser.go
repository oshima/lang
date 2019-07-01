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
	tk     *token.Token   // current token (tokens[idx])
}

func (p *parser) next() {
	p.idx++
	p.tk = p.tokens[p.idx]
}

func (p *parser) peekTk() *token.Token {
	if p.tk.Type == token.EOF {
		p.error("%s: unexpected eof", p.tk.Pos)
	}
	return p.tokens[p.idx+1]
}

func (p *parser) lookPrec() int {
	if prec, ok := precOf[p.tk.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *parser) expect(ty token.Type) {
	if p.tk.Type != ty {
		p.error("%s: expected %s, but got %s", p.tk.Pos, ty, p.tk.Type)
	}
}

func (p *parser) consume(ty token.Type) {
	if p.tk.Type != ty {
		p.error("%s: expected %s, but got %s", p.tk.Pos, ty, p.tk.Type)
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
		p.error("%s: expected , or %s, but got %s", p.tk.Pos, terminator, p.tk.Type)
	}
}

func (p *parser) error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// ----------------------------------------------------------------
// Program

func (p *parser) parseProgram() *ast.Program {
	prog := new(ast.Program)
	for p.tk.Type != token.EOF {
		prog.Stmts = append(prog.Stmts, p.parseStmt())
	}
	return prog
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
	stmt := new(ast.BlockStmt)
	stmt.SetPos(p.tk.Pos)
	p.next()
	for p.tk.Type != token.RBRACE {
		stmt.Stmts = append(stmt.Stmts, p.parseStmt())
	}
	p.next()
	return stmt
}

func (p *parser) parseVarStmt() *ast.VarStmt {
	stmt := new(ast.VarStmt)
	stmt.SetPos(p.tk.Pos)
	p.next()
	for p.tk.Type != token.SEMICOLON {
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
	stmt.SetPos(p.tk.Pos)
	p.next()
	stmt.Func = p.parseFuncDecl()
	return stmt
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	stmt := new(ast.IfStmt)
	stmt.SetPos(p.tk.Pos)
	p.next()
	stmt.Cond = p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	stmt.Body = p.parseBlockStmt()
	if p.tk.Type != token.ELSE {
		return stmt
	}
	p.next()
	switch p.tk.Type {
	case token.LBRACE:
		stmt.Else = p.parseBlockStmt()
	case token.IF:
		stmt.Else = p.parseIfStmt()
	default:
		p.error("%s: expected { or if, but got %s", p.tk.Pos, p.tk.Type)
	}
	return stmt
}

func (p *parser) parseWhileStmt() *ast.WhileStmt {
	stmt := new(ast.WhileStmt)
	stmt.SetPos(p.tk.Pos)
	p.next()
	stmt.Cond = p.parseExpr(LOWEST)
	p.expect(token.LBRACE)
	stmt.Body = p.parseBlockStmt()
	return stmt
}

func (p *parser) parseForStmt() *ast.ForStmt {
	stmt := new(ast.ForStmt)
	stmt.SetPos(p.tk.Pos)
	p.next()
	p.expect(token.IDENT)
	stmt.Elem = &ast.VarDecl{Name: p.tk.Literal}
	p.next()
	if p.tk.Type == token.COMMA {
		p.next()
		p.expect(token.IDENT)
		stmt.Index = &ast.VarDecl{Name: p.tk.Literal}
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
	stmt.SetPos(p.tk.Pos)
	p.next()
	p.consume(token.SEMICOLON)
	return stmt
}

func (p *parser) parseBreakStmt() *ast.BreakStmt {
	stmt := new(ast.BreakStmt)
	stmt.SetPos(p.tk.Pos)
	p.next()
	p.consume(token.SEMICOLON)
	return stmt
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	stmt := new(ast.ReturnStmt)
	stmt.SetPos(p.tk.Pos)
	p.next()
	if p.tk.Type == token.SEMICOLON {
		p.next()
		return stmt
	}
	stmt.Value = p.parseExpr(LOWEST)
	p.consume(token.SEMICOLON)
	return stmt
}

func (p *parser) parseAssignStmtOrExprStmt() ast.Stmt {
	pos := p.tk.Pos
	expr := p.parseExpr(LOWEST)
	// AssignStmt
	if _, ok := assignOps[p.tk.Type]; ok {
		stmt := new(ast.AssignStmt)
		switch expr.(type) {
		case *ast.Ident, *ast.IndexExpr:
			stmt.Target = expr
		default:
			p.error("%s: invalid target in assignment", expr.Pos())
		}
		stmt.SetPos(p.tk.Pos)
		stmt.Op = p.tk.Literal
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
		p.error("%s: unexpected %s", p.tk.Pos, p.tk.Type)
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
	expr := new(ast.PrefixExpr)
	expr.SetPos(p.tk.Pos)
	expr.Op = p.tk.Literal
	p.next()
	expr.Right = p.parseExpr(PREFIX)
	return expr
}

func (p *parser) parseInfixExpr(left ast.Expr) *ast.InfixExpr {
	expr := new(ast.InfixExpr)
	expr.Left = left
	expr.SetPos(p.tk.Pos)
	expr.Op = p.tk.Literal
	prec := p.lookPrec()
	p.next()
	expr.Right = p.parseExpr(prec)
	return expr
}

func (p *parser) parseIndexExpr(left ast.Expr) *ast.IndexExpr {
	expr := new(ast.IndexExpr)
	expr.Left = left
	expr.SetPos(p.tk.Pos)
	p.next()
	expr.Index = p.parseExpr(LOWEST)
	p.consume(token.RBRACK)
	return expr
}

func (p *parser) parseCallExprOrLibCallExpr(left ast.Expr) ast.Expr {
	pos := p.tk.Pos
	p.next()
	params := make([]ast.Expr, 0, 4)
	for p.tk.Type != token.RPAREN {
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
	expr.SetPos(p.tk.Pos)
	expr.Name = p.tk.Literal
	p.next()
	return expr
}

func (p *parser) parseIntLit() *ast.IntLit {
	expr := new(ast.IntLit)
	expr.SetPos(p.tk.Pos)
	value, err := strconv.Atoi(p.tk.Literal)
	if err != nil {
		p.error("%s: cannot parse %s as integer", p.tk.Pos, p.tk.Literal)
	}
	expr.Value = value
	p.next()
	return expr
}

func (p *parser) parseBoolLit() *ast.BoolLit {
	expr := new(ast.BoolLit)
	expr.SetPos(p.tk.Pos)
	expr.Value = p.tk.Type == token.TRUE
	p.next()
	return expr
}

func (p *parser) parseStringLit() *ast.StringLit {
	expr := new(ast.StringLit)
	expr.SetPos(p.tk.Pos)
	escaped := false
	for _, ch := range p.tk.Literal {
		if escaped {
			if raw, ok := unescape[ch]; ok {
				expr.Value += string(raw)
				escaped = false
			} else {
				p.error("%s: unknown escape sequence \\%c", p.tk.Pos, ch)
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
	expr.SetPos(p.tk.Pos)
	p.next()
	expr.Upper = p.parseExpr(BETWEEN)
	return expr
}

func (p *parser) parseArrayLitOrArrayShortLit() ast.Expr {
	pos := p.tk.Pos
	p.next()
	pick := p.parseExpr(LOWEST)
	// ArrayShortLit
	if p.tk.Type == token.RBRACK {
		if _, ok := typeStart[p.peekTk().Type]; ok {
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
			if p.tk.Type != token.RPAREN {
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
	for p.tk.Type != token.RBRACK {
		expr.Elems = append(expr.Elems, p.parseExpr(LOWEST))
		p.consumeComma(token.RBRACK)
	}
	p.next()
	return expr
}

func (p *parser) parseFuncLitOrGroupedExpr() ast.Expr {
	pos := p.tk.Pos
	p.next()
	// grouped expression
	if p.tk.Type != token.RPAREN && p.peekTk().Type != token.COLON {
		expr := p.parseExpr(LOWEST)
		p.consume(token.RPAREN)
		return expr
	}
	// FuncLit
	expr := new(ast.FuncLit)
	expr.SetPos(pos)
	for p.tk.Type != token.RPAREN {
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
	if p.tk.Type != token.LBRACE {
		expr.ReturnType = p.parseType()
		p.expect(token.LBRACE)
	}
	expr.Body = p.parseBlockStmt()
	return expr
}

// ----------------------------------------------------------------
// Decl

func (p *parser) parseVarDecl() *ast.VarDecl {
	p.expect(token.IDENT)
	decl := new(ast.VarDecl)
	decl.SetPos(p.tk.Pos)
	decl.Name = p.tk.Literal
	p.next()
	if p.tk.Type != token.COLON && p.tk.Type != token.ASSIGN {
		p.error("%s: unexpected %s", p.tk.Pos, p.tk.Type)
	}
	if p.tk.Type == token.COLON {
		p.next()
		decl.VarType = p.parseType()
	}
	if p.tk.Type == token.ASSIGN {
		p.next()
		decl.Value = p.parseExpr(LOWEST)
	}
	return decl
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	p.expect(token.IDENT)
	decl := new(ast.FuncDecl)
	decl.SetPos(p.tk.Pos)
	decl.Name = p.tk.Literal
	p.next()
	p.consume(token.LPAREN)
	for p.tk.Type != token.RPAREN {
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
	p.consume(token.ARROW)
	if p.tk.Type != token.LBRACE {
		decl.ReturnType = p.parseType()
		p.expect(token.LBRACE)
	}
	decl.Body = p.parseBlockStmt()
	return decl
}

// ----------------------------------------------------------------
// Type

func (p *parser) parseType() types.Type {
	switch p.tk.Type {
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
		p.error("%s: unexpected %s", p.tk.Pos, p.tk.Type)
		return nil // unreachable
	}
}

func (p *parser) parseArray() *types.Array {
	typ := new(types.Array)
	p.next()
	p.expect(token.NUMBER)
	len, err := strconv.Atoi(p.tk.Literal)
	if err != nil {
		p.error("%s: cannot parse %s as integer", p.tk.Pos, p.tk.Literal)
	}
	if len < 0 {
		p.error("%s: array length must be non-negative number", p.tk.Pos)
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
	for p.tk.Type != token.RPAREN {
		typ.ParamTypes = append(typ.ParamTypes, p.parseType())
		p.consumeComma(token.RPAREN)
	}
	p.next()
	p.consume(token.ARROW)
	if p.tk.Type == token.LBRACE {
		p.next()
		p.consume(token.RBRACE)
	} else {
		typ.ReturnType = p.parseType()
	}
	return typ
}
