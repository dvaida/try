package selector

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tobi/try/golang-api/internal/ui"
	"golang.org/x/term"
)

type TryInfo struct {
	Name     string
	Basename string
	Path     string
	IsNew    bool
	Ctime    time.Time
	Mtime    time.Time
	Score    float64
}

type TrySelector struct {
	SearchTerm     string
	CursorPos      int
	ScrollOffset   int
	InputBuffer    string
	Selected       map[string]interface{}
	AllTries       []TryInfo
	BasePath       string
	DeleteStatus   string
	TestRenderOnce bool
	TestNoCls      bool
	TestKeys       []string
	TestConfirm    string
}

func NewTrySelector(searchTerm, basePath string, options map[string]interface{}) *TrySelector {
	ts := &TrySelector{
		SearchTerm:  strings.ReplaceAll(searchTerm, " ", "-"),
		BasePath:    basePath,
		CursorPos:   0,
		ScrollOffset: 0,
		InputBuffer: strings.ReplaceAll(searchTerm, " ", "-"),
	}

	if initialInput, ok := options["initial_input"].(string); ok && initialInput != "" {
		ts.InputBuffer = strings.ReplaceAll(initialInput, " ", "-")
	} else if ts.SearchTerm != "" {
		ts.InputBuffer = ts.SearchTerm
	}

	if renderOnce, ok := options["test_render_once"].(bool); ok {
		ts.TestRenderOnce = renderOnce
	}
	if noCls, ok := options["test_no_cls"].(bool); ok {
		ts.TestNoCls = noCls
	}
	if keys, ok := options["test_keys"].([]string); ok {
		ts.TestKeys = keys
	}
	if confirm, ok := options["test_confirm"].(string); ok {
		ts.TestConfirm = confirm
	}

	os.MkdirAll(basePath, 0755)
	return ts
}

func (ts *TrySelector) Run() map[string]interface{} {
	if ts.TestRenderOnce {
		tries := ts.GetTries()
		ts.render(tries)
		return nil
	}

	// Determine if we're in test mode
	isTestMode := len(ts.TestKeys) > 0 || ts.TestNoCls

	// Setup terminal raw mode for interactive input (only in non-test mode)
	var oldState *term.State
	if !isTestMode {
		var err error
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up terminal: %v\n", err)
			return nil
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		// Hide cursor
		fmt.Fprint(os.Stderr, "\x1b[?25l")
		defer fmt.Fprint(os.Stderr, "\x1b[?25h")

		// Clear screen initially
		ui.Cls()
	}

	for {
		tries := ts.GetTries()
		totalItems := len(tries) + 1

		if ts.CursorPos < 0 {
			ts.CursorPos = 0
		}
		if ts.CursorPos >= totalItems {
			ts.CursorPos = totalItems - 1
		}

		ts.render(tries)

		key := ts.readKey()

		switch key {
		case "\r":
			// Clear screen before exit (only in non-test mode)
			if !isTestMode {
				ui.Cls()
			}
			if ts.CursorPos < len(tries) {
				ts.Selected = map[string]interface{}{
					"type": "cd",
					"path": tries[ts.CursorPos].Path,
				}
				return ts.Selected
			} else {
				return ts.handleCreateNew()
			}
		case "\x1b[A", "\x10", "\x0B":
			if ts.CursorPos > 0 {
				ts.CursorPos--
			}
		case "\x1b[B", "\x0E", "\n":
			if ts.CursorPos < totalItems-1 {
				ts.CursorPos++
			}
		case "\x7F", "\b":
			if len(ts.InputBuffer) > 0 {
				ts.InputBuffer = ts.InputBuffer[:len(ts.InputBuffer)-1]
				ts.CursorPos = 0
			}
		case "\x04":
			if ts.CursorPos < len(tries) {
				ts.handleDelete(tries[ts.CursorPos])
				ts.AllTries = nil
			}
		case "\x03", "\x1b":
			// Clear screen before exit (only in non-test mode)
			if !isTestMode {
				ui.Cls()
			}
			ts.Selected = nil
			return nil
		default:
			if len(key) == 1 && regexp.MustCompile(`[a-zA-Z0-9\-\_\. ]`).MatchString(key) {
				ts.InputBuffer += key
				ts.CursorPos = 0
			}
		}
	}
}

func (ts *TrySelector) LoadAllTries() []TryInfo {
	if ts.AllTries != nil {
		return ts.AllTries
	}

	var tries []TryInfo
	entries, err := ioutil.ReadDir(ts.BasePath)
	if err != nil {
		return tries
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(ts.BasePath, entry.Name())
		stat, err := os.Stat(path)
		if err != nil {
			continue
		}

		tries = append(tries, TryInfo{
			Name:     "📁 " + entry.Name(),
			Basename: entry.Name(),
			Path:     path,
			IsNew:    false,
			Ctime:    stat.ModTime(),
			Mtime:    stat.ModTime(),
		})
	}

	ts.AllTries = tries
	return tries
}

func (ts *TrySelector) GetTries() []TryInfo {
	allTries := ts.LoadAllTries()

	scored := make([]TryInfo, len(allTries))
	for i, try := range allTries {
		score := ts.CalculateScore(try.Basename, ts.InputBuffer, try.Ctime.Unix(), try.Mtime.Unix())
		tryWithScore := try
		tryWithScore.Score = score
		scored[i] = tryWithScore
	}

	if ts.InputBuffer == "" {
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Score > scored[j].Score
		})
		return scored
	}

	var filtered []TryInfo
	for _, try := range scored {
		if try.Score > 0 {
			filtered = append(filtered, try)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	return filtered
}

func (ts *TrySelector) CalculateScore(text, query string, ctime, mtime int64) float64 {
	score := 0.0

	if regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`).MatchString(text) {
		score += 2.0
	}

	if query != "" {
		textLower := strings.ToLower(text)
		queryLower := strings.ToLower(query)
		queryChars := []rune(queryLower)

		lastPos := -1
		queryIdx := 0

		for pos, char := range textLower {
			if queryIdx >= len(queryChars) {
				break
			}
			if char != queryChars[queryIdx] {
				continue
			}

			score += 1.0
			if pos == 0 || !regexp.MustCompile(`\w`).MatchString(string(textLower[pos-1])) {
				score += 1.0
			}

			if lastPos >= 0 {
				gap := pos - lastPos - 1
				score += 1.0 / math.Sqrt(float64(gap+1))
			}

			lastPos = pos
			queryIdx++
		}

		if queryIdx < len(queryChars) {
			return 0.0
		}

		if lastPos >= 0 {
			score *= float64(len(queryChars)) / float64(lastPos+1)
		}

		score *= 10.0 / float64(len(text)+10)
	}

	now := time.Now().Unix()
	if ctime > 0 {
		daysOld := float64(now-ctime) / 86400.0
		score += 2.0 / math.Sqrt(daysOld+1)
	}

	if mtime > 0 {
		hoursSinceAccess := float64(now-mtime) / 3600.0
		score += 3.0 / math.Sqrt(hoursSinceAccess+1)
	}

	return score
}

func (ts *TrySelector) FormatRelativeTime(t time.Time) string {
	seconds := time.Since(t).Seconds()
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	if seconds < 10 {
		return "just now"
	} else if minutes < 60 {
		return fmt.Sprintf("%dm ago", int(minutes))
	} else if hours < 24 {
		return fmt.Sprintf("%dh ago", int(hours))
	} else if days < 30 {
		return fmt.Sprintf("%dd ago", int(days))
	} else if days < 365 {
		return fmt.Sprintf("%dmo ago", int(days/30))
	} else {
		return fmt.Sprintf("%dy ago", int(days/365))
	}
}

func (ts *TrySelector) HighlightMatches(text, query string) string {
	if query == "" {
		return text
	}

	result := ""
	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)
	queryChars := []rune(queryLower)
	queryIdx := 0

	for i, char := range text {
		if queryIdx < len(queryChars) && rune(textLower[i]) == queryChars[queryIdx] {
			result += "{highlight}" + string(char) + "{text}"
			queryIdx++
		} else {
			result += string(char)
		}
	}

	return result
}

func (ts *TrySelector) TruncateWithANSI(text string, maxLength int) string {
	visibleCount := 0
	result := ""
	inAnsi := false

	for _, char := range text {
		if char == '\x1b' {
			inAnsi = true
			result += string(char)
		} else if inAnsi {
			result += string(char)
			if char == 'm' {
				inAnsi = false
			}
		} else {
			if visibleCount >= maxLength {
				break
			}
			result += string(char)
			visibleCount++
		}
	}

	return result
}

func (ts *TrySelector) render(tries []TryInfo) {
	ui.Puts("{h1}📁 Try Directory Selection")
	ui.Puts("{dim_text}────────────────────────────────────────")
	ui.Puts(fmt.Sprintf("{highlight}Search: {reset}%s", ts.InputBuffer))
	ui.Puts("{dim_text}────────────────────────────────────────")

	totalItems := len(tries) + 1

	for idx := 0; idx < totalItems; idx++ {
		if idx == len(tries) && len(tries) > 0 {
			ui.Puts("")
		}

		isSelected := idx == ts.CursorPos
		if isSelected {
			ui.Print("{highlight}→ {reset_fg}")
		} else {
			ui.Print("  ")
		}

		if idx < len(tries) {
			try := tries[idx]
			ui.Print("📁 ")

			if isSelected {
				ui.Print("{start_selected}")
			}

			if regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`).MatchString(try.Basename) {
				parts := strings.SplitN(try.Basename, "-", 4)
				if len(parts) >= 4 {
					datepart := strings.Join(parts[0:3], "-")
					namepart := strings.Join(parts[3:], "-")

					ui.Print("{dim_text}" + datepart + "{reset_fg}")
					if ts.InputBuffer != "" && strings.Contains(ts.InputBuffer, "-") {
						ui.Print("{highlight}-{reset_fg}")
					} else {
						ui.Print("{dim_text}-{reset_fg}")
					}

					if ts.InputBuffer != "" {
						ui.Print(ts.HighlightMatches(namepart, ts.InputBuffer))
					} else {
						ui.Print(namepart)
					}
				}
			} else {
				if ts.InputBuffer != "" {
					ui.Print(ts.HighlightMatches(try.Basename, ts.InputBuffer))
				} else {
					ui.Print(try.Basename)
				}
			}

			timeText := ts.FormatRelativeTime(try.Mtime)
			scoreText := fmt.Sprintf("%.1f", try.Score)
			metaText := fmt.Sprintf("%s, %s", timeText, scoreText)

			ui.Print(" ")
			if isSelected {
				ui.Print("{end_selected}")
			}
			ui.Print("{dim_text}" + metaText + "{reset_fg}")
		} else {
			ui.Print("+ ")
			if isSelected {
				ui.Print("{start_selected}")
			}

			if ts.InputBuffer == "" {
				ui.Print("Create new")
			} else {
				ui.Print(fmt.Sprintf("Create new: %s", ts.InputBuffer))
			}

			if isSelected {
				ui.Print("{end_selected}")
			}
		}

		ui.Puts("")
	}

	ui.Puts("{dim_text}────────────────────────────────────────")

	if ts.DeleteStatus != "" {
		ui.Puts("{h1}Delete Directory")
		ui.Puts("{highlight}" + ts.DeleteStatus + "{reset}")
		ts.DeleteStatus = ""
	} else {
		ui.Puts("{dim_text}↑↓/Ctrl-P,N,J,K: Navigate  Enter: Select  Ctrl-D: Delete  ESC: Cancel{reset}")
	}

	// Use TTY mode unless we're in test mode
	isTTY := !ts.TestNoCls && len(ts.TestKeys) == 0
	ui.Flush(isTTY)
}

func (ts *TrySelector) readKey() string {
	// For testing mode, use pre-defined keys
	if len(ts.TestKeys) > 0 {
		key := ts.TestKeys[0]
		ts.TestKeys = ts.TestKeys[1:]
		return key
	}

	// Read from stdin in raw mode
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return "\x1b" // ESC on error
	}

	// Handle escape sequences
	if n == 1 {
		return string(buf[0])
	}

	// Multi-byte escape sequence
	if buf[0] == '\x1b' {
		if n == 2 && buf[1] == '\x1b' {
			// Double ESC, treat as single ESC
			return "\x1b"
		}
		if n >= 3 && buf[1] == '[' {
			// ANSI escape sequence
			switch buf[2] {
			case 'A': // Up arrow
				return "\x1b[A"
			case 'B': // Down arrow
				return "\x1b[B"
			case 'C': // Right arrow
				return "\x1b[C"
			case 'D': // Left arrow
				return "\x1b[D"
			}
		}
		// Unknown escape sequence, return ESC
		return "\x1b"
	}

	return string(buf[:n])
}

func (ts *TrySelector) handleCreateNew() map[string]interface{} {
	datePrefix := time.Now().Format("2006-01-02")

	var finalName string
	if ts.InputBuffer != "" {
		finalName = datePrefix + "-" + ts.InputBuffer
		finalName = strings.ReplaceAll(finalName, " ", "-")
		fullPath := filepath.Join(ts.BasePath, finalName)
		return map[string]interface{}{
			"type": "mkdir",
			"path": fullPath,
		}
	}

	return map[string]interface{}{
		"type": "cancel",
		"path": "",
	}
}

func (ts *TrySelector) handleDelete(try TryInfo) {
	if ts.TestConfirm == "YES" {
		os.RemoveAll(try.Path)
		ts.DeleteStatus = fmt.Sprintf("Deleted: %s", try.Basename)
		ts.AllTries = nil
	} else {
		ts.DeleteStatus = "Delete cancelled"
	}
}
