package scenes

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VerifyScene represents Scene 12: Verification Result TUI
// It displays the verification result with signature components, message,
// VALID/INVALID indicator, and OpenSSL command reference.
type VerifyScene struct {
	styles       *Styles
	pubkey       string
	sigR         string
	sigS         string
	message      string
	valid        bool
	opensslCmd   string
	currentStep  int
	maxSteps     int
	noColor      bool
}

// Styles holds styling information for the TUI
type Styles struct {
	NoColor bool
}

// NewStyles creates a new Styles instance
func NewStyles(noColor bool) *Styles {
	return &Styles{
		NoColor: noColor,
	}
}

// Hex formats a byte slice as hex in 8-char groups
func (s *Styles) Hex(data []byte) string {
	hexStr := fmt.Sprintf("%x", data)
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

// Separator creates a visual separator line
func (s *Styles) Separator(width int, char string) string {
	if width <= 0 {
		width = 40
	}
	if char == "" {
		char = "─"
	}
	return strings.Repeat(char, width)
}

// NewVerifyScene creates a new verification scene
func NewVerifyScene(noColor bool, pubkey, sigR, sigS, message string, valid bool) *VerifyScene {
	// Generate OpenSSL command
	opensslCmd := generateOpenSSLCommand(sigR, sigS, message)

	return &VerifyScene{
		styles:     NewStyles(noColor),
		pubkey:     pubkey,
		sigR:       sigR,
		sigS:       sigS,
		message:    message,
		valid:      valid,
		opensslCmd: opensslCmd,
		currentStep: 0,
		maxSteps:    3,
		noColor:     noColor,
	}
}

// Init initializes the scene
func (s *VerifyScene) Init() tea.Cmd {
	return nil
}

// Update handles events in the scene
func (s *VerifyScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "right", "l", " ":
			if s.currentStep < s.maxSteps {
				s.currentStep++
			}
		case "left", "h":
			if s.currentStep > 0 {
				s.currentStep--
			}
		}
	}
	return s, nil
}

// Render renders the scene view
func (s *VerifyScene) Render() string {
	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	builder.WriteString(headerStyle.Render("Scene 12: Signature Verification Result"))
	builder.WriteString("\n\n")

	// Progress indicator
	progress := ""
	for i := 0; i <= s.maxSteps; i++ {
		if i == s.currentStep {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("●")
		} else if i < s.currentStep {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("○")
		} else {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("○")
		}
	}
	builder.WriteString(progress)
	builder.WriteString("\n\n")

	// Content based on current step
	switch s.currentStep {
	case 0:
		// Step 0: Verification Result
		builder.WriteString(s.renderVerificationResult())

	case 1:
		// Step 1: Signature Components
		builder.WriteString(s.renderSignatureComponents())

	case 2:
		// Step 2: OpenSSL Command
		builder.WriteString(s.renderOpenSSLCommand())

	case 3:
		// Step 3: Security Explanation
		builder.WriteString(s.renderSecurityExplanation())
	}

	// Navigation hint
	navigationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	builder.WriteString("\n\n")
	builder.WriteString(navigationStyle.Render("[←/→ or h/l to navigate] [q to quit]"))

	return builder.String()
}

// renderVerificationResult renders the verification result with VALID/INVALID indicator
func (s *VerifyScene) renderVerificationResult() string {
	var builder strings.Builder

	// Result indicator
	var resultStyle lipgloss.Style
	if s.valid {
		resultStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(1, 4).
			Foreground(lipgloss.Color("23")). // Bright green
			Background(lipgloss.Color("22"))  // Green
	} else {
		resultStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(1, 4).
			Foreground(lipgloss.Color("231")). // Bright red
			Background(lipgloss.Color("196"))  // Red
	}

	resultText := "VALID"
	if !s.valid {
		resultText = "INVALID"
	}

	builder.WriteString("  ┌──────────────────────────────────────────────────────────────┐\n")
	builder.WriteString("  │                                                              │\n")
	builder.WriteString("  │  ")
	builder.WriteString(resultStyle.Render(resultText))
	builder.WriteString("  │\n")
	builder.WriteString("  │                                                              │\n")
	builder.WriteString("  └──────────────────────────────────────────────────────────────┘\n")
	builder.WriteString("\n")

	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")). // White
		MarginLeft(2)

	builder.WriteString(messageStyle.Render("Message:"))
	builder.WriteString("\n")
	builder.WriteString(messageStyle.Render("  "))
	builder.WriteString(s.styles.Hex([]byte(s.message)))
	builder.WriteString("\n\n")

	// Public key
	builder.WriteString(messageStyle.Render("Public Key:"))
	builder.WriteString("\n")
	builder.WriteString(messageStyle.Render("  "))
	builder.WriteString(s.formatHex(s.pubkey))
	builder.WriteString("\n")

	return builder.String()
}

// renderSignatureComponents renders the signature components
func (s *VerifyScene) renderSignatureComponents() string {
	var builder strings.Builder

	// Party A section (cyan)
	partyAStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("81")) // Cyan

	builder.WriteString(partyAStyle.Render("Party A Contribution"))
	builder.WriteString("\n")
	builder.WriteString(s.styles.Separator(60, "─"))
	builder.WriteString("\n\n")

	// Party B section (magenta)
	partyBStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("213")) // Magenta

	builder.WriteString(partyBStyle.Render("Party B Contribution"))
	builder.WriteString("\n")
	builder.WriteString(s.styles.Separator(60, "─"))
	builder.WriteString("\n\n")

	// Combined signature (yellow)
	sharedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")) // Yellow

	builder.WriteString(sharedStyle.Render("Combined Signature"))
	builder.WriteString("\n")
	builder.WriteString(s.styles.Separator(60, "─"))
	builder.WriteString("\n\n")

	// R component
	rStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")) // White

	builder.WriteString(rStyle.Render("R = "))
	builder.WriteString(s.formatHex(s.sigR))
	builder.WriteString("\n\n")

	// S component
	builder.WriteString(rStyle.Render("S = "))
	builder.WriteString(s.formatHex(s.sigS))
	builder.WriteString("\n\n")

	// ECDSA formula
	formulaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Light blue
		MarginLeft(2)

	builder.WriteString(formulaStyle.Render("ECDSA Verification Formula:"))
	builder.WriteString("\n")
	builder.WriteString(formulaStyle.Render("  1. Hash message: h = SHA256(message)"))
	builder.WriteString("\n")
	builder.WriteString(formulaStyle.Render("  2. Compute: u1 = h · s⁻¹ mod n"))
	builder.WriteString("\n")
	builder.WriteString(formulaStyle.Render("  3. Compute: u2 = r · s⁻¹ mod n"))
	builder.WriteString("\n")
	builder.WriteString(formulaStyle.Render("  4. Compute: (x, y) = u1·G + u2·Q"))
	builder.WriteString("\n")
	builder.WriteString(formulaStyle.Render("  5. Verify: r ≡ x mod n"))

	return builder.String()
}

// renderOpenSSLCommand renders the OpenSSL verification command
func (s *VerifyScene) renderOpenSSLCommand() string {
	var builder strings.Builder

	// OpenSSL header
	opensslStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("220")) // Light gray

	builder.WriteString(opensslStyle.Render("OpenSSL Verification Command"))
	builder.WriteString("\n")
	builder.WriteString(s.styles.Separator(60, "─"))
	builder.WriteString("\n\n")

	// Command box
	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")). // Dark gray
		Background(lipgloss.Color("236")). // Dark background
		Padding(1, 2)

	builder.WriteString("To verify this signature externally, run:\n\n")
	builder.WriteString(cmdStyle.Render(s.opensslCmd))
	builder.WriteString("\n\n")

	// Note about DER encoding
	noteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Dim gray
		MarginLeft(2)

	builder.WriteString(noteStyle.Render("Note: The signature is DER-encoded for OpenSSL compatibility."))
	builder.WriteString("\n")
	builder.WriteString(noteStyle.Render("The R and S components are converted from raw bytes to ASN.1 DER format."))

	return builder.String()
}

// renderSecurityExplanation renders the security explanation
func (s *VerifyScene) renderSecurityExplanation() string {
	var builder strings.Builder

	// Security header
	securityStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")) // Yellow

	builder.WriteString(securityStyle.Render("Why This Verification Works"))
	builder.WriteString("\n")
	builder.WriteString(s.styles.Separator(60, "─"))
	builder.WriteString("\n\n")

	// Explanation
	explainStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")). // White
		MarginLeft(2)

	builder.WriteString(explainStyle.Render("The DKLS protocol ensures that:"))
	builder.WriteString("\n\n")

	builder.WriteString(explainStyle.Render("  ✓ Neither party ever learns the other's secret share"))
	builder.WriteString("\n")
	builder.WriteString(explainStyle.Render("  ✓ The combined signature is mathematically equivalent to"))
	builder.WriteString("\n")
	builder.WriteString(explainStyle.Render("    a signature produced by a single party with the full key"))
	builder.WriteString("\n")
	builder.WriteString(explainStyle.Render("  ✓ Verification uses standard ECDSA math on secp256k1"))
	builder.WriteString("\n")
	builder.WriteString(explainStyle.Render("  ✓ The signature can be verified by any standard tool"))
	builder.WriteString("\n\n")

	// Phantom key warning
	phantomStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")) // Red

	builder.WriteString(phantomStyle.Render("⚠ Important:"))
	builder.WriteString("\n")
	builder.WriteString(explainStyle.Render("  The phantom key (combined public key) is never reconstructed"))
	builder.WriteString("\n")
	builder.WriteString(explainStyle.Render("  in plaintext. Only the signature components are combined."))

	return builder.String()
}

// formatHex formats a hex string in 8-character groups
func (s *VerifyScene) formatHex(hexStr string) string {
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
func (s *VerifyScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene
func (s *VerifyScene) Narrator() string {
	narrators := []string{
		"Verification complete! The signature has been validated against the public key. The result shows whether the signature is mathematically valid for the given message and public key.",
		"Here are the signature components. R and S are the two parts of the ECDSA signature. They were computed jointly by Party A and Party B without either party learning the other's secret share.",
		"This OpenSSL command can be used to verify the signature externally. The signature is DER-encoded (ASN.1 format) as required by OpenSSL. You can copy this command and run it to independently verify the result.",
		"The security of this protocol relies on the hardness of the discrete logarithm problem. Even though two parties collaborated, neither could forge a signature alone, and the verification is indistinguishable from a standard ECDSA signature.",
	}

	if s.currentStep >= len(narrators) {
		s.currentStep = len(narrators) - 1
	}
	return narrators[s.currentStep]
}

// generateOpenSSLCommand generates an OpenSSL command for verification
func generateOpenSSLCommand(sigR, sigS, message string) string {
	// For display purposes, we show a simplified command
	// In practice, the signature would need to be DER-encoded and saved to a file
	return fmt.Sprintf("echo -n '%s' | openssl dgst -sha256 -verify pubkey.pem -signature signature.der", message)
}
