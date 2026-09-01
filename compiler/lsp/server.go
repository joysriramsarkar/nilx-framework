// Package lsp implements the Language Server Protocol (LSP 3.17) for NilLang.
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/joysriramsarkar/nilx-framework/compiler/formatter"
	"github.com/joysriramsarkar/nilx-framework/compiler/lexer"
	"github.com/joysriramsarkar/nilx-framework/compiler/parser"
	"github.com/joysriramsarkar/nilx-framework/compiler/types"
)

// Server is the NilLang language server (nills).
type Server struct {
	mu        sync.RWMutex
	documents map[string]string
	in        *bufio.Reader
	out       io.Writer
}

// NewServer creates a new NilLang LSP server.
func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{
		documents: make(map[string]string),
		in:        bufio.NewReader(in),
		out:       out,
	}
}

// Serve runs the standard LSP stdio transport loop.
func (s *Server) Serve() error {
	for {
		req, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		s.handleRequest(req)
	}
}

func (s *Server) readMessage() (*RequestMessage, error) {
	var contentLength int
	for {
		line, err := s.in.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				contentLength, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("empty content length")
	}

	body := make([]byte, contentLength)
	_, err := io.ReadFull(s.in, body)
	if err != nil {
		return nil, err
	}

	var req RequestMessage
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *Server) sendResponse(id interface{}, result interface{}, rpcErr *RPCError) {
	resp := ResponseMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   rpcErr,
	}
	data, _ := json.Marshal(resp)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	_, _ = s.out.Write([]byte(header))
	_, _ = s.out.Write(data)
}

func (s *Server) sendNotification(method string, params interface{}) {
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	data, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	_, _ = s.out.Write([]byte(header))
	_, _ = s.out.Write(data)
}

func (s *Server) handleRequest(req *RequestMessage) {
	switch req.Method {
	case "initialize":
		s.sendResponse(req.ID, map[string]interface{}{
			"capabilities": map[string]interface{}{
				"textDocumentSync": 1, // Full
				"completionProvider": map[string]interface{}{
					"triggerCharacters": []string{".", "@", ":"},
				},
				"hoverProvider":              true,
				"documentFormattingProvider": true,
				"definitionProvider":        true,
			},
			"serverInfo": map[string]interface{}{
				"name":    "nills (NilLang Language Server)",
				"version": "0.1.0",
			},
		}, nil)

	case "initialized":
		// Ready

	case "textDocument/didOpen":
		var params struct {
			TextDocument struct {
				URI  string `json:"uri"`
				Text string `json:"text"`
			} `json:"textDocument"`
		}
		pBytes, _ := json.Marshal(req.Params)
		_ = json.Unmarshal(pBytes, &params)
		s.mu.Lock()
		s.documents[params.TextDocument.URI] = params.TextDocument.Text
		s.mu.Unlock()
		s.publishDiagnostics(params.TextDocument.URI, params.TextDocument.Text)

	case "textDocument/didChange":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		pBytes, _ := json.Marshal(req.Params)
		_ = json.Unmarshal(pBytes, &params)
		if len(params.ContentChanges) > 0 {
			newText := params.ContentChanges[len(params.ContentChanges)-1].Text
			s.mu.Lock()
			s.documents[params.TextDocument.URI] = newText
			s.mu.Unlock()
			s.publishDiagnostics(params.TextDocument.URI, newText)
		}

	case "textDocument/completion":
		s.sendResponse(req.ID, s.getCompletions(), nil)

	case "textDocument/hover":
		s.sendResponse(req.ID, s.getHoverInfo(), nil)

	case "textDocument/formatting":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		pBytes, _ := json.Marshal(req.Params)
		_ = json.Unmarshal(pBytes, &params)
		s.mu.RLock()
		src := s.documents[params.TextDocument.URI]
		s.mu.RUnlock()

		formatted := formatter.Format(src)
		lines := strings.Split(src, "\n")
		lastLine := len(lines) - 1
		lastChar := 0
		if lastLine >= 0 {
			lastChar = len(lines[lastLine])
		}

		edits := []TextEdit{
			{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: lastLine, Character: lastChar},
				},
				NewText: formatted,
			},
		}
		s.sendResponse(req.ID, edits, nil)

	case "shutdown":
		s.sendResponse(req.ID, nil, nil)

	case "exit":
		// Clean exit
	}
}

func (s *Server) publishDiagnostics(uri, src string) {
	var diagnostics []Diagnostic

	l := lexer.New(uri, src)
	tokens := l.Tokenize()
	for _, errStr := range l.Errors() {
		diagnostics = append(diagnostics, Diagnostic{
			Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 10}},
			Severity: 1,
			Source:   "nil-lexer",
			Message:  errStr,
		})
	}

	p := parser.New(uri, tokens)
	prog := p.Parse()
	for _, errStr := range p.Errors() {
		diagnostics = append(diagnostics, Diagnostic{
			Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 10}},
			Severity: 1,
			Source:   "nil-parser",
			Message:  errStr,
		})
	}

	if prog != nil {
		checker := types.New()
		checker.CheckProgram(prog)
		for _, warnStr := range checker.Errors() {
			diagnostics = append(diagnostics, Diagnostic{
				Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 10}},
				Severity: 2,
				Source:   "nil-types",
				Message:  warnStr,
			})
		}
	}

	s.sendNotification("textDocument/publishDiagnostics", map[string]interface{}{
		"uri":         uri,
		"diagnostics": diagnostics,
	})
}

func (s *Server) getCompletions() []CompletionItem {
	return []CompletionItem{
		// Keywords
		{Label: "function", Kind: 14, Detail: "function declaration", InsertText: "function ${1:name}(${2:params}): ${3:void} {\n\t$0\n}"},
		{Label: "component", Kind: 7, Detail: "declarative UI component", InsertText: "component ${1:Name} {\n\tbuild() {\n\t\t$0\n\t}\n}"},
		{Label: "let", Kind: 14, Detail: "variable declaration", InsertText: "let ${1:name}: ${2:type} = ${3:value}"},
		{Label: "const", Kind: 14, Detail: "constant declaration", InsertText: "const ${1:NAME}: ${2:type} = ${3:value}"},
		{Label: "if", Kind: 14, Detail: "conditional statement", InsertText: "if ${1:condition} {\n\t$0\n}"},
		{Label: "while", Kind: 14, Detail: "loop statement", InsertText: "while ${1:condition} {\n\t$0\n}"},
		{Label: "match", Kind: 14, Detail: "pattern matching", InsertText: "match ${1:expr} {\n\t${2:pattern} => { $0 }\n}"},
		{Label: "struct", Kind: 7, Detail: "struct type declaration", InsertText: "struct ${1:Name} {\n\t${2:field}: ${3:type}\n}"},
		{Label: "enum", Kind: 13, Detail: "enum type declaration", InsertText: "enum ${1:Name} {\n\t${2:Value},\n}"},
		{Label: "actor", Kind: 7, Detail: "actor concurrency primitive", InsertText: "actor ${1:Name} {\n\ton ${2:Message}(${3:params}) {\n\t\t$0\n\t}\n}"},

		// UI Components
		{Label: "Column", Kind: 7, Detail: "Vertical flex container", InsertText: "Column {\n\t$0\n}"},
		{Label: "Row", Kind: 7, Detail: "Horizontal flex container", InsertText: "Row {\n\t$0\n}"},
		{Label: "Text", Kind: 3, Detail: "Text widget", InsertText: "Text(\"${1:Content}\")"},
		{Label: "Button", Kind: 3, Detail: "Button widget", InsertText: "Button(\"${1:Click Me}\") {\n\tonClick => {\n\t\t$0\n\t}\n}"},
		{Label: "TextInput", Kind: 3, Detail: "Text input field", InsertText: "TextInput(\"${1:placeholder}\")"},
		{Label: "Card", Kind: 7, Detail: "Card container with elevation", InsertText: "Card {\n\t$0\n}"},
		{Label: "ScrollView", Kind: 7, Detail: "Scrollable viewport container", InsertText: "ScrollView {\n\t$0\n}"},
		{Label: "Divider", Kind: 3, Detail: "Horizontal rule separator", InsertText: "Divider()"},
		{Label: "Spacer", Kind: 3, Detail: "Flexible layout spacer", InsertText: "Spacer()"},

		// Standard Library functions
		{Label: "time_now", Kind: 3, Detail: "() -> i64 (Unix timestamp ms)", InsertText: "time_now()"},
		{Label: "time_sleep", Kind: 3, Detail: "(ms: i64) -> void", InsertText: "time_sleep(${1:1000})"},
		{Label: "math_sqrt", Kind: 3, Detail: "(x: f64) -> f64", InsertText: "math_sqrt(${1:16})"},
		{Label: "math_abs", Kind: 3, Detail: "(x: f64) -> f64", InsertText: "math_abs(${1:x})"},
		{Label: "crypto_sha256", Kind: 3, Detail: "(data: string) -> string", InsertText: "crypto_sha256(\"${1:data}\")"},
		{Label: "crypto_uuid", Kind: 3, Detail: "() -> string (UUID v4)", InsertText: "crypto_uuid()"},
		{Label: "json_stringify", Kind: 3, Detail: "(val: any) -> string", InsertText: "json_stringify(${1:obj})"},
		{Label: "json_parse", Kind: 3, Detail: "(json: string) -> any", InsertText: "json_parse(${1:str})"},
		{Label: "fs_read", Kind: 3, Detail: "(path: string) -> string", InsertText: "fs_read(\"${1:path}\")"},
		{Label: "fs_write", Kind: 3, Detail: "(path: string, content: string) -> bool", InsertText: "fs_write(\"${1:path}\", \"${2:content}\")"},
		{Label: "http_get", Kind: 3, Detail: "(url: string) -> string", InsertText: "http_get(\"${1:url}\")"},
		{Label: "log_info", Kind: 3, Detail: "(msg: string) -> void", InsertText: "log_info(\"${1:message}\")"},
	}
}

func (s *Server) getHoverInfo() Hover {
	return Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: "**NilLang Language & Framework**\n\nDeclarative UI, Go-inspired concurrency, static typing.",
		},
	}
}
