package parser

import (
	"fmt"
	"strconv"

	"github.com/valsov/gointerpreter/ast"
	"github.com/valsov/gointerpreter/lexer"
	"github.com/valsov/gointerpreter/token"
)

const (
	LOWEST        int = iota
	EQUALS            // ==
	LESSORGREATER     // < >
	SUM               // +
	PRODUCT           // *
	PREFIX            // -x !x
	CALL              // functionCall(x)
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l                    *lexer.Lexer
	currentToken         token.Token
	peekToken            token.Token
	errors               []string
	prefixParseFunctions map[token.TokenType]prefixParseFn
	infixParseFunctions  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                    l,
		errors:               []string{},
		prefixParseFunctions: make(map[token.TokenType]prefixParseFn),
	}

	// Register expression parsers
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	// Init token cursors
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken() // Get from lexer
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFunctions[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFunctions[tokenType] = fn
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: []ast.Statement{}}

	for !p.currentTokenIs(token.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() ast.Statement {
	statement := ast.LetStatement{Token: p.currentToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// todo [...]
	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return &statement // Need to use a pointer here because LetStatement only implements Statement interface with a pointer receiver
}

func (p *Parser) parseReturnStatement() ast.Statement {
	statement := ast.ReturnStatement{Token: p.currentToken}
	p.nextToken()

	// todo [...]
	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return &statement
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: p.currentToken}
	statement.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return statement
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFunctions[p.currentToken.Type]
	if prefix == nil {
		p.errors = append(p.errors, fmt.Sprintf("no prefix parse function for '%s' found", p.currentToken.Type))
		return nil
	}
	return prefix()
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	intLit := &ast.IntegerLiteral{Token: p.currentToken}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parase %q as integer", p.currentToken.Literal))
		return nil
	}

	intLit.Value = value
	return intLit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
	}
	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if !p.peekTokenIs(t) {
		err := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
		p.errors = append(p.errors, err)
		return false
	}
	p.nextToken()
	return true
}