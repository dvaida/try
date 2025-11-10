package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runCmdWithEnv(t *testing.T, env map[string]string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command("./try", args...)
	for k, v := range env {
		cmd.Env = append(os.Environ(), k+"="+v)
	}
	stdout, err := cmd.Output()
	var stderr []byte
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	return string(stdout), string(stderr), err
}

func TestInitEmitsBashFunctionWithPath(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := runCmdWithEnv(t, map[string]string{"SHELL": "/bin/bash"}, "init", dir)
	if err != nil {
		t.Fatalf("init should exit successfully: %v", err)
	}
	if !strings.Contains(stdout, "try()") {
		t.Error("should emit try function")
	}
	if !strings.Contains(stdout, "cd --path") {
		t.Error("should contain cd --path")
	}
	if !strings.Contains(stdout, "case") {
		t.Error("should contain case statement")
	}
	if !strings.Contains(stdout, `eval "$cmd"`) {
		t.Error("should contain eval statement")
	}
}

func TestInitEmitsFishFunctionWithPath(t *testing.T) {
	dir := t.TempDir()
	stdout, _, err := runCmdWithEnv(t, map[string]string{"SHELL": "/usr/bin/fish"}, "init", dir)
	if err != nil {
		t.Fatalf("init should exit successfully: %v", err)
	}
	if !strings.Contains(stdout, "function try") {
		t.Error("should emit fish function")
	}
	if !strings.Contains(stdout, "cd --path") {
		t.Error("should contain cd --path")
	}
	if !strings.Contains(stdout, "string collect") {
		t.Error("should contain string collect")
	}
	if !strings.Contains(stdout, "switch") {
		t.Error("should contain switch statement")
	}
}
