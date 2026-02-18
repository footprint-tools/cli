package style

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// ColorConfig holds all configurable colors for the UI.
// Values can be ANSI color numbers (0-255) or "bold" for bold styling.
type ColorConfig struct {
	Success  string
	Warning  string
	Error    string
	Info     string
	Muted    string
	Header   string
	Border   string // Interactive delimiters (scrollbars, card borders, etc.)
	UIActive string // Interactive UI elements when focused/active (scrollbar thumb, focused borders)
	UIDim    string // Interactive UI elements when unfocused/inactive (dimmed state)
	Color1   string // post-commit
	Color2   string // post-rewrite
	Color3   string // post-checkout
	Color4   string // post-merge
	Color5   string // pre-push
	Color6   string // backfill
	Color7   string // manual
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
		Success:  "10",  // bright green
		Warning:  "11",  // bright yellow
		Error:    "9",   // bright red
		Info:     "14",  // bright cyan
		Muted:    "245", // medium gray
		Header:   "bold",
		Border:   "244", // medium gray for borders
		UIActive: "14",  // bright cyan for active UI elements
		UIDim:    "240", // dark gray for inactive UI elements
		Color1:   "10",  // post-commit (bright green)
		Color2:   "13",  // post-rewrite (bright magenta)
		Color3:   "12",  // post-checkout (bright blue)
		Color4:   "14",  // post-merge (bright cyan)
		Color5:   "11",  // pre-push (bright yellow)
		Color6:   "8",   // backfill (dark gray)
		Color7:   "15",  // manual (white)
	},

	// Classic light - dark saturated colors for light/white backgrounds.
	// Each color is dark enough to contrast with white text background.
	"default-light": {
		Success:  "28",  // dark green
		Warning:  "130", // dark orange
		Error:    "124", // dark red
		Info:     "27",  // dark blue
		Muted:    "243", // medium-dark gray
		Header:   "bold",
		Border:   "250", // light gray for borders
		UIActive: "27",  // dark blue for active UI elements
		UIDim:    "252", // very light gray for inactive UI elements
		Color1:   "28",  // post-commit (dark green)
		Color2:   "90",  // post-rewrite (dark magenta)
		Color3:   "27",  // post-checkout (dark blue)
		Color4:   "30",  // post-merge (dark cyan)
		Color5:   "130", // pre-push (dark orange)
		Color6:   "240", // backfill (dark gray)
		Color7:   "235", // manual (near black)
	},

	// Neon dark - vivid saturated colors, cyberpunk aesthetic.
	// High-contrast bright colors that pop on dark backgrounds.
	"neon-dark": {
		Success:  "48",  // bright teal
		Warning:  "220", // gold
		Error:    "197", // hot pink
		Info:     "51",  // electric cyan
		Muted:    "244", // gray
		Header:   "bold",
		Border:   "93",  // purple for neon borders
		UIActive: "201", // hot magenta for active UI elements
		UIDim:    "93",  // purple for inactive UI elements
		Color1:   "46",  // post-commit (neon green)
		Color2:   "201", // post-rewrite (hot magenta)
		Color3:   "39",  // post-checkout (deep sky blue)
		Color4:   "51",  // post-merge (cyan)
		Color5:   "226", // pre-push (yellow)
		Color6:   "242", // backfill (gray)
		Color7:   "231", // manual (white)
	},

	// Neon light - deep saturated colors for light backgrounds.
	// Rich jewel tones that remain vibrant but readable.
	"neon-light": {
		Success:  "29",  // deep teal
		Warning:  "166", // dark orange
		Error:    "161", // dark pink
		Info:     "32",  // deep blue
		Muted:    "245", // gray
		Header:   "bold",
		Border:   "99",  // medium purple for borders
		UIActive: "127", // dark magenta for active UI elements
		UIDim:    "99",  // medium purple for inactive UI elements
		Color1:   "28",  // post-commit (forest green)
		Color2:   "127", // post-rewrite (dark magenta)
		Color3:   "26",  // post-checkout (navy)
		Color4:   "37",  // post-merge (teal)
		Color5:   "166", // pre-push (dark orange)
		Color6:   "241", // backfill (gray)
		Color7:   "236", // manual (dark gray)
	},

	// Aurora dark - northern lights inspired palette for dark backgrounds.
	// Dreamy purples, teals, and soft pinks.
	"aurora-dark": {
		Success:  "121", // mint green
		Warning:  "222", // soft gold
		Error:    "204", // salmon pink
		Info:     "147", // lavender
		Muted:    "246", // light gray
		Header:   "bold",
		Border:   "141", // light purple for aurora borders
		UIActive: "147", // lavender for active UI elements
		UIDim:    "141", // light purple for inactive UI elements
		Color1:   "121", // post-commit (mint)
		Color2:   "183", // post-rewrite (orchid)
		Color3:   "111", // post-checkout (sky blue)
		Color4:   "123", // post-merge (turquoise)
		Color5:   "222", // pre-push (gold)
		Color6:   "245", // backfill (gray)
		Color7:   "189", // manual (light lavender)
	},

	// Aurora light - deep jewel tones for light backgrounds.
	// Rich purples, teals, and magentas with good contrast.
	"aurora-light": {
		Success:  "30",  // dark teal
		Warning:  "136", // amber
		Error:    "125", // dark magenta
		Info:     "62",  // purple
		Muted:    "244", // gray
		Border:   "103", // medium purple for borders
		Header:   "bold",
		UIActive: "62",  // purple for active UI elements
		UIDim:    "103", // medium purple for inactive UI elements
		Color1:   "30",  // post-commit (dark teal)
		Color2:   "133", // post-rewrite (medium orchid)
		Color3:   "61",  // post-checkout (slate blue)
		Color4:   "37",  // post-merge (teal)
		Color5:   "136", // pre-push (amber)
		Color6:   "241", // backfill (dark gray)
		Color7:   "96",  // manual (plum)
	},

	// Mono dark - minimalist grayscale with cyan accent.
	// Clean, distraction-free aesthetic.
	"mono-dark": {
		Success:  "50",  // cyan (the one accent)
		Warning:  "229", // pale yellow
		Error:    "210", // light red
		Info:     "50",  // cyan
		Muted:    "245", // gray
		Header:   "bold",
		Border:   "247", // light gray for mono borders
		UIActive: "50",  // cyan for active UI elements
		UIDim:    "247", // light gray for inactive UI elements
		Color1:   "50",  // post-commit (cyan)
		Color2:   "251", // post-rewrite (light gray)
		Color3:   "248", // post-checkout (gray)
		Color4:   "50",  // post-merge (cyan)
		Color5:   "229", // pre-push (pale yellow)
		Color6:   "243", // backfill (dim gray)
		Color7:   "255", // manual (white)
	},

	// Mono light - minimalist grayscale with teal accent.
	// Clean, professional look for light backgrounds.
	"mono-light": {
		Success:  "30",  // dark teal (the one accent)
		Warning:  "136", // amber
		Error:    "124", // dark red
		Info:     "30",  // dark teal
		Muted:    "244", // gray
		Header:   "bold",
		Border:   "249", // light gray for mono borders
		UIActive: "30",  // dark teal for active UI elements
		UIDim:    "249", // light gray for inactive UI elements
		Color1:   "30",  // post-commit (teal)
		Color2:   "241", // post-rewrite (dark gray)
		Color3:   "244", // post-checkout (gray)
		Color4:   "30",  // post-merge (teal)
		Color5:   "136", // pre-push (amber)
		Color6:   "247", // backfill (light gray)
		Color7:   "235", // manual (near black)
	},

	// Ocean dark - cool blues and teals, like deep water.
	// Unified aquatic palette.
	"ocean-dark": {
		Success:  "43",  // turquoise
		Warning:  "221", // light gold
		Error:    "174", // light coral
		Info:     "75",  // sky blue
		Muted:    "245", // gray
		Header:   "bold",
		Border:   "68",  // steel blue for ocean borders
		UIActive: "75",  // sky blue for active UI elements
		UIDim:    "68",  // steel blue for inactive UI elements
		Color1:   "43",  // post-commit (turquoise)
		Color2:   "105", // post-rewrite (slate blue)
		Color3:   "75",  // post-checkout (sky blue)
		Color4:   "80",  // post-merge (medium turquoise)
		Color5:   "221", // pre-push (gold)
		Color6:   "67",  // backfill (steel blue)
		Color7:   "159", // manual (light cyan)
	},

	// Ocean light - deep sea colors for light backgrounds.
	// Navy, teal, and deep blues.
	"ocean-light": {
		Success:  "30",  // dark cyan
		Warning:  "130", // dark orange
		Error:    "124", // dark red
		Info:     "25",  // dark blue
		Muted:    "244", // gray
		Header:   "bold",
		Border:   "74",  // medium cyan for ocean borders
		UIActive: "25",  // dark blue for active UI elements
		UIDim:    "74",  // medium cyan for inactive UI elements
		Color1:   "30",  // post-commit (dark cyan)
		Color2:   "61",  // post-rewrite (slate blue)
		Color3:   "25",  // post-checkout (dark blue)
		Color4:   "37",  // post-merge (teal)
		Color5:   "130", // pre-push (dark orange)
		Color6:   "66",  // backfill (grayish cyan)
		Color7:   "17",  // manual (navy)
	},

	// Sunset dark - warm gradient from orange to magenta to purple.
	// Dusk vibes.
	"sunset-dark": {
		Success:  "216", // light salmon
		Warning:  "221", // light goldenrod
		Error:    "204", // hot pink
		Info:     "183", // plum
		Muted:    "245", // gray
		Header:   "bold",
		Border:   "175", // light pink for sunset borders
		UIActive: "213", // orchid for active UI elements
		UIDim:    "175", // light pink for inactive UI elements
		Color1:   "216", // post-commit (salmon)
		Color2:   "213", // post-rewrite (orchid)
		Color3:   "183", // post-checkout (plum)
		Color4:   "209", // post-merge (coral)
		Color5:   "221", // pre-push (gold)
		Color6:   "139", // backfill (dusty rose)
		Color7:   "224", // manual (misty rose)
	},

	// Sunset light - deep warm tones for light backgrounds.
	// Rich oranges, magentas, and purples.
	"sunset-light": {
		Success:  "166", // dark orange
		Warning:  "136", // dark goldenrod
		Error:    "125", // dark pink
		Info:     "90",  // dark magenta
		Muted:    "244", // gray
		Header:   "bold",
		Border:   "132", // medium orchid for sunset borders
		UIActive: "127", // medium violet for active UI elements
		UIDim:    "132", // medium orchid for inactive UI elements
		Color1:   "166", // post-commit (dark orange)
		Color2:   "127", // post-rewrite (medium violet)
		Color3:   "90",  // post-checkout (dark magenta)
		Color4:   "130", // post-merge (dark coral)
		Color5:   "136", // pre-push (dark gold)
		Color6:   "95",  // backfill (dusty purple)
		Color7:   "52",  // manual (dark red)
	},

	// Candy dark - sweet pastel colors on dark background.
	// Playful and soft.
	"candy-dark": {
		Success:  "158", // mint
		Warning:  "222", // light peach
		Error:    "211", // light pink
		Info:     "153", // baby blue
		Muted:    "250", // light gray
		Header:   "bold",
		Border:   "183", // light orchid for candy borders
		UIActive: "153", // baby blue for active UI elements
		UIDim:    "183", // light orchid for inactive UI elements
		Color1:   "158", // post-commit (mint)
		Color2:   "218", // post-rewrite (pink)
		Color3:   "153", // post-checkout (baby blue)
		Color4:   "158", // post-merge (aquamarine)
		Color5:   "222", // pre-push (peach)
		Color6:   "188", // backfill (light lavender)
		Color7:   "231", // manual (white)
	},

	// Candy light - deeper candy colors for light backgrounds.
	// Still playful but readable.
	"candy-light": {
		Success:  "36",  // dark mint
		Warning:  "172", // dark peach
		Error:    "168", // dark pink
		Info:     "68",  // medium blue
		Muted:    "244", // gray
		Header:   "bold",
		Border:   "139", // medium pink for candy borders
		UIActive: "68",  // medium blue for active UI elements
		UIDim:    "139", // medium pink for inactive UI elements
		Color1:   "36",  // post-commit (dark mint)
		Color2:   "132", // post-rewrite (medium orchid)
		Color3:   "68",  // post-checkout (medium blue)
		Color4:   "73",  // post-merge (cadet blue)
		Color5:   "172", // pre-push (dark peach)
		Color6:   "103", // backfill (medium purple)
		Color7:   "240", // manual (dark gray)
	},

	// Contrast dark - maximum readability with pure primaries.
	// High contrast, accessibility-focused.
	"contrast-dark": {
		Success:  "46",  // pure bright green
		Warning:  "226", // pure bright yellow
		Error:    "196", // pure bright red
		Info:     "51",  // pure bright cyan
		Muted:    "250", // bright gray
		Header:   "bold",
		Border:   "255", // white for high contrast borders
		UIActive: "51",  // pure bright cyan for active UI elements
		UIDim:    "255", // white for inactive UI elements
		Color1:   "46",  // post-commit (green)
		Color2:   "201", // post-rewrite (magenta)
		Color3:   "21",  // post-checkout (blue)
		Color4:   "51",  // post-merge (cyan)
		Color5:   "226", // pre-push (yellow)
		Color6:   "245", // backfill (gray)
		Color7:   "231", // manual (white)
	},

	// Contrast light - maximum readability for light backgrounds.
	// Pure dark primaries, very accessible.
	"contrast-light": {
		Success:  "22",  // dark green
		Warning:  "130", // dark orange (yellow hard to read on white)
		Error:    "124", // dark red
		Info:     "21",  // dark blue
		Muted:    "240", // dark gray
		Header:   "bold",
		Border:   "238", // dark gray for high contrast borders
		UIActive: "21",  // dark blue for active UI elements
		UIDim:    "238", // dark gray for inactive UI elements
		Color1:   "22",  // post-commit (dark green)
		Color2:   "90",  // post-rewrite (dark magenta)
		Color3:   "19",  // post-checkout (dark blue)
		Color4:   "30",  // post-merge (dark cyan)
		Color5:   "130", // pre-push (dark orange)
		Color6:   "243", // backfill (gray)
		Color7:   "232", // manual (near black)
	},
}

// colorConfigKeys maps config/env key names to ColorConfig field names.
var colorConfigKeys = map[string]string{
	"color_success":   "Success",
	"color_warning":   "Warning",
	"color_error":     "Error",
	"color_info":      "Info",
	"color_muted":     "Muted",
	"color_header":    "Header",
	"color_border":    "Border",
	"color_ui_active": "UIActive",
	"color_ui_dim":    "UIDim",
	"color_1":         "Color1",
	"color_2":         "Color2",
	"color_3":         "Color3",
	"color_4":         "Color4",
	"color_5":         "Color5",
	"color_6":         "Color6",
	"color_7":         "Color7",
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
// 3. Theme value (from theme config)
// 4. Default theme (auto-detected based on terminal background)
func LoadColorConfig(cfg map[string]string) ColorConfig {
	// Start with auto-detected default
	themeName := ResolveThemeName("default")

	// Check env for theme override
	if envTheme := os.Getenv("FP_COLOR_THEME"); envTheme != "" {
		themeName = ResolveThemeName(envTheme)
	} else if cfgTheme, ok := cfg["theme"]; ok && cfgTheme != "" {
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
	case "Border":
		c.Border = value
	case "UIActive":
		c.UIActive = value
	case "UIDim":
		c.UIDim = value
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
	return strings.ToUpper(s)
}
