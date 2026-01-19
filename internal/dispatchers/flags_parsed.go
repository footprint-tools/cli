package dispatchers

import (
	"strconv"
	"strings"
	"time"
)

// ParsedFlags provides typed access to command-line flags.
type ParsedFlags struct {
	raw []string
}

// NewParsedFlags creates a ParsedFlags from a slice of flag strings.
func NewParsedFlags(flags []string) *ParsedFlags {
	return &ParsedFlags{raw: flags}
}

// Raw returns the underlying flag strings.
func (f *ParsedFlags) Raw() []string {
	return f.raw
}

// Has returns true if the flag is present (for boolean flags).
func (f *ParsedFlags) Has(name string) bool {
	for _, flag := range f.raw {
		if flag == name {
			return true
		}
	}
	return false
}

// String returns the value of a flag, or defaultVal if not present.
// Supports both --flag=value and --flag value formats.
func (f *ParsedFlags) String(name, defaultVal string) string {
	prefix := name + "="
	for _, flag := range f.raw {
		if strings.HasPrefix(flag, prefix) {
			return strings.TrimPrefix(flag, prefix)
		}
	}
	return defaultVal
}

// Int returns the integer value of a flag, or defaultVal if not present or invalid.
func (f *ParsedFlags) Int(name string, defaultVal int) int {
	str := f.String(name, "")
	if str == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(str)
	if err != nil {
		return defaultVal
	}
	return n
}

// Date returns the time.Time value of a flag (YYYY-MM-DD format), or nil if not present or invalid.
func (f *ParsedFlags) Date(name string) *time.Time {
	str := f.String(name, "")
	if str == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return nil
	}
	return &t
}
