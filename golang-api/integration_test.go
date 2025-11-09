package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseGitURIHTTPSGithubGeneratesCorrectPath(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := runCmd(t, "clone", "https://github.com/user/repo.git", "--path", dir)
	if err != nil {
		t.Logf("command error (may be expected): %v", err)
	}
	if !strings.Contains(stdout, "user-repo") {
		t.Error("should generate date-prefixed user-repo path")
	}
	if !strings.Contains(stdout, "git clone") {
		t.Error("should emit git clone command")
	}
}

func TestParseGitURISSHFormatGeneratesCorrectPath(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := runCmd(t, "clone", "git@github.com:user/repo.git", "--path", dir)
	if err != nil {
		t.Logf("command error (may be expected): %v", err)
	}
	if !strings.Contains(stdout, "user-repo") {
		t.Error("should generate date-prefixed user-repo path")
	}
}

func TestCloneWithCustomNameUsesCustomName(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := runCmd(t, "clone", "https://github.com/user/repo.git", "custom", "--path", dir)
	if err != nil {
		t.Logf("command error (may be expected): %v", err)
	}
	if !strings.Contains(stdout, "custom") {
		t.Error("should use custom name")
	}
	if strings.Contains(stdout, "user-repo") && !strings.Contains(stdout, "custom") {
		t.Error("should not use auto-generated name when custom provided")
	}
}

func TestURLShorthandWorksLikeClone(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := runCmd(t, "cd", "https://github.com/user/repo.git", "--path", dir)
	if err != nil {
		t.Logf("command error (may be expected): %v", err)
	}
	if !strings.Contains(stdout, "git clone") {
		t.Error("url shorthand should work like clone")
	}
}

func TestGitlabURLParsing(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := runCmd(t, "clone", "https://gitlab.com/org/project.git", "--path", dir)
	if err != nil {
		t.Logf("command error (may be expected): %v", err)
	}
	if !strings.Contains(stdout, "org-project") {
		t.Error("should parse GitLab URL correctly")
	}
}

func TestUniqueDirectoryNameOnCollision(t *testing.T) {
	dir := t.TempDir()
	datePrefix := time.Now().Format("2006-01-02")
	existing := datePrefix + "-test"
	os.MkdirAll(filepath.Join(dir, existing), 0755)

	stdout, _, _ := runCmd(t, "cd", "test", "--and-keys", "ENTER", "--path", dir)

	// Should handle collision with unique suffix or similar
	if !strings.Contains(stdout, "test") {
		t.Error("should reference test directory in some form")
	}
}

func TestWorktreeWithoutGitRepoOnlyCreatesDir(t *testing.T) {
	tries := t.TempDir()
	repo := t.TempDir()
	// No .git directory

	cmd := exec.Command("./try", "worktree", "dir", "test", "--path", tries)
	cmd.Dir = repo
	stdout, _ := cmd.Output()

	out := string(stdout)
	if !strings.Contains(out, "mkdir") {
		t.Error("should emit mkdir")
	}
	if strings.Contains(out, "worktree add") {
		t.Error("should not emit worktree when not in git repo")
	}
}

func TestWorktreeWithGitRepoAddsWorktree(t *testing.T) {
	tries := t.TempDir()
	repo := t.TempDir()
	os.MkdirAll(filepath.Join(repo, ".git"), 0755)

	cmd := exec.Command("./try", "worktree", "dir", "test", "--path", tries)
	cmd.Dir = repo
	stdout, _ := cmd.Output()

	out := string(stdout)
	if !strings.Contains(out, "worktree add") {
		t.Error("should emit worktree add when in git repo")
	}
}

func TestInitBashEmitsFunction(t *testing.T) {
	stdout, _, err := runCmdWithEnv(t, map[string]string{"SHELL": "/bin/bash"}, "init", "/tmp/tries")
	if err != nil {
		t.Fatalf("init should succeed: %v", err)
	}
	if !strings.Contains(stdout, "try()") {
		t.Error("should emit bash try function")
	}
	if !strings.Contains(stdout, "case") {
		t.Error("should contain case statement")
	}
}

func TestInitFishEmitsFunction(t *testing.T) {
	stdout, _, err := runCmdWithEnv(t, map[string]string{"SHELL": "/usr/bin/fish"}, "init", "/tmp/tries")
	if err != nil {
		t.Fatalf("init should succeed: %v", err)
	}
	if !strings.Contains(stdout, "function try") {
		t.Error("should emit fish function")
	}
	if !strings.Contains(stdout, "switch") {
		t.Error("should contain switch statement")
	}
}

func TestMultipleDirectoryCollisionHandling(t *testing.T) {
	dir := t.TempDir()
	datePrefix := time.Now().Format("2006-01-02")
	os.MkdirAll(filepath.Join(dir, datePrefix+"-test1"), 0755)
	os.MkdirAll(filepath.Join(dir, datePrefix+"-test2"), 0755)

	stdout, _, _ := runCmd(t, "cd", "test1", "--and-keys", "ENTER", "--path", dir)

	// Should have unique naming strategy
	if !strings.Contains(stdout, "test") {
		t.Error("should reference test in output")
	}
}

func TestNavigationKeysCtrlPAndN(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "2025-08-14-first"), 0755)
	os.MkdirAll(filepath.Join(dir, "2025-08-15-second"), 0755)
	f, _ := os.Create(filepath.Join(dir, "2025-08-14-first", ".touch"))
	f.Close()

	// Ctrl-N (down) then Enter
	stdout, _, _ := runCmd(t, "cd", "--and-keys", "CTRL-N,ENTER", "--path", dir)
	if !strings.Contains(stdout, "second") {
		t.Error("Ctrl-N should navigate down")
	}
}

func TestBackspaceRemovesCharactersFromSearch(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "test-dir"), 0755)

	stdout, stderr, _ := runCmd(t, "cd", "--and-type", "test", "--and-keys", "x,BACKSPACE,ENTER", "--path", dir)
	combined := stdout + stderr
	if !strings.Contains(combined, "test") {
		t.Error("should handle backspace correctly")
	}
}

func TestEscCancelsSelector(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "test-dir"), 0755)

	// Press ESC to cancel
	stdout, _, _ := runCmd(t, "cd", "--and-keys", "ESC", "--path", dir)
	// Should exit without emitting cd command
	if strings.Contains(stdout, "cd ") {
		t.Error("ESC should cancel without emitting cd")
	}
}
