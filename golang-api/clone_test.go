package main

import (
	"strings"
	"testing"
)

func TestCloneGeneratesScript(t *testing.T) {
	dir := t.TempDir()
	stdout, _, _ := runCmd(t, "clone", "https://github.com/tobi/try.git", "my-fork", "--path", dir)
	if !strings.Contains(stdout, "mkdir -p") || !strings.Contains(stdout, "my-fork") {
		t.Error("should emit mkdir with my-fork")
	}
	if !strings.Contains(stdout, "git clone 'https://github.com/tobi/try.git'") {
		t.Error("should emit git clone command")
	}
	if !strings.Contains(stdout, "cd") || !strings.Contains(stdout, "my-fork") {
		t.Error("should emit cd into my-fork")
	}
}

func TestCdUrlShorthandWithName(t *testing.T) {
	dir := t.TempDir()
	stdout, _, _ := runCmd(t, "cd", "https://github.com/tobi/try.git", "my-fork", "--path", dir)
	if !strings.Contains(stdout, "git clone 'https://github.com/tobi/try.git'") {
		t.Error("should emit git clone command")
	}
	if !strings.Contains(stdout, "my-fork") {
		t.Error("should use the provided custom name in path")
	}
}

func TestCdCloneWrapperEmitsCloneScript(t *testing.T) {
	dir := t.TempDir()
	stdout, _, _ := runCmd(t, "cd", "clone", "https://github.com/tobi/try.git", "my-fork", "--path", dir)
	if !strings.Contains(stdout, "mkdir -p") || !strings.Contains(stdout, "my-fork") {
		t.Error("should emit mkdir with my-fork")
	}
	if !strings.Contains(stdout, "git clone") {
		t.Error("should emit git clone command")
	}
	if !strings.Contains(stdout, "cd") || !strings.Contains(stdout, "my-fork") {
		t.Error("should emit cd command")
	}
}

func TestCloneEchoMessagePresent(t *testing.T) {
	dir := t.TempDir()
	stdout, _, _ := runCmd(t, "clone", "https://github.com/tobi/try.git", "my-fork", "--path", dir)
	lower := strings.ToLower(stdout)
	if !strings.Contains(lower, "git clone") || !strings.Contains(lower, "create this trial") {
		t.Error("should contain echo message about git clone creating trial")
	}
}

func TestCdUrlEchoMessagePresent(t *testing.T) {
	dir := t.TempDir()
	stdout, _, _ := runCmd(t, "cd", "https://github.com/tobi/try.git", "my-fork", "--path", dir)
	lower := strings.ToLower(stdout)
	if !strings.Contains(lower, "git clone") || !strings.Contains(lower, "create this trial") {
		t.Error("should contain echo message about git clone creating trial")
	}
}
