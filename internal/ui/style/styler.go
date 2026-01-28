package style

import "github.com/footprint-tools/cli/internal/domain"

// Styler implements domain.Styler using the global style functions.
type Styler struct{}

// NewStyler creates a new Styler instance.
func NewStyler() *Styler {
	return &Styler{}
}

// Enabled returns true if styling is enabled.
func (s *Styler) Enabled() bool {
	return Enabled()
}

// Success styles text as success.
func (s *Styler) Success(text string) string {
	return Success(text)
}

// Warning styles text as warning.
func (s *Styler) Warning(text string) string {
	return Warning(text)
}

// Error styles text as error.
func (s *Styler) Error(text string) string {
	return Error(text)
}

// Info styles text as info.
func (s *Styler) Info(text string) string {
	return Info(text)
}

// Muted styles text as muted.
func (s *Styler) Muted(text string) string {
	return Muted(text)
}

// Header styles text as header.
func (s *Styler) Header(text string) string {
	return Header(text)
}

// NopStyler is a no-op styler that returns text unchanged.
// Useful for testing or when styling is disabled.
type NopStyler struct{}

func (NopStyler) Enabled() bool           { return false }
func (NopStyler) Success(text string) string { return text }
func (NopStyler) Warning(text string) string { return text }
func (NopStyler) Error(text string) string   { return text }
func (NopStyler) Info(text string) string    { return text }
func (NopStyler) Muted(text string) string   { return text }
func (NopStyler) Header(text string) string  { return text }

// Verify Styler and NopStyler implement domain.Styler
var _ domain.Styler = (*Styler)(nil)
var _ domain.Styler = NopStyler{}
