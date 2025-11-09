package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorktreeDirWithName(t *testing.T) {
	tries := t.TempDir()
	repo := t.TempDir()
	os.MkdirAll(filepath.Join(repo, ".git"), 0755)

	cmd := exec.Command("./try", "worktree", "dir", "xyz", "--path", tries)
	cmd.Dir = repo
	stdout, err := cmd.Output()
	if err != nil {
		t.Logf("command failed (may be expected): %v", err)
	}

	out := string(stdout)
	if !strings.Contains(out, "mkdir -p") || !strings.Contains(out, "xyz") {
		t.Error("should emit mkdir for date-prefixed xyz directory")
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "git worktree") || !strings.Contains(lower, "create this trial") {
		t.Error("should contain git worktree echo message")
	}
	if !strings.Contains(out, "worktree add --detach") {
		t.Error("should emit worktree add command")
	}
	if !strings.Contains(out, "cd") {
		t.Error("should emit cd command")
	}
}

func TestWorktreeDirWithoutGitSkipsWorktreeAndEcho(t *testing.T) {
	tries := t.TempDir()
	repo := t.TempDir()
	// No .git directory

	cmd := exec.Command("./try", "worktree", "dir", "xyz", "--path", tries)
	cmd.Dir = repo
	stdout, _ := cmd.Output()

	out := string(stdout)
	if strings.Contains(out, "worktree add --detach") {
		t.Error("should not emit worktree add when not in git repo")
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "git worktree") && strings.Contains(lower, "create this trial") {
		t.Error("should not contain git worktree echo message when not in repo")
	}
	if !strings.Contains(out, "mkdir -p") {
		t.Error("should still emit mkdir")
	}
}

func TestTryDotEmitsWorktreeStepAndUsesCwdName(t *testing.T) {
	proj := filepath.Join(t.TempDir(), "myproj")
	os.MkdirAll(filepath.Join(proj, ".git"), 0755)
	tries := t.TempDir()

	cmd := exec.Command("./try", "cd", "./", "--path", tries)
	cmd.Dir = proj
	stdout, _ := cmd.Output()

	out := string(stdout)
	if !strings.Contains(out, "worktree add --detach") {
		t.Error("should include git worktree step")
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "git worktree") || !strings.Contains(lower, "create this trial") {
		t.Error("should echo intent to use worktree")
	}
	if !strings.Contains(out, "myproj") {
		t.Error("should include date-prefixed cwd name")
	}
}

func TestTryDotWithNameOverridesBasename(t *testing.T) {
	proj := filepath.Join(t.TempDir(), "myproj")
	os.MkdirAll(filepath.Join(proj, ".git"), 0755)
	tries := t.TempDir()

	cmd := exec.Command("./try", "cd", ".", "custom-name", "--path", tries)
	cmd.Dir = proj
	stdout, _ := cmd.Output()

	out := string(stdout)
	if !strings.Contains(out, "worktree add --detach") || !strings.Contains(out, "custom-name") {
		t.Error("should use custom name in worktree command")
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "git worktree") || !strings.Contains(lower, "create this trial") {
		t.Error("should contain echo message")
	}
}

func TestTryDotWithoutGitSkipsWorktree(t *testing.T) {
	proj := filepath.Join(t.TempDir(), "plain")
	os.MkdirAll(proj, 0755)
	tries := t.TempDir()

	cmd := exec.Command("./try", "cd", ".", "--path", tries)
	cmd.Dir = proj
	stdout, _ := cmd.Output()

	out := string(stdout)
	if strings.Contains(out, "worktree add --detach") {
		t.Error("should not emit worktree when not in git repo")
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "git worktree") && strings.Contains(lower, "create this trial") {
		t.Error("should not contain git worktree echo message")
	}
	if !strings.Contains(out, "mkdir -p") || !strings.Contains(out, "plain") {
		t.Error("should emit mkdir for date-prefixed plain directory")
	}
}
