// Package hir defines the High-Level Intermediate Representation and optimization passes for NilLang.
package hir

import (
	"fmt"
	"strings"
)

// OpKind represents the operation type in HIR.
type OpKind int

const (
	OpConstInt OpKind = iota
	OpConstFloat
	OpConstString
	OpConstBool
	OpConstNull
	OpLoadVar
	OpStoreVar
	OpBinary
	OpUnary
	OpCall
	OpReturn
	OpBranch
	OpJump
	OpUICreate
	OpUIProp
	OpUIEnd
)

// Value represents a typed operand in HIR.
type Value struct {
	Type     string
	IntVal   int64
	FloatVal float64
	StrVal   string
	BoolVal  bool
	IsConst  bool
}

// Instruction is a single intermediate representation statement.
type Instruction struct {
	Op       OpKind
	ResultVar string
	Left     Value
	Right    Value
	Operator string
	Args     []Value
	Target   int
}

// Function represents a compiled/optimized routine in HIR.
type Function struct {
	Name         string
	Params       []string
	Instructions []Instruction
}

// Module represents the complete lowered program.
type Module struct {
	Name      string
	Functions []*Function
}

func (f *Function) Dump() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func %s(%s):\n", f.Name, strings.Join(f.Params, ", ")))
	for i, ins := range f.Instructions {
		sb.WriteString(fmt.Sprintf("  %3d: %v\n", i, ins))
	}
	return sb.String()
}
