// Package lsp implements the Language Server Protocol (LSP 3.17) for NilLang.
package lsp

// Position in a text document expressed as zero-based line and character offset.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range in a text document expressed as (start, end) positions.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic represents a compiler error or warning.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1: Error, 2: Warning, 3: Info, 4: Hint
	Source   string `json:"source"`
	Message  string `json:"message"`
}

// CompletionItem represents a code autocompletion suggestion.
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"` // 1: Text, 2: Method, 3: Function, 6: Variable, 7: Class, 14: Keyword
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// Hover represents hover documentation at a position.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown" or "plaintext"
	Value string `json:"value"`
}

// TextEdit represents a text change in a document.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// JSON-RPC 2.0 structures
type RequestMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type ResponseMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
