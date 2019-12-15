package parser

import (
	"compiler/ast"
	"compiler/lexer"
	"errors"
	"fmt"
	"io"
)

// Parser holds the Lexer to generate the AST
type Parser struct {
	l *lexer.Lexer
}

// NewParser creates a new parser
func NewParser(lexer *lexer.Lexer) *Parser {
	return &Parser{
		l: lexer,
	}
}

// NextValidToken finds the next non whitespace or line return token
func (p *Parser) NextValidToken() (*lexer.Token, error) {
	for {
		t := p.l.Next()
		err := p.l.Err()
		if err != nil {
			return nil, err
		}
		if t.Type != lexer.WhitespaceToken && t.Type != lexer.LineTerminatorToken {
			return t, nil
		}
	}
	return nil, errors.New("Could not find next token")
}

// ParseReturnStatement will return a Statement from a set of tokens
// It follows this grammar
// <return_statement> ::= "return" <exp>
func (p *Parser) ParseReturnStatement(tokens []*lexer.Token) (ast.Statement, error) {
	exp, _, err := p.ParseExpression(tokens)
	if err != nil {
		return nil, err
	}
	stmt, err := ast.NewReturnStatement(exp)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

// ParseStatement will return the correct Statement for the tokens to follow
// It will get all tokens until the next ";"
// Right now only return statements exists
// <statement> ::= <return_statement>
func (p *Parser) ParseStatement(t *lexer.Token) (ast.Statement, error) {
	// Get all tokens
	tokens, err := p.GetStatementTokens()
	if err != nil {
		return nil, err
	}
	// Merge the detection token with the received list
	tokens = append([]*lexer.Token{t}, tokens...)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("Could not parse statement, empty token list")
	}
	// Get first token to decide what type of statement this will be
	t = tokens[0]
	var stmt ast.Statement
	if t.Type != lexer.IdentifierToken {
		return nil, fmt.Errorf("Parse error, %s, %s", string(t.Value), t.Type)
	}
	if string(t.Value) == "return" {
		// Return statement detected, remove the "return" token and parse the rest
		tokens = tokens[1:]
		s, err := p.ParseReturnStatement(tokens)
		if err != nil {
			return nil, err
		}
		stmt = s
	} else {
		return nil, fmt.Errorf("Statement not supported", string(t.Value))
	}
	return stmt, nil
}

// ParseBlockStatement will return a statement list of all statements in a block
// <block_statement> ::= { <statement> }
func (p *Parser) ParseBlockStatement() (*ast.BlockStatement, error) {
	stmts, err := ast.NewStatementList()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.NextValidToken()
		if err != nil {
			return nil, err
		}
		// Indicates the end of the block
		if string(t.Value) == "}" {
			break
		}
		stmt, err := p.ParseStatement(t)
		if err != nil {
			return nil, err
		}
		stmts, err = ast.AppendStatement(stmts, stmt)
		if err != nil {
			return nil, err
		}
	}
	block, err := ast.NewBlockStatement(stmts)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// ParseProgram will parse the entire source by consuming all tokens from the lexer
// and building an AST with a Program as the root
func (p *Parser) ParseProgram() (*ast.Program, error) {
	// Prepare the function list of the program
	fns, err := ast.NewStatementList()
	if err != nil {
		return nil, err
	}
	// Prepare the statement list of the program
	stmts, err := ast.NewStatementList()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.NextValidToken()
		// Stop if hit the end of the file
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// We only support top level functions with int return type so far
		if string(t.Value) == "int" {
			f, err := p.ParseFunction(t)
			if err != nil {
				return nil, err
			}
			fns, err = ast.AppendStatement(fns, f)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}

	}
	program, err := ast.NewProgram(fns, stmts)
	if err != nil {
		return nil, err
	}
	return program, nil
}

// ParseFunction will return a Function node from the next tokens in the lexer
// <function> ::= "int" <identifier> "(" ")" <block_statement>
func (p *Parser) ParseFunction(token *lexer.Token) (ast.Statement, error) {
	if token.Type != lexer.IdentifierToken {
		return nil, fmt.Errorf("Failed to parse: Expected identifier got %s", token.Value)
	}
	// Second token in the list is the function name
	nameToken, err := p.NextValidToken()
	if err != nil {
		return nil, err
	}
	if nameToken.Type != lexer.IdentifierToken {
		return nil, errors.New("Failed to parse")
	}
	t, err := p.NextValidToken()
	if err != nil {
		return nil, err
	}
	if t.Value[0] != '(' {
		return nil, fmt.Errorf("Unexpected %s, expected (", string(t.Value))
	}
	// Ignore this for now, we don't parse the function parameters yet
	t, err = p.NextValidToken()
	if err != nil {
		return nil, err
	}
	if t.Value[0] != ')' {
		return nil, fmt.Errorf("Unexpected %s, expected )", string(t.Value))
	}
	t, err = p.NextValidToken()
	if err != nil {
		return nil, err
	}
	// Make sure we got the beginning of a block statement
	if t.Value[0] != '{' {
		return nil, fmt.Errorf("Unexpected %s, expected {", string(t.Value))
	}
	// This will get the function body as a block statement
	blockStmt, err := p.ParseBlockStatement()
	if err != nil {
		return nil, err
	}
	fun, err := ast.NewFunctionStatement(nameToken, nil, token, blockStmt)
	if err != nil {
		return nil, fmt.Errorf("Error")
	}
	return fun, nil
}

// GetStatementTokens will read the valid tokens from the lexer until it finds a ";"
func (p *Parser) GetStatementTokens() ([]*lexer.Token, error) {
	tokens := make([]*lexer.Token, 0)
	for {
		t, err := p.NextValidToken()
		if err != nil {
			return nil, err
		}
		if string(t.Value) == ";" {
			break
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}
