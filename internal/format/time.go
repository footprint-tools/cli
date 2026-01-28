package format

import (
	"strings"
	"time"

	"github.com/footprint-tools/cli/internal/config"
)

// DateTime formats a time with both date and time according to config.
// Example output: "23/01/2024 15:04" or "01/23/2024 3:04 PM"
func DateTime(t time.Time) string {
	return Date(t) + " " + Time(t)
}

// DateTimeShort formats a time with short date and time (no year).
// Example output: "23/01 15:04" or "01/23 3:04 PM"
func DateTimeShort(t time.Time) string {
	return DateShort(t) + " " + Time(t)
}

// Date formats only the date portion according to config.
// Example output: "23/01/2024" or "01/23/2024" or "2024-01-23"
func Date(t time.Time) string {
	format := getDateFormat()
	return t.Format(format)
}

// DateShort formats date without year.
// Example output: "23/01" or "01/23"
func DateShort(t time.Time) string {
	format := getDateFormatShort()
	return t.Format(format)
}

// Time formats only the time portion according to config.
// Example output: "15:04" or "3:04 PM"
func Time(t time.Time) string {
	format := getTimeFormat()
	return t.Format(format)
}

// TimeFull formats time with seconds.
// Example output: "15:04:05" or "3:04:05 PM"
func TimeFull(t time.Time) string {
	format := getTimeFormatFull()
	return t.Format(format)
}

// Full formats with full date and time with seconds.
// Example output: "23/01/2024 15:04:05"
func Full(t time.Time) string {
	return Date(t) + " " + TimeFull(t)
}

// getDateFormat returns the Go time format string for dates.
func getDateFormat() string {
	displayDate, _ := config.Get("display_date")
	if displayDate == "" {
		displayDate = "Jan 02"
	}

	// Check for preset formats
	switch displayDate {
	case "mm/dd/yyyy":
		return "01/02/2006"
	case "yyyy-mm-dd":
		return "2006-01-02"
	case "dd/mm/yyyy":
		return "02/01/2006"
	default:
		// Assume it's a custom Go time format (e.g., "Jan 02")
		return displayDate
	}
}

// getDateFormatShort returns the Go time format string for short dates (no year).
func getDateFormatShort() string {
	displayDate, _ := config.Get("display_date")
	if displayDate == "" {
		displayDate = "Jan 02"
	}

	// Check for preset formats
	switch displayDate {
	case "mm/dd/yyyy":
		return "01/02"
	case "yyyy-mm-dd":
		return "01-02"
	case "dd/mm/yyyy":
		return "02/01"
	default:
		// For custom formats, try to derive a short version by removing year patterns
		short := displayDate
		// Remove common year patterns
		short = strings.ReplaceAll(short, "2006", "")
		short = strings.ReplaceAll(short, "/06", "")
		short = strings.ReplaceAll(short, "-06", "")
		short = strings.ReplaceAll(short, " 06", "")
		short = strings.TrimSpace(short)
		short = strings.Trim(short, "/-")
		if short == "" {
			return "Jan 02" // fallback
		}
		return short
	}
}

// getTimeFormat returns the Go time format string for times.
func getTimeFormat() string {
	displayTime, _ := config.Get("display_time")
	if displayTime == "" {
		displayTime = "24h"
	}

	switch displayTime {
	case "12h":
		return "3:04 PM"
	case "24h":
		fallthrough
	default:
		return "15:04"
	}
}

// getTimeFormatFull returns the Go time format string for times with seconds.
func getTimeFormatFull() string {
	displayTime, _ := config.Get("display_time")
	if displayTime == "" {
		displayTime = "24h"
	}

	switch displayTime {
	case "12h":
		return "3:04:05 PM"
	case "24h":
		fallthrough
	default:
		return "15:04:05"
	}
}
