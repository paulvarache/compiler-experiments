package ast

import "compiler/lexer"

// Attrib represent any node of the tree
type Attrib interface{}

// Program is the rrot node
type Program struct {
	Statements []Statement `json:"statements"`
	Functions  []Statement `json:"functions"`
}

// Node represent an abstract node in the tree
type Node interface {
	TokenLiteral() string
}

// Statement is a statement node in the AST
type Statement interface {
	Node
	statementNode()
}

// Expression is an expression node in the tree
type Expression interface {
	Node
	expressionNode()
}

type Identifier struct {
	Token *lexer.Token `json:"-"`
	Value string       `json:"value"`
}

type DeclStatement struct {
	Token *lexer.Token `json:"-"`
	Left  Identifier   `json:"left"`
	Right Expression   `json:"right"`
	Type  string       `json:"type"`
}

type AssignExpression struct {
	Token    *lexer.Token `json:"-"`
	Operator string       `json:"operator"`
	Left     Identifier   `json:"left"`
	Right    Expression   `json:"right"`
}

type ExpStatement struct {
	Token      *lexer.Token `json:"-"`
	Expression Expression   `json:"expression"`
}

type FunctionStatement struct {
	Token      *lexer.Token    `json:"-"`
	Name       string          `json:"name"`
	Parameters []FormalArg     `json:"params"`
	Body       *BlockStatement `json:"body"`
	Return     string          `json:"return"`
}

type FormalArg struct {
	Arg  string `json:"arg"`
	Type string `json:"type"`
}

type ReturnStatement struct {
	Token       *lexer.Token `json:"-"`
	ReturnValue Expression   `json:"return"`
}

type BlockStatement struct {
	Token      *lexer.Token `json:"-"`
	Statements []Statement  `json:"statements"`
}

type IfStatement struct {
	Condition Expression `json:"condition"`
	Body      Statement  `json:"statement"`
	ElseBody  Statement  `json:"else"`
}

type IntegerLiteral struct {
	Token *lexer.Token `json:"-"`
	Value string       `json:"value"`
}

type PrefixExpression struct {
	Token      *lexer.Token `json:"-"`
	Operator   string       `json:"operator"`
	Expression Expression   `json:"expression"`
}

type InfixExpression struct {
	Token    *lexer.Token `json:"-"`
	Operator string       `json:"operator"`
	Left     Expression   `json:"left"`
	Right    Expression   `json:"right"`
}
