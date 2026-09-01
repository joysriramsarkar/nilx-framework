package abi

import (
	"strings"
	"testing"

	"github.com/joysriramsarkar/nilx-framework/compiler/codegen"
	"github.com/joysriramsarkar/nilx-framework/compiler/lexer"
	"github.com/joysriramsarkar/nilx-framework/compiler/parser"
)

func TestSerializeDeserializeRoundtrip(t *testing.T) {
	src := `
component CounterApp {
    build() {
        Column {
            Text("NilX ABI Test")
            Button("Increment")
        }
    }
}
`
	l := lexer.New("test.nil", src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		t.Fatalf("lex errors: %v", l.Errors())
	}
	p := parser.New("test.nil", tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	gen := codegen.New("test")
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		t.Fatalf("codegen errors: %v", gen.Errors())
	}

	bytes := codegen.Serialize(gen.Module())
	if len(bytes) == 0 {
		t.Fatalf("expected non-empty serialized bytes")
	}

	res, err := RunBytecodeInGo(bytes)
	if err != nil {
		t.Fatalf("failed to run deserialized bytecode: %v", err)
	}

	if !strings.Contains(res, "NilX ABI Test") {
		t.Errorf("expected UI output to contain 'NilX ABI Test', got: %s", res)
	}
}
