// Package fs provides filesystem input/output utilities for NilLang.
package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadTextFile reads the full content of a file as a string.
func ReadTextFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("fs.readTextFile: %w", err)
	}
	return string(bytes), nil
}

// WriteTextFile writes string data to a file (creating parent directories if needed).
func WriteTextFile(path, content string) error {
	dir := filepath.Dir(path)
	if dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("fs.writeTextFile: %w", err)
	}
	return nil
}

// AppendTextFile appends string data to a file.
func AppendTextFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("fs.appendTextFile: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// Exists checks if a file or directory exists at the path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Remove deletes a file or empty directory.
func Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll deletes a file or directory and all its children.
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Mkdir creates a directory (and parents).
func Mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ReadDir returns list of filenames inside a directory.
func ReadDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("fs.readDir: %w", err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names, nil
}
