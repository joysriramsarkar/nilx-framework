package manager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageManagerInstall(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nilpm_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manifest := `name: test-app
version: 0.1.0
dependencies:
  "@nilx/ui-charts": "^1.2.0"
  "@nilx/sqlite": "^0.4.0"
`
	err = os.WriteFile(filepath.Join(tempDir, "nilx.yaml"), []byte(manifest), 0644)
	if err != nil {
		t.Fatalf("failed writing manifest: %v", err)
	}

	pm := New(tempDir)
	if err := pm.Install(); err != nil {
		t.Fatalf("pm.Install failed: %v", err)
	}

	lockBytes, err := os.ReadFile(filepath.Join(tempDir, "nilx.lock"))
	if err != nil {
		t.Fatalf("expected nilx.lock to be generated: %v", err)
	}

	lockStr := string(lockBytes)
	if !strings.Contains(lockStr, "@nilx/ui-charts") || !strings.Contains(lockStr, "sha256-") {
		t.Errorf("expected lockfile to contain packages and integrity hash, got:\n%s", lockStr)
	}
}
