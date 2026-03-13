package scenes

import (
	"sync/atomic"
	"time"
)

// CeremonyData holds computed ceremony values for display in scenes.
// These are populated by the protocol layer before the TUI starts.
type CeremonyData struct {
	// Keygen
	PartyASecretHex string
	PartyBSecretHex string
	PartyAPubHex    string
	PartyBPubHex    string
	CombinedPubHex  string

	// Message
	MessageText string
	MessageHash string

	// Nonces
	NonceAHex        string
	NonceBHex        string
	NonceAPubHex     string
	NonceBPubHex     string
	CombinedRPubHex  string
	RHex             string

	// OT
	OTInput0Hex string
	OTInput1Hex string
	OTChoiceBit int
	OTOutputHex string

	// MtA
	AlphaHex string
	BetaHex  string

	// Partial signatures
	PartialSigAHex string
	PartialSigBHex string

	// Final signature
	SignatureRHex string
	SignatureSHex string

	// Verification
	Valid bool
}

// Config holds the TUI configuration
type Config struct {
	FixedMode bool
	Message   string
	Speed     string
	NoColor   bool
	Ceremony  *CeremonyData
}

// fixedModeCounter is a package-level atomic counter used to produce
// deterministic "random" values when fixed mode is active.  It is
// reset to 0 whenever ResetFixedCounter is called (e.g. at scene init).
var fixedModeCounter uint64

// ResetFixedCounter resets the deterministic animation counter to zero.
// Call this at the start of each scene Init() when fixed mode is active so
// that every run produces the same sequence of animation frames.
func ResetFixedCounter() {
	atomic.StoreUint64(&fixedModeCounter, 0)
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

// getDeterministicHexChar returns a deterministic hex character derived from
// the package-level counter.  The counter is atomically incremented on each
// call so successive calls produce different (but reproducible) characters.
func getDeterministicHexChar() rune {
	hexChars := "0123456789abcdef"
	idx := atomic.AddUint64(&fixedModeCounter, 1) - 1
	return rune(hexChars[idx%16])
}

// getDeterministicInt returns a deterministic integer in [0, max) derived
// from the package-level counter.
func getDeterministicInt(max int) int {
	idx := atomic.AddUint64(&fixedModeCounter, 1) - 1
	return int(idx) % max
}

// pickHexChar returns a hex character that is deterministic when fixedMode is
// true, and random otherwise.  Scenes should call this instead of
// getRandomHexChar directly so that fixed mode propagates.
func pickHexChar(fixedMode bool) rune {
	if fixedMode {
		return getDeterministicHexChar()
	}
	return getRandomHexChar()
}

// pickInt returns an integer in [0, max) that is deterministic when
// fixedMode is true, and random otherwise.
func pickInt(fixedMode bool, max int) int {
	if fixedMode {
		return getDeterministicInt(max)
	}
	return getRandomInt(max)
}
