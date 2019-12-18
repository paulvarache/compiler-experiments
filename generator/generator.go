package generator

import (
	"compiler/ast"
	"fmt"
)

type AssemblyGenerator struct {
	LabelGenerator *LabelGenerator
	Variables      *VariableManager
	Lines          [][]string
	Depth          int
}

func NewAssemblyGenerator() *AssemblyGenerator {
	return &AssemblyGenerator{LabelGenerator: &LabelGenerator{}, Variables: NewVariableManager(), Lines: make([][]string, 0), Depth: 0}
}

func (g *AssemblyGenerator) AddLine(els ...string) {
	lines := make([]string, 0)
	for i := 0; i < g.Depth; i++ {
		lines = append(lines, "")
	}
	lines = append(lines, els...)
	g.Lines = append(g.Lines, lines)
}

func (g *AssemblyGenerator) EnterContext() {
	g.Depth++
}

func (g *AssemblyGenerator) LeaveContext() {
	g.Depth--
}

func (g *AssemblyGenerator) GetString() string {
	src := ""
	longest := 0
	// Find longest line
	for _, line := range g.Lines {
		if len(line) > longest {
			longest = len(line)
		}
	}
	longestItem := make([]int, longest)
	// Find the longest item for each position in line
	for _, line := range g.Lines {
		for i, item := range line {
			if len(item) > longestItem[i] {
				longestItem[i] = len(item)
			}
		}
	}

	// Concat using right pad to allign everything
	for _, line := range g.Lines {
		lineString := ""
		for i, item := range line {
			padding := longestItem[i] + 4
			lineString += fmt.Sprintf("%-*v", padding, item)
		}
		src += lineString + "\n"
	}
	return src
}

func (g *AssemblyGenerator) FromProgram(p *ast.Program) (string, error) {
	for _, fn := range p.Functions {
		err := g.FromStatement(fn)
		if err != nil {
			return "", err
		}
	}
	return g.GetString(), nil
}

func (g *AssemblyGenerator) FromFunction(f ast.FunctionStatement) error {
	g.AddLine(fmt.Sprintf(".globl %s", f.Name))
	g.AddLine(fmt.Sprintf("%s:", f.Name))
	g.EnterContext()
	g.AddLine("pushq", "%rbp", "/* Save value of the bottom of the current frame */")
	g.AddLine("movq", "%rsp, %rbp", "/* Top of stack is now bottom of new frame */")
	err := g.FromStatement(f.Body)
	g.LeaveContext()
	if err != nil {
		return err
	}
	return nil
}

func (g *AssemblyGenerator) FromReturnStatement(r ast.ReturnStatement) error {
	err := g.FromExpression(r.ReturnValue)
	if err != nil {
		return err
	}
	g.AddLine("movq", "%rbp, %rsp", "/* restore esp now it points to the old ebp */")
	g.AddLine("popq", "%rbp", "/* restore old ebp, esp is now where it was before */")
	g.AddLine("ret")
	return nil
}

func (g *AssemblyGenerator) FromBlockStatement(block ast.BlockStatement) error {
	for _, stmt := range block.Statements {
		err := g.FromStatement(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *AssemblyGenerator) FromDeclStatement(s ast.DeclStatement) error {
	if s.Right != nil {
		err := g.FromExpression(s.Right)
		if err != nil {
			return err
		}
	} else {
		g.AddLine("mov", "$0, %rax", "/* default variable value */")
	}
	g.AddLine("push", "%rax", "/* Save variable value to stack */")
	return g.Variables.CreateVariable(s.Left.Value)
}

func (g *AssemblyGenerator) FromExpStatement(s ast.ExpStatement) error {
	return g.FromExpression(s.Expression)
}

func (g *AssemblyGenerator) FromIfStatement(s ast.IfStatement) error {
	err := g.FromExpression(s.Condition)
	if err != nil {
		return err
	}
	clauseName := g.LabelGenerator.GetNextLabel("clause")
	postName := g.LabelGenerator.GetNextLabel("post")
	elseName := g.LabelGenerator.GetNextLabel("else")
	g.AddLine("cmp", "$0, %rax", "/* Set ZF to 0 if condition is false */")
	if s.ElseBody == nil {
		g.AddLine("je", postName, "/* Go to post if consition is false */")
		g.AddLine("jmp", clauseName, "/* Go to clause if condition otherwise */")
	} else {
		g.AddLine("je", elseName)
		g.AddLine("jmp", postName)
	}
	// Add else if there is any
	g.LeaveContext()
	g.AddLine(fmt.Sprintf("%s:", clauseName))
	g.EnterContext()
	err = g.FromStatement(s.Body)
	if err != nil {
		return err
	}
	g.AddLine("jmp", postName, "/* Clause ran, go back to post */")
	if s.ElseBody != nil {
		g.LeaveContext()
		g.AddLine(fmt.Sprintf("%s:", elseName))
		g.EnterContext()
		err = g.FromStatement(s.ElseBody)
		if err != nil {
			return err
		}
	}
	g.LeaveContext()
	g.AddLine(fmt.Sprintf("%s:", postName))
	g.EnterContext()
	return nil
}

func (g *AssemblyGenerator) FromStatement(s ast.Statement) error {
	switch s := s.(type) {
	case *ast.ReturnStatement:
		return g.FromReturnStatement(*s)
	case *ast.BlockStatement:
		return g.FromBlockStatement(*s)
	case *ast.FunctionStatement:
		return g.FromFunction(*s)
	case *ast.DeclStatement:
		return g.FromDeclStatement(*s)
	case *ast.ExpStatement:
		return g.FromExpStatement(*s)
	case *ast.IfStatement:
		return g.FromIfStatement(*s)
	default:
		return fmt.Errorf("Failed with %s", s.TokenLiteral())
	}
}
