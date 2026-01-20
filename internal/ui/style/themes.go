package style

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// ColorConfig holds all configurable colors for the UI.
// Values can be ANSI color numbers (0-255) or "bold" for bold styling.
type ColorConfig struct {
	Success string
	Warning string
	Error   string
	Info    string
	Muted   string
	Header  string
	Color1  string // POST-COMMIT
	Color2  string // POST-REWRITE
	Color3  string // POST-CHECKOUT
	Color4  string // POST-MERGE
	Color5  string // PRE-PUSH
	Color6  string // BACKFILL
	Color7  string // MANUAL
}

// BaseThemeNames lists available theme bases (auto-detects dark/light).
var BaseThemeNames = []string{
	"default",
	"neon",
	"aurora",
	"mono",
	"ocean",
	"sunset",
	"candy",
	"contrast",
}

// ThemeNames lists all themes with explicit dark/light variants.
var ThemeNames = []string{
	"default-dark", "default-light",
	"neon-dark", "neon-light",
	"aurora-dark", "aurora-light",
	"mono-dark", "mono-light",
	"ocean-dark", "ocean-light",
	"sunset-dark", "sunset-light",
	"candy-dark", "candy-light",
	"contrast-dark", "contrast-light",
}

// Themes contains the built-in color themes.
// Dark themes use BRIGHT colors (high contrast on dark backgrounds).
// Light themes use DARK colors (high contrast on light/white backgrounds).
var Themes = map[string]ColorConfig{
	// Classic dark - traditional bright terminal colors for dark backgrounds.
	// Uses the standard 16-color palette for maximum compatibility.
	"default-dark": {
		Success: "10",  // bright green
		Warning: "11",  // bright yellow
		Error:   "9",   // bright red
		Info:    "14",  // bright cyan
		Muted:   "245", // medium gray
		Header:  "bold",
		Color1:  "10",  // POST-COMMIT (bright green)
		Color2:  "13",  // POST-REWRITE (bright magenta)
		Color3:  "12",  // POST-CHECKOUT (bright blue)
		Color4:  "14",  // POST-MERGE (bright cyan)
		Color5:  "11",  // PRE-PUSH (bright yellow)
		Color6:  "8",   // BACKFILL (dark gray)
		Color7:  "15",  // MANUAL (white)
	},

	// Classic light - dark saturated colors for light/white backgrounds.
	// Each color is dark enough to contrast with white text background.
	"default-light": {
		Success: "28",  // dark green
		Warning: "130", // dark orange
		Error:   "124", // dark red
		Info:    "27",  // dark blue
		Muted:   "243", // medium-dark gray
		Header:  "bold",
		Color1:  "28",  // POST-COMMIT (dark green)
		Color2:  "90",  // POST-REWRITE (dark magenta)
		Color3:  "27",  // POST-CHECKOUT (dark blue)
		Color4:  "30",  // POST-MERGE (dark cyan)
		Color5:  "130", // PRE-PUSH (dark orange)
		Color6:  "240", // BACKFILL (dark gray)
		Color7:  "235", // MANUAL (near black)
	},

	// Neon dark - vivid saturated colors, cyberpunk aesthetic.
	// High-contrast bright colors that pop on dark backgrounds.
	"neon-dark": {
		Success: "48",  // bright teal
		Warning: "220", // gold
		Error:   "197", // hot pink
		Info:    "51",  // electric cyan
		Muted:   "244", // gray
		Header:  "bold",
		Color1:  "46",  // POST-COMMIT (neon green)
		Color2:  "201", // POST-REWRITE (hot magenta)
		Color3:  "39",  // POST-CHECKOUT (deep sky blue)
		Color4:  "51",  // POST-MERGE (cyan)
		Color5:  "226", // PRE-PUSH (yellow)
		Color6:  "242", // BACKFILL (gray)
		Color7:  "231", // MANUAL (white)
	},

	// Neon light - deep saturated colors for light backgrounds.
	// Rich jewel tones that remain vibrant but readable.
	"neon-light": {
		Success: "29",  // deep teal
		Warning: "166", // dark orange
		Error:   "161", // dark pink
		Info:    "32",  // deep blue
		Muted:   "245", // gray
		Header:  "bold",
		Color1:  "28",  // POST-COMMIT (forest green)
		Color2:  "127", // POST-REWRITE (dark magenta)
		Color3:  "26",  // POST-CHECKOUT (navy)
		Color4:  "37",  // POST-MERGE (teal)
		Color5:  "166", // PRE-PUSH (dark orange)
		Color6:  "241", // BACKFILL (gray)
		Color7:  "236", // MANUAL (dark gray)
	},

	// Aurora dark - northern lights inspired palette for dark backgrounds.
	// Dreamy purples, teals, and soft pinks.
	"aurora-dark": {
		Success: "121", // mint green
		Warning: "222", // soft gold
		Error:   "204", // salmon pink
		Info:    "147", // lavender
		Muted:   "246", // light gray
		Header:  "bold",
		Color1:  "121", // POST-COMMIT (mint)
		Color2:  "183", // POST-REWRITE (orchid)
		Color3:  "111", // POST-CHECKOUT (sky blue)
		Color4:  "123", // POST-MERGE (turquoise)
		Color5:  "222", // PRE-PUSH (gold)
		Color6:  "245", // BACKFILL (gray)
		Color7:  "189", // MANUAL (light lavender)
	},

	// Aurora light - deep jewel tones for light backgrounds.
	// Rich purples, teals, and magentas with good contrast.
	"aurora-light": {
		Success: "30",  // dark teal
		Warning: "136", // amber
		Error:   "125", // dark magenta
		Info:    "62",  // purple
		Muted:   "244", // gray
		Header:  "bold",
		Color1:  "30",  // POST-COMMIT (dark teal)
		Color2:  "133", // POST-REWRITE (medium orchid)
		Color3:  "61",  // POST-CHECKOUT (slate blue)
		Color4:  "37",  // POST-MERGE (teal)
		Color5:  "136", // PRE-PUSH (amber)
		Color6:  "241", // BACKFILL (dark gray)
		Color7:  "96",  // MANUAL (plum)
	},

	// Mono dark - minimalist grayscale with cyan accent.
	// Clean, distraction-free aesthetic.
	"mono-dark": {
		Success: "50",  // cyan (the one accent)
		Warning: "229", // pale yellow
		Error:   "210", // light red
		Info:    "50",  // cyan
		Muted:   "245", // gray
		Header:  "bold",
		Color1:  "50",  // POST-COMMIT (cyan)
		Color2:  "251", // POST-REWRITE (light gray)
		Color3:  "248", // POST-CHECKOUT (gray)
		Color4:  "50",  // POST-MERGE (cyan)
		Color5:  "229", // PRE-PUSH (pale yellow)
		Color6:  "243", // BACKFILL (dim gray)
		Color7:  "255", // MANUAL (white)
	},

	// Mono light - minimalist grayscale with teal accent.
	// Clean, professional look for light backgrounds.
	"mono-light": {
		Success: "30",  // dark teal (the one accent)
		Warning: "136", // amber
		Error:   "124", // dark red
		Info:    "30",  // dark teal
		Muted:   "244", // gray
		Header:  "bold",
		Color1:  "30",  // POST-COMMIT (teal)
		Color2:  "241", // POST-REWRITE (dark gray)
		Color3:  "244", // POST-CHECKOUT (gray)
		Color4:  "30",  // POST-MERGE (teal)
		Color5:  "136", // PRE-PUSH (amber)
		Color6:  "247", // BACKFILL (light gray)
		Color7:  "235", // MANUAL (near black)
	},

	// Ocean dark - cool blues and teals, like deep water.
	// Unified aquatic palette.
	"ocean-dark": {
		Success: "43",  // turquoise
		Warning: "221", // light gold
		Error:   "174", // light coral
		Info:    "75",  // sky blue
		Muted:   "245", // gray
		Header:  "bold",
		Color1:  "43",  // POST-COMMIT (turquoise)
		Color2:  "105", // POST-REWRITE (slate blue)
		Color3:  "75",  // POST-CHECKOUT (sky blue)
		Color4:  "80",  // POST-MERGE (medium turquoise)
		Color5:  "221", // PRE-PUSH (gold)
		Color6:  "67",  // BACKFILL (steel blue)
		Color7:  "159", // MANUAL (light cyan)
	},

	// Ocean light - deep sea colors for light backgrounds.
	// Navy, teal, and deep blues.
	"ocean-light": {
		Success: "30",  // dark cyan
		Warning: "130", // dark orange
		Error:   "124", // dark red
		Info:    "25",  // dark blue
		Muted:   "244", // gray
		Header:  "bold",
		Color1:  "30",  // POST-COMMIT (dark cyan)
		Color2:  "61",  // POST-REWRITE (slate blue)
		Color3:  "25",  // POST-CHECKOUT (dark blue)
		Color4:  "37",  // POST-MERGE (teal)
		Color5:  "130", // PRE-PUSH (dark orange)
		Color6:  "66",  // BACKFILL (grayish cyan)
		Color7:  "17",  // MANUAL (navy)
	},

	// Sunset dark - warm gradient from orange to magenta to purple.
	// Dusk vibes.
	"sunset-dark": {
		Success: "216", // light salmon
		Warning: "221", // light goldenrod
		Error:   "204", // hot pink
		Info:    "183", // plum
		Muted:   "245", // gray
		Header:  "bold",
		Color1:  "216", // POST-COMMIT (salmon)
		Color2:  "213", // POST-REWRITE (orchid)
		Color3:  "183", // POST-CHECKOUT (plum)
		Color4:  "209", // POST-MERGE (coral)
		Color5:  "221", // PRE-PUSH (gold)
		Color6:  "139", // BACKFILL (dusty rose)
		Color7:  "224", // MANUAL (misty rose)
	},

	// Sunset light - deep warm tones for light backgrounds.
	// Rich oranges, magentas, and purples.
	"sunset-light": {
		Success: "166", // dark orange
		Warning: "136", // dark goldenrod
		Error:   "125", // dark pink
		Info:    "90",  // dark magenta
		Muted:   "244", // gray
		Header:  "bold",
		Color1:  "166", // POST-COMMIT (dark orange)
		Color2:  "127", // POST-REWRITE (medium violet)
		Color3:  "90",  // POST-CHECKOUT (dark magenta)
		Color4:  "130", // POST-MERGE (dark coral)
		Color5:  "136", // PRE-PUSH (dark gold)
		Color6:  "95",  // BACKFILL (dusty purple)
		Color7:  "52",  // MANUAL (dark red)
	},

	// Candy dark - sweet pastel colors on dark background.
	// Playful and soft.
	"candy-dark": {
		Success: "158", // mint
		Warning: "222", // light peach
		Error:   "211", // light pink
		Info:    "153", // baby blue
		Muted:   "250", // light gray
		Header:  "bold",
		Color1:  "158", // POST-COMMIT (mint)
		Color2:  "218", // POST-REWRITE (pink)
		Color3:  "153", // POST-CHECKOUT (baby blue)
		Color4:  "158", // POST-MERGE (aquamarine)
		Color5:  "222", // PRE-PUSH (peach)
		Color6:  "188", // BACKFILL (light lavender)
		Color7:  "231", // MANUAL (white)
	},

	// Candy light - deeper candy colors for light backgrounds.
	// Still playful but readable.
	"candy-light": {
		Success: "36",  // dark mint
		Warning: "172", // dark peach
		Error:   "168", // dark pink
		Info:    "68",  // medium blue
		Muted:   "244", // gray
		Header:  "bold",
		Color1:  "36",  // POST-COMMIT (dark mint)
		Color2:  "132", // POST-REWRITE (medium orchid)
		Color3:  "68",  // POST-CHECKOUT (medium blue)
		Color4:  "73",  // POST-MERGE (cadet blue)
		Color5:  "172", // PRE-PUSH (dark peach)
		Color6:  "103", // BACKFILL (medium purple)
		Color7:  "240", // MANUAL (dark gray)
	},

	// Contrast dark - maximum readability with pure primaries.
	// High contrast, accessibility-focused.
	"contrast-dark": {
		Success: "46",  // pure bright green
		Warning: "226", // pure bright yellow
		Error:   "196", // pure bright red
		Info:    "51",  // pure bright cyan
		Muted:   "250", // bright gray
		Header:  "bold",
		Color1:  "46",  // POST-COMMIT (green)
		Color2:  "201", // POST-REWRITE (magenta)
		Color3:  "21",  // POST-CHECKOUT (blue)
		Color4:  "51",  // POST-MERGE (cyan)
		Color5:  "226", // PRE-PUSH (yellow)
		Color6:  "245", // BACKFILL (gray)
		Color7:  "231", // MANUAL (white)
	},

	// Contrast light - maximum readability for light backgrounds.
	// Pure dark primaries, very accessible.
	"contrast-light": {
		Success: "22",  // dark green
		Warning: "130", // dark orange (yellow hard to read on white)
		Error:   "124", // dark red
		Info:    "21",  // dark blue
		Muted:   "240", // dark gray
		Header:  "bold",
		Color1:  "22",  // POST-COMMIT (dark green)
		Color2:  "90",  // POST-REWRITE (dark magenta)
		Color3:  "19",  // POST-CHECKOUT (dark blue)
		Color4:  "30",  // POST-MERGE (dark cyan)
		Color5:  "130", // PRE-PUSH (dark orange)
		Color6:  "243", // BACKFILL (gray)
		Color7:  "232", // MANUAL (near black)
	},
}

// colorConfigKeys maps config/env key names to ColorConfig field names.
var colorConfigKeys = map[string]string{
	"color_success": "Success",
	"color_warning": "Warning",
	"color_error":   "Error",
	"color_info":    "Info",
	"color_muted":   "Muted",
	"color_header":  "Header",
	"color_1":       "Color1",
	"color_2":       "Color2",
	"color_3":       "Color3",
	"color_4":       "Color4",
	"color_5":       "Color5",
	"color_6":       "Color6",
	"color_7":       "Color7",
}

// IsDarkBackground returns true if the terminal has a dark background.
// Uses termenv to query the terminal. Returns true if detection fails.
func IsDarkBackground() bool {
	return termenv.HasDarkBackground()
}

// ResolveThemeName takes a theme name and returns the full theme name.
// If the name doesn't have a -dark/-light suffix, it appends one based
// on terminal background detection.
func ResolveThemeName(name string) string {
	// If already has suffix, return as-is
	if strings.HasSuffix(name, "-dark") || strings.HasSuffix(name, "-light") {
		return name
	}

	// Auto-detect and append suffix
	if IsDarkBackground() {
		return name + "-dark"
	}
	return name + "-light"
}

// LoadColorConfig builds a ColorConfig from the given configuration map.
// Resolution priority:
// 1. Environment variable (FP_COLOR_*)
// 2. Config file value
// 3. Theme value (from color_theme config)
// 4. Default theme (auto-detected based on terminal background)
func LoadColorConfig(cfg map[string]string) ColorConfig {
	// Start with auto-detected default
	themeName := ResolveThemeName("default")

	// Check env for theme override
	if envTheme := os.Getenv("FP_COLOR_THEME"); envTheme != "" {
		themeName = ResolveThemeName(envTheme)
	} else if cfgTheme, ok := cfg["color_theme"]; ok && cfgTheme != "" {
		themeName = ResolveThemeName(cfgTheme)
	}

	// Get base theme (fall back to default-dark if unknown)
	theme, ok := Themes[themeName]
	if !ok {
		theme = Themes["default-dark"]
	}

	// Apply overrides from config and env
	result := theme

	for configKey, fieldName := range colorConfigKeys {
		// Check env first (highest priority)
		envKey := "FP_" + toUpperSnake(configKey)
		if envVal := os.Getenv(envKey); envVal != "" {
			setColorField(&result, fieldName, envVal)
			continue
		}

		// Check config file
		if cfgVal, ok := cfg[configKey]; ok && cfgVal != "" {
			setColorField(&result, fieldName, cfgVal)
		}
	}

	return result
}

// setColorField sets a field on ColorConfig by name.
func setColorField(c *ColorConfig, field, value string) {
	switch field {
	case "Success":
		c.Success = value
	case "Warning":
		c.Warning = value
	case "Error":
		c.Error = value
	case "Info":
		c.Info = value
	case "Muted":
		c.Muted = value
	case "Header":
		c.Header = value
	case "Color1":
		c.Color1 = value
	case "Color2":
		c.Color2 = value
	case "Color3":
		c.Color3 = value
	case "Color4":
		c.Color4 = value
	case "Color5":
		c.Color5 = value
	case "Color6":
		c.Color6 = value
	case "Color7":
		c.Color7 = value
	}
}

// toUpperSnake converts "color_success" to "COLOR_SUCCESS".
func toUpperSnake(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 'a' + 'A'
		} else {
			result[i] = c
		}
	}
	return string(result)
}
