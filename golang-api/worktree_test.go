package main

import (
	"io"
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

	stdout, _, err := runCmdInDir(t, repo, "worktree", "dir", "xyz", "--path", tries)
	if err != nil {
		t.Logf("command failed (may be expected): %v", err)
	}

	out := stdout
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

	stdout, _, _ := runCmdInDir(t, repo, "worktree", "dir", "xyz", "--path", tries)

	out := stdout
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

	stdout, _, _ := runCmdInDir(t, proj, "cd", "./", "--path", tries)

	out := stdout
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

	stdout, _, _ := runCmdInDir(t, proj, "cd", ".", "custom-name", "--path", tries)

	out := stdout
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

	stdout, _, _ := runCmdInDir(t, proj, "cd", ".", "--path", tries)

	out := stdout
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

func runCmdInDir(t *testing.T, dir string, args ...string) (string, string, error) {
	t.Helper()
	tryPath, _ := filepath.Abs("./try")
	cmd := exec.Command(tryPath, args...)
	cmd.Dir = dir

	var stdoutBuf, stderrBuf []byte
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", "", err
	}

	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})

	go func() {
		stdoutBuf, _ = io.ReadAll(stdoutPipe)
		close(stdoutDone)
	}()

	go func() {
		stderrBuf, _ = io.ReadAll(stderrPipe)
		close(stderrDone)
	}()

	<-stdoutDone
	<-stderrDone

	err := cmd.Wait()

	return string(stdoutBuf), string(stderrBuf), err
}
