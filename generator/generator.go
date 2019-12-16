package generator

import (
	"compiler/ast"
	"fmt"
	"strings"
)

type AssemblyGenerator struct {
	LabelGenerator *LabelGenerator
}

func NewAssemblyGenerator() *AssemblyGenerator {
	return &AssemblyGenerator{LabelGenerator: &LabelGenerator{}}
}

func (g *AssemblyGenerator) FromProgram(p *ast.Program) (string, error) {
	src := ""
	for _, fn := range p.Functions {
		fnString, err := g.FromStatement(fn)
		if err != nil {
			return "", err
		}
		src += fmt.Sprintf("%s\n", fnString)
	}
	return src, nil
}

func (g *AssemblyGenerator) FromFunction(f ast.FunctionStatement) (string, error) {
	wrapper := `.globl %s
%s:
push    %%rbp         /* Save value of the bottom of the current frame */
mov     %%rsp, %%rbp  /* Top of stack is now bottom of new frame */
%s`
	body, err := g.FromStatement(f.Body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(wrapper, f.Name, f.Name, body), nil
}

func (g *AssemblyGenerator) FromReturnStatement(r ast.ReturnStatement) (string, error) {
	expString, err := g.FromExpression(r.ReturnValue)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
%s
mov  %%rbp, %%rsp     /* restore esp now it points to the old ebp */
pop  %%rbp            /* restore old ebp, esp is now where it was before */
ret
`, expString)), nil
}

func (g *AssemblyGenerator) FromBlockStatement(block ast.BlockStatement) (string, error) {
	src := ""
	for _, stmt := range block.Statements {
		stmtString, err := g.FromStatement(stmt)
		if err != nil {
			return "", err
		}
		src += fmt.Sprintf("%s\n", stmtString)
	}
	return src, nil
}

func (g *AssemblyGenerator) FromStatement(s ast.Statement) (string, error) {
	switch s := s.(type) {
	case *ast.ReturnStatement:
		return g.FromReturnStatement(*s)
	case *ast.BlockStatement:
		return g.FromBlockStatement(*s)
	case *ast.FunctionStatement:
		return g.FromFunction(*s)
	default:
		return "", fmt.Errorf("Failed with %s", s.TokenLiteral())
	}
}
