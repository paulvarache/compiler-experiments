package generator

import (
	"compiler/ast"
	"fmt"
)

func (g *AssemblyGenerator) FromIntegerLiteral(e ast.IntegerLiteral) {
	g.AddLine("mov", fmt.Sprintf("$%s, %%rax", e.Value), "/* Push the int constant to the RAX register */")
}

func (g *AssemblyGenerator) FromPrefixExpression(e ast.PrefixExpression) error {
	err := g.FromExpression(e.Expression)
	if err != nil {
		return err
	}
	if e.Operator == "-" || e.Operator == "~" {
		g.AddLine("eg", "%rax", "/* Negates the value in RAX */")
		return nil
	} else if e.Operator == "!" {
		g.AddLine("cmp", "$0, %rax", "/* Set ZF to 0 if expression is equal to 0 */")
		g.AddLine("mov", "$0, %rax", "/* Clear the EAX register */")
		g.AddLine("sete", "%al", "/* Set the AL register to the value in ZF */")
		return nil
	}
	return fmt.Errorf("Could not generate. Operator '%s' is not supported", e.Operator)
}

// GenerateAddAssembly will output the string for an addition operation between two expressions
func (g *AssemblyGenerator) GenerateAddAssembly(e1 ast.Expression, e2 ast.Expression) error {
	err := g.FromExpression(e1)
	if err != nil {
		return err
	}
	g.AddLine("push", "%rax", "/* Push the previous expression (e1) result to the RAX register */")
	err = g.FromExpression(e2)
	if err != nil {
		return err
	}
	g.AddLine("pop", "%rcx", "/* Extract the second expression (e2) result from the RAX register */")
	g.AddLine("add", "%rcx, %rax", "/* Add e1 and e2 and push it to the RAX register */")
	return nil
}

// GenerateSubAssembly will output the string for an subtraction operation between two expressions
func (g *AssemblyGenerator) GenerateSubAssembly(e1 ast.Expression, e2 ast.Expression) error {
	err := g.FromExpression(e2)
	if err != nil {
		return err
	}
	g.AddLine("push", "%rax", "/* Push the previous expression (e2) result to the stack */")
	err = g.FromExpression(e1)
	if err != nil {
		return err
	}
	g.AddLine("pop", "%rcx", "/* Extract the second expression (e2) result from the the stack onto RCX */")
	g.AddLine("sub", "%rcx, %rax", "/* Subtract e2 from e1 and push it to the RAX register */")
	return nil
}

// GenerateMultAssembly will output the assembly string multiplying two expressions
func (g *AssemblyGenerator) GenerateMultAssembly(e1 ast.Expression, e2 ast.Expression) error {
	err := g.FromExpression(e1)
	if err != nil {
		return err
	}
	g.AddLine("push", "%rax", "/* Push the previous expression (e1) result to the RAX register */")
	err = g.FromExpression(e2)
	if err != nil {
		return err
	}
	g.AddLine("pop", "%rcx", "/* Extract the second expression (e2) result from the RAX register */")
	g.AddLine("imul", "%rcx, %rax", "/* Multiply e1 and e2 and push it to the RAX register */")
	return nil
}

// GenerateDivAssembly will output the assembly string dividing two expressions
func (g *AssemblyGenerator) GenerateDivAssembly(e1 ast.Expression, e2 ast.Expression) error {
	err := g.FromExpression(e2)
	if err != nil {
		return err
	}
	g.AddLine("push", "%rax", "/* Push (e2) to the stack */")
	err = g.FromExpression(e1)
	if err != nil {
		return err
	}
	g.AddLine("cdq", "/* Expand rax (e1) into rdx to become a 64bit number */")
	g.AddLine("pop", "%rcx", "/* Grab e2 form the stack to become the dividend */")
	g.AddLine("div", "%rcx", "/* Divide (e1) by (e2), sending the quotient to rax */")
	return nil
}

// GenerateModuloAssembly will output the assembly string dividing two expressions
func (g *AssemblyGenerator) GenerateModuloAssembly(e1 ast.Expression, e2 ast.Expression) error {
	err := g.GenerateDivAssembly(e1, e2)
	if err != nil {
		return err
	}
	// Grab the division assembly code and move the value of rdx (remainder) to rax
	g.AddLine("mov", "%rdx, %rax", "/* Grab the remainder and move it to the rax register */")
	return nil
}

func (g *AssemblyGenerator) GenerateLogicalAndAssembly(e1 ast.Expression, e2 ast.Expression) error {
	clauseName := g.LabelGenerator.GetNextLabel("clause")
	endName := g.LabelGenerator.GetNextLabel("end")
	err := g.FromExpression(e1)
	if err != nil {
		return err
	}
	g.AddLine("cmp", "$0, %rax", "/* Check if e1 is 0 */")
	g.AddLine("jne", clauseName, "/* If it is not, jump to the second clause */")
	g.AddLine("jmp", endName, "/* Go to end otherwise */")
	g.LeaveContext()
	g.AddLine(fmt.Sprintf("%s:", clauseName))
	g.EnterContext()
	err = g.FromExpression(e2)
	if err != nil {
		return err
	}
	g.AddLine("cmp", "$0, %rax", "/* Check if e2 is 0 */")
	g.AddLine("mov", "$0, %rax", "/* Reset rax, we'll use AL */")
	g.AddLine("setne", "%al", "/* Set result of cmp to al */")
	g.LeaveContext()
	g.AddLine(fmt.Sprintf("%s:", endName))
	g.EnterContext()
	return nil
}

func (g *AssemblyGenerator) GenerateLogicalOrAssembly(e1 ast.Expression, e2 ast.Expression) error {
	clauseName := g.LabelGenerator.GetNextLabel("clause")
	endName := g.LabelGenerator.GetNextLabel("end")
	err := g.FromExpression(e1)
	if err != nil {
		return err
	}
	g.AddLine("cmp", "$0, %rax", "/* Check if e1 is 0 */")
	g.AddLine("je", clauseName, "/* If it is not, jump to the second clause */")
	g.AddLine("mov", "$1, %rax", "/* If not true, skip and return 1 */")
	g.AddLine("jmp", endName, "/* Go to end */")
	g.LeaveContext()
	g.AddLine(fmt.Sprintf("%s:", clauseName))
	g.EnterContext()
	err = g.FromExpression(e2)
	if err != nil {
		return err
	}
	g.AddLine("cmp", "$0, %rax", "/* Check if e2 is 0 */")
	g.AddLine("mov", "$0, %rax", "/* Reset rax, we'll use AL */")
	g.AddLine("setne", "%al", "/* Set result of cmp to al */")
	g.LeaveContext()
	g.AddLine(fmt.Sprintf("%s:", endName))
	g.EnterContext()
	return nil
}

func (g *AssemblyGenerator) GenerateComparatorAssembly(instr string, e1 ast.Expression, e2 ast.Expression) error {
	err := g.FromExpression(e1)
	if err != nil {
		return err
	}
	g.AddLine("push", "%rax")
	err = g.FromExpression(e2)
	if err != nil {
		return err
	}
	g.AddLine("pop", "%rcx")
	g.AddLine("cmp", "%rax, %rcx", "/* Set ZF on if e1 == e2 otherwise off */")
	g.AddLine("mov", "$0, %rax", "/* Reset rax, does not affect ZF */")
	g.AddLine(instr, "%al", "/* Set the lower half of rax to 1 based on ZF */")
	return nil
}

func (g *AssemblyGenerator) FromAssignExpression(e ast.AssignExpression) error {
	err := g.FromExpression(e.Right)
	if err != nil {
		return err
	}
	stackIndex, err := g.Variables.GetVariableStackIndex(e.Left.Value)
	if err != nil {
		return err
	}
	switch e.Operator {
	case "":
		break
	case "+=":
		g.AddLine("add", fmt.Sprintf("%d(%%rbp), %%rax", stackIndex), "/* Add the expression result and the variable. add puch to RAX */")
		break
	case "-=":
		g.AddLine("push", "%rax", "/* Stack the value to subtract */")
		g.AddLine("mov", fmt.Sprintf("%d(%%rbp), %%rax", stackIndex), "/* Move the variable into RAX */")
		g.AddLine("pop", "%rcx", "/* Move the value to subtract into RCX */")
		g.AddLine("sub", "%rcx, %rax", "/* Subtract RCX from the variable */")
		break
	case "*=":
		g.AddLine("push", "%rax", "/* Stack the multiplier */")
		g.AddLine("mov", fmt.Sprintf("%d(%%rbp), %%rax", stackIndex), "/* Move the variable into RAX */")
		g.AddLine("pop", "%rcx", "/* Move the multiplier into RCX */")
		g.AddLine("imul", "%rcx, %rax", "/* Multiply the var by the multipler */")
		break
	case "/=":
		g.AddLine("push", "%rax", "/* Stack the divisor */")
		g.AddLine("mov", fmt.Sprintf("%d(%%rbp), %%rax", stackIndex), "/* Move the variable into RAX */")
		g.AddLine("cdq", "/* Expand RAX into RDX */")
		g.AddLine("pop", "%rcx", "/* Move the divisor into RCX */")
		g.AddLine("div", "%rcx", "/* Divide the var by the divisor in RAX:RDX */")
		break
	default:
		fmt.Errorf("Expected a valid assignment operator, got '%s'", e.Operator)
	}
	// Always move the result into the variable
	g.AddLine("mov", fmt.Sprintf("%%rax, %d(%%rbp)", stackIndex), "/* Move the result into the variable */")
	return nil
}

func (g *AssemblyGenerator) FromIdentifier(i ast.Identifier) error {
	stackIndex, err := g.Variables.GetVariableStackIndex(i.Value)
	if err != nil {
		return err
	}
	g.AddLine("mov", fmt.Sprintf("%d(%%rbp), %%rax", stackIndex), "/* Move the variable into the rax register */")
	return nil
}

// GenerateFromInfixExpression outputs assembly for a InfixExpression node
func (g *AssemblyGenerator) FromInfixExpression(e ast.InfixExpression) error {
	l, r := e.Left, e.Right
	switch e.Operator {
	case "+":
		return g.GenerateAddAssembly(l, r)
	case "*":
		return g.GenerateMultAssembly(l, r)
	case "-":
		return g.GenerateSubAssembly(l, r)
	case "/":
		return g.GenerateDivAssembly(l, r)
	case "%":
		return g.GenerateModuloAssembly(l, r)
	case "==":
		return g.GenerateComparatorAssembly("sete", l, r)
	case "!=":
		return g.GenerateComparatorAssembly("setne", l, r)
	case ">":
		return g.GenerateComparatorAssembly("setg", l, r)
	case ">=":
		return g.GenerateComparatorAssembly("setge", l, r)
	case "<":
		return g.GenerateComparatorAssembly("setl", l, r)
	case "<=":
		return g.GenerateComparatorAssembly("setle", l, r)
	case "&&":
		return g.GenerateLogicalAndAssembly(l, r)
	case "||":
		return g.GenerateLogicalOrAssembly(l, r)
	default:
		return fmt.Errorf("Unsupported infix operation with operator '%s'", e.Operator)
	}
}

func (g *AssemblyGenerator) FromExpression(e ast.Expression) error {
	switch e := e.(type) {
	case *ast.IntegerLiteral:
		g.FromIntegerLiteral(*e)
		return nil
	case *ast.PrefixExpression:
		return g.FromPrefixExpression(*e)
	case *ast.InfixExpression:
		return g.FromInfixExpression(*e)
	case *ast.AssignExpression:
		return g.FromAssignExpression(*e)
	case *ast.Identifier:
		return g.FromIdentifier(*e)
	default:
		return fmt.Errorf("Failed with %s", e.TokenLiteral())
	}
}
