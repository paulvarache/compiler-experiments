package ast

import (
	"compiler/lexer"
	"fmt"
)

func (p *Program) String() string {
	return "Program"
}

func (fs FunctionStatement) statementNode()       {}
func (fs FunctionStatement) TokenLiteral() string { return "FunctionStatement" }

func (bs BlockStatement) statementNode()       {}
func (bs BlockStatement) TokenLiteral() string { return "BlockStatement" }

func (rs ReturnStatement) statementNode()       {}
func (rs ReturnStatement) TokenLiteral() string { return "ReturnStatement" }

func (il IntegerLiteral) expressionNode()      {}
func (il IntegerLiteral) TokenLiteral() string { return "IntegerLiteral" }

func (pe PrefixExpression) expressionNode()      {}
func (pe PrefixExpression) TokenLiteral() string { return "PrefixExpression" }

func (ie InfixExpression) expressionNode()      {}
func (ie InfixExpression) TokenLiteral() string { return "InfixExpression" }

func NewProgram(funcs, stmts Attrib) (*Program, error) {
	s, ok := stmts.([]Statement)
	if !ok {
		return nil, fmt.Errorf("NewProgram", "[]Statement", "stmts", stmts)
	}
	f, ok := funcs.([]Statement)
	if !ok {
		return nil, fmt.Errorf("NewProgram", "[]Statement", "funcs", funcs)
	}
	return &Program{Functions: f, Statements: s}, nil
}

func NewStatementList() ([]Statement, error) {
	return []Statement{}, nil
}

func AppendStatement(stmtList, stmt Attrib) ([]Statement, error) {
	s, ok := stmt.(Statement)
	if !ok {
		return nil, fmt.Errorf("AppendStatement", "Statement", "stmt", stmt)
	}
	return append(stmtList.([]Statement), s), nil
}

func NewBlockStatement(stmts Attrib) (*BlockStatement, error) {
	s, ok := stmts.([]Statement)
	if !ok {
		return nil, fmt.Errorf("NewBlockStatement", "[]Statement", "stmts", stmts)
	}
	return &BlockStatement{Statements: s}, nil
}

func NewReturnStatement(exp Attrib) (Statement, error) {
	e, ok := exp.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewReturnStatement", "Expression", "exp", exp)
	}
	return &ReturnStatement{ReturnValue: e}, nil
}

func NewIntegerLiteral(integer Attrib) (*IntegerLiteral, error) {
	intLit, ok := integer.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewIntegerLiteral", "*lexer.Token", "integer", integer)
	}
	return &IntegerLiteral{Token: intLit, Value: string(intLit.Value)}, nil
}

func NewPrefixExpression(operator, expression Attrib) (*PrefixExpression, error) {
	op, ok := operator.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewPrefixExpression", "*lexer.Token", "operator", operator)
	}
	exp, ok := expression.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewPrefixExpression", "*lexer.Token", "expression", expression)
	}
	return &PrefixExpression{Operator: string(op.Value), Expression: exp}, nil
}

func NewInfixExpression(operator, left Attrib, right Attrib) (*InfixExpression, error) {
	op, ok := operator.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewInfixExpression", "*lexer.Token", "operator", operator)
	}
	l, ok := left.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewInfixExpression", "*lexer.Token", "left", left)
	}
	r, ok := right.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewInfixExpression", "*lexer.Token", "right", right)
	}
	return &InfixExpression{Operator: string(op.Value), Left: l, Right: r}, nil
}

func NewFunctionStatement(name, args, ret, block Attrib) (Statement, error) {
	n, ok := name.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewFunctionStatement", "*lexer.Token", "name", name)
	}
	b, ok := block.(*BlockStatement)
	if !ok {
		return nil, fmt.Errorf("NewFunctionStatement", "*BlockStatement", "block", block)
	}
	a := []FormalArg{}
	if args != nil {
		a, ok = args.([]FormalArg)
		if !ok {
			return nil, fmt.Errorf("NewFunctionStatement", "[]FormalArg", "args", args)
		}
	}

	r, ok := ret.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewFunctionStatement", "*lexer.Token", "ret", ret)
	}
	return &FunctionStatement{Name: string(n.Value), Body: b, Parameters: a, Return: string(r.Value)}, nil
}
