package scenes

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CombineScene represents the combined public key with phantom private key (Scene 4)
type CombineScene struct {
	config   *Config
	styles   *Styles
	phase    int // 0: computing, 1: complete
	step     int
	started  bool
	duration time.Duration
}

// NewCombineScene creates a new combine scene
func NewCombineScene(config *Config, styles *Styles) *CombineScene {
	return &CombineScene{
		config:   config,
		styles:   styles,
		phase:    0,
		step:     0,
		started:  false,
		duration: getStepDuration(config.Speed),
	}
}

// Init initializes the scene
func (s *CombineScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.step = 0
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *CombineScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "enter", "right", "l", " ", "n", "j", "down":
			// Advance to next scene - handled by parent model
			return s, nil
		}

	case tickMsg:
		// Animation tick
		s.step++
		if s.phase == 0 && s.step >= 10 {
			s.phase = 1
			s.step = 0
		}
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// Render renders the scene view
func (s *CombineScene) Render() string {
	// Build the view
	var builder strings.Builder

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")). // Yellow
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		MarginRight(2)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Header
	builder.WriteString(headerStyle.Render("Combined Public Key") + "\n\n")

	// Separator
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	// Show the combination
	builder.WriteString(labelStyle.Render("Public Key Combination:") + "\n\n")

	// Party A's public share
	builder.WriteString("  ")
	builder.WriteString(PartyAColor + "A" + Reset + " = a × G (Party A's public share)")
	builder.WriteString("\n")

	// Party B's public share
	builder.WriteString("  ")
	builder.WriteString(PartyBColor + "B" + Reset + " = b × G (Party B's public share)")
	builder.WriteString("\n\n")

	// Combination arrow
	if s.phase == 0 {
		arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		builder.WriteString("  ")
		for i := 0; i < 10; i++ {
			if i < s.step {
				builder.WriteString(arrowStyle.Render("↓"))
			} else {
				builder.WriteString("·")
			}
		}
		builder.WriteString("\n")
	} else {
		builder.WriteString("  ↓ (point addition)\n")
	}

	// Combined public key
	combinedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Yellow/Gold
		Bold(true)

	builder.WriteString("  ")
	builder.WriteString(combinedStyle.Render("P = A + B"))
	builder.WriteString("\n")
	builder.WriteString("  ")
	builder.WriteString(combinedStyle.Render("P = (a + b) × G"))
	builder.WriteString("\n\n")

	// Phantom private key
	phantomStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red
		Strikethrough(true)

	builder.WriteString(labelStyle.Render("Phantom Private Key:") + "\n\n")
	builder.WriteString("  ")
	builder.WriteString(phantomStyle.Render("x = a + b"))
	builder.WriteString("\n")
	builder.WriteString("  ")
	builder.WriteString(phantomStyle.Render("(NEVER COMPUTED - never exists as a single value)"))
	builder.WriteString("\n\n")

	// Explanation
	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(explanationStyle.Render("The combined public key P is computed by adding the two public shares."))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("The phantom private key x = a + b is the sum of both secrets."))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("Crucially, x is NEVER computed - it only exists conceptually."))
	builder.WriteString("\n")
	builder.WriteString(explanationStyle.Render("This is the magic of threshold cryptography!"))
	builder.WriteString("\n\n")

	// Status
	var status string
	if s.phase == 0 {
		status = "Computing combined public key..."
	} else {
		status = "Combined public key computed!"
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

// View renders the scene view (required by tea.Model interface)
func (s *CombineScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *CombineScene) Narrator() string {
	if s.phase == 0 {
		return "The combined public key is computed by adding the two public shares using elliptic curve point addition. This gives us P = A + B = (a + b) × G."
	}
	return "The phantom private key x = a + b is the conceptual sum of both secrets. Crucially, this value is NEVER actually computed - it only exists as a mathematical concept. This is the fundamental magic of threshold cryptography: we can sign with x without ever knowing x."
}
