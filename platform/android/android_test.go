package android

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAndroidProjectGenerator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "android_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := New()
	bytecode := []byte("NABC_ANDROID_TEST_BYTECODE")
	if err := adapter.GenerateProject(tempDir, bytecode); err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	// Verify settings.gradle.kts
	settings, err := os.ReadFile(filepath.Join(tempDir, "settings.gradle.kts"))
	if err != nil || !strings.Contains(string(settings), "include(\":app\")") {
		t.Errorf("expected settings.gradle.kts with :app include, got: %s", string(settings))
	}

	// Verify NilXActivity.kt
	activity, err := os.ReadFile(filepath.Join(tempDir, "app", "src", "main", "java", "io", "nilx", "app", "NilXActivity.kt"))
	if err != nil || !strings.Contains(string(activity), "class NilXActivity") {
		t.Errorf("expected NilXActivity.kt generated, got: %s", string(activity))
	}

	// Verify bundled bytecode in assets
	assetNabc, err := os.ReadFile(filepath.Join(tempDir, "app", "src", "main", "assets", "main.nabc"))
	if err != nil || string(assetNabc) != string(bytecode) {
		t.Errorf("expected main.nabc in assets matching bytecode")
	}
}
