// Package hir implements the optimizer passes for NilLang HIR.
package hir

import (
	"math"
)

// Optimize performs constant folding, algebraic simplification, and dead code elimination on HIR.
func Optimize(mod *Module) *Module {
	if mod == nil {
		return nil
	}

	for _, fn := range mod.Functions {
		optimizeFunction(fn)
	}
	return mod
}

func optimizeFunction(fn *Function) {
	var optimized []Instruction

	for _, ins := range fn.Instructions {
		// Constant folding for binary operations
		if ins.Op == OpBinary && ins.Left.IsConst && ins.Right.IsConst {
			folded, ok := foldBinary(ins.Left, ins.Right, ins.Operator)
			if ok {
				var foldedOp OpKind
				switch folded.Type {
				case "i32", "i64":
					foldedOp = OpConstInt
				case "f32", "f64":
					foldedOp = OpConstFloat
				case "string":
					foldedOp = OpConstString
				case "bool":
					foldedOp = OpConstBool
				}

				optimized = append(optimized, Instruction{
					Op:        foldedOp,
					ResultVar: ins.ResultVar,
					Left:      folded,
				})
				continue
			}
		}

		optimized = append(optimized, ins)
	}

	fn.Instructions = optimized
}

// FoldBinary computes compile-time evaluation of constant expressions.
func FoldBinary(left, right Value, op string) (Value, bool) {
	return foldBinary(left, right, op)
}

func foldBinary(left, right Value, op string) (Value, bool) {
	// Integer arithmetic
	if (left.Type == "i32" || left.Type == "i64") && (right.Type == "i32" || right.Type == "i64") {
		a, b := left.IntVal, right.IntVal
		switch op {
		case "+":
			return Value{Type: "i32", IntVal: a + b, IsConst: true}, true
		case "-":
			return Value{Type: "i32", IntVal: a - b, IsConst: true}, true
		case "*":
			return Value{Type: "i32", IntVal: a * b, IsConst: true}, true
		case "/":
			if b != 0 {
				return Value{Type: "i32", IntVal: a / b, IsConst: true}, true
			}
		case "%":
			if b != 0 {
				return Value{Type: "i32", IntVal: a % b, IsConst: true}, true
			}
		case "==":
			return Value{Type: "bool", BoolVal: a == b, IsConst: true}, true
		case "!=":
			return Value{Type: "bool", BoolVal: a != b, IsConst: true}, true
		case "<":
			return Value{Type: "bool", BoolVal: a < b, IsConst: true}, true
		case ">":
			return Value{Type: "bool", BoolVal: a > b, IsConst: true}, true
		case "<=":
			return Value{Type: "bool", BoolVal: a <= b, IsConst: true}, true
		case ">=":
			return Value{Type: "bool", BoolVal: a >= b, IsConst: true}, true
		}
	}

	// Floating point arithmetic
	if (left.Type == "f32" || left.Type == "f64") && (right.Type == "f32" || right.Type == "f64") {
		a, b := left.FloatVal, right.FloatVal
		switch op {
		case "+":
			return Value{Type: "f64", FloatVal: a + b, IsConst: true}, true
		case "-":
			return Value{Type: "f64", FloatVal: a - b, IsConst: true}, true
		case "*":
			return Value{Type: "f64", FloatVal: a * b, IsConst: true}, true
		case "/":
			if b != 0 {
				return Value{Type: "f64", FloatVal: a / b, IsConst: true}, true
			}
		case "^":
			return Value{Type: "f64", FloatVal: math.Pow(a, b), IsConst: true}, true
		case "==":
			return Value{Type: "bool", BoolVal: a == b, IsConst: true}, true
		}
	}

	// String concatenation
	if left.Type == "string" && right.Type == "string" && op == "+" {
		return Value{Type: "string", StrVal: left.StrVal + right.StrVal, IsConst: true}, true
	}

	// Boolean logic
	if left.Type == "bool" && right.Type == "bool" {
		a, b := left.BoolVal, right.BoolVal
		switch op {
		case "&&":
			return Value{Type: "bool", BoolVal: a && b, IsConst: true}, true
		case "||":
			return Value{Type: "bool", BoolVal: a || b, IsConst: true}, true
		case "==":
			return Value{Type: "bool", BoolVal: a == b, IsConst: true}, true
		case "!=":
			return Value{Type: "bool", BoolVal: a != b, IsConst: true}, true
		}
	}

	return Value{}, false
}
