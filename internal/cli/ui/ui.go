// Package ui provides a tiny, dependency-free helper for consistent terminal
// output across the repo's developer tooling (e.g. the cmd/scaffold audit and
// cmd/scaffold drift subcommands). It renders emoji status markers, progress
// counters, section headers and summaries, and automatically falls back to
// plain ASCII when the output is not an interactive terminal or when NO_COLOR
// is set, so CI logs stay clean.
package ui

import (
	"fmt"
	"io"
	"os"
)

// Out and Err are the writers used by the package-level helpers. They default
// to the process stdout/stderr but can be swapped in tests.
var (
	Out io.Writer = os.Stdout
	Err io.Writer = os.Stderr
)

// enabled reports whether decorated (emoji) output should be used. It is
// computed once at init from the real stdout: emoji are used only on an
// interactive terminal and when NO_COLOR is unset.
var enabled = detectEnabled()

func detectEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isTerminal(os.Stdout)
}

func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// symbol returns the emoji form when decorations are enabled, otherwise the
// ASCII fallback.
func symbol(emoji, ascii string) string {
	if enabled {
		return emoji
	}
	return ascii
}

// Status markers. Kept as functions (not consts) so they respect the
// runtime-detected `enabled` flag.
func success() string { return symbol("✅", "OK  ") }
func failure() string { return symbol("❌", "FAIL") }
func warning() string { return symbol("⚠️ ", "WARN") }
func info() string    { return symbol("🔍", "--  ") }

// Arrow renders a right-pointing arrow suitable for inline "from -> to" text.
func Arrow() string { return symbol("→", "->") }

// Header prints a blank line followed by a decorated section title.
func Header(title string) {
	if enabled {
		_, _ = fmt.Fprintf(Out, "\n🚀 %s\n", title)
		return
	}
	_, _ = fmt.Fprintf(Out, "\n== %s ==\n", title)
}

// Step prints a progress line of the form "[cur/total] <message>".
func Step(cur, total int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(Out, "[%d/%d] %s\n", cur, total, msg)
}

// Successf prints a success-prefixed line to Out.
func Successf(format string, args ...any) {
	_, _ = fmt.Fprintf(Out, "%s %s\n", success(), fmt.Sprintf(format, args...))
}

// Infof prints an info-prefixed line to Out.
func Infof(format string, args ...any) {
	_, _ = fmt.Fprintf(Out, "%s %s\n", info(), fmt.Sprintf(format, args...))
}

// Warnf prints a warning-prefixed line to Err.
func Warnf(format string, args ...any) {
	_, _ = fmt.Fprintf(Err, "%s %s\n", warning(), fmt.Sprintf(format, args...))
}

// Errorf prints an error-prefixed line to Err.
func Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(Err, "%s %s\n", failure(), fmt.Sprintf(format, args...))
}

// Summary prints a one-line tally of clean / warning / error counts.
func Summary(clean, warnings, errors int) {
	_, _ = fmt.Fprintf(Out, "\n%s %d clean   %s %d warnings   %s %d errors\n",
		success(), clean, warning(), warnings, failure(), errors)
}
