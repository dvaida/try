package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tobi/try/golang-api/internal/ui"
)

func TestTokenMapContainsCoreTokens(t *testing.T) {
	coreTokens := []string{"{text}", "{dim_text}", "{h1}", "{h2}", "{highlight}", "{reset}", "{reset_bg}", "{reset_fg}"}
	for _, tok := range coreTokens {
		if !ui.HasToken(tok) {
			t.Errorf("missing token %s", tok)
		}
	}
}

func TestExpandTokensSubstitutesSequences(t *testing.T) {
	sample := "{h1}Title{reset}"
	expanded := ui.ExpandTokens(sample)
	if sample == expanded {
		t.Error("should have expanded tokens")
	}
	if !strings.Contains(expanded, "\x1b[") {
		t.Error("expanded should contain ANSI")
	}
}

func TestFlushStripsTokensForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	ui.SetOutput(&buf)
	ui.Puts("{h2}Hello{reset}")
	ui.Flush(false) // false = non-TTY
	out := buf.String()
	if !strings.Contains(out, "Hello") {
		t.Error("output should contain Hello")
	}
	if strings.Contains(out, "{h2}") || strings.Contains(out, "{reset}") {
		t.Error("should have stripped tokens")
	}
}

func TestClsClearsBuffersAndScreen(t *testing.T) {
	var buf bytes.Buffer
	ui.SetOutput(&buf)
	ui.Puts("some content")
	ui.Cls()
	out := buf.String()
	if !strings.Contains(out, "\x1b[2J") || !strings.Contains(out, "\x1b[H") {
		t.Error("cls should emit clear screen and home sequences")
	}
}

func TestHeightReturnsPositiveInteger(t *testing.T) {
	height := ui.Height()
	if height <= 0 {
		t.Error("height should return positive value")
	}
}

func TestWidthReturnsPositiveInteger(t *testing.T) {
	width := ui.Width()
	if width <= 0 {
		t.Error("width should return positive value")
	}
}

func TestExpandTokensRaisesOnUnknownToken(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic on unknown token")
		}
	}()
	ui.ExpandTokens("{unknown_token}")
}

func TestPrintAccumulatesToCurrentLine(t *testing.T) {
	ui.Reset()
	ui.Print("hello")
	ui.Print(" world")
	line := ui.GetCurrentLine()
	if line != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", line)
	}
}

func TestPutsAddsToBuffer(t *testing.T) {
	ui.Reset()
	ui.Puts("line1")
	ui.Puts("line2")
	buffer := ui.GetBuffer()
	if len(buffer) != 2 || buffer[0] != "line1" || buffer[1] != "line2" {
		t.Errorf("expected [line1, line2], got %v", buffer)
	}
}
