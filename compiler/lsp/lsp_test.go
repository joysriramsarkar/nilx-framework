package lsp

import (
	"bytes"
	"testing"
)

func TestLSPDefinitionAndSymbols(t *testing.T) {
	src := `
struct User {
    id: string
    name: string
}

function calculateScore(u: User): i32 {
    let base: i32 = 100
    return base
}
`
	var inBuf, outBuf bytes.Buffer
	server := NewServer(&inBuf, &outBuf)
	server.documents["file:///test.nil"] = src

	// Test Document Symbols
	symbols := server.getDocumentSymbols("file:///test.nil")
	if len(symbols) < 2 {
		t.Errorf("expected at least 2 symbols (User struct and calculateScore function), got %d", len(symbols))
	}

	// Test Definition lookup for calculateScore
	loc := server.findDefinition("file:///test.nil", Position{Line: 6, Character: 12})
	if loc == nil {
		t.Errorf("expected definition found for calculateScore")
	}

	// Test Completions
	completions := server.getCompletions()
	if len(completions) < 10 {
		t.Errorf("expected at least 10 completions, got %d", len(completions))
	}
}
