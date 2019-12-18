package ast

import (
	"compiler/lexer"
	"fmt"
)

func (p *Program) String() string {
	return "Program"
}

func (ds DeclStatement) statementNode()       {}
func (ds DeclStatement) TokenLiteral() string { return "DeclStatement" }

func (es ExpStatement) statementNode()       {}
func (es ExpStatement) TokenLiteral() string { return "ExpStatement" }

func (ae AssignExpression) expressionNode()      {}
func (ae AssignExpression) TokenLiteral() string { return "AssignExpression" }

func (fs FunctionStatement) statementNode()       {}
func (fs FunctionStatement) TokenLiteral() string { return "FunctionStatement" }

func (bs BlockStatement) statementNode()       {}
func (bs BlockStatement) TokenLiteral() string { return "BlockStatement" }

func (rs ReturnStatement) statementNode()       {}
func (rs ReturnStatement) TokenLiteral() string { return "ReturnStatement" }

func (is IfStatement) statementNode()       {}
func (is IfStatement) TokenLiteral() string { return "IfStatement" }

func (il IntegerLiteral) expressionNode()      {}
func (il IntegerLiteral) TokenLiteral() string { return "IntegerLiteral" }

func (pe PrefixExpression) expressionNode()      {}
func (pe PrefixExpression) TokenLiteral() string { return "PrefixExpression" }

func (ie InfixExpression) expressionNode()      {}
func (ie InfixExpression) TokenLiteral() string { return "InfixExpression" }

func (i Identifier) expressionNode()      {}
func (i Identifier) TokenLiteral() string { return "Identifier" }

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

func NewDeclStatement(varType, left, right Attrib) (Statement, error) {
	t, ok := varType.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewDeclStatement", "*lexer.Token", "varType", varType)
	}
	l, ok := left.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewDeclStatement", "*lexer.Token", "left", left)
	}
	id := Identifier{Token: l, Value: string(l.Value)}
	stmt := &DeclStatement{Token: t, Left: id, Type: string(t.Value)}
	if right == nil {
		return stmt, nil
	}
	r, ok := right.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewDeclStatement", "Expression", "right", right)
	}
	stmt.Right = r
	return stmt, nil
}

func NewAssignExpression(operator, left, right Attrib) (Expression, error) {
	op, ok := operator.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewAssignStatement", "*lexer.Token", "operator", operator)
	}
	l, ok := left.(*lexer.Token)
	if !ok {
		return nil, fmt.Errorf("NewAssignStatement", "*lexer.Token", "left", left)
	}
	r, ok := right.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewAssignStatement", "Expression", "right", right)
	}
	return &AssignExpression{Token: l, Operator: string(op.Value), Left: Identifier{Token: l, Value: string(l.Value)}, Right: r}, nil
}

func NewIdentifier(id *lexer.Token) Expression {
	return &Identifier{Token: id, Value: string(id.Value)}
}

func NewExpStatement(exp Attrib) (Statement, error) {
	e, ok := exp.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewExpStatement", "Expression", "exp", exp)
	}
	return &ExpStatement{Expression: e}, nil
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
		return nil, fmt.Errorf("NewPrefixExpression", "Expression", "expression", expression)
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

func NewIfStatement(cond Attrib, body Attrib, elseBody Attrib) (Statement, error) {
	c, ok := cond.(Expression)
	if !ok {
		return nil, fmt.Errorf("NewIfStatement", c)
	}
	b, ok := body.(Statement)
	if !ok {
		return nil, fmt.Errorf("NewIfStatement", b)
	}
	stmt := &IfStatement{Condition: c, Body: b}
	if elseBody == nil {
		return stmt, nil
	}
	e, ok := elseBody.(Statement)
	if !ok {
		return nil, fmt.Errorf("Fail")
	}
	stmt.ElseBody = e
	return stmt, nil
}
