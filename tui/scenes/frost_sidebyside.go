package scenes

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Scene represents the FROST side-by-side comparison scene (Scene 17)
// It displays the FROST ceremony structure alongside the DKLS ceremony,
// demonstrating how FROST collapses complexity.
type Scene struct {
	currentStep int
	maxSteps    int
}

// NewScene creates a new FROST side-by-side scene
func NewScene() *Scene {
	return &Scene{
		currentStep: 0,
		maxSteps:    6,
	}
}

// Init initializes the scene
func (s *Scene) Init() tea.Cmd {
	return nil
}

// Update handles events in the scene
func (s *Scene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *Scene) Render() string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("226")) // Yellow

	frostStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("202")) // Magenta
	dklsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")) // Cyan
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Dark gray
	stepStyle := lipgloss.NewStyle().
		Bold(true).
		Margin(0, 1)

	// Define the ceremony steps for comparison
	frostSteps := []string{
		"1. Key Generation\n   - Generate secret shares\n   - Distribute via DKG\n   - Verify shares",
		"2. Nonce Generation\n   - Each party generates nonce\n   - Broadcast nonce commitment\n   - No OT required",
		"3. Partial Signatures\n   - Compute partial sig\n   - No MtA needed\n   - Direct Schnorr partial",
		"4. Aggregation\n   - Sum partial signatures\n   - Simple addition\n   - No complex combination",
		"5. Verification\n   - Verify aggregated sig\n   - Standard Schnorr check\n   - Single verification",
		"6. Result\n   - Valid Schnorr signature\n   - Compact and efficient\n   - No phantom key needed",
	}

	dklsSteps := []string{
		"1. Key Generation\n   - Generate secret shares\n   - Distribute via DKG\n   - Verify shares",
		"2. Nonce Generation\n   - Generate nonce shares\n   - OT for privacy\n   - Complex coordination",
		"3. Partial Signatures\n   - Compute partial sig\n   - MtA for k⁻¹·x term\n   - ECDSA complexity",
		"4. Aggregation\n   - Combine partial sigs\n   - Complex formula\n   - Phantom key derivation",
		"5. Verification\n   - Verify ECDSA signature\n   - Standard ECDSA check\n   - Compatible with existing",
		"6. Result\n   - Valid ECDSA signature\n   - Compatible with Bitcoin/Ethereum\n   - Legacy infrastructure",
	}

	// Build the view
	var builder string

	// Header
	builder += headerStyle.Render("FROST vs DKLS: Side-by-Side Comparison") + "\n\n"

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
	builder += progress + "\n\n"

	// Column headers
	builder += frostStyle.Render("FROST (Schnorr)") + "  " +
		separatorStyle.Render("│") + "  " +
		dklsStyle.Render("DKLS (ECDSA)") + "\n"
	builder += separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"

	// Current step
	currentFrost := frostStyle.Render(frostSteps[s.currentStep])
	currentDkls := dklsStyle.Render(dklsSteps[s.currentStep])

	// Calculate column widths
	frostWidth := 30
	dklsWidth := 30

	// Split lines for side-by-side display
	frostLines := splitLines(currentFrost, frostWidth)
	dklsLines := splitLines(currentDkls, dklsWidth)

	maxLines := len(frostLines)
	if len(dklsLines) > maxLines {
		maxLines = len(dklsLines)
	}

	for i := 0; i < maxLines; i++ {
		frostLine := ""
		dklsLine := ""
		if i < len(frostLines) {
			frostLine = frostLines[i]
		}
		if i < len(dklsLines) {
			dklsLine = dklsLines[i]
		}
		builder += fmt.Sprintf("%-*s  %s  %-*s\n", frostWidth, frostLine, separatorStyle.Render("│"), dklsWidth, dklsLine)
	}

	builder += "\n" + separatorStyle.Render("────────────────────────────────────────────────────────────────────────────────") + "\n\n"

	// Key differences summary
	builder += stepStyle.Render("Key Differences:") + "\n\n"
	builder += frostStyle.Render("FROST Advantages:") + "\n"
	builder += "  • Simpler partial signature computation\n"
	builder += "  • No Oblivious Transfer required\n"
	builder += "  • No Multiplication-in-the-Exponent (MtA)\n"
	builder += "  • Direct signature aggregation\n"
	builder += "\n" + dklsStyle.Render("DKLS Advantages:") + "\n"
	builder += "  • ECDSA compatibility (Bitcoin, Ethereum)\n"
	builder += "  • Works with existing infrastructure\n"
	builder += "  • No protocol migration needed\n"

	builder += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("[←/→ or h/l to navigate] [q to quit]")

	return builder
}

// View renders the scene view (required by tea.Model interface)
func (s *Scene) View() string {
	return s.Render()
}

// splitLines splits text into lines of approximately max width
func splitLines(text string, max int) []string {
	var lines []string
	currentLine := ""
	
	for _, char := range text {
		if len(currentLine)+1 > max {
			lines = append(lines, currentLine)
			currentLine = ""
		}
		currentLine += string(char)
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return lines
}

// Narrator returns the narrator text for this scene
func (s *Scene) Narrator() string {
	narrators := []string{
		"Welcome to the side-by-side comparison. Here we see how FROST and DKLS ceremonies differ in structure and complexity.",
		"Step 1: Key Generation. Both protocols start similarly with distributed key generation. The DKG phase is comparable in both.",
		"Step 2: Nonce Generation. FROST uses simple nonce commitments, while DKLS requires Oblivious Transfer for privacy - adding significant complexity.",
		"Step 3: Partial Signatures. FROST computes simple Schnorr partials. DKLS must handle the k⁻¹·x term via MtA - the most complex part of ECDSA threshold signing.",
		"Step 4: Aggregation. FROST simply adds partial signatures. DKLS requires complex combination formulas and phantom key derivation.",
		"Step 5: Verification. Both produce verifiable signatures, but FROST uses Schnorr verification while DKLS produces standard ECDSA.",
		"Step 6: The Result. FROST gives you a clean Schnorr signature. DKLS gives you ECDSA compatibility - crucial for Bitcoin and Ethereum integration.",
	}
	
	if s.currentStep >= len(narrators) {
		s.currentStep = len(narrators) - 1
	}
	return narrators[s.currentStep]
}
