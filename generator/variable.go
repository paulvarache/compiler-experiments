package generator

import (
	"fmt"
)

type VariableManager struct {
	Variables  map[string]int
	StackIndex int
}

func NewVariableManager() *VariableManager {
	return &VariableManager{Variables: make(map[string]int), StackIndex: -8}
}

func (v *VariableManager) VariableExists(name string) bool {
	_, ok := v.Variables[name]
	return ok
}

func (v *VariableManager) GetVariableStackIndex(name string) (int, error) {
	value, ok := v.Variables[name]
	if !ok {
		return 0, fmt.Errorf("Undecraled variable '%s'", name)
	}
	return value, nil
}

func (v *VariableManager) CreateVariable(name string) error {
	if v.VariableExists(name) {
		return fmt.Errorf("Could not re-declare variable '%s'", name)
	}
	v.Variables[name] = v.StackIndex
	v.StackIndex = v.StackIndex - 8
	return nil
}
