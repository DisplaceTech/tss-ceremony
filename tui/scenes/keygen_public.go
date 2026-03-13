package scenes

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// PublicShareScene represents the public share computation and exchange (Scene 3)
type PublicShareScene struct {
	config      *Config
	styles      *Styles
	phase       int // 0: computing A, 1: computing B, 2: exchange, 3: complete
	step        int // animation step within phase
	started     bool
	duration    time.Duration
	partyAPub   string // Hex string of Party A's public share
	partyBPub   string // Hex string of Party B's public share
	secretA     []byte // Party A's secret (for computing public share)
	secretB     []byte // Party B's secret (for computing public share)
}

// NewPublicShareScene creates a new public share scene
func NewPublicShareScene(config *Config, styles *Styles) *PublicShareScene {
	return &PublicShareScene{
		config:   config,
		styles:   styles,
		phase:    0,
		step:     0,
		started:  false,
		duration: getStepDuration(config.Speed),
	}
}

// SetSecrets sets the secret shares for both parties (called by parent model)
func (s *PublicShareScene) SetSecrets(secretA, secretB []byte) {
	s.secretA = secretA
	s.secretB = secretB
}

// ComputePublicShares computes the public shares from secrets using protocol logic
func (s *PublicShareScene) ComputePublicShares() error {
	if s.secretA != nil && len(s.secretA) == 32 {
		// Compute Party A's public share: A = a × G
		privKey := secp256k1.PrivKeyFromBytes(s.secretA)
		if privKey != nil {
			pubKey := privKey.PubKey()
			s.partyAPub = formatPublicKeyHex(pubKey)
		}
	}
	if s.secretB != nil && len(s.secretB) == 32 {
		// Compute Party B's public share: B = b × G
		privKey := secp256k1.PrivKeyFromBytes(s.secretB)
		if privKey != nil {
			pubKey := privKey.PubKey()
			s.partyBPub = formatPublicKeyHex(pubKey)
		}
	}
	return nil
}

// formatPublicKeyHex formats a public key as hex string with 8-char groups
func formatPublicKeyHex(pubKey *secp256k1.PublicKey) string {
	// Use uncompressed format for display (0x04 prefix + x + y)
	bytes := pubKey.SerializeUncompressed()
	hexStr := hex.EncodeToString(bytes)
	// Format with spaces every 8 characters
	var result strings.Builder
	for i, c := range hexStr {
		if i > 0 && i%8 == 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// Init initializes the scene
func (s *PublicShareScene) Init() tea.Cmd {
	s.started = true
	s.phase = 0
	s.step = 0
	s.partyAPub = ""
	s.partyBPub = ""
	return tea.Tick(s.duration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles events in the scene
func (s *PublicShareScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		switch s.phase {
		case 0:
			// Computing Party A's public share
			if s.step >= 10 {
				s.phase = 1
				s.step = 0
			}
		case 1:
			// Computing Party B's public share
			if s.step >= 10 {
				s.phase = 2
				s.step = 0
			}
		case 2:
			// Exchange animation
			if s.step >= 15 {
				s.phase = 3
				s.step = 0
			}
		case 3:
			// Complete, keep ticking for auto-advance
		}
		return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

// renderThreeColumns renders a three-column layout using │ separators.
// Each column is padded to its fixed width.
func renderThreeColumns(leftLines, sharedLines, rightLines []string, leftW, sharedW, rightW int) string {
	padRight := func(s string, w int) string {
		if len(s) >= w {
			return s[:w]
		}
		return s + strings.Repeat(" ", w-len(s))
	}
	rowAt := func(lines []string, i int) string {
		if i < len(lines) {
			return lines[i]
		}
		return ""
	}
	maxLen := len(leftLines)
	if len(sharedLines) > maxLen {
		maxLen = len(sharedLines)
	}
	if len(rightLines) > maxLen {
		maxLen = len(rightLines)
	}

	var sb strings.Builder
	// Top border
	sb.WriteString("┌" + strings.Repeat("─", leftW) + "┬" +
		strings.Repeat("─", sharedW) + "┬" +
		strings.Repeat("─", rightW) + "┐\n")
	for i := 0; i < maxLen; i++ {
		sb.WriteString("│")
		sb.WriteString(padRight(rowAt(leftLines, i), leftW))
		sb.WriteString("│")
		sb.WriteString(padRight(rowAt(sharedLines, i), sharedW))
		sb.WriteString("│")
		sb.WriteString(padRight(rowAt(rightLines, i), rightW))
		sb.WriteString("│\n")
	}
	// Bottom border
	sb.WriteString("└" + strings.Repeat("─", leftW) + "┴" +
		strings.Repeat("─", sharedW) + "┴" +
		strings.Repeat("─", rightW) + "┘\n")
	return sb.String()
}

// Render renders the scene view
func (s *PublicShareScene) Render() string {
	const leftW, sharedW, rightW = 24, 32, 24

	// Party A column
	leftLines := []string{"PARTY A", strings.Repeat("-", leftW)}
	switch s.phase {
	case 0:
		leftLines = append(leftLines, "a = secret", "Computing...")
	default:
		leftLines = append(leftLines, "a = secret", "A = a x G", "(done)")
	}

	// Shared column
	sharedLines := []string{"SHARED STATE", strings.Repeat("-", sharedW)}
	switch s.phase {
	case 0:
		sharedLines = append(sharedLines, "Waiting for A...")
	case 1:
		sharedLines = append(sharedLines, "A computed", "Waiting for B...")
	case 2:
		sharedLines = append(sharedLines, "Exchanging...", "A -> B", "B -> A")
	default:
		sharedLines = append(sharedLines, "Exchange done", "A = a x G", "B = b x G")
	}

	// Party B column
	rightLines := []string{"PARTY B", strings.Repeat("-", rightW)}
	switch s.phase {
	case 0:
		rightLines = append(rightLines, "b = secret", "Waiting...")
	case 1:
		rightLines = append(rightLines, "b = secret", "Computing...")
	default:
		rightLines = append(rightLines, "b = secret", "B = b x G", "(done)")
	}

	threeCol := renderThreeColumns(leftLines, sharedLines, rightLines, leftW, sharedW, rightW)

	var status string
	switch s.phase {
	case 0:
		status = "Computing Party A's public share: A = a x G"
	case 1:
		status = "Computing Party B's public share: B = b x G"
	case 2:
		status = "Exchanging public shares..."
	default:
		status = "Public shares exchanged!"
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	return threeCol + statusStyle.Render(status) + "\n"
}


// View renders the scene view (required by tea.Model interface)
func (s *PublicShareScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *PublicShareScene) Narrator() string {
	if s.phase == 0 {
		return "Party A computes their public share by multiplying their secret by the generator point G. This is scalar multiplication on the elliptic curve."
	} else if s.phase == 1 {
		return "Party B computes their public share the same way. Scalar multiplication is easy to compute but hard to reverse - this is the one-way property of elliptic curves."
	} else if s.phase == 2 {
		return "Now both parties exchange their public shares. Unlike secrets, public shares can be shared openly - anyone can see them."
	}
	return "Both parties now have each other's public shares. These will be used to compute the combined public key."
}
