// Package ui provides terminal UI utilities including pager support.
//
// SECURITY NOTE: The pager functionality intentionally allows execution of
// arbitrary commands specified via --pager flag or config. This is standard
// behavior for CLI tools (similar to git, less, man) and requires local
// access to exploit. Users should only configure pagers they trust.
package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/footprint-tools/cli/internal/config"
	"golang.org/x/term"
)

var (
	pagerDisabled bool
	pagerOverride string
)

// DisablePager disables the pager globally (used by --no-pager flag).
func DisablePager() {
	pagerDisabled = true
}

// SetPager sets a pager override for this invocation (used by --pager flag).
func SetPager(cmd string) {
	pagerOverride = cmd
}

// isBypassPager returns true if the pager command means "bypass pager".
func isBypassPager(cmd string) bool {
	return cmd == "cat"
}

// Pager displays content through a pager if appropriate.
//
// Precedence:
//  1. --no-pager flag → direct output
//  2. stdout not a TTY → direct output
//  3. --pager=<cmd> flag → uses flag pager, "cat" bypasses
//  4. fp config pager → uses configured pager, "cat" bypasses
//  5. $PAGER env var → uses env pager, "cat" bypasses
//  6. Default: "less -FRSX"
func Pager(content string) {
	// 1. --no-pager flag
	if pagerDisabled {
		fmt.Print(content)
		return
	}

	// 2. Not a TTY
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Print(content)
		return
	}

	// 3. --pager=<cmd> flag override
	if pagerOverride != "" {
		if isBypassPager(pagerOverride) {
			fmt.Print(content)
			return
		}
		runPagerCmd(pagerOverride, content)
		return
	}

	// 4. Check config override
	if configPager, ok := config.Get("pager"); ok {
		if isBypassPager(configPager) {
			fmt.Print(content)
			return
		}
		runPagerCmd(configPager, content)
		return
	}

	// 5. $PAGER environment variable
	if envPager := os.Getenv("PAGER"); envPager != "" {
		if isBypassPager(envPager) {
			fmt.Print(content)
			return
		}
		runPagerCmd(envPager, content)
		return
	}

	// 6. Default: less with standard flags
	runPager("less", []string{"-FRSX"}, content)
}

// runPagerCmd parses a pager command string (e.g., "less -R") and executes it.
func runPagerCmd(pagerCmd string, content string) {
	parts := strings.Fields(pagerCmd)
	if len(parts) == 0 {
		fmt.Print(content)
		return
	}

	runPager(parts[0], parts[1:], content)
}

// runPager executes the pager command with the given content.
// Falls back to direct output on error.
func runPager(pager string, args []string, content string) {
	cmd := exec.Command(pager, args...)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Print(content)
	}
}
