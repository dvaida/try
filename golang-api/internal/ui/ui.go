package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var tokenMap = map[string]string{
	"{text}":           "\x1b[39m",
	"{dim_text}":       "\x1b[90m",
	"{h1}":             "\x1b[1;33m",
	"{h2}":             "\x1b[1;36m",
	"{highlight}":      "\x1b[1;33m",
	"{reset}":          "\x1b[0m\x1b[39m\x1b[49m",
	"{reset_bg}":       "\x1b[49m",
	"{reset_fg}":       "\x1b[39m",
	"{clear_screen}":   "\x1b[2J",
	"{clear_line}":     "\x1b[2K",
	"{home}":           "\x1b[H",
	"{clear_below}":    "\x1b[0J",
	"{hide_cursor}":    "\x1b[?25l",
	"{show_cursor}":    "\x1b[?25h",
	"{start_selected}": "\x1b[1m",
	"{end_selected}":   "\x1b[0m",
	"{bold}":           "\x1b[1m",
}

var buffer []string
var lastBuffer []string
var currentLine string
var output io.Writer = os.Stderr

func HasToken(token string) bool {
	_, exists := tokenMap[token]
	return exists
}

func ExpandTokens(text string) string {
	re := regexp.MustCompile(`\{.*?\}`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		if val, ok := tokenMap[match]; ok {
			return val
		}
		panic(fmt.Sprintf("Unknown token: %s", match))
	})
}

func SetOutput(w io.Writer) {
	output = w
}

func Print(text string) {
	if text == "" {
		return
	}
	currentLine += text
}

func Puts(text string) {
	currentLine += text
	buffer = append(buffer, currentLine)
	currentLine = ""
}

func Flush(isTTY bool) {
	if currentLine != "" {
		buffer = append(buffer, currentLine)
		currentLine = ""
	}

	if !isTTY {
		re := regexp.MustCompile(`\{.*?\}`)
		plain := re.ReplaceAllString(strings.Join(buffer, "\n"), "")
		fmt.Fprint(output, plain)
		if !strings.HasSuffix(plain, "\n") {
			fmt.Fprint(output, "\n")
		}
		lastBuffer = nil
		buffer = nil
		currentLine = ""
		return
	}

	fmt.Fprint(output, "\x1b[H")
	maxLines := len(buffer)
	if len(lastBuffer) > maxLines {
		maxLines = len(lastBuffer)
	}
	reset := tokenMap["{reset}"]

	for i := 0; i < maxLines; i++ {
		var currentBufLine, lastBufLine string
		if i < len(buffer) {
			currentBufLine = buffer[i]
		}
		if i < len(lastBuffer) {
			lastBufLine = lastBuffer[i]
		}

		if currentBufLine != lastBufLine {
			fmt.Fprintf(output, "\x1b[%d;1H\x1b[2K", i+1)
			if currentBufLine != "" {
				processed := ExpandTokens(currentBufLine)
				fmt.Fprint(output, processed)
				fmt.Fprint(output, reset)
			}
		}
	}

	lastBuffer = make([]string, len(buffer))
	copy(lastBuffer, buffer)
	buffer = nil
	currentLine = ""
}

func Cls() {
	currentLine = ""
	buffer = nil
	lastBuffer = nil
	fmt.Fprint(output, "\x1b[2J\x1b[H")
}

func Height() int {
	cmd := exec.Command("tput", "lines")
	out, err := cmd.Output()
	if err != nil {
		return 24
	}
	h, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || h <= 0 {
		return 24
	}
	return h
}

func Width() int {
	cmd := exec.Command("tput", "cols")
	out, err := cmd.Output()
	if err != nil {
		return 80
	}
	w, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

func Reset() {
	buffer = nil
	lastBuffer = nil
	currentLine = ""
}

func GetCurrentLine() string {
	return currentLine
}

func GetBuffer() []string {
	return buffer
}
