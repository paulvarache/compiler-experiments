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
	l           *lexer.Lexer
	TokenBuffer []*lexer.Token
}

// NewParser creates a new parser
func NewParser(l *lexer.Lexer) *Parser {
	return &Parser{
		l:           l,
		TokenBuffer: make([]*lexer.Token, 0),
	}
}

func (p *Parser) PeekNextValidToken() (*lexer.Token, error) {
	if len(p.TokenBuffer) != 0 {
		t := p.TokenBuffer[0]
		return t, nil
	}
	t, err := p.NextValidToken()
	if err != nil {
		return nil, err
	}
	p.TokenBuffer = append([]*lexer.Token{t}, p.TokenBuffer...)
	return t, nil
}

// NextValidToken finds the next non whitespace or line return token
func (p *Parser) NextValidToken() (*lexer.Token, error) {
	if len(p.TokenBuffer) != 0 {
		t := p.TokenBuffer[0]
		p.TokenBuffer = p.TokenBuffer[1:]
		return t, nil
	}
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
// <return_statement> ::= "return" <exp> ";"
func (p *Parser) ParseReturnStatement(token *lexer.Token) (ast.Statement, error) {
	tokens, err := p.GetTokensUntil(";", false)
	if err != nil {
		return nil, err
	}
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

// ParseExpressionStatement will return a Statement from a set of tokens
// It follows this grammar
// <expression_statement> ::= <exp> ";"
func (p *Parser) ParseExpressionStatement(token *lexer.Token) (ast.Statement, error) {
	tokens, err := p.GetTokensUntil(";", false)
	if err != nil {
		return nil, err
	}
	// Create an array with the first token and the rest of the line
	tokens = append([]*lexer.Token{token}, tokens...)
	exp, tokens, err := p.ParseExpression(tokens)
	if err != nil {
		return nil, err
	}
	return ast.NewExpStatement(exp)
}

// ParseDeclStatement will return a Statement from a set of tokens
// It follows this grammar
// <decl_statement> ::= "int" <id> [ = <exp> ] ";"
func (p *Parser) ParseDeclStatement(token *lexer.Token) (ast.Statement, error) {
	tokens, err := p.GetTokensUntil(";", false)
	if err != nil {
		return nil, err
	}
	tName, tokens := tokens[0], tokens[1:]
	if tName.Type != lexer.IdentifierToken {
		return nil, fmt.Errorf("Expected identifier, got '%s'", tName.Value)
	}
	if len(tokens) == 0 {
		return ast.NewDeclStatement(token, tName, nil)
	}
	t, tokens := tokens[0], tokens[1:]
	if string(t.Value) != "=" {
		return nil, fmt.Errorf("Expected '=' got '%s'", t.Value)
	}
	exp, tokens, err := p.ParseExpression(tokens)
	if err != nil {
		return nil, err
	}
	return ast.NewDeclStatement(token, tName, exp)
}

func (p *Parser) GetTokensBetween(start string, end string) ([]*lexer.Token, error) {
	t, err := p.NextValidToken()
	if err != nil {
		return nil, err
	}
	if string(t.Value) != start {
		return nil, fmt.Errorf("Expected '%s', got '%s'", start, t.Value)
	}
	tokens := make([]*lexer.Token, 0)
	depth := 0
	for {
		t, err = p.NextValidToken()
		if err != nil {
			return nil, err
		}
		// Same as start, add to list, increase depth
		if string(t.Value) == start {
			tokens = append(tokens, t)
			depth++
		} else if string(t.Value) == end {
			// Found end token, but maybe it's matching another start
			if depth == 0 {
				return tokens, nil
			}
			tokens = append(tokens, t)
			depth--
		} else {
			tokens = append(tokens, t)
		}
	}
}

func (p *Parser) ParseIfStatement() (ast.Statement, error) {
	tokens, err := p.GetTokensBetween("(", ")")
	if err != nil {
		return nil, err
	}
	exp, tokens, err := p.ParseExpression(tokens)
	if err != nil {
		return nil, err
	}
	t, err := p.NextValidToken()
	if err != nil {
		return nil, err
	}
	s, err := p.ParseStatement(t)
	if err != nil {
		return nil, err
	}
	stmt, err := ast.NewIfStatement(exp, s, nil)
	if err != nil {
		return nil, err
	}
	t, err = p.PeekNextValidToken()
	if err != nil {
		return nil, err
	}
	if string(t.Value) != "else" {
		return stmt, nil
	}
	// Consume the "else" from the buffer
	t, err = p.NextValidToken()
	if err != nil {
		return nil, err
	}
	// Get the real next token
	t, err = p.NextValidToken()
	if err != nil {
		return nil, err
	}
	elseS, err := p.ParseStatement(t)
	if err != nil {
		return nil, err
	}
	stmt, err = ast.NewIfStatement(exp, s, elseS)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) ParseBlockItem(t *lexer.Token) (ast.Statement, error) {
	switch string(t.Value) {
	case "int":
		s, err := p.ParseDeclStatement(t)
		if err != nil {
			return nil, err
		}
		return s, nil
	default:
		s, err := p.ParseStatement(t)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
}

// ParseStatement will return the correct Statement for the tokens to follow
// It will get all tokens until the next ";"
// Right now only return statements exists
// <statement> ::= <return_statement>
func (p *Parser) ParseStatement(t *lexer.Token) (ast.Statement, error) {
	switch string(t.Value) {
	case "{":
		s, err := p.ParseBlockStatement()
		if err != nil {
			return nil, err
		}
		return s, nil
	case "if":
		s, err := p.ParseIfStatement()
		if err != nil {
			return nil, err
		}
		return s, nil
	case "return":
		s, err := p.ParseReturnStatement(t)
		if err != nil {
			return nil, err
		}
		return s, nil
	default:
		s, err := p.ParseExpressionStatement(t)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
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
	stmts, err := ast.NewStatementList()
	if err != nil {
		return nil, err
	}
	for {
		t, err := p.NextValidToken()
		if err != nil {
			return nil, err
		}
		// Found the end of the function
		if string(t.Value) == "}" {
			break
		}
		// This will get the function body as a block statement
		blockItem, err := p.ParseBlockItem(t)
		if err != nil {
			return nil, err
		}
		stmts, err = ast.AppendStatement(stmts, blockItem)
		if err != nil {
			return nil, err
		}
	}
	body, err := ast.NewBlockStatement(stmts)
	if err != nil {
		return nil, err
	}
	fun, err := ast.NewFunctionStatement(nameToken, nil, token, body)
	if err != nil {
		return nil, fmt.Errorf("Error")
	}
	return fun, nil
}

// GetTokensUntil will read the valid tokens from the lexer until it finds the token provided
func (p *Parser) GetTokensUntil(val string, include bool) ([]*lexer.Token, error) {
	tokens := make([]*lexer.Token, 0)
	for {
		t, err := p.NextValidToken()
		if err != nil {
			return nil, err
		}
		if string(t.Value) == val {
			if include {
				tokens = append(tokens, t)
			}
			break
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (p *Parser) GetStatementTokens() ([]*lexer.Token, error) {
	return p.GetTokensUntil(";", false)
}
