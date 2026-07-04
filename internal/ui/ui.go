// Package ui provides small colored-logging helpers for the CLI.
package ui

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes. Disabled automatically when stdout is not a terminal.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
)

// UI wraps stdout with leveled, colored output.
type UI struct {
	color bool
}

// New builds a UI, auto-detecting color support (honors NO_COLOR).
func New() *UI {
	color := true
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		color = false
	}
	if fi, err := os.Stdout.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		color = false
	}
	return &UI{color: color}
}

func (u *UI) paint(code, s string) string {
	if !u.color {
		return s
	}
	return code + s + reset
}

// Section prints a bold header for a group of work.
func (u *UI) Section(title string) {
	fmt.Println()
	fmt.Println(u.paint(bold+cyan, "▸ "+title))
}

// Step prints the module currently being evaluated.
func (u *UI) Step(id, title string) {
	fmt.Printf("%s %s %s\n", u.paint(bold+blue, "»"), u.paint(bold, id), u.paint(dim, title))
}

// OK prints a success line.
func (u *UI) OK(format string, a ...any) {
	fmt.Printf("  %s %s\n", u.paint(green, "✓"), fmt.Sprintf(format, a...))
}

// Skip prints an "already satisfied" line.
func (u *UI) Skip(format string, a ...any) {
	fmt.Printf("  %s %s\n", u.paint(green, "✓"), u.paint(dim, fmt.Sprintf(format, a...)))
}

// Info prints an informational line.
func (u *UI) Info(format string, a ...any) {
	fmt.Printf("  %s %s\n", u.paint(blue, "•"), fmt.Sprintf(format, a...))
}

// Warn prints a warning line.
func (u *UI) Warn(format string, a ...any) {
	fmt.Printf("  %s %s\n", u.paint(yellow, "!"), fmt.Sprintf(format, a...))
}

// Fail prints an error line.
func (u *UI) Fail(format string, a ...any) {
	fmt.Printf("  %s %s\n", u.paint(red, "✗"), u.paint(red, fmt.Sprintf(format, a...)))
}

// Cmd echoes a command that is about to run (or would run in dry-run).
func (u *UI) Cmd(prefix, line string) {
	fmt.Printf("    %s %s\n", u.paint(dim, prefix), u.paint(dim, line))
}

// Banner prints the program title.
func (u *UI) Banner(version string) {
	line := strings.Repeat("─", 52)
	fmt.Println(u.paint(cyan, line))
	fmt.Println(u.paint(bold+cyan, "  DevaStation") + u.paint(dim, "  "+version))
	fmt.Println(u.paint(dim, "  Ubuntu → estação de desenvolvimento Go"))
	fmt.Println(u.paint(cyan, line))
}
