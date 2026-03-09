package scenes

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigScene represents the protocol parameters display (Scene 1)
type ConfigScene struct {
	config   *Config
	styles   *Styles
	started  bool
	duration time.Duration
}

// NewConfigScene creates a new config scene
func NewConfigScene(config *Config, styles *Styles) *ConfigScene {
	return &ConfigScene{
		config:   config,
		styles:   styles,
		started:  false,
		duration: getSceneDuration(config.Speed),
	}
}

// Init initializes the scene
func (s *ConfigScene) Init() tea.Cmd {
	s.started = true
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *ConfigScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Auto-advance timer
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// Render renders the scene view
func (s *ConfigScene) Render() string {
	// Protocol parameters
	curveName := "secp256k1"
	fieldOrder := "FFFFFFFF FFFFFFFF FFFFFFFF FFFFFFFF FFFFFFFF FFFFFFFF FFFFFFFE FFFFFC2F"
	generatorX := "79BE667E F9DCBBAC 55A06295 CE870B07 029BFCDB 2DCE28D9 59F2815B 16F81798"
	generatorY := "483ADA77 26A3C465 5DA4FBFC 0E1108A8 FD17B448 A6855419 9C47D08F FB10D4B8"

	// Mode indicator
	var modeText string
	if s.config.FixedMode {
		modeText = "FIXED MODE"
	} else {
		modeText = "RANDOM MODE"
	}

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

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Blue
		MarginBottom(1)

	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")). // Yellow
		MarginBottom(2)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Header
	builder.WriteString(headerStyle.Render("Protocol Parameters") + "\n\n")

	// Separator
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	// Curve name
	builder.WriteString(labelStyle.Render("Curve:") + "\n")
	builder.WriteString(valueStyle.Render(curveName) + "\n\n")

	// Field order
	builder.WriteString(labelStyle.Render("Field Order (n):") + "\n")
	builder.WriteString(valueStyle.Render(fieldOrder) + "\n\n")

	// Generator point
	builder.WriteString(labelStyle.Render("Generator Point (G):") + "\n")
	builder.WriteString(valueStyle.Render("  Gx: " + generatorX) + "\n")
	builder.WriteString(valueStyle.Render("  Gy: " + generatorY) + "\n\n")

	// Separator
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	// Ceremony config
	builder.WriteString(headerStyle.Render("Ceremony Configuration") + "\n\n")

	// Mode
	builder.WriteString(modeStyle.Render("Mode: " + modeText) + "\n\n")

	// Message
	var messageText string
	if s.config.Message != "" {
		messageText = s.config.Message
	} else {
		messageText = "(none)"
	}
	builder.WriteString(labelStyle.Render("Message:") + "\n")
	builder.WriteString(valueStyle.Render(messageText) + "\n\n")

	// Speed
	builder.WriteString(labelStyle.Render("Animation Speed:") + "\n")
	builder.WriteString(valueStyle.Render(s.config.Speed) + "\n\n")

	// Navigation hint
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(hintStyle.Render("Press Enter to continue..."))

	return builder.String()
}

// View renders the scene view (required by tea.Model interface)
func (s *ConfigScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *ConfigScene) Narrator() string {
	return "These are the protocol parameters for our ceremony. We use secp256k1, the same elliptic curve as Bitcoin and Ethereum. The field order n defines the size of the scalar field, and G is the generator point used for all scalar multiplications."
}
