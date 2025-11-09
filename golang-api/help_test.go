package main

import (
	"io"
	"os/exec"
	"strings"
	"testing"
)

func runCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command("./try", args...)

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

func TestHelpFlagPrintsUsage(t *testing.T) {
	stdout, _, err := runCmd(t, "--help")
	if err != nil && !strings.Contains(err.Error(), "exit status 0") {
		t.Fatalf("expected --help to exit successfully, got error: %v", err)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Error("help should print usage to stdout")
	}
}

func TestNoArgsPrintsUsage(t *testing.T) {
	stdout, _, err := runCmd(t)
	if err == nil {
		t.Log("command succeeded (expected non-zero exit for no args)")
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Error("running without args should print usage")
	}
}
