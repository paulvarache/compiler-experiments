package generator

import (
	"compiler/ast"
	"fmt"
	"strings"
)

func GenerateFromProgram(p *ast.Program) (string, error) {
	src := ""
	for _, fn := range p.Functions {
		fnString, err := GenerateFromStatement(fn)
		if err != nil {
			return "", err
		}
		src += fmt.Sprintf("%s\n", fnString)
	}
	return src, nil
}

func GenerateFromFunction(f ast.FunctionStatement) (string, error) {
	wrapper := `.globl %s
%s:
%s`
	body, err := GenerateFromStatement(f.Body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(wrapper, f.Name, f.Name, body), nil
}

func GenerateFromReturnStatement(r ast.ReturnStatement) (string, error) {
	expString, err := GenerateFromExpression(r.ReturnValue)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\nret", expString), nil
}

func GenerateFromBlockStatement(block ast.BlockStatement) (string, error) {
	src := ""
	for _, stmt := range block.Statements {
		stmtString, err := GenerateFromStatement(stmt)
		if err != nil {
			return "", err
		}
		src += fmt.Sprintf("%s\n", stmtString)
	}
	return src, nil
}

func GenerateFromStatement(s ast.Statement) (string, error) {
	switch s := s.(type) {
	case *ast.ReturnStatement:
		return GenerateFromReturnStatement(*s)
	case *ast.BlockStatement:
		return GenerateFromBlockStatement(*s)
	case *ast.FunctionStatement:
		return GenerateFromFunction(*s)
	default:
		return "", fmt.Errorf("Failed with %s", s.TokenLiteral())
	}
}

func GenerateFromIntegerLiteral(e ast.IntegerLiteral) (string, error) {
	return fmt.Sprintf("mov    $%s, %%rax    /* Push the int constant to the RAX register */", e.Value), nil
}

func GenerateFromPrefixExpression(e ast.PrefixExpression) (string, error) {
	c, err := GenerateFromExpression(e.Expression)
	if err != nil {
		return "", err
	}
	if e.Operator == "-" || e.Operator == "~" {
		return fmt.Sprintf(`%s
neg    %%rax    /* Negates the value in RAX */`, c), nil
	} else if e.Operator == "!" {
		return fmt.Sprintf(`%s
cmp    $0, %%rax  /* Set ZF to 0 if expression is equal to 0 */
mov    $0, %%rax  /* Clear the RAX register */
sete   %%al       /* Set the AL register (low bit of RAX) to value of ZF */`, c), nil
	}
	return "", fmt.Errorf("Could not generate. Operator '%s' is not supported", e.Operator)
}

// GenerateAddAssembly will output the string for an addition operation between two expressions
func GenerateAddAssembly(e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push   %%rax        /* Push the previous expression (e1) result to the RAX register */
%s
pop    %%rcx        /* Extract the second expression (e2) result from the RAX register */
add    %%rcx, %%rax  /* Add e1 and e2 and push it to the RAX register */
`, e1, e2))
}

// GenerateSubAssembly will output the string for an subtraction operation between two expressions
func GenerateSubAssembly(e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push   %%rax        /* Push the previous expression (e2) result to the RAX register */
%s
pop    %%rcx        /* Extract the second expression (e1) result from the RAX register */
sub    %%rcx, %%rax  /* Subtract e2 from e1 and push it to the RAX register */
`, e2, e1))
}

// GenerateMultAssembly will output the assembly string multiplying two expressions
func GenerateMultAssembly(e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push   %%rax        /* Push the previous expression (e1) result to the RAX register */
%s
pop    %%rcx        /* Extract the second expression (e2) result from the RAX register */
imul   %%rcx, %%rax  /* Multiply e1 and e2 and push it to the RAX register */
	`, e1, e2))
}

// GenerateDivAssembly will output the assembly string dividing two expressions
func GenerateDivAssembly(e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push   %%rax       /* Push (e2) to the stack */
%s
cdq                /* Expand rax (e1) into rdx to become a 64bit number */
pop    %%rcx        /* Grab e2 form the stack to become the dividend */
div    %%rcx        /* Divide (e1) by (e2), sending the quotient to rax */
	`, e2, e1))
}

// GenerateModuloAssembly will output the assembly string dividing two expressions
func GenerateModuloAssembly(e1 string, e2 string) string {
	divSrc := GenerateDivAssembly(e1, e2)
	// Grab the division assembly code and move the value of rdx (remainder) to rax
	return strings.TrimSpace(fmt.Sprintf(`
%s
mov    %%rdx, %%rax /* Grab the remainder and move it to the rax register */
	`, divSrc))
}

// GenerateFromInfixExpression outputs assembly for a InfixExpression node
func GenerateFromInfixExpression(e ast.InfixExpression) (string, error) {
	// Extract the left operand
	l, err := GenerateFromExpression(e.Left)
	if err != nil {
		return "", err
	}
	// Extract the right operand
	r, err := GenerateFromExpression(e.Right)
	if err != nil {
		return "", err
	}
	switch e.Operator {
	case "+":
		return GenerateAddAssembly(l, r), nil
	case "*":
		return GenerateMultAssembly(l, r), nil
	case "-":
		return GenerateSubAssembly(l, r), nil
	case "/":
		return GenerateDivAssembly(l, r), nil
	case "%":
		return GenerateModuloAssembly(l, r), nil
	default:
		return "", fmt.Errorf("Unsupported infix operation with operator '%s'", e.Operator)
	}
}

func GenerateFromExpression(e ast.Expression) (string, error) {
	switch e := e.(type) {
	case *ast.IntegerLiteral:
		return GenerateFromIntegerLiteral(*e)
	case *ast.PrefixExpression:
		return GenerateFromPrefixExpression(*e)
	case *ast.InfixExpression:
		return GenerateFromInfixExpression(*e)
	default:
		return "", fmt.Errorf("Failed with %s", e.TokenLiteral())
	}
}
