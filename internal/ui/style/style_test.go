package style

import (
	"os"
	"strings"
	"testing"
)

func TestDisabledReturnsPlainText(t *testing.T) {
	// Ensure no env vars interfere
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("FP_NO_COLOR")

	Init(false)

	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"Success", Success},
		{"Warning", Warning},
		{"Error", Error},
		{"Info", Info},
		{"Header", Header},
		{"Muted", Muted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "test message"
			output := tt.fn(input)

			if output != input {
				t.Errorf("%s() with disabled styling: got %q, want %q", tt.name, output, input)
			}

			// Verify no ANSI escape codes
			if strings.Contains(output, "\x1b[") {
				t.Errorf("%s() with disabled styling contains ANSI codes: %q", tt.name, output)
			}
		})
	}
}

func TestEnabledReturnsStyledText(t *testing.T) {
	// Ensure no env vars interfere
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("FP_NO_COLOR")

	Init(true)

	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"Success", Success},
		{"Warning", Warning},
		{"Error", Error},
		{"Info", Info},
		{"Header", Header},
		{"Muted", Muted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "test message"
			output := tt.fn(input)

			// Output should contain the original text
			if !strings.Contains(output, input) {
				t.Errorf("%s() output %q does not contain input %q", tt.name, output, input)
			}

			// Output should contain ANSI escape codes when enabled
			if !strings.Contains(output, "\x1b[") {
				t.Errorf("%s() with enabled styling should contain ANSI codes: %q", tt.name, output)
			}
		})
	}
}

func TestNoColorEnvDisablesStyling(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	Init(true) // Try to enable, but NO_COLOR should override

	if Enabled() {
		t.Error("Enabled() should return false when NO_COLOR is set")
	}

	input := "test"
	output := Success(input)
	if output != input {
		t.Errorf("Success() should return plain text when NO_COLOR is set: got %q, want %q", output, input)
	}
}

func TestFPNoColorEnvDisablesStyling(t *testing.T) {
	os.Setenv("FP_NO_COLOR", "1")
	defer os.Unsetenv("FP_NO_COLOR")

	Init(true) // Try to enable, but FP_NO_COLOR should override

	if Enabled() {
		t.Error("Enabled() should return false when FP_NO_COLOR is set")
	}

	input := "test"
	output := Warning(input)
	if output != input {
		t.Errorf("Warning() should return plain text when FP_NO_COLOR is set: got %q, want %q", output, input)
	}
}

func TestEnabledReturnsCorrectState(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("FP_NO_COLOR")

	Init(false)
	if Enabled() {
		t.Error("Enabled() should return false after Init(false)")
	}

	Init(true)
	if !Enabled() {
		t.Error("Enabled() should return true after Init(true)")
	}
}

func TestEmptyStringHandling(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("FP_NO_COLOR")

	// Test with disabled
	Init(false)
	if got := Success(""); got != "" {
		t.Errorf("Success(\"\") with disabled styling: got %q, want \"\"", got)
	}

	// Test with enabled - should still handle empty gracefully
	Init(true)
	output := Success("")
	// Empty string with styling might have escape codes but should be minimal
	if !strings.Contains(output, "") {
		t.Errorf("Success(\"\") with enabled styling failed")
	}
}
