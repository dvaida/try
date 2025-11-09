package ui

import "io"

var tokenMap = map[string]string{
	"{text}":       "\x1b[39m",
	"{dim_text}":   "\x1b[90m",
	"{h1}":         "\x1b[1;33m",
	"{h2}":         "\x1b[1;36m",
	"{highlight}":  "\x1b[1;33m",
	"{reset}":      "\x1b[0m\x1b[39m\x1b[49m",
	"{reset_bg}":   "\x1b[49m",
	"{reset_fg}":   "\x1b[39m",
	"{clear_screen}": "\x1b[2J",
	"{clear_line}":   "\x1b[2K",
	"{home}":         "\x1b[H",
	"{clear_below}":  "\x1b[0J",
	"{hide_cursor}":  "\x1b[?25l",
	"{show_cursor}":  "\x1b[?25h",
	"{start_selected}": "\x1b[1m",
	"{end_selected}":   "\x1b[0m",
	"{bold}":           "\x1b[1m",
}

var buffer []string
var currentLine string
var output io.Writer

func HasToken(token string) bool {
	panic("not implemented")
}

func ExpandTokens(text string) string {
	panic("not implemented")
}

func SetOutput(w io.Writer) {
	panic("not implemented")
}

func Print(text string) {
	panic("not implemented")
}

func Puts(text string) {
	panic("not implemented")
}

func Flush(isTTY bool) {
	panic("not implemented")
}

func Cls() {
	panic("not implemented")
}

func Height() int {
	panic("not implemented")
}

func Width() int {
	panic("not implemented")
}

func Reset() {
	panic("not implemented")
}

func GetCurrentLine() string {
	panic("not implemented")
}

func GetBuffer() []string {
	panic("not implemented")
}
