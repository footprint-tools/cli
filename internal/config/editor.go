package config

import "strings"

func Set(lines []string, key, value string) ([]string, bool) {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		if strings.TrimSpace(parts[0]) == key {
			// Check for inline comment after the value and preserve it
			oldValue := parts[1]
			commentIdx := strings.Index(oldValue, "#")
			if commentIdx >= 0 {
				comment := strings.TrimSpace(oldValue[commentIdx:])
				lines[i] = key + "=" + value + " " + comment
			} else {
				lines[i] = key + "=" + value
			}
			return lines, true
		}
	}

	lines = append(lines, key+"="+value)
	return lines, false
}

func Unset(lines []string, key string) ([]string, bool) {
	var out []string
	removed := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			out = append(out, line)
			continue
		}

		if strings.TrimSpace(parts[0]) == key {
			removed = true
			continue
		}

		out = append(out, line)
	}

	return out, removed
}
