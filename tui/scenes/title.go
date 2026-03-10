package scenes

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TitleScene represents the ASCII art title screen (Scene 0)
type TitleScene struct {
	config   *Config
	styles   *Styles
	started  bool
	duration time.Duration
}

// NewTitleScene creates a new title scene
func NewTitleScene(config *Config, styles *Styles) *TitleScene {
	return &TitleScene{
		config:   config,
		styles:   styles,
		started:  false,
		duration: getSceneDuration(config.Speed),
	}
}

// Init initializes the scene
func (s *TitleScene) Init() tea.Cmd {
	s.started = true
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *TitleScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *TitleScene) Render() string {
	// ASCII art logo
	asciiArt := `
   ██████╗ ██████╗ ██████╗ ███████╗██████╗ 
  ██╔═══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗
  ██║   ██║██████╔╝██████╔╝█████╗  ██████╔╝
  ██║   ██║██╔══██╗██╔══██╗██╔══╝  ██╔══██╗
  ╚██████╔╝██║  ██║██║  ██║███████╗██║  ██║
   ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
`

	// Mode indicator
	var modeText string
	if s.config.FixedMode {
		modeText = "FIXED MODE"
	} else {
		modeText = "RANDOM MODE"
	}

	// Message preview
	var messagePreview string
	if s.config.Message != "" {
		messagePreview = fmt.Sprintf("Message: %s", s.config.Message)
	} else {
		messagePreview = "Message: (none)"
	}

	// Build the view
	var builder strings.Builder

	// Styles
	var titleStyle lipgloss.Style
	if s.styles.NoColor {
		titleStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	} else {
		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")). // Yellow
			MarginBottom(1)
	}

	var modeStyle lipgloss.Style
	if s.styles.NoColor {
		modeStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	} else {
		modeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")). // Blue
			MarginBottom(1)
	}

	var messageStyle lipgloss.Style
	if s.styles.NoColor {
		messageStyle = lipgloss.NewStyle().MarginBottom(2)
	} else {
		messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")). // Gray
			MarginBottom(2)
	}

	var separatorStyle lipgloss.Style
	if s.styles.NoColor {
		separatorStyle = lipgloss.NewStyle()
	} else {
		separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	}

	// Render ASCII art
	builder.WriteString(asciiArt)
	builder.WriteString("\n")

	// Title
	builder.WriteString(titleStyle.Render("DKLS23 Threshold ECDSA Signing Ceremony") + "\n\n")

	// Separator
	builder.WriteString(separatorStyle.Render(strings.Repeat("─", 50)) + "\n\n")

	// Mode indicator
	builder.WriteString(modeStyle.Render("Mode: " + modeText) + "\n\n")

	// Message preview
	builder.WriteString(messageStyle.Render(messagePreview) + "\n\n")

	// Navigation hint
	var hintStyle lipgloss.Style
	if s.styles.NoColor {
		hintStyle = lipgloss.NewStyle()
	} else {
		hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))
	}
	builder.WriteString(hintStyle.Render("Press Enter to begin the ceremony..."))

	return builder.String()
}

// View renders the scene view (required by tea.Model interface)
func (s *TitleScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *TitleScene) Narrator() string {
	return "Welcome to the DKLS23 Threshold ECDSA Signing Ceremony. This interactive demonstration shows how two parties can jointly sign a message without ever revealing their private keys to each other."
}
