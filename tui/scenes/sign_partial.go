package scenes

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PartialSigScene represents Scene 10: Partial Signature Computation
// It shows the partial signature formula for each party and the real values.
type PartialSigScene struct {
	config   *Config
	styles   *Styles
	phase    int // 0: show_formula_A, 1: show_formula_B, 2: show_values
	step     int
	started  bool
	duration time.Duration
}

// NewPartialSigScene creates a new partial signature computation scene
func NewPartialSigScene(config *Config, styles *Styles) *PartialSigScene {
	return &PartialSigScene{
		config:   config,
		styles:   styles,
		phase:    0,
		step:     0,
		started:  false,
		duration: getStepDuration(config.Speed),
	}
}

// Init initializes the scene
func (s *PartialSigScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.step = 0
	if s.config.FixedMode {
		ResetFixedCounter()
	}
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *PartialSigScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "enter", "right", "l", " ", "n", "j", "down":
			return s, nil
		}

	case tickMsg:
		s.step++
		if s.phase == 0 && s.step >= 8 {
			s.phase = 1
			s.step = 0
		} else if s.phase == 1 && s.step >= 8 {
			s.phase = 2
			s.step = 0
		}
		if s.phase < 2 {
			return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
	}
	return s, nil
}

// Render renders the scene view
func (s *PartialSigScene) Render() string {
	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")).
		MarginBottom(1)

	builder.WriteString(headerStyle.Render("Scene 10: Partial Signature Computation"))
	builder.WriteString("\n\n")

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	partyAStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Bold(true)

	partyBStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("213")).
		Bold(true)

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		MarginLeft(2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Party A formula (visible in phase 0+)
	if s.phase >= 0 {
		builder.WriteString(partyAStyle.Render("Party A computes:"))
		builder.WriteString("\n")
		builder.WriteString("  ")
		builder.WriteString(partyAStyle.Render("s_a = k_a · z + α · a  (mod n)"))
		builder.WriteString("\n\n")

		if s.phase == 0 {
			// Show computing animation
			computing := "  Computing"
			for i := 0; i < (s.step % 4); i++ {
				computing += "."
			}
			builder.WriteString(labelStyle.Render(computing))
			builder.WriteString("\n\n")
		}
	}

	// Party B formula (visible in phase 1+)
	if s.phase >= 1 {
		builder.WriteString(partyBStyle.Render("Party B computes:"))
		builder.WriteString("\n")
		builder.WriteString("  ")
		builder.WriteString(partyBStyle.Render("s_b = k_b · z + β · b  (mod n)"))
		builder.WriteString("\n\n")

		if s.phase == 1 {
			// Show computing animation
			computing := "  Computing"
			for i := 0; i < (s.step % 4); i++ {
				computing += "."
			}
			builder.WriteString(labelStyle.Render(computing))
			builder.WriteString("\n\n")
		}
	}

	// Show values (phase 2)
	if s.phase >= 2 {
		builder.WriteString(separatorStyle.Render(strings.Repeat("─", 60)))
		builder.WriteString("\n\n")

		builder.WriteString(partyAStyle.Render("Partial Signature A (s_a):"))
		builder.WriteString("\n")
		builder.WriteString("  ")
		builder.WriteString(partyAStyle.Render(s.formatHex(s.config.Ceremony.PartialSigAHex)))
		builder.WriteString("\n\n")

		builder.WriteString(partyBStyle.Render("Partial Signature B (s_b):"))
		builder.WriteString("\n")
		builder.WriteString("  ")
		builder.WriteString(partyBStyle.Render(s.formatHex(s.config.Ceremony.PartialSigBHex)))
		builder.WriteString("\n\n")

		builder.WriteString(explanationStyle.Render("Each partial signature is computed locally."))
		builder.WriteString("\n")
		builder.WriteString(explanationStyle.Render("Neither party reveals their secret key share."))
		builder.WriteString("\n\n")
	}

	// Status and navigation
	var status string
	switch s.phase {
	case 0:
		status = "Party A computing partial signature..."
	case 1:
		status = "Party B computing partial signature..."
	default:
		status = "Both partial signatures computed!"
	}
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(statusStyle.Render(status))
	builder.WriteString("\n\n")

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(hintStyle.Render("Press Enter to continue..."))

	return builder.String()
}

// formatHex formats a hex string in 8-character groups
func (s *PartialSigScene) formatHex(hexStr string) string {
	var result strings.Builder
	for i := 0; i < len(hexStr); i += 8 {
		end := i + 8
		if end > len(hexStr) {
			end = len(hexStr)
		}
		if i > 0 {
			result.WriteString(" ")
		}
		result.WriteString(hexStr[i:end])
	}
	return result.String()
}

// View renders the scene view (required by tea.Model interface)
func (s *PartialSigScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *PartialSigScene) Narrator() string {
	switch s.phase {
	case 0:
		return "Party A computes their partial signature using their nonce share k_a, the message hash z, their MtA share α, and their secret key share a. The formula s_a = k_a · z + α · a ensures that Party A's contribution is bound to the message."
	case 1:
		return "Party B computes their partial signature using the same structure but with their own secret values. Party B uses k_b, β, and b. Neither party needs to know the other's secrets to produce their partial signature."
	default:
		return "Both partial signatures are now computed. Each one is useless on its own — it reveals nothing about either party's secret key share. But when combined, they will form a valid ECDSA signature that can be verified with the combined public key."
	}
}
