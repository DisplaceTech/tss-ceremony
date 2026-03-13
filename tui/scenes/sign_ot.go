package scenes

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SignOTScene represents Scenes 7-8: Oblivious Transfer.
// Step 0 explains the OT concept with an envelope metaphor.
// Step 1 shows OT extension with bit-by-bit accumulation and a progress bar.
// Step 2 shows the real OT values from ceremony data.
type SignOTScene struct {
	config      *Config
	styles      *Styles
	currentStep int // 0: concept, 1: extension animation, 2: real values
	maxSteps    int
	started     bool

	// Animation state for step 1
	bitProgress int  // how many bits have been processed
	totalBits   int  // total bits to process (e.g. 256)
	animTick    int  // animation tick counter
	duration    time.Duration
}

// NewSignOTScene creates a new oblivious transfer scene.
func NewSignOTScene(config *Config, styles *Styles) *SignOTScene {
	return &SignOTScene{
		config:      config,
		styles:      styles,
		currentStep: 0,
		maxSteps:    2,
		started:     false,
		totalBits:   256,
		duration:    getStepDuration(config.Speed),
	}
}

// Init initializes the scene.
func (s *SignOTScene) Init() tea.Cmd {
	s.started = true
	s.currentStep = 0
	s.bitProgress = 0
	s.animTick = 0
	if s.config.FixedMode {
		ResetFixedCounter()
	}
	return nil
}

// Update handles events in the scene.
func (s *SignOTScene) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "right", "l", " ":
			if s.currentStep < s.maxSteps {
				s.currentStep++
				if s.currentStep == 1 {
					// Start the animation for step 1
					s.bitProgress = 0
					s.animTick = 0
					return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
						return tickMsg(t)
					})
				}
			}
		case "left", "h":
			if s.currentStep > 0 {
				s.currentStep--
			}
		case "enter", "n", "j", "down":
			if s.currentStep >= s.maxSteps {
				// At last step, advance to next scene
				return s, nil
			}
			s.currentStep++
			if s.currentStep == 1 {
				s.bitProgress = 0
				s.animTick = 0
				return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
					return tickMsg(t)
				})
			}
		}

	case tickMsg:
		if s.currentStep == 1 && s.bitProgress < s.totalBits {
			// Advance bit progress in chunks for smoother animation
			s.bitProgress += 4
			if s.bitProgress > s.totalBits {
				s.bitProgress = s.totalBits
			}
			s.animTick++
			return s, tea.Tick(s.duration, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
	}
	return s, nil
}

// Render renders the scene view.
func (s *SignOTScene) Render() string {
	var builder strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")).
		MarginBottom(1)

	builder.WriteString(headerStyle.Render("Oblivious Transfer (OT)") + "\n\n")

	// Progress dots
	progress := ""
	for i := 0; i <= s.maxSteps; i++ {
		if i == s.currentStep {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("\u25cf")
		} else if i < s.currentStep {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("\u25cb")
		} else {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\u25cb")
		}
	}
	builder.WriteString(progress + "\n\n")

	switch s.currentStep {
	case 0:
		builder.WriteString(s.renderConcept())
	case 1:
		builder.WriteString(s.renderExtension())
	case 2:
		builder.WriteString(s.renderRealValues())
	}

	// Navigation hint
	builder.WriteString("\n\n")
	navStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	builder.WriteString(navStyle.Render("[\u2190/\u2192 or h/l to navigate] [q to quit]"))

	return builder.String()
}

// renderConcept renders step 0: OT concept with envelope metaphor.
func (s *SignOTScene) renderConcept() string {
	var builder strings.Builder

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	builder.WriteString(labelStyle.Render("What is Oblivious Transfer?") + "\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("\u2500", 50)) + "\n\n")

	explainStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	builder.WriteString(explainStyle.Render("OT lets a receiver choose ONE of two values from a sender,") + "\n")
	builder.WriteString(explainStyle.Render("without the sender learning which value was chosen.") + "\n\n")

	// ASCII envelope diagram
	partyAStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("81")) // Cyan
	partyBStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("213")) // Magenta
	sharedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")) // Yellow

	builder.WriteString(partyAStyle.Render("  Sender (Party A)") + "\n")
	builder.WriteString("  \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510   \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510\n")
	builder.WriteString("  \u2502 " + sharedStyle.Render("Envelope 0") + " \u2502   \u2502 " + sharedStyle.Render("Envelope 1") + " \u2502\n")
	builder.WriteString("  \u2502  m\u2080 = ???  \u2502   \u2502  m\u2081 = ???  \u2502\n")
	builder.WriteString("  \u2502   [LOCKED]  \u2502   \u2502   [LOCKED]  \u2502\n")
	builder.WriteString("  \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518   \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518\n")
	builder.WriteString("         \\         /\n")
	builder.WriteString("          \\       /\n")
	builder.WriteString("           v     v\n")
	builder.WriteString(partyBStyle.Render("     Receiver (Party B)") + "\n")
	builder.WriteString("     Chooses bit b \u2208 {0, 1}\n")
	builder.WriteString("     Learns m_b, nothing else\n\n")

	builder.WriteString(separatorStyle.Render(strings.Repeat("\u2500", 50)) + "\n\n")

	builder.WriteString(explainStyle.Render("Key properties:") + "\n")
	builder.WriteString(explainStyle.Render("  \u2022 Sender does NOT learn which envelope was opened") + "\n")
	builder.WriteString(explainStyle.Render("  \u2022 Receiver learns ONLY the chosen value") + "\n")
	builder.WriteString(explainStyle.Render("  \u2022 This is the building block for MtA (Multiply-to-Add)") + "\n")

	return builder.String()
}

// renderExtension renders step 1: OT extension with progress bar.
func (s *SignOTScene) renderExtension() string {
	var builder strings.Builder

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	builder.WriteString(labelStyle.Render("OT Extension: Bit-by-Bit Transfer") + "\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("\u2500", 50)) + "\n\n")

	explainStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	builder.WriteString(explainStyle.Render("OT extension converts a small number of base OTs") + "\n")
	builder.WriteString(explainStyle.Render("into many OTs efficiently, one bit at a time.") + "\n\n")

	// Progress bar
	barWidth := 40
	filled := 0
	if s.totalBits > 0 {
		filled = (s.bitProgress * barWidth) / s.totalBits
	}
	if filled > barWidth {
		filled = barWidth
	}

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")) // Yellow
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	builder.WriteString("  Progress: [")
	builder.WriteString(progressStyle.Render(strings.Repeat("\u2588", filled)))
	builder.WriteString(dimStyle.Render(strings.Repeat("\u2591", barWidth-filled)))
	builder.WriteString(fmt.Sprintf("] %d/%d bits\n\n", s.bitProgress, s.totalBits))

	// Show accumulating bits
	bitsToShow := s.bitProgress
	if bitsToShow > 64 {
		bitsToShow = 64 // Show first 64 bits for display
	}
	builder.WriteString(labelStyle.Render("  Transferred bits:") + "\n  ")
	for i := 0; i < bitsToShow; i++ {
		if i > 0 && i%8 == 0 {
			builder.WriteString(" ")
		}
		bit := pickInt(s.config.FixedMode, 2)
		if bit == 0 {
			builder.WriteString("0")
		} else {
			builder.WriteString("1")
		}
	}
	if s.bitProgress > 64 {
		builder.WriteString(" ...")
	}
	builder.WriteString("\n\n")

	// Status
	if s.bitProgress >= s.totalBits {
		doneStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226"))
		builder.WriteString(doneStyle.Render("  OT extension complete!") + "\n")
	} else {
		builder.WriteString(explainStyle.Render(fmt.Sprintf("  Processing bit %d of %d...", s.bitProgress, s.totalBits)) + "\n")
	}

	return builder.String()
}

// renderRealValues renders step 2: real OT values from ceremony data.
func (s *SignOTScene) renderRealValues() string {
	var builder strings.Builder

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	builder.WriteString(labelStyle.Render("OT Results") + "\n")
	builder.WriteString(separatorStyle.Render(strings.Repeat("\u2500", 50)) + "\n\n")

	partyAStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("81")) // Cyan
	partyBStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("213")) // Magenta
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Sender's inputs
	builder.WriteString(partyAStyle.Render("Sender (Party A) inputs:") + "\n\n")

	input0 := "(not available)"
	input1 := "(not available)"
	if s.config.Ceremony != nil {
		if s.config.Ceremony.OTInput0Hex != "" {
			input0 = formatHexGroups(s.config.Ceremony.OTInput0Hex)
		}
		if s.config.Ceremony.OTInput1Hex != "" {
			input1 = formatHexGroups(s.config.Ceremony.OTInput1Hex)
		}
	}

	builder.WriteString(labelStyle.Render("  m\u2080 = ") + valueStyle.Render(input0) + "\n")
	builder.WriteString(labelStyle.Render("  m\u2081 = ") + valueStyle.Render(input1) + "\n\n")

	// Receiver's choice
	builder.WriteString(partyBStyle.Render("Receiver (Party B) choice:") + "\n\n")

	choiceBit := 0
	if s.config.Ceremony != nil {
		choiceBit = s.config.Ceremony.OTChoiceBit
	}

	choiceStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226"))
	builder.WriteString(labelStyle.Render("  Choice bit: ") + choiceStyle.Render(fmt.Sprintf("%d", choiceBit)) + "\n\n")

	// Output
	builder.WriteString(separatorStyle.Render(strings.Repeat("\u2500", 50)) + "\n\n")

	sharedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226"))
	builder.WriteString(sharedStyle.Render("OT Output:") + "\n\n")

	output := "(not available)"
	if s.config.Ceremony != nil && s.config.Ceremony.OTOutputHex != "" {
		output = formatHexGroups(s.config.Ceremony.OTOutputHex)
	}

	builder.WriteString(labelStyle.Render(fmt.Sprintf("  m_%d = ", choiceBit)) + valueStyle.Render(output) + "\n\n")

	// Explanation
	explainStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	builder.WriteString(explainStyle.Render("Party B received m_b without Party A learning b.") + "\n")
	builder.WriteString(explainStyle.Render("This OT output feeds into the MtA (Multiply-to-Add) protocol.") + "\n")

	return builder.String()
}

// View renders the scene view (required by tea.Model interface).
func (s *SignOTScene) View() string {
	return s.Render()
}

// Narrator returns the narrator text for this scene.
func (s *SignOTScene) Narrator() string {
	switch s.currentStep {
	case 0:
		return "Oblivious Transfer (OT) is a fundamental cryptographic primitive. It allows Party B to receive exactly one of two values from Party A, without Party A learning which value was chosen. Think of it as two locked envelopes -- the receiver gets a key to only one."
	case 1:
		return "OT Extension is an efficiency technique. Instead of running expensive public-key OT for every bit, we use a small number of base OTs and then extend them using symmetric cryptography. This processes all 256 bits needed for the MtA protocol."
	default:
		return "Here are the actual OT values from the ceremony. Party A provided two inputs (m0, m1), Party B chose one with their choice bit, and received the corresponding value. Party A never learns which value was selected. This output will be used in the multiplicative-to-additive share conversion."
	}
}
