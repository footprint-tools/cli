package hooks

import "strings"

// shellQuote escapes a string for safe use in shell scripts
func shellQuote(s string) string {
	// Use single quotes and escape any single quotes within
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func Script(fpPath string, source string) string {
	// Run fp record with the source environment variable
	// Redirect stdout to /dev/null (suppress normal output)
	// Errors are now logged internally by fp record via the logger
	// Use proper shell quoting to prevent injection
	return "#!/bin/sh\n" +
		"FP_SOURCE=" + shellQuote(source) + " " +
		shellQuote(fpPath) + " record >/dev/null 2>&1 || true\n"
}
