package ios

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIOSProjectGenerator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ios_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := New()
	bytecode := []byte("NABC_IOS_TEST_BYTECODE")
	if err := adapter.GenerateProject(tempDir, bytecode); err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	// Verify AlapViewController.swift
	vcPath := filepath.Join(tempDir, "ios", adapter.AppName, "AlapViewController.swift")
	vcContent, err := os.ReadFile(vcPath)
	if err != nil || !strings.Contains(string(vcContent), "class AlapViewController") {
		t.Errorf("expected AlapViewController.swift generated, got: %s", string(vcContent))
	}

	// Verify Info.plist
	plistPath := filepath.Join(tempDir, "ios", adapter.AppName, "Info.plist")
	plistContent, err := os.ReadFile(plistPath)
	if err != nil || !strings.Contains(string(plistContent), "CFBundleIdentifier") {
		t.Errorf("expected Info.plist generated, got: %s", string(plistContent))
	}
}
