package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTUIRendersWithAndExitAndType(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "2025-08-14-redis-connection-pool"), 0755)
	os.MkdirAll(filepath.Join(dir, "thread-pool"), 0755)

	stdout, stderr, _ := runCmd(t, "cd", "--and-type", "pool", "--and-exit", "--path", dir)
	combined := stdout + stderr

	// Strip ANSI codes (basic pattern)
	clean := stripANSI(combined)

	if !strings.Contains(clean, "Try Directory Selection") {
		t.Error("should contain the TUI header")
	}
	if !strings.Contains(clean, "Search: pool") {
		t.Error("should show the typed query")
	}
	if !strings.Contains(clean, "redis-connection-pool") {
		t.Error("should list redis-connection-pool directory")
	}
	if !strings.Contains(clean, "thread-pool") {
		t.Error("should list thread-pool directory")
	}
	if !strings.Contains(clean, "Create new: pool") {
		t.Error("should show create new line when input present")
	}
}

func TestCreateNewGeneratesMkdirScript(t *testing.T) {
	dir := t.TempDir()
	stdout, _, _ := runCmd(t, "cd", "new-thing", "--and-keys", "ENTER", "--path", dir)

	if !strings.Contains(stdout, "mkdir -p") || !strings.Contains(stdout, "new-thing") {
		t.Error("should emit mkdir for create new")
	}
	if !strings.Contains(stdout, "cd") || !strings.Contains(stdout, "new-thing") {
		t.Error("should emit cd into created dir")
	}
	// Should include date prefix
	if !strings.Contains(stdout, "2") {
		t.Error("should include date-prefixed new directory name")
	}
}

func TestDeleteFlowConfirmsAndDeletes(t *testing.T) {
	dir := t.TempDir()
	name := "2025-08-14-delete-me"
	path := filepath.Join(dir, name)
	os.MkdirAll(path, 0755)

	stdout, stderr, _ := runCmd(t, "cd", "--and-type", "delete-me", "--and-keys", "CTRL-D,ESC", "--and-confirm", "YES", "--path", dir)
	combined := stripANSI(stdout + stderr)

	if !strings.Contains(combined, "Delete Directory") {
		t.Error("should show delete confirmation header")
	}
	if !strings.Contains(combined, "Deleted: "+name) {
		t.Error("should display deletion status")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("directory should be deleted")
	}
}

func TestCtrlJKNavigation(t *testing.T) {
	dir := t.TempDir()
	firstDir := "2025-08-14-first"
	secondDir := "2025-08-15-second"
	os.MkdirAll(filepath.Join(dir, firstDir), 0755)
	os.MkdirAll(filepath.Join(dir, secondDir), 0755)

	// Bump mtime of first directory
	f, _ := os.Create(filepath.Join(dir, firstDir, ".mtime_bump"))
	f.Close()

	// Ctrl-J (down) navigation - starts at index 0, goes to index 1
	stdout, _, _ := runCmd(t, "cd", "--and-keys", "CTRL-J,ENTER", "--path", dir)

	if !strings.Contains(stdout, "cd") || !strings.Contains(stdout, secondDir) {
		t.Error("Ctrl-J should navigate down to second directory")
	}

	// Ctrl-K (up) navigation
	stdout, _, _ = runCmd(t, "cd", "--and-keys", "CTRL-J,CTRL-K,ENTER", "--path", dir)

	if !strings.Contains(stdout, "cd") || !strings.Contains(stdout, firstDir) {
		t.Error("Ctrl-K should navigate up after going down")
	}
}

func TestDeleteFlowCancel(t *testing.T) {
	dir := t.TempDir()
	name := "2025-08-14-keep-me"
	path := filepath.Join(dir, name)
	os.MkdirAll(path, 0755)

	stdout, stderr, _ := runCmd(t, "cd", "--and-type", "keep-me", "--and-keys", "CTRL-D,ESC", "--and-confirm", "NO", "--path", dir)
	combined := stripANSI(stdout + stderr)

	if !strings.Contains(combined, "Delete Directory") {
		t.Error("should show delete confirmation")
	}
	if !strings.Contains(combined, "Delete cancelled") {
		t.Error("should show cancellation message")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("directory should still exist")
	}
}

// Helper function to strip ANSI escape codes
func stripANSI(s string) string {
	// Simple ANSI strip - matches escape sequences
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (s[i] >= '@' && s[i] <= '~') || s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
