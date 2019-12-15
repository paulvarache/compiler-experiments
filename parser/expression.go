package parser

import (
	"compiler/ast"
	"compiler/lexer"
	"fmt"
)

// ParseExpression returns an expression tree based on provided tokens
// Returns an Expression, a new array of left over tokens and an optional error
// Handles the following scenario
// <exp> ::= <term> { ("+" | "-") <term> }
func (p *Parser) ParseExpression(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	// First tokens have to be a term
	exp, tokens, err := p.ParseTerm(tokens)
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
		s := string(t.Value)
		// Not a valid character, exit
		if s != "+" && s != "-" {
			break
		}
		// Remove the accepted token
		tokens = tokens[1:]
		// Prepare a variable to host the expression
		// This is to make sure we can use the = operator and not :=
		// The second one would create a copy of the tokens array in the scope
		var nextTerm ast.Expression
		nextTerm, tokens, err = p.ParseTerm(tokens)
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

// ParseTerm will build and expression matching the following grammar for given tokens
// <term> ::= <factor> { ("*" | "/") <factor> }
// It will return an Expression, the remaining tokens and an optional error
func (p *Parser) ParseTerm(tokens []*lexer.Token) (ast.Expression, []*lexer.Token, error) {
	// Extract the non-optional factor
	exp, tokens, err := p.ParseFactor(tokens)
	if err != nil {
		return nil, tokens, err
	}
	// Loop to get every following factors
	for {
		// No more tokens, reached the end
		if len(tokens) == 0 {
			return exp, tokens, nil
		}
		// Extract the first token
		t := tokens[0]
		s := string(t.Value)
		// Not a valid operator for the infix expression,
		// This is the end of the current term
		if s != "*" && s != "/" && s != "%" {
			break
		}
		// Remove the accepted token
		tokens = tokens[1:]
		// Parse the expected following factor
		var nextFact ast.Expression
		nextFact, tokens, err = p.ParseFactor(tokens)
		if err != nil {
			return nil, tokens, err
		}
		// Build the infix expression with the two factors
		exp, err = ast.NewInfixExpression(t, exp, nextFact)
		if err != nil {
			return nil, tokens, err
		}
	}
	// Return the accepted expression
	return exp, tokens, nil
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
