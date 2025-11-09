package main

import (
	"os/exec"
	"strings"
	"testing"
)

func runCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command("./try", args...)
	stdout, err := cmd.Output()
	var stderr []byte
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	return string(stdout), string(stderr), err
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
