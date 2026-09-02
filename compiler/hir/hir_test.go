package hir

import (
	"testing"
)

func TestConstantFoldingInt(t *testing.T) {
	left := Value{Type: "i32", IntVal: 10, IsConst: true}
	right := Value{Type: "i32", IntVal: 20, IsConst: true}

	res, ok := FoldBinary(left, right, "+")
	if !ok || res.IntVal != 30 {
		t.Errorf("expected 10 + 20 = 30, got %v", res)
	}

	resSub, ok := FoldBinary(right, left, "-")
	if !ok || resSub.IntVal != 10 {
		t.Errorf("expected 20 - 10 = 10, got %v", resSub)
	}
}

func TestConstantFoldingString(t *testing.T) {
	left := Value{Type: "string", StrVal: "Hello ", IsConst: true}
	right := Value{Type: "string", StrVal: "Onuron", IsConst: true}

	res, ok := FoldBinary(left, right, "+")
	if !ok || res.StrVal != "Hello Onuron" {
		t.Errorf("expected string concat 'Hello NilOS', got %q", res.StrVal)
	}
}

func TestConstantFoldingBool(t *testing.T) {
	left := Value{Type: "bool", BoolVal: true, IsConst: true}
	right := Value{Type: "bool", BoolVal: false, IsConst: true}

	resAnd, ok := FoldBinary(left, right, "&&")
	if !ok || resAnd.BoolVal != false {
		t.Errorf("expected true && false = false, got %v", resAnd)
	}

	resOr, ok := FoldBinary(left, right, "||")
	if !ok || resOr.BoolVal != true {
		t.Errorf("expected true || false = true, got %v", resOr)
	}
}
