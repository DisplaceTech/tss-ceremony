package scenes

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SignNonceScene represents Scene 6: Nonce Generation.
// It shows Party A generating nonce k_a, Party B generating k_b,
// then computing the nonce public points R_a and R_b.
type SignNonceScene struct {
	config      *Config
	styles      *Styles
	phase       int // 0: nonce_a, 1: nonce_b, 2: compute_R, 3: complete
	step        int // animation step within phase
	started     bool
	duration    time.Duration
	nonceAChars []rune // rolling hex chars for nonce A
	nonceBChars []rune // rolling hex chars for nonce B
	nonceAIndex int    // how many chars settled for nonce A
	nonceBIndex int    // how many chars settled for nonce B
}

// NewSignNonceScene creates a new nonce generation scene.
func NewSignNonceScene(config *Config, styles *Styles) *SignNonceScene {
	return &SignNonceScene{
		config:      config,
		styles:      styles,
		phase:       0,
		step:        0,
		started:     false,
		duration:    getCharDuration(config.Speed),
		nonceAChars: make([]rune, 64), // 32 bytes = 64 hex chars
		nonceBChars: make([]rune, 64),
	}
}

// Init initializes the scene.
func (s *SignNonceScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.step = 0
	s.nonceAIndex = 0
	s.nonceBIndex = 0
	if s.config.FixedMode {
		ResetFixedCounter()
	}
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene.
func (s *SignNonceScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "enter", "right", "l", " ", "n", "j", "down":
			return s, nil
		}

	case tickMsg:
		switch s.phase {
		case 0:
			// Generate nonce A with rolling hex
			if s.nonceAIndex < 64 {
				s.nonceAChars[s.nonceAIndex] = pickHexChar(s.config.FixedMode)
				s.nonceAIndex++
				if s.nonceAIndex >= 64 {
					s.phase = 1
				}
			}
		case 1:
			// Generate nonce B with rolling hex
			if s.nonceBIndex < 64 {
				s.nonceBChars[s.nonceBIndex] = pickHexChar(s.config.FixedMode)
				s.nonceBIndex++
				if s.nonceBIndex >= 64 {
					s.phase = 2
					s.step = 0
				}
			}
		case 2:
			// Show R computation animation
			s.step++
			if s.step >= 10 {
				s.phase = 3
				s.step = 0
			}
		case 3:
			// Complete, keep ticking
		}
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// Render renders the scene view.
func (s *SignNonceScene) Render() string {
	var builder strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginRight(2)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Header
	builder.WriteString(headerStyle.Render("Nonce Generation") + "\n\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	// Party A nonce
	partyAStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("81")) // Cyan
	builder.WriteString(partyAStyle.Render("Party A") + "\n")
	builder.WriteString(labelStyle.Render("  Nonce k_a:") + "\n")
	builder.WriteString("  " + s.renderNonceHex(s.nonceAChars, s.nonceAIndex, PartyAColor) + "\n\n")

	// Party B nonce
	partyBStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("213")) // Magenta
	builder.WriteString(partyBStyle.Render("Party B") + "\n")
	builder.WriteString(labelStyle.Render("  Nonce k_b:") + "\n")
	builder.WriteString("  " + s.renderNonceHex(s.nonceBChars, s.nonceBIndex, PartyBColor) + "\n\n")

	// Show nonce public points in phase 2+
	if s.phase >= 2 {
		builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")
		builder.WriteString(labelStyle.Render("Nonce Public Points:") + "\n\n")

		// R_a = k_a * G
		builder.WriteString("  " + partyAStyle.Render("R_a") + " = k_a \u00d7 G\n")
		if s.config.Ceremony != nil && s.config.Ceremony.NonceAPubHex != "" {
			builder.WriteString("  " + formatHexGroups(s.config.Ceremony.NonceAPubHex) + "\n\n")
		} else {
			builder.WriteString("  (computing...)\n\n")
		}

		// R_b = k_b * G
		builder.WriteString("  " + partyBStyle.Render("R_b") + " = k_b \u00d7 G\n")
		if s.config.Ceremony != nil && s.config.Ceremony.NonceBPubHex != "" {
			builder.WriteString("  " + formatHexGroups(s.config.Ceremony.NonceBPubHex) + "\n\n")
		} else {
			builder.WriteString("  (computing...)\n\n")
		}
	}

	// Show combined R in phase 3
	if s.phase >= 3 {
		sharedStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")) // Yellow
		builder.WriteString("  " + sharedStyle.Render("R") + " = R_a + R_b  (combined nonce point)\n")
		if s.config.Ceremony != nil && s.config.Ceremony.CombinedRPubHex != "" {
			builder.WriteString("  " + formatHexGroups(s.config.Ceremony.CombinedRPubHex) + "\n\n")
		}
		builder.WriteString(labelStyle.Render("  r = R.x mod n") + "\n")
		if s.config.Ceremony != nil && s.config.Ceremony.RHex != "" {
			rStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
			builder.WriteString("  " + rStyle.Render(formatHexGroups(s.config.Ceremony.RHex)) + "\n\n")
		}
	}

	// Status
	var status string
	switch s.phase {
	case 0:
		status = "Generating Party A's nonce..."
	case 1:
		status = "Generating Party B's nonce..."
	case 2:
		status = "Computing nonce public points..."
	case 3:
		status = "Nonce generation complete!"
	}
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(statusStyle.Render(status) + "\n\n")

	// Navigation hint
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(hintStyle.Render("Press Enter to continue..."))

	return builder.String()
}

// renderNonceHex renders nonce hex characters with color, showing progress.
func (s *SignNonceScene) renderNonceHex(chars []rune, index int, color string) string {
	var builder strings.Builder
	for i := 0; i < 64; i++ {
		if i < index {
			if !s.styles.NoColor {
				builder.WriteString(color)
			}
			builder.WriteRune(chars[i])
			if !s.styles.NoColor {
				builder.WriteString(Reset)
			}
		} else {
			builder.WriteRune('.')
		}
		// Space every 8 characters
		if (i+1)%8 == 0 && i < 63 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

// View renders the scene view (required by tea.Model interface).
func (s *SignNonceScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene.
func (s *SignNonceScene) Narrator() string {
	switch s.phase {
	case 0:
		return "Each signing session requires fresh random nonces. Party A generates a secret nonce k_a. This nonce must NEVER be reused -- reusing a nonce in ECDSA leaks the private key."
	case 1:
		return "Party B independently generates their own secret nonce k_b. Like the key shares, neither party reveals their nonce to the other."
	case 2:
		return "Each party computes their nonce public point by scalar multiplication with the generator G. These public points R_a and R_b can be safely exchanged."
	default:
		return "The combined nonce point R = R_a + R_b is computed via point addition. The x-coordinate of R gives us 'r', the first component of the ECDSA signature (r, s). Both parties can compute r since they exchange R_a and R_b."
	}
}
