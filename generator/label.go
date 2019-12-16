package generator

import (
	"fmt"
)

type LabelGenerator struct {
	Count int
}

func NewLabelGenerator() *LabelGenerator {
	return &LabelGenerator{Count: 0}
}

func (g *LabelGenerator) GetNextLabel(label string) string {
	l := fmt.Sprintf("%s%d", label, g.Count)
	g.Count++
	return l
}
