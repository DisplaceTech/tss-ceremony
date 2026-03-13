package scenes

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SignMessageScene represents Scene 5: Message & Hash Display.
// It shows the plaintext message, its byte-by-byte hex conversion,
// and the SHA-256 hash computation with a rolling hex animation.
type SignMessageScene struct {
	config   *Config
	styles   *Styles
	phase    int // 0: show message, 1: show bytes, 2: show hash
	step     int // animation step within phase
	started  bool
	duration time.Duration

	// Rolling hash animation state
	hashChars   []rune // current display characters for the hash
	hashSettled int    // how many characters have settled to final value
}

// NewSignMessageScene creates a new sign message scene.
func NewSignMessageScene(config *Config, styles *Styles) *SignMessageScene {
	return &SignMessageScene{
		config:    config,
		styles:    styles,
		phase:     0,
		step:      0,
		started:   false,
		duration:  getCharDuration(config.Speed),
		hashChars: make([]rune, 64), // SHA-256 = 32 bytes = 64 hex chars
	}
}

// Init initializes the scene.
func (s *SignMessageScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.step = 0
	s.hashSettled = 0
	if s.config.FixedMode {
		ResetFixedCounter()
	}
	// Initialize hash chars with placeholders
	for i := range s.hashChars {
		s.hashChars[i] = '.'
	}
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene.
func (s *SignMessageScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Show message phase — brief pause then advance
			s.step++
			if s.step >= 8 {
				s.phase = 1
				s.step = 0
			}
		case 1:
			// Show bytes phase — brief pause then advance
			s.step++
			if s.step >= 8 {
				s.phase = 2
				s.step = 0
			}
		case 2:
			// Hash rolling animation
			finalHash := s.getFinalHash()
			if s.hashSettled < 64 {
				// Roll unsettled characters
				for i := s.hashSettled; i < 64; i++ {
					s.hashChars[i] = pickHexChar(s.config.FixedMode)
				}
				// Settle the next character
				if s.hashSettled < len(finalHash) {
					s.hashChars[s.hashSettled] = rune(finalHash[s.hashSettled])
				}
				s.hashSettled++
			}
		}
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// getFinalHash returns the final hash hex string from ceremony data.
func (s *SignMessageScene) getFinalHash() string {
	if s.config.Ceremony != nil && s.config.Ceremony.MessageHash != "" {
		return s.config.Ceremony.MessageHash
	}
	return strings.Repeat("0", 64)
}

// getMessageText returns the message being signed.
func (s *SignMessageScene) getMessageText() string {
	if s.config.Ceremony != nil && s.config.Ceremony.MessageText != "" {
		return s.config.Ceremony.MessageText
	}
	if s.config.Message != "" {
		return s.config.Message
	}
	return "Hello, threshold signatures!"
}

// formatHexGroups formats a hex string in 8-character groups.
func formatHexGroups(hexStr string) string {
	var result strings.Builder
	for i := 0; i < len(hexStr); i++ {
		if i > 0 && i%8 == 0 {
			result.WriteString(" ")
		}
		result.WriteByte(hexStr[i])
	}
	return result.String()
}

// Render renders the scene view.
func (s *SignMessageScene) Render() string {
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

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Header
	builder.WriteString(headerStyle.Render("Message & Hash") + "\n\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	messageText := s.getMessageText()

	// Phase 0+: Always show the plaintext message
	builder.WriteString(labelStyle.Render("Message:") + "\n")
	msgStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)
	builder.WriteString("  " + msgStyle.Render(fmt.Sprintf("%q", messageText)) + "\n\n")

	// Phase 1+: Show byte-by-byte hex conversion
	if s.phase >= 1 {
		builder.WriteString(labelStyle.Render("Message bytes (hex):") + "\n")
		hexBytes := hex.EncodeToString([]byte(messageText))
		builder.WriteString("  " + valueStyle.Render(formatHexGroups(hexBytes)) + "\n\n")
	}

	// Phase 2: Show SHA-256 hash with rolling animation
	if s.phase >= 2 {
		builder.WriteString(labelStyle.Render("SHA-256 Hash:") + "\n")
		builder.WriteString("  " + s.renderHashChars() + "\n\n")

		// Show hash formula
		formulaStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))
		builder.WriteString(formulaStyle.Render("  h = SHA-256(message)") + "\n\n")
	}

	// Status
	var status string
	switch s.phase {
	case 0:
		status = "Displaying message to sign..."
	case 1:
		status = "Converting message to bytes..."
	case 2:
		if s.hashSettled >= 64 {
			status = "Hash computed!"
		} else {
			status = "Computing SHA-256 hash..."
		}
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

// renderHashChars renders the hash characters with color coding for settled vs rolling.
func (s *SignMessageScene) renderHashChars() string {
	var builder strings.Builder
	settledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")) // Yellow for settled
	rollingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")) // Dim for rolling

	for i := 0; i < 64; i++ {
		if i < s.hashSettled {
			builder.WriteString(settledStyle.Render(string(s.hashChars[i])))
		} else if s.hashChars[i] != '.' {
			builder.WriteString(rollingStyle.Render(string(s.hashChars[i])))
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
func (s *SignMessageScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene.
func (s *SignMessageScene) Narrator() string {
	switch s.phase {
	case 0:
		return "Before signing, we need to prepare the message. ECDSA does not sign the raw message directly — it signs a hash of the message."
	case 1:
		return "The message is first converted to its byte representation. Each character becomes one or more bytes in UTF-8 encoding."
	default:
		if s.hashSettled >= 64 {
			return "The SHA-256 hash is complete. This 256-bit digest is what both parties will jointly sign using the DKLS protocol. The hash ensures that any change to the message produces a completely different value to sign."
		}
		return "SHA-256 processes the message bytes through 64 rounds of compression, producing a fixed 256-bit (32-byte) digest. This hash will be the input to the ECDSA signing equation."
	}
}
