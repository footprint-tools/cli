package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Skryensya/footprint/internal/domain"
	"golang.org/x/term"
)

// Writer implements domain.OutputWriter for stdout.
type Writer struct {
	out           io.Writer
	pagerDisabled bool
	pagerOverride string
	configGetter  func(string) (string, bool)
	envGetter     func(string) string
}

// WriterOption configures a Writer.
type WriterOption func(*Writer)

// WithPagerDisabled disables the pager.
func WithPagerDisabled() WriterOption {
	return func(w *Writer) {
		w.pagerDisabled = true
	}
}

// WithPagerOverride sets a pager command override.
func WithPagerOverride(cmd string) WriterOption {
	return func(w *Writer) {
		w.pagerOverride = cmd
	}
}

// WithConfigGetter sets the config getter function.
func WithConfigGetter(fn func(string) (string, bool)) WriterOption {
	return func(w *Writer) {
		w.configGetter = fn
	}
}

// WithEnvGetter sets the environment variable getter function.
func WithEnvGetter(fn func(string) string) WriterOption {
	return func(w *Writer) {
		w.envGetter = fn
	}
}

// NewWriter creates a new Writer that writes to stdout.
func NewWriter(opts ...WriterOption) *Writer {
	w := &Writer{
		out:       os.Stdout,
		envGetter: os.Getenv,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// NewWriterTo creates a new Writer that writes to the specified writer.
func NewWriterTo(out io.Writer, opts ...WriterOption) *Writer {
	w := &Writer{
		out:       out,
		envGetter: os.Getenv,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.out.Write(p)
}

// Printf formats and prints to the output.
func (w *Writer) Printf(format string, args ...any) (int, error) {
	return fmt.Fprintf(w.out, format, args...)
}

// Println prints a line to the output.
func (w *Writer) Println(args ...any) (int, error) {
	return fmt.Fprintln(w.out, args...)
}

// Pager displays content through a pager if appropriate.
func (w *Writer) Pager(content string) {
	// 1. Pager disabled
	if w.pagerDisabled {
		fmt.Fprint(w.out, content)
		return
	}

	// 2. Not a TTY (check if output supports Fd())
	if f, ok := w.out.(*os.File); ok {
		if !term.IsTerminal(int(f.Fd())) {
			fmt.Fprint(w.out, content)
			return
		}
	} else {
		// Non-file outputs (like bytes.Buffer) - just print
		fmt.Fprint(w.out, content)
		return
	}

	// 3. Pager override flag
	if w.pagerOverride != "" {
		if w.isBypassPager(w.pagerOverride) {
			fmt.Fprint(w.out, content)
			return
		}
		w.runPagerCmd(w.pagerOverride, content)
		return
	}

	// 4. Config pager
	if w.configGetter != nil {
		if configPager, ok := w.configGetter("pager"); ok && configPager != "" {
			if w.isBypassPager(configPager) {
				fmt.Fprint(w.out, content)
				return
			}
			w.runPagerCmd(configPager, content)
			return
		}
	}

	// 5. $PAGER environment variable
	if w.envGetter != nil {
		if envPager := w.envGetter("PAGER"); envPager != "" {
			if w.isBypassPager(envPager) {
				fmt.Fprint(w.out, content)
				return
			}
			w.runPagerCmd(envPager, content)
			return
		}
	}

	// 6. Default: less with standard flags
	w.runPager("less", []string{"-FRSX"}, content)
}

func (w *Writer) isBypassPager(cmd string) bool {
	return cmd == "cat"
}

func (w *Writer) runPagerCmd(pagerCmd string, content string) {
	parts := strings.Fields(pagerCmd)
	if len(parts) == 0 {
		fmt.Fprint(w.out, content)
		return
	}
	w.runPager(parts[0], parts[1:], content)
}

func (w *Writer) runPager(pager string, args []string, content string) {
	cmd := exec.Command(pager, args...)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprint(w.out, content)
	}
}

// Verify Writer implements domain.OutputWriter
var _ domain.OutputWriter = (*Writer)(nil)
