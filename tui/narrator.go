// Package tui provides the terminal user interface for the TSS ceremony.
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// NarratorState represents the current state of the narrator.
type NarratorState int

const (
	NarratorIdle NarratorState = iota
	NarratorSpeaking
	NarratorPaused
)

// NarratorMessage represents a message to be displayed by the narrator.
type NarratorMessage struct {
	Text       string
	Timestamp  time.Time
	IsComplete bool
}

// NarratorManager handles narrator logic and message queue.
type NarratorManager struct {
	Name         string
	Avatar       string
	State        NarratorState
	CurrentMsg   NarratorMessage
	MessageQueue []NarratorMessage
	Styles       Style
	IsSpeaking   bool
}

// NewNarratorManager creates a new NarratorManager with the given name and styles.
func NewNarratorManager(name string, styles Style) *NarratorManager {
	return &NarratorManager{
		Name:   name,
		Avatar: "🎭",
		State:  NarratorIdle,
		Styles: styles,
	}
}

// QueueMessage adds a message to the narrator's queue.
func (nm *NarratorManager) QueueMessage(text string) {
	nm.MessageQueue = append(nm.MessageQueue, NarratorMessage{
		Text:      text,
		Timestamp: time.Now(),
	})
}

// NextMessage retrieves the next message from the queue.
func (nm *NarratorManager) NextMessage() (NarratorMessage, bool) {
	if len(nm.MessageQueue) == 0 {
		return NarratorMessage{}, false
	}
	msg := nm.MessageQueue[0]
	nm.MessageQueue = nm.MessageQueue[1:]
	nm.CurrentMsg = msg
	nm.State = NarratorSpeaking
	nm.IsSpeaking = true
	return msg, true
}

// CompleteMessage marks the current message as complete.
func (nm *NarratorManager) CompleteMessage() {
	if nm.State == NarratorSpeaking {
		nm.CurrentMsg.IsComplete = true
		nm.State = NarratorIdle
		nm.IsSpeaking = false
	}
}

// PauseSpeaking pauses the narrator.
func (nm *NarratorManager) PauseSpeaking() {
	if nm.State == NarratorSpeaking {
		nm.State = NarratorPaused
		nm.IsSpeaking = false
	}
}

// ResumeSpeaking resumes the narrator.
func (nm *NarratorManager) ResumeSpeaking() {
	if nm.State == NarratorPaused {
		nm.State = NarratorSpeaking
		nm.IsSpeaking = true
	}
}

// HasMessages returns true if there are messages in the queue.
func (nm *NarratorManager) HasMessages() bool {
	return len(nm.MessageQueue) > 0
}

// MessageCount returns the number of messages in the queue.
func (nm *NarratorManager) MessageCount() int {
	return len(nm.MessageQueue)
}

// ClearQueue clears all messages from the queue.
func (nm *NarratorManager) ClearQueue() {
	nm.MessageQueue = nil
	nm.CurrentMsg = NarratorMessage{}
	nm.State = NarratorIdle
	nm.IsSpeaking = false
}

// Render renders the narrator's current state.
func (nm *NarratorManager) Render() string {
	var sb strings.Builder

	if nm.State == NarratorIdle && nm.CurrentMsg.Text == "" {
		return ""
	}

	// Avatar and name
	sb.WriteString(nm.Styles.Info.Render(fmt.Sprintf("%s %s: ", nm.Avatar, nm.Name)))

	// Current message
	if nm.CurrentMsg.Text != "" {
		if nm.State == NarratorSpeaking {
			sb.WriteString(nm.Styles.Description.Render(nm.CurrentMsg.Text))
		} else if nm.State == NarratorPaused {
			sb.WriteString(nm.Styles.Warning.Render("[" + nm.CurrentMsg.Text + " (paused)]"))
		} else {
			sb.WriteString(nm.Styles.Success.Render("✓ " + nm.CurrentMsg.Text))
		}
	} else {
		sb.WriteString(nm.Styles.Info.Render("(idle)"))
	}

	sb.WriteString("\n")

	// Queue status
	if nm.HasMessages() {
		sb.WriteString(nm.Styles.Info.Render(fmt.Sprintf("  Queue: %d message(s) pending", nm.MessageCount())))
		sb.WriteString("\n")
	}

	return sb.String()
}

// NarratorTickCmd is a command for narrator timing.
type NarratorTickCmd struct{}

// NarratorTick creates a command that triggers after a duration.
func NarratorTick(duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(duration)
		return NarratorTickCmd{}
	}
}

// NarratorUpdateCmd is a command for narrator state updates.
type NarratorUpdateCmd struct {
	Action string
}

// NarratorCompleteCmd creates a command to complete the current message.
func NarratorCompleteCmd() tea.Cmd {
	return func() tea.Msg {
		return NarratorUpdateCmd{Action: "complete"}
	}
}

// NarratorNextCmd creates a command to get the next message.
func NarratorNextCmd() tea.Cmd {
	return func() tea.Msg {
		return NarratorUpdateCmd{Action: "next"}
	}
}

// UpdateNarrator handles narrator-related commands.
func UpdateNarrator(nm *NarratorManager, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NarratorUpdateCmd:
		switch msg.Action {
		case "complete":
			nm.CompleteMessage()
			if nm.HasMessages() {
				return nil, NarratorNextCmd()
			}
		case "next":
			if nm.HasMessages() {
				nm.NextMessage()
			}
		}
	case NarratorTickCmd:
		// Timer tick - can be used for auto-advancing messages
	}

	return nil, nil
}

// CreateNarratorScript creates a sequence of messages for a scene.
func CreateNarratorScript(scene Scene) []string {
	switch scene.(type) {
	case WelcomeScene:
		return []string{
			"Welcome to the TSS Ceremony guide!",
			"This interactive demonstration will show you how threshold signatures work.",
			"Let's begin our journey through the DKLS23 protocol.",
		}
	case SetupScene:
		return []string{
			"First, we need to set up the ceremony parameters.",
			"We'll define the threshold and total number of participants.",
			"For this demo, we're using a 2-of-2 threshold scheme.",
		}
	case KeyGenerationScene:
		return []string{
			"Now we generate the key shares for each participant.",
			"Each participant will receive a secret share of the private key.",
			"No single participant knows the complete private key.",
		}
	case SigningScene:
		return []string{
			"It's time to perform the threshold signature operation.",
			"Each participant signs with their key share.",
			"The partial signatures are combined to produce the final signature.",
		}
	case CompletionScene:
		return []string{
			"Congratulations! The ceremony is complete.",
			"The threshold signature has been successfully generated.",
			"Thank you for participating in this TSS demonstration.",
		}
	default:
		return []string{"An unknown scene was encountered."}
	}
}

// LoadSceneScript loads all messages for a scene into the narrator's queue.
func LoadSceneScript(nm *NarratorManager, scene Scene) {
	messages := CreateNarratorScript(scene)
	for _, msg := range messages {
		nm.QueueMessage(msg)
	}
}
