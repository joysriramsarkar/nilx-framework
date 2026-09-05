package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWebPlatformAdapter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "alap_web_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := New()
	bytecode := []byte("WASM_BYTECODE_DUMMY")

	err = adapter.GenerateProject(tempDir, bytecode)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	// Verify index.html
	indexFile, err := os.ReadFile(filepath.Join(tempDir, "web", "index.html"))
	if err != nil || !strings.Contains(string(indexFile), "AlapRuntime.boot") {
		t.Errorf("expected valid index.html with runtime boot, got: %s", string(indexFile))
	}

	// Verify runtime.js
	runtimeFile, err := os.ReadFile(filepath.Join(tempDir, "web", "js", "runtime.js"))
	if err != nil || !strings.Contains(string(runtimeFile), "AlapRuntime =") {
		t.Errorf("expected valid runtime.js, got: %s", string(runtimeFile))
	}

	// Verify manifest.json
	manifestFile, err := os.ReadFile(filepath.Join(tempDir, "web", "manifest.json"))
	if err != nil || !strings.Contains(string(manifestFile), "AlapWebApp") {
		t.Errorf("expected valid manifest.json, got: %s", string(manifestFile))
	}
}
