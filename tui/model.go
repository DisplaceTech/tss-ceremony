// Package tui implements the ceremony animation as a single continuous
// protocol trace that progressively reveals real cryptographic values.
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// Animation phases — each phase reveals one element of the protocol trace.
const (
	phaseStart    = iota // title only
	phaseSecretA         // animate Party A secret
	phaseSecretB         // animate Party B secret
	phasePubA            // animate Party A public key
	phasePubB            // animate Party B public key
	phaseCombined        // animate combined key P
	phaseMsg             // show message text (instant)
	phaseHash            // animate SHA-256 hash
	phaseNonceA          // animate nonce k_a
	phaseNonceB          // animate nonce k_b
	phaseR               // show R_a, R_b, R, r (instant)
	phaseOT              // show OT values (instant)
	phaseMtA             // show MtA values (instant)
	phasePartialA        // animate partial sig s_a
	phasePartialB        // animate partial sig s_b
	phaseSig             // animate final s
	phaseVerify          // show verification (instant)
	phaseDone            // animation complete
)

type tickMsg time.Time

// styles holds lipgloss styles, respecting --no-color.
type styles struct {
	cyan    lipgloss.Style
	magenta lipgloss.Style
	yellow  lipgloss.Style
	green   lipgloss.Style
	red     lipgloss.Style
	dim     lipgloss.Style
	bold    lipgloss.Style
}

func newStyles(noColor bool) styles {
	if noColor {
		return styles{
			cyan: lipgloss.NewStyle(), magenta: lipgloss.NewStyle(),
			yellow: lipgloss.NewStyle(), green: lipgloss.NewStyle(),
			red: lipgloss.NewStyle(), dim: lipgloss.NewStyle(),
			bold: lipgloss.NewStyle(),
		}
	}
	return styles{
		cyan:    lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		magenta: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		yellow:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		green:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		red:     lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		bold:    lipgloss.NewStyle().Bold(true),
	}
}

// Model is the top-level bubbletea model for the ceremony animation.
type Model struct {
	config *scenes.Config
	data   *scenes.CeremonyData
	s      styles

	phase     int // current animation phase
	animPos   int // character position within current hex animation
	waitTicks int // pause ticks between phases

	width  int
	height int
	paused bool
}

// NewModel creates the animation model.
func NewModel(config *scenes.Config, _ *protocol.Ceremony) *Model {
	return &Model{
		config: config,
		data:   config.Ceremony,
		s:      newStyles(config.NoColor),
		phase:  phaseStart,
	}
}

// Init starts the animation ticker.
func (m *Model) Init() tea.Cmd {
	return m.tick()
}

func (m *Model) tick() tea.Cmd {
	d := 30 * time.Millisecond
	switch m.config.Speed {
	case "slow":
		d = 60 * time.Millisecond
	case "fast":
		d = 12 * time.Millisecond
	}
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Update handles input and animation ticks.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
			if !m.paused {
				return m, m.tick()
			}
			return m, nil
		case "enter":
			if m.phase < phaseDone {
				m.phase++
				m.animPos = 0
				m.waitTicks = 0
			}
			return m, m.tick()
		}

	case tickMsg:
		if m.paused || m.phase >= phaseDone {
			return m, nil
		}
		m.advance()
		return m, m.tick()
	}
	return m, nil
}

// advance progresses the animation by one tick.
func (m *Model) advance() {
	// Inter-phase pause
	if m.waitTicks > 0 {
		m.waitTicks--
		return
	}

	target := m.phaseHex()
	if target == "" {
		// Instant phase
		m.phase++
		m.animPos = 0
		m.waitTicks = 5
		return
	}

	// Animated phase: advance character position
	m.animPos++
	if m.animPos >= len(target) {
		m.phase++
		m.animPos = 0
		m.waitTicks = 8
	}
}

// phaseHex returns the hex string to animate for the current phase,
// or "" if the phase is an instant reveal.
func (m *Model) phaseHex() string {
	switch m.phase {
	case phaseSecretA:
		return m.data.PartyASecretHex
	case phaseSecretB:
		return m.data.PartyBSecretHex
	case phasePubA:
		return m.data.PartyAPubHex
	case phasePubB:
		return m.data.PartyBPubHex
	case phaseCombined:
		return m.data.CombinedPubHex
	case phaseHash:
		return m.data.MessageHash
	case phaseNonceA:
		return m.data.NonceAHex
	case phaseNonceB:
		return m.data.NonceBHex
	case phasePartialA:
		return m.data.PartialSigAHex
	case phasePartialB:
		return m.data.PartialSigBHex
	case phaseSig:
		return m.data.SignatureSHex
	default:
		return ""
	}
}

// View renders the current animation frame.
func (m *Model) View() string {
	content := m.renderTrace()
	lines := strings.Split(content, "\n")
	status := m.statusLine()

	avail := m.height - 1
	if avail < 1 {
		avail = 24
	}

	// Auto-scroll: show the bottom N lines
	if len(lines) > avail {
		lines = lines[len(lines)-avail:]
	}
	// Pad short content
	for len(lines) < avail {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + status
}

func (m *Model) statusLine() string {
	d := m.s.dim
	if m.phase >= phaseDone {
		return d.Render("  ceremony complete · [q] quit")
	}
	if m.paused {
		return d.Render("  ▐▐ paused · [space] resume  [enter] skip step  [q] quit")
	}
	return d.Render("  ▶ [space] pause  [enter] skip step  [q] quit")
}

// ---------------------------------------------------------------------------
// Rendering — builds the protocol trace as a single string
// ---------------------------------------------------------------------------

func (m *Model) section(title string) string {
	w := m.width
	if w < 80 {
		w = 80
	}
	pad := w - len(title) - 8
	if pad < 4 {
		pad = 4
	}
	return m.s.bold.Render("  ─── "+title+" ") + m.s.dim.Render(strings.Repeat("─", pad))
}

func (m *Model) renderTrace() string {
	var b strings.Builder

	// Title — always visible
	b.WriteString(m.s.bold.Render("  DKLS23 2-of-2 Threshold ECDSA"))
	b.WriteString(m.s.dim.Render(" · secp256k1") + "\n")
	w := m.width
	if w < 80 {
		w = 80
	}
	b.WriteString(m.s.dim.Render("  "+strings.Repeat("═", w-4)) + "\n")

	if m.phase <= phaseStart {
		return b.String()
	}

	// === Key Generation ===
	b.WriteString("\n" + m.section("Key Generation") + "\n")
	b.WriteString(m.s.cyan.Render("  a ") + m.animHex(phaseSecretA, m.data.PartyASecretHex))
	b.WriteString(m.s.dim.Render("  Party A secret") + "\n")

	if m.phase <= phaseSecretA {
		return b.String()
	}

	b.WriteString(m.s.magenta.Render("  b ") + m.animHex(phaseSecretB, m.data.PartyBSecretHex))
	b.WriteString(m.s.dim.Render("  Party B secret") + "\n")

	if m.phase <= phaseSecretB {
		return b.String()
	}

	b.WriteString(m.s.cyan.Render("  A") + m.s.dim.Render(" = a·G") + "\n")
	b.WriteString("    " + m.animHex(phasePubA, m.data.PartyAPubHex) + "\n")

	if m.phase <= phasePubA {
		return b.String()
	}

	b.WriteString(m.s.magenta.Render("  B") + m.s.dim.Render(" = b·G") + "\n")
	b.WriteString("    " + m.animHex(phasePubB, m.data.PartyBPubHex) + "\n")

	if m.phase <= phasePubB {
		return b.String()
	}

	b.WriteString(m.s.yellow.Render("  P") + m.s.dim.Render(" = A + B"))
	b.WriteString(m.s.red.Render("  ← no one knows p where P = p·G") + "\n")
	b.WriteString("    " + m.animHex(phaseCombined, m.data.CombinedPubHex) + "\n")

	if m.phase <= phaseCombined {
		return b.String()
	}

	// === Signing ===
	b.WriteString("\n" + m.section("Signing") + "\n")
	b.WriteString(m.s.dim.Render("  m ") + fmt.Sprintf(`"%s"`, m.data.MessageText) + "\n")

	if m.phase <= phaseMsg {
		return b.String()
	}

	b.WriteString(m.s.dim.Render("  H = SHA-256(m)") + "\n")
	b.WriteString("    " + m.animHex(phaseHash, m.data.MessageHash) + "\n")

	if m.phase <= phaseHash {
		return b.String()
	}

	// === Nonces ===
	b.WriteString("\n" + m.section("Nonces") + "\n")
	b.WriteString(m.s.cyan.Render("  k_a ") + m.animHex(phaseNonceA, m.data.NonceAHex) + "\n")

	if m.phase <= phaseNonceA {
		return b.String()
	}

	b.WriteString(m.s.magenta.Render("  k_b ") + m.animHex(phaseNonceB, m.data.NonceBHex) + "\n")

	if m.phase <= phaseNonceB {
		return b.String()
	}

	// R values — instant reveal
	if m.data.NonceAPubHex != "" {
		b.WriteString(m.s.cyan.Render("  R_a") + m.s.dim.Render(" = k_a·G = ") + fmtHex(m.data.NonceAPubHex) + "\n")
	}
	if m.data.NonceBPubHex != "" {
		b.WriteString(m.s.magenta.Render("  R_b") + m.s.dim.Render(" = k_b·G = ") + fmtHex(m.data.NonceBPubHex) + "\n")
	}
	if m.data.CombinedRPubHex != "" {
		b.WriteString(m.s.yellow.Render("  R  ") + m.s.dim.Render(" = R_a + R_b = ") + fmtHex(m.data.CombinedRPubHex) + "\n")
	}
	if m.data.RHex != "" {
		b.WriteString(m.s.yellow.Render("  r  ") + m.s.dim.Render(" = R.x mod n = ") + fmtHex(m.data.RHex) + "\n")
	}

	if m.phase <= phaseR {
		return b.String()
	}

	// === Oblivious Transfer ===
	b.WriteString("\n" + m.section("Oblivious Transfer") + "\n")
	b.WriteString(m.s.dim.Render("  sender:   ") +
		fmt.Sprintf("x₀ = %s", truncHex(m.data.OTInput0Hex, 16)) + "\n")
	b.WriteString(m.s.dim.Render("            ") +
		fmt.Sprintf("x₁ = %s", truncHex(m.data.OTInput1Hex, 16)) + "\n")
	b.WriteString(m.s.dim.Render("  receiver: ") +
		fmt.Sprintf("c  = %d    x_c = %s", m.data.OTChoiceBit, truncHex(m.data.OTOutputHex, 16)) + "\n")

	if m.phase <= phaseOT {
		return b.String()
	}

	// === MtA ===
	b.WriteString("\n" + m.section("Multiplicative to Additive") + "\n")
	b.WriteString(m.s.dim.Render("  α + β ≡ k_a · k_b  (mod n)") + "\n")
	b.WriteString(m.s.cyan.Render("  α ") + fmtHex(m.data.AlphaHex) + "\n")
	b.WriteString(m.s.magenta.Render("  β ") + fmtHex(m.data.BetaHex) + "\n")

	if m.phase <= phaseMtA {
		return b.String()
	}

	// === Partial Signatures ===
	b.WriteString("\n" + m.section("Partial Signatures") + "\n")
	b.WriteString(m.s.cyan.Render("  s_a") + m.s.dim.Render(" = k_a⁻¹·(H + r·a) + α") + "\n")
	b.WriteString("      " + m.animHex(phasePartialA, m.data.PartialSigAHex) + "\n")

	if m.phase <= phasePartialA {
		return b.String()
	}

	b.WriteString(m.s.magenta.Render("  s_b") + m.s.dim.Render(" = k_b⁻¹·(H + r·b) + β") + "\n")
	b.WriteString("      " + m.animHex(phasePartialB, m.data.PartialSigBHex) + "\n")

	if m.phase <= phasePartialB {
		return b.String()
	}

	// === Combine ===
	b.WriteString("\n" + m.section("Combine") + "\n")
	b.WriteString(m.s.yellow.Render("  s") + m.s.dim.Render(" = s_a + s_b mod n") + "\n")
	b.WriteString("    " + m.animHex(phaseSig, m.data.SignatureSHex) + "\n")

	if m.phase <= phaseSig {
		return b.String()
	}

	// === Result ===
	b.WriteString("\n" + m.section("Signature") + "\n")
	b.WriteString(m.s.yellow.Render("  r ") + fmtHex(m.data.SignatureRHex) + "\n")
	b.WriteString(m.s.yellow.Render("  s ") + fmtHex(m.data.SignatureSHex) + "\n")
	b.WriteString("\n")

	// === Verify ===
	if m.data.Valid {
		b.WriteString("  ECDSA.Verify(P, H(m), r, s) → " + m.s.green.Render("✓ VALID") + "\n")
	} else {
		b.WriteString("  ECDSA.Verify(P, H(m), r, s) → " + m.s.red.Render("✗ INVALID") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(m.s.dim.Render("  The signature (r, s) is valid under combined key P.") + "\n")
	b.WriteString(m.s.dim.Render("  No single party ever held the private key p.") + "\n")

	return b.String()
}

// ---------------------------------------------------------------------------
// Hex rendering helpers
// ---------------------------------------------------------------------------

// animHex returns a formatted hex string with rolling animation for the
// current phase, completed hex for past phases, or dots for future phases.
func (m *Model) animHex(phase int, realHex string) string {
	if realHex == "" {
		return m.s.dim.Render("(n/a)")
	}
	if m.phase > phase {
		return fmtHex(realHex)
	}
	if m.phase < phase {
		return fmtDots(len(realHex))
	}
	// Currently animating
	var sb strings.Builder
	for i := 0; i < len(realHex); i++ {
		if i > 0 && i%8 == 0 {
			sb.WriteRune(' ')
		}
		if i < m.animPos {
			sb.WriteByte(realHex[i])
		} else if i < m.animPos+3 {
			sb.WriteRune(rollingHexChar())
		} else {
			sb.WriteRune('·')
		}
	}
	return sb.String()
}

// fmtHex formats a hex string with spaces every 8 characters.
func fmtHex(hex string) string {
	var sb strings.Builder
	for i := 0; i < len(hex); i++ {
		if i > 0 && i%8 == 0 {
			sb.WriteRune(' ')
		}
		sb.WriteByte(hex[i])
	}
	return sb.String()
}

// fmtDots returns a dot string matching the display length of n hex chars.
func fmtDots(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 && i%8 == 0 {
			sb.WriteRune(' ')
		}
		sb.WriteRune('·')
	}
	return sb.String()
}

// truncHex shows at most maxBytes bytes of hex, adding "…" if truncated.
func truncHex(hex string, maxBytes int) string {
	max := maxBytes * 2
	if len(hex) <= max {
		return fmtHex(hex)
	}
	return fmtHex(hex[:max]) + "…"
}

// rollingHexChar returns a pseudo-random hex character for animation.
func rollingHexChar() rune {
	const hexChars = "0123456789abcdef"
	return rune(hexChars[time.Now().UnixNano()%16])
}
