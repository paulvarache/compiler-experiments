package generator

import (
	"compiler/ast"
	"fmt"
	"strings"
)

func (g *AssemblyGenerator) FromIntegerLiteral(e ast.IntegerLiteral) (string, error) {
	return fmt.Sprintf("mov    $%s, %%rax    /* Push the int constant to the RAX register */", e.Value), nil
}

func (g *AssemblyGenerator) FromPrefixExpression(e ast.PrefixExpression) (string, error) {
	c, err := g.FromExpression(e.Expression)
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
func (g *AssemblyGenerator) GenerateAddAssembly(e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push   %%rax        /* Push the previous expression (e1) result to the RAX register */
%s
pop    %%rcx        /* Extract the second expression (e2) result from the RAX register */
add    %%rcx, %%rax  /* Add e1 and e2 and push it to the RAX register */
`, e1, e2))
}

// GenerateSubAssembly will output the string for an subtraction operation between two expressions
func (g *AssemblyGenerator) GenerateSubAssembly(e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push   %%rax        /* Push the previous expression (e2) result to the RAX register */
%s
pop    %%rcx        /* Extract the second expression (e1) result from the RAX register */
sub    %%rcx, %%rax  /* Subtract e2 from e1 and push it to the RAX register */
`, e2, e1))
}

// GenerateMultAssembly will output the assembly string multiplying two expressions
func (g *AssemblyGenerator) GenerateMultAssembly(e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push   %%rax        /* Push the previous expression (e1) result to the RAX register */
%s
pop    %%rcx        /* Extract the second expression (e2) result from the RAX register */
imul   %%rcx, %%rax  /* Multiply e1 and e2 and push it to the RAX register */
	`, e1, e2))
}

// GenerateDivAssembly will output the assembly string dividing two expressions
func (g *AssemblyGenerator) GenerateDivAssembly(e1 string, e2 string) string {
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
func (g *AssemblyGenerator) GenerateModuloAssembly(e1 string, e2 string) string {
	divSrc := g.GenerateDivAssembly(e1, e2)
	// Grab the division assembly code and move the value of rdx (remainder) to rax
	return strings.TrimSpace(fmt.Sprintf(`
%s
mov    %%rdx, %%rax /* Grab the remainder and move it to the rax register */
	`, divSrc))
}

func (g *AssemblyGenerator) GenerateLogicalAndAssembly(e1 string, e2 string) string {
	clauseName := g.LabelGenerator.GetNextLabel("clause")
	endName := g.LabelGenerator.GetNextLabel("end")
	return strings.TrimSpace(fmt.Sprintf(`
%s
cmp   $0, %%rax   /* Check if e1 is 0 */
jne   %s         /* If it is not, jump to the second clause */
jmp   %s
%s:
	%s
	cmp   $0, %%rax   /* Check if e2 is 0 */
	mov   $0, %%rax   /* reset rax, we'll use AL to set the return value */
	setne %%al
%s:
`, e1, clauseName, endName, clauseName, e2, endName))
}

func (g *AssemblyGenerator) GenerateLogicalOrAssembly(e1 string, e2 string) string {
	clauseName := g.LabelGenerator.GetNextLabel("clause")
	endName := g.LabelGenerator.GetNextLabel("end")
	return strings.TrimSpace(fmt.Sprintf(`
%s
cmp   $0, %%rax   /* Check if e1 is 0 */
je    %s          /* If it is, jump to the second clause */
mov   $1, %%rax   /* If it's not skip and return 1 */
jmp   %s          /* jump to the end */
%s:
	%s
	cmp   $0, %%rax   /* Check if e2 is 0 */
	mov   $0, %%rax   /* reset rax, we'll use AL to set the return value */
	setne %%al
%s:
`, e1, clauseName, endName, clauseName, e2, endName))
}

func (g *AssemblyGenerator) GenerateComparatorAssembly(instr string, e1 string, e2 string) string {
	return strings.TrimSpace(fmt.Sprintf(`
%s
push    %%rax
%s
pop		%%rcx
cmp    %%rax, %%rcx    /* Set ZF on if e1 == e2 otherwise off */
mov    $0, %%rax       /* Reset rax, does not affect ZF */
%s    %%al           /* Set the lower half of rax to 1 iff ZF is on */
	`, e1, e2, instr))
}

// GenerateFromInfixExpression outputs assembly for a InfixExpression node
func (g *AssemblyGenerator) FromInfixExpression(e ast.InfixExpression) (string, error) {
	// Extract the left operand
	l, err := g.FromExpression(e.Left)
	if err != nil {
		return "", err
	}
	// Extract the right operand
	r, err := g.FromExpression(e.Right)
	if err != nil {
		return "", err
	}
	switch e.Operator {
	case "+":
		return g.GenerateAddAssembly(l, r), nil
	case "*":
		return g.GenerateMultAssembly(l, r), nil
	case "-":
		return g.GenerateSubAssembly(l, r), nil
	case "/":
		return g.GenerateDivAssembly(l, r), nil
	case "%":
		return g.GenerateModuloAssembly(l, r), nil
	case "==":
		return g.GenerateComparatorAssembly("sete", l, r), nil
	case "!=":
		return g.GenerateComparatorAssembly("setne", l, r), nil
	case ">":
		return g.GenerateComparatorAssembly("setg", l, r), nil
	case ">=":
		return g.GenerateComparatorAssembly("setge", l, r), nil
	case "<":
		return g.GenerateComparatorAssembly("setl", l, r), nil
	case "<=":
		return g.GenerateComparatorAssembly("setle", l, r), nil
	case "&&":
		return g.GenerateLogicalAndAssembly(l, r), nil
	case "||":
		return g.GenerateLogicalOrAssembly(l, r), nil
	default:
		return "", fmt.Errorf("Unsupported infix operation with operator '%s'", e.Operator)
	}
}

func (g *AssemblyGenerator) FromExpression(e ast.Expression) (string, error) {
	switch e := e.(type) {
	case *ast.IntegerLiteral:
		return g.FromIntegerLiteral(*e)
	case *ast.PrefixExpression:
		return g.FromPrefixExpression(*e)
	case *ast.InfixExpression:
		return g.FromInfixExpression(*e)
	default:
		return "", fmt.Errorf("Failed with %s", e.TokenLiteral())
	}
}
