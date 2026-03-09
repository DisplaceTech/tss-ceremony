package scenes

import "time"

// Config holds the TUI configuration
type Config struct {
	FixedMode bool
	Message   string
	Speed     string
	NoColor   bool
}

// Styles holds styling information for the TUI
type Styles struct {
	NoColor bool
}

// NewStyles creates a new Styles instance
func NewStyles(noColor bool) *Styles {
	return &Styles{
		NoColor: noColor,
	}
}

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	// Foreground colors
	FGBlack   = "\033[30m"
	FGRed     = "\033[31m"
	FGGreen   = "\033[32m"
	FGYellow  = "\033[33m"
	FGBlue    = "\033[34m"
	FGMagenta = "\033[35m"
	FGCyan    = "\033[36m"
	FGWhite   = "\033[37m"

	// Bright foreground colors
	FGBrightRed    = "\033[91m"
	FGBrightGreen  = "\033[92m"
	FGBrightYellow = "\033[93m"
	FGBrightBlue   = "\033[94m"
	FGBrightMagenta = "\033[95m"
	FGBrightCyan   = "\033[96m"
	FGBrightWhite  = "\033[97m"

	// Background colors
	BGBlack   = "\033[40m"
	BGRed     = "\033[41m"
	BGGreen   = "\033[42m"
	BGYellow  = "\033[43m"
	BGBlue    = "\033[44m"
	BGMagenta = "\033[45m"
	BGCyan    = "\033[46m"
	BGWhite   = "\033[47m"
)

// Party colors
var (
	PartyAColor   = FGCyan
	PartyBColor   = FGMagenta
	SharedColor   = FGYellow
	PhantomColor  = FGRed
	NarratorColor = Dim
)

// tickMsg is a message type for animation ticks
type tickMsg time.Time

// getSceneDuration returns the duration based on speed setting
func getSceneDuration(speed string) time.Duration {
	switch speed {
	case "slow":
		return 3 * time.Second
	case "fast":
		return 1 * time.Second
	default:
		return 2 * time.Second
	}
}

// getCharDuration returns the duration per character based on speed setting
func getCharDuration(speed string) time.Duration {
	switch speed {
	case "slow":
		return 100 * time.Millisecond
	case "fast":
		return 20 * time.Millisecond
	default:
		return 50 * time.Millisecond
	}
}

// getStepDuration returns the duration per step based on speed setting
func getStepDuration(speed string) time.Duration {
	switch speed {
	case "slow":
		return 300 * time.Millisecond
	case "fast":
		return 100 * time.Millisecond
	default:
		return 200 * time.Millisecond
	}
}

// getRandomHexChar returns a random hex character
func getRandomHexChar() rune {
	hexChars := "0123456789abcdef"
	return rune(hexChars[getRandomInt(16)])
}

// getRandomInt returns a random integer in [0, max)
func getRandomInt(max int) int {
	// Simple pseudo-random for animation
	return int(time.Now().Nanosecond()) % max
}
