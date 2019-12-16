package parser

import (
	"compiler/ast"
	"compiler/lexer"
	"fmt"
)

// ExpressionParser parses an expression
type ExpressionParser interface {
	parseExpression(p *Parser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error)
	isValidToken(token *lexer.Token) bool
}

// AdditiveExpressionParser handles
// <additive_exp> ::= <term> { ("+" | "-") <term> }
type AdditiveExpressionParser struct{}

func (lep AdditiveExpressionParser) parseExpression(p *Parser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.ParseTerm(tokens)
}

func (lep AdditiveExpressionParser) isValidToken(token *lexer.Token) bool {
	s := string(token.Value)
	return s == "+" || s == "-"
}

// RelationalExpressionParser handles
type RelationalExpressionParser struct{}

func (rep RelationalExpressionParser) parseExpression(p *Parser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.ParseAdditiveExpression(tokens)
}

func (rep RelationalExpressionParser) isValidToken(token *lexer.Token) bool {
	s := string(token.Value)
	return s == "<" || s == ">" || s == "<=" || s == ">="
}

// EqualityExpressionParser handles
type EqualityExpressionParser struct{}

func (eep EqualityExpressionParser) parseExpression(p *Parser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.ParseRelationalExpression(tokens)
}

func (eep EqualityExpressionParser) isValidToken(token *lexer.Token) bool {
	s := string(token.Value)
	return s == "!=" || s == "=="
}

// LogicalAndExpressionParser handles
type LogicalAndExpressionParser struct{}

func (laep LogicalAndExpressionParser) parseExpression(p *Parser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.ParseEqualityExpression(tokens)
}

func (laep LogicalAndExpressionParser) isValidToken(token *lexer.Token) bool {
	s := string(token.Value)
	return s == "&&"
}

// LogicalOrExpressionParser handles
type LogicalOrExpressionParser struct{}

func (loep LogicalOrExpressionParser) parseExpression(p *Parser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.ParseLogicalAndExpression(tokens)
}

func (loep LogicalOrExpressionParser) isValidToken(token *lexer.Token) bool {
	s := string(token.Value)
	return s == "||"
}

// TermExpressionParser handles
type TermExpressionParser struct{}

func (tep TermExpressionParser) parseExpression(p *Parser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.ParseFactor(tokens)
}

func (tep TermExpressionParser) isValidToken(token *lexer.Token) bool {
	s := string(token.Value)
	return s == "*" || s == "/" || s == "%"
}

// GetExpressionFromParser will return the expression for a given set of tokens and a parser
func (p *Parser) GetExpressionFromParser(parser ExpressionParser, tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	// First tokens have to be a term
	exp, tokens, err := parser.parseExpression(p, tokens)
	if err != nil {
		return nil, tokens, err
	}
	// There will be more terms if the following tokens are + or -
	for {
		// Ignore if we reached the end
		if len(tokens) == 0 {
			return exp, tokens, nil
		}
		// Get the first token
		t := tokens[0]
		isValid := parser.isValidToken(t)
		if !isValid {
			break
		}
		// Remove the accepted token
		tokens = tokens[1:]
		// Prepare a variable to host the expression
		// This is to make sure we can use the = operator and not :=
		// The second one would create a copy of the tokens array in the scope
		var nextTerm ast.Expression
		nextTerm, tokens, err = parser.parseExpression(p, tokens)
		if err != nil {
			return nil, tokens, err
		}
		// Got all the valid bits, build the infox expression
		exp, err = ast.NewInfixExpression(t, exp, nextTerm)
		if err != nil {
			return nil, tokens, err
		}
	}
	return exp, tokens, nil
}

// ParseExpression parses the grammar as follow
// <exp> ::= <id> "=" <exp> | <logical_or_exp>
func (p *Parser) ParseExpression(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	tName := tokens[0]
	if tName.Type == lexer.IdentifierToken {
		if len(tokens) == 1 {
			// No more token, it's a lonely identifier
			return p.ParseLogicalOrExpression(tokens)
		}
		tokens := tokens[1:]
		tEq := tokens[0]
		if string(tEq.Value) != "=" {
			return nil, tokens, fmt.Errorf("Expected '=' got '%s'", tEq.Value)
		}
		// consume the "="
		tokens = tokens[1:]
		exp, tokens, err := p.ParseExpression(tokens)
		if err != nil {
			return nil, tokens, err
		}
		nextExp, err := ast.NewAssignExpression(tName, exp)
		if err != nil {
			return nil, tokens, err
		}
		return nextExp, tokens, nil
	}
	return p.ParseLogicalOrExpression(tokens)
}

// ParseLogicalOrExpression parses the grammar as follow
// <logical_or_exp> ::= <logical_and_exp> { "||" <logical_and_exp> }
func (p *Parser) ParseLogicalOrExpression(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.GetExpressionFromParser(&LogicalOrExpressionParser{}, tokens)
}

// ParseLogicalAndExpression parses the grammar as follow
// <logical_and_exp> ::= <equality_exp> { "&&" <equality_exp> }
func (p *Parser) ParseLogicalAndExpression(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.GetExpressionFromParser(&LogicalAndExpressionParser{}, tokens)
}

// ParseEqualityExpression parses the grammar as follow
// <equality_exp> ::= <relational_exp> { ("!=" | "==") <relational_exp> }
func (p *Parser) ParseEqualityExpression(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.GetExpressionFromParser(&EqualityExpressionParser{}, tokens)
}

// ParseRelationalExpression parses the grammar as follow
// <relational_exp> ::= <additive_exp> { ("<" | ">" | "<=" | ">=") <additive_exp> }
func (p *Parser) ParseRelationalExpression(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.GetExpressionFromParser(&RelationalExpressionParser{}, tokens)
}

// ParseAdditiveExpression returns an expression tree based on provided tokens
// Returns an Expression, a new array of left over tokens and an optional error
// Handles the following scenario
// <additive_exp> ::= <term> { ("+" | "-") <term> }
func (p *Parser) ParseAdditiveExpression(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.GetExpressionFromParser(&AdditiveExpressionParser{}, tokens)
}

// ParseTerm will build and expression matching the following grammar for given tokens
// <term> ::= <factor> { ("*" | "/") <factor> }
// It will return an Expression, the remaining tokens and an optional error
func (p *Parser) ParseTerm(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	return p.GetExpressionFromParser(&TermExpressionParser{}, tokens)
}

// ParseFactor will return an Expression and the remaining tokens
// for a given array of tokens following this grammar
// <factor> ::= "(" <exp> ")" | <unary_op> <factor> | <const>
func (p *Parser) ParseFactor(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	// Extract the first token to try to match one of the 3 options
	t := tokens[0]
	var exp ast.Expression
	var err error
	if string(t.Value) == "(" {
		// Matches the expression in parenthesis
		// "(" <exp> ")"
		tokens = tokens[1:]
		exp, tokens, err = p.ParseExpression(tokens)
		if err != nil {
			return nil, tokens, err
		}
		t = tokens[0]
		if string(t.Value) != ")" {
			return nil, tokens, fmt.Errorf("Expected ')' got '%s'", t.Value)
		}
		tokens = tokens[1:]
		return exp, tokens, nil
	} else if IsUnaryOp(t) {
		// Matches an unary operation
		// <unary_op> <factor>
		tokens = tokens[1:]
		var fact ast.Expression
		fact, tokens, err = p.ParseFactor(tokens)
		if err != nil {
			return nil, tokens, err
		}
		exp, err = ast.NewPrefixExpression(t, fact)
		if err != nil {
			return nil, tokens, err
		}
		return exp, tokens, nil
	} else if IsConstant(t) {
		// Matches a constant
		// <const>
		tokens = tokens[1:]
		intLit, err := ast.NewIntegerLiteral(t)
		if err != nil {
			return nil, tokens, err
		}
		return intLit, tokens, nil
	} else if t.Type == lexer.IdentifierToken {
		tokens = tokens[1:]
		id := ast.NewIdentifier(t)
		return id, tokens, nil
	} else {
		return nil, tokens, fmt.Errorf("Failed to parse factor. Unexpected token %s '%s'", t, t.Value)
	}
}

// IsUnaryOp will return a boolean indicating whether or not a given token
// starts an unary operation
func IsUnaryOp(t *lexer.Token) bool {
	s := string(t.Value)
	return s == "-" || s == "!" || s == "~"
}

// IsConstant will returna boolean indicating whether or not a given token
// starts a constant
func IsConstant(t *lexer.Token) bool {
	// TODO: support more than just numbers
	return t.Type == lexer.NumericToken
}
