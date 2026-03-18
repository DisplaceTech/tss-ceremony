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
	phaseDERMsg          // show message hex encoding (instant)
	phaseDERKey          // show public key DER breakdown (instant)
	phaseDERSig          // show signature DER breakdown (instant)
	phaseDERCmd          // show openssl command (instant)
	phaseDone            // DKLS animation complete
	phaseFrostIntro      // equation comparison
	phaseFrostKeygen     // FROST keygen values
	phaseFrostSign       // FROST signing (nonces + challenge + partials)
	phaseFrostResult     // FROST signature + verify
	phaseFrostSummary    // protocol complexity comparison
	phaseAllDone         // comparison complete
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
			} else if m.phase >= phaseFrostIntro && m.phase < phaseAllDone {
				m.phase++
			}
			return m, m.tick()
		case "c":
			if m.phase == phaseDone && m.data.FrostCombinedPubHex != "" {
				m.phase = phaseFrostIntro
			}
			return m, nil
		}

	case tickMsg:
		if m.paused {
			return m, nil
		}
		if m.phase >= phaseDone {
			if m.phase == phaseDone && m.config.AutoQuit {
				if m.waitTicks <= 0 {
					return m, tea.Quit
				}
				m.waitTicks--
				return m, m.tick()
			}
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
		if m.phase == phaseDone && m.config.AutoQuit {
			m.waitTicks = 100 // ~3s pause before auto-quit
		} else {
			m.waitTicks = 5
		}
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
	if m.phase >= phaseAllDone {
		return d.Render("  comparison complete · [q] quit")
	}
	if m.phase >= phaseFrostIntro {
		return d.Render("  [enter] next  [q] quit")
	}
	if m.phase >= phaseDone {
		if m.data.FrostCombinedPubHex != "" {
			return d.Render("  ceremony complete · [c] compare to FROST  [q] quit")
		}
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
	b.WriteString("    " + m.animHex(phasePubA, m.data.PartyAPubHex))
	b.WriteString(m.s.dim.Render("  Party A public") + "\n")

	if m.phase <= phasePubA {
		return b.String()
	}

	b.WriteString(m.s.magenta.Render("  B") + m.s.dim.Render(" = b·G") + "\n")
	b.WriteString("    " + m.animHex(phasePubB, m.data.PartyBPubHex))
	b.WriteString(m.s.dim.Render("  Party B public") + "\n")

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
	b.WriteString(m.s.cyan.Render("  k_a ") + m.animHex(phaseNonceA, m.data.NonceAHex))
	b.WriteString(m.s.dim.Render("  random") + "\n")

	if m.phase <= phaseNonceA {
		return b.String()
	}

	b.WriteString(m.s.magenta.Render("  k_b ") + m.animHex(phaseNonceB, m.data.NonceBHex))
	b.WriteString(m.s.dim.Render("  random") + "\n")

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
		fmt.Sprintf("x₀ = %s", truncHex(m.data.OTInput0Hex, 16)) +
		m.s.dim.Render("  random") + "\n")
	b.WriteString(m.s.dim.Render("            ") +
		fmt.Sprintf("x₁ = %s", truncHex(m.data.OTInput1Hex, 16)) +
		m.s.dim.Render("  random") + "\n")
	b.WriteString(m.s.dim.Render("  receiver: ") +
		fmt.Sprintf("c  = %d", m.data.OTChoiceBit) + "\n")
	b.WriteString(m.s.dim.Render("            ") +
		fmt.Sprintf("x_c = %s", truncHex(m.data.OTOutputHex, 16)) + "\n")

	if m.phase <= phaseOT {
		return b.String()
	}

	// === MtA ===
	b.WriteString("\n" + m.section("Multiplicative to Additive") + "\n")
	b.WriteString(m.s.dim.Render("  product = k_a · k_b mod n") + "\n")
	b.WriteString(m.s.dim.Render("  β = random scalar in [1, n-1]") + "\n")
	b.WriteString(m.s.dim.Render("  α = product - β mod n") + "\n")
	b.WriteString(m.s.dim.Render("  so α + β ≡ k_a · k_b  (mod n)") + "\n")
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

	if m.phase <= phaseVerify || m.data.OpenSSLVerify == "" {
		return b.String()
	}

	b.WriteString("\n" + m.section("Verify with OpenSSL") + "\n")
	b.WriteString(m.s.dim.Render("  OpenSSL needs DER-encoded inputs. Here's how our values map:") + "\n")
	b.WriteString("\n")

	// Message
	b.WriteString(m.s.dim.Render("  message: ") + m.s.green.Render(`"`+m.data.MessageText+`"`) + "\n")

	if m.phase <= phaseDERMsg {
		return b.String()
	}

	// Public key DER breakdown
	b.WriteString("\n")
	b.WriteString(m.s.dim.Render("  public key (DER):") + "\n")
	b.WriteString(m.s.dim.Render("    3056 3010                          ") + m.s.dim.Render("SEQUENCE { SEQUENCE {") + "\n")
	b.WriteString(m.s.dim.Render("      0607 2a8648ce3d0201              ") + m.s.dim.Render("OID ecPublicKey") + "\n")
	b.WriteString(m.s.dim.Render("      0605 2b8104000a                  ") + m.s.dim.Render("OID secp256k1 }") + "\n")
	b.WriteString(m.s.dim.Render("    034200 04                          ") + m.s.dim.Render("BIT STRING, uncompressed point") + "\n")
	b.WriteString("    " + m.s.cyan.Render("P.x  ") + m.s.cyan.Render(fmtHex(m.data.CombinedPubHex)) + "\n")
	b.WriteString("    " + m.s.magenta.Render("P.y  ") + m.s.magenta.Render(fmtHex(m.data.PubKeyYHex)) + "\n")

	if m.phase <= phaseDERKey {
		return b.String()
	}

	// Signature DER breakdown
	b.WriteString("\n")
	b.WriteString(m.s.dim.Render("  signature (DER):") + "\n")
	b.WriteString(m.s.dim.Render("    30 <len>                            ") + m.s.dim.Render("SEQUENCE {") + "\n")
	b.WriteString(m.s.dim.Render("      02 <len>                          ") + m.s.dim.Render("INTEGER r") + "\n")
	b.WriteString("      " + m.s.yellow.Render("r  ") + m.s.yellow.Render(fmtHex(m.data.SignatureRHex)) + "\n")
	b.WriteString(m.s.dim.Render("      02 <len>                          ") + m.s.dim.Render("INTEGER s }") + "\n")
	b.WriteString("      " + m.s.yellow.Render("s  ") + m.s.yellow.Render(fmtHex(m.data.SignatureSHex)) + "\n")

	if m.phase <= phaseDERSig {
		return b.String()
	}

	// Colorized openssl command
	b.WriteString("\n")
	b.WriteString(m.renderColorizedCmd())

	if m.phase < phaseFrostIntro {
		return b.String()
	}

	// === FROST Comparison ===
	b.WriteString("\n\n")
	b.WriteString(m.s.bold.Render("  FROST (Schnorr) Comparison"))
	b.WriteString(m.s.dim.Render(" · same keys, same nonces, simpler math") + "\n")
	b.WriteString(m.s.dim.Render("  "+strings.Repeat("═", w-4)) + "\n")

	b.WriteString("\n" + m.section("The Key Insight") + "\n")
	b.WriteString(m.s.dim.Render("  ECDSA:   s = k") + m.s.red.Render("⁻¹") + m.s.dim.Render("·(z + r·x)") + "\n")
	b.WriteString(m.s.dim.Render("                 ↑ this inverse breaks linearity") + "\n")
	b.WriteString(m.s.dim.Render("                   → needs OT + MtA to split") + "\n")
	b.WriteString("\n")
	b.WriteString(m.s.dim.Render("  Schnorr: s = k + e·x") + "\n")
	b.WriteString(m.s.green.Render("               ↑ pure addition — threshold-friendly!") + "\n")

	if m.phase <= phaseFrostIntro {
		return b.String()
	}

	// FROST Keygen — same keys as DKLS
	b.WriteString("\n" + m.section("FROST Key Generation") + "\n")
	b.WriteString(m.s.dim.Render("  Same keys as DKLS — direct comparison:") + "\n")
	b.WriteString(m.s.cyan.Render("  a ") + fmtHex(m.data.FrostPartyASecretHex))
	b.WriteString(m.s.dim.Render("  ← same Party A secret") + "\n")
	b.WriteString(m.s.magenta.Render("  b ") + fmtHex(m.data.FrostPartyBSecretHex))
	b.WriteString(m.s.dim.Render("  ← same Party B secret") + "\n")
	b.WriteString(m.s.yellow.Render("  P") + m.s.dim.Render(" = a·G + b·G = ") + fmtHex(m.data.FrostCombinedPubHex))
	b.WriteString(m.s.dim.Render("  ← identical combined key") + "\n")

	if m.phase <= phaseFrostKeygen {
		return b.String()
	}

	// FROST Signing — same nonces as DKLS
	b.WriteString("\n" + m.section("FROST Signing") + "\n")
	b.WriteString(m.s.dim.Render("  Same nonces as DKLS:") + "\n")
	b.WriteString(m.s.cyan.Render("  k_a ") + fmtHex(m.data.FrostNonceAHex))
	b.WriteString(m.s.dim.Render("  ← same") + "\n")
	b.WriteString(m.s.magenta.Render("  k_b ") + fmtHex(m.data.FrostNonceBHex))
	b.WriteString(m.s.dim.Render("  ← same") + "\n")
	b.WriteString(m.s.yellow.Render("  R  ") + m.s.dim.Render(" = R_a + R_b"))
	b.WriteString(m.s.dim.Render("                                          ← identical R point") + "\n")
	b.WriteString("\n")
	b.WriteString(m.s.dim.Render("  Challenge") + m.s.yellow.Render(" (differs from ECDSA's z = SHA-256(m))") + m.s.dim.Render(":") + "\n")
	b.WriteString(m.s.dim.Render("  e = H(R || m || P) = ") + fmtHex(m.data.FrostChallengeHex) + "\n")
	b.WriteString("\n")
	b.WriteString(m.s.dim.Render("  Partial signatures") + m.s.green.Render("  ← no OT, no MtA — just addition!") + "\n")
	b.WriteString(m.s.cyan.Render("  s_a") + m.s.dim.Render(" = k_a + e·a = ") + fmtHex(m.data.FrostPartialSigAHex) + "\n")
	b.WriteString(m.s.magenta.Render("  s_b") + m.s.dim.Render(" = k_b + e·b = ") + fmtHex(m.data.FrostPartialSigBHex) + "\n")

	if m.phase <= phaseFrostSign {
		return b.String()
	}

	// FROST Signature Result
	b.WriteString("\n" + m.section("FROST Signature") + "\n")
	b.WriteString(m.s.yellow.Render("  R ") + fmtHex(m.data.FrostSignatureRHex) + "\n")
	b.WriteString(m.s.yellow.Render("  s ") + fmtHex(m.data.FrostSignatureSHex) + "\n")
	b.WriteString("\n")
	if m.data.FrostValid {
		b.WriteString("  Schnorr.Verify(P, m, R, s) → " + m.s.green.Render("✓ VALID") + "\n")
	} else {
		b.WriteString("  Schnorr.Verify(P, m, R, s) → " + m.s.red.Render("✗ INVALID") + "\n")
	}

	if m.phase <= phaseFrostResult {
		return b.String()
	}

	// Protocol Complexity Comparison
	b.WriteString("\n" + m.section("Protocol Complexity") + "\n")
	b.WriteString("  " + m.s.bold.Render("DKLS (ECDSA)") + "              " + m.s.bold.Render("FROST (Schnorr)") + "\n")
	b.WriteString(m.s.dim.Render("  ──────────────────        ──────────────────") + "\n")
	b.WriteString("  Key generation            Key generation\n")
	b.WriteString("  Nonce generation          Nonce generation\n")
	b.WriteString("  " + m.s.red.Render("Oblivious Transfer") + "        " + m.s.green.Render("── not needed ──") + "\n")
	b.WriteString("  " + m.s.red.Render("MtA conversion") + "            " + m.s.green.Render("── not needed ──") + "\n")
	b.WriteString("  Partial signatures        Partial signatures\n")
	b.WriteString("  Signature combine         Signature combine\n")
	b.WriteString("  " + m.s.red.Render("4+ rounds") + "                 " + m.s.green.Render("2 rounds") + "\n")
	b.WriteString("\n")
	b.WriteString(m.s.dim.Render("  ECDSA: Bitcoin, Ethereum, TLS — the infrastructure of today") + "\n")
	b.WriteString(m.s.dim.Render("  Schnorr: Taproot, Ed25519 — the elegant future") + "\n")
	b.WriteString(m.s.dim.Render("  Same security. Same curve. Different trade-offs.") + "\n")

	return b.String()
}

// renderColorizedCmd renders the openssl command with colored DER components.
// ANSI colors don't affect copy-paste, so the command remains valid bash.
func (m *Model) renderColorizedCmd() string {
	d := m.data
	dim := m.s.dim

	// Public key DER: fixed header (48 hex chars) + P.x (64) + P.y (64) = 176
	pubHeader := "3056301006072a8648ce3d020106052b8104000a03420004"
	colorPub := dim.Render(pubHeader) +
		m.s.cyan.Render(d.CombinedPubHex) +
		m.s.magenta.Render(d.PubKeyYHex)

	// Signature DER: find r and s within the DER hex
	colorSig := m.colorizeSigDER(d.SigDERHex)

	// Message (plaintext)
	colorMsg := m.s.green.Render(d.MessageText)

	var b strings.Builder
	b.WriteString(dim.Render("  echo -n '") + colorMsg + dim.Render("' | \\") + "\n")
	b.WriteString(dim.Render("      openssl dgst -sha256 \\") + "\n")
	b.WriteString(dim.Render("      -verify <(echo '") + colorPub + dim.Render("' | \\") + "\n")
	b.WriteString(dim.Render("        xxd -r -p | openssl ec -pubin -inform DER 2>/dev/null) \\") + "\n")
	b.WriteString(dim.Render("      -signature <(echo '") + colorSig + dim.Render("' | xxd -r -p)") + "\n")
	return b.String()
}

// colorizeSigDER highlights r and s values within a DER-encoded signature hex.
func (m *Model) colorizeSigDER(derHex string) string {
	// DER structure: 30 <len> 02 <rlen> <r_bytes> 02 <slen> <s_bytes>
	// All positions in hex characters (2 per byte)
	if len(derHex) < 8 {
		return m.s.dim.Render(derHex)
	}

	// Parse r length at offset 3 (byte index), hex position 6
	rLen := hexByte(derHex[6:8]) * 2 // length in hex chars
	rStart := 8
	rEnd := rStart + rLen

	// After r: 02 <slen> <s_bytes>
	sLenPos := rEnd + 2
	if sLenPos+2 > len(derHex) {
		return m.s.dim.Render(derHex)
	}
	sLen := hexByte(derHex[sLenPos:sLenPos+2]) * 2
	sStart := sLenPos + 2
	sEnd := sStart + sLen

	dim := m.s.dim
	var sb strings.Builder
	sb.WriteString(dim.Render(derHex[:rStart]))                        // 30<len>02<rlen>
	sb.WriteString(m.s.yellow.Render(derHex[rStart:rEnd]))             // r bytes
	sb.WriteString(dim.Render(derHex[rEnd:sStart]))                    // 02<slen>
	sb.WriteString(m.s.yellow.Render(derHex[sStart:sEnd]))             // s bytes
	if sEnd < len(derHex) {
		sb.WriteString(dim.Render(derHex[sEnd:]))
	}
	return sb.String()
}

// hexByte parses a 2-char hex string as an integer.
func hexByte(s string) int {
	var v int
	for _, c := range s {
		v *= 16
		switch {
		case c >= '0' && c <= '9':
			v += int(c - '0')
		case c >= 'a' && c <= 'f':
			v += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v += int(c-'A') + 10
		}
	}
	return v
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
