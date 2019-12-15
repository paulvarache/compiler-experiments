package lexer

import "strconv"

// TokenState represents a state the lexer can be while reading the source
type TokenState uint32

// Holds all the possible states for the lexer
const (
	ExprState TokenState = iota
	StmtParensState
	SubscriptState
	PropNameState
)

// ParsingContext represent the context the lexer can be in during parsing
// This handles block statements and nested parenthesis
type ParsingContext uint32

// All the possible contexts
const (
	GlobalContext ParsingContext = iota
	StmtParensContext
	ExprParensContext
	BracesContext
	TemplateContext
)

// TokenType represents the type of a token
type TokenType uint32

// Lists all possible tokens
const (
	ErrToken TokenType = iota
	UnknownToken
	IdentifierToken
	WhitespaceToken
	PunctuatorToken
	TemplateToken
	LineTerminatorToken
	NumericToken
)

func (tt TokenType) String() string {
	switch tt {
	case ErrToken:
		return "Error"
	case UnknownToken:
		return "Unknown"
	case IdentifierToken:
		return "Identifier"
	case WhitespaceToken:
		return "Whitespace"
	case PunctuatorToken:
		return "Punctuator"
	case TemplateToken:
		return "Template"
	case LineTerminatorToken:
		return "LineTerminator"
	case NumericToken:
		return "Numeric"
	}
	return "Invalid(" + strconv.Itoa(int(tt)) + ")"
}

// Token represents a found token with a type and its value
type Token struct {
	Type  TokenType
	Value []byte
}
