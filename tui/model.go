package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// Scene represents a single scene in the ceremony.
// Concrete scenes (from tui/scenes/) implement the full interface including
// Render and Narrator; placeholder scenes provide stub implementations.
type Scene interface {
	tea.Model
	Render() string
	Narrator() string
}

// Footer holds footer display information
type Footer struct {
	CurrentScene int
	TotalScenes  int
	KeyBindings  string
}

// Header holds header display information
type Header struct {
	SceneNum   int
	TotalScenes int
	PhaseName  string
}

// Config holds the TUI configuration
type Config struct {
	FixedMode bool
	Message   string
	Speed     string
	NoColor   bool
}

// Model represents the main TUI model that manages scenes
type Model struct {
	config       *scenes.Config
	ceremony     *protocol.Ceremony
	currentScene int
	scenes       []Scene
	quit         bool
	speedDelay   time.Duration
	styles       *scenes.Styles
	header       Header
	footer       Footer
}

// NewModel creates a new TUI model with all scenes wired up
func NewModel(config *scenes.Config, ceremony *protocol.Ceremony) Model {
	m := Model{
		config:     config,
		ceremony:   ceremony,
		quit:       false,
		speedDelay: getSpeedDelay(config.Speed),
		styles:     scenes.NewStyles(config.NoColor),
	}

	m.scenes = m.createScenes()
	m.updateHeaderFooter()

	return m
}

// getSpeedDelay returns the delay duration based on speed setting
func getSpeedDelay(speed string) time.Duration {
	switch speed {
	case "slow":
		return 200 * time.Millisecond
	case "fast":
		return 50 * time.Millisecond
	default:
		return 100 * time.Millisecond
	}
}

// updateHeaderFooter updates the header and footer data structures
func (m *Model) updateHeaderFooter() {
	m.header = Header{
		SceneNum:    m.currentScene,
		TotalScenes: len(m.scenes),
		PhaseName:   sceneNames[m.currentScene],
	}
	m.footer = Footer{
		CurrentScene: m.currentScene,
		TotalScenes:  len(m.scenes),
		KeyBindings:  "Enter/←/→/q",
	}
}

// sceneNames maps scene indices to display names.
var sceneNames = [20]string{
	"Scene 0: Title Screen",
	"Scene 1: Protocol Parameters",
	"Scene 2: Secret Generation",
	"Scene 3: Public Share Exchange",
	"Scene 4: Combined Public Key",
	"Scene 5: Message & Hash",
	"Scene 6: Nonce Generation",
	"Scene 7-8: Oblivious Transfer",
	"Scene 7-8: Oblivious Transfer",
	"Scene 9: MtA Conversion",
	"Scene 10: Partial Signatures",
	"Scene 11: Signature Assembly",
	"Scene 12: Verification",
	"Scene 13: Security Proof",
	"Scene 14: Ceremony Summary",
	"Scene 15: The Reveal",
	"Scene 16: Schnorr vs ECDSA",
	"Scene 17: FROST Side-by-Side",
	"Scene 18: Animated FROST",
	"Scene 19: Why Both Exist",
}

// createScenes creates all ceremony scenes.
func (m Model) createScenes() []Scene {
	s := make([]Scene, 20)

	// Core DKLS ceremony scenes (0-4): Key Generation
	s[0] = scenes.NewTitleScene(m.config, m.styles)
	s[1] = scenes.NewConfigScene(m.config, m.styles)
	s[2] = scenes.NewSecretGenScene(m.config, m.styles)
	s[3] = scenes.NewPublicShareScene(m.config, m.styles)
	s[4] = scenes.NewCombineScene(m.config, m.styles)

	// Core DKLS ceremony scenes (5-11): Signing
	s[5] = scenes.NewSignMessageScene(m.config, m.styles)
	s[6] = scenes.NewSignNonceScene(m.config, m.styles)
	s[7] = scenes.NewSignOTScene(m.config, m.styles)
	s[8] = scenes.NewSignOTScene(m.config, m.styles) // OT continued (same scene type)
	s[9] = scenes.NewMtAScene(m.config, m.styles)
	s[10] = scenes.NewPartialSigScene(m.config, m.styles)
	s[11] = scenes.NewCombineSigScene(m.config, m.styles)

	// Core DKLS ceremony scenes (12-14): Verification & Summary
	noColor := m.config != nil && m.config.NoColor
	pubkey, sigR, sigS, message := "", "", "", ""
	valid := true
	if m.config != nil && m.config.Ceremony != nil {
		cd := m.config.Ceremony
		pubkey = cd.CombinedPubHex
		sigR = cd.SignatureRHex
		sigS = cd.SignatureSHex
		message = cd.MessageText
		valid = cd.Valid
	}
	s[12] = scenes.NewVerifyScene(noColor, pubkey, sigR, sigS, message, valid)
	s[13] = scenes.NewImpossibilityScene(m.config, m.styles)
	s[14] = scenes.NewSummaryScene(m.config, m.styles)

	// Bonus scenes (15-19): Schnorr/FROST comparison
	s[15] = scenes.NewRevealScene()
	s[16] = scenes.NewSchnorrCompareScene()
	s[17] = scenes.NewScene()
	s[18] = scenes.NewFrostAnimatedScene(m.config)
	s[19] = scenes.NewWhyBothScene()

	return s
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Tick(m.speedDelay, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// tickMsg is a message type for animation ticks
type tickMsg time.Time

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		case "right", "l", " ", "n", "j", "down":
			if m.currentScene < len(m.scenes)-1 {
				m.currentScene++
				m.updateHeaderFooter()
				return m, m.scenes[m.currentScene].Init()
			}
		case "left", "h", "p", "k", "up":
			if m.currentScene > 0 {
				m.currentScene--
				m.updateHeaderFooter()
				return m, m.scenes[m.currentScene].Init()
			}
		}

	case tea.WindowSizeMsg:
		// Handle window resize if needed

	case tickMsg:
		return m, tea.Tick(m.speedDelay, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	// Update current scene
	if m.currentScene < len(m.scenes) {
		updated, cmd := m.scenes[m.currentScene].Update(msg)
		m.scenes[m.currentScene] = updated.(Scene)
		return m, cmd
	}

	return m, nil
}

// View renders the current scene
func (m Model) View() string {
	if m.quit {
		return "Goodbye!\n"
	}

	var view string

	if m.config.FixedMode {
		view += fixedModeBanner() + "\n"
	}

	if m.currentScene < len(m.scenes) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			MarginBottom(1).
			Foreground(lipgloss.Color("226"))

		// Render header using Header struct
		headerText := fmt.Sprintf("Scene %d/%d · %s",
			m.header.SceneNum, m.header.TotalScenes, m.header.PhaseName)
		view += headerStyle.Render(headerText) + "\n"

		scene := m.scenes[m.currentScene]
		view += scene.Render() + "\n"

		narrator := scene.Narrator()
		if narrator != "" {
			narratorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Bold(true).
				MarginTop(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("241")).
				Padding(0, 1)
			view += "\n" + narratorStyle.Render("Narrator: "+narrator) + "\n"
		}
	}

	// Render footer using Footer struct
	navigationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	view += "\n" + navigationStyle.Render(fmt.Sprintf("[%d/%d] [Key bindings: %s]",
		m.footer.CurrentScene, m.footer.TotalScenes, m.footer.KeyBindings))

	return view
}

// GetCurrentScene returns the current scene index
func (m Model) GetCurrentScene() int {
	return m.currentScene
}

// GetSceneCount returns the total number of scenes
func (m Model) GetSceneCount() int {
	return len(m.scenes)
}

// fixedModeBanner returns a banner for fixed mode
func fixedModeBanner() string {
	return "=== FIXED MODE ==="
}

// PlaceholderScene is a placeholder for scene implementations not yet built.
type PlaceholderScene struct {
	SceneNum int
	Config   *scenes.Config
	Styles   *scenes.Styles
}

func (s *PlaceholderScene) Init() tea.Cmd { return nil }

func (s *PlaceholderScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

func (s *PlaceholderScene) View() string {
	if s.Config != nil && s.Config.NoColor {
		return fmt.Sprintf("Scene %d placeholder - implement me!", s.SceneNum)
	}
	return fmt.Sprintf("%sScene %d placeholder - implement me!%s", "\033[36m", s.SceneNum, "\033[0m")
}

func (s *PlaceholderScene) Render() string {
	return s.View()
}

func (s *PlaceholderScene) Narrator() string {
	return ""
}
