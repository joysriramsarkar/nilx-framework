package linux

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxProjectGenerator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "linux_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := New()
	bytecode := []byte("NABC_LINUX_TEST_BYTECODE")
	if err := adapter.GenerateProject(tempDir, bytecode); err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	// Verify AppRun
	appRun, err := os.ReadFile(filepath.Join(tempDir, "linux", "AppDir", "AppRun"))
	if err != nil || !strings.Contains(string(appRun), "exec") {
		t.Errorf("expected AppRun script, got: %s", string(appRun))
	}

	// Verify .desktop file
	desktopFile, err := os.ReadFile(filepath.Join(tempDir, "linux", "AppDir", adapter.AppID+".desktop"))
	if err != nil || !strings.Contains(string(desktopFile), "[Desktop Entry]") {
		t.Errorf("expected .desktop file, got: %s", string(desktopFile))
	}

	// Verify Flatpak manifest
	manifest, err := os.ReadFile(filepath.Join(tempDir, "linux", adapter.AppID+".json"))
	if err != nil || !strings.Contains(string(manifest), "org.freedesktop.Platform") {
		t.Errorf("expected Flatpak manifest, got: %s", string(manifest))
	}
}
