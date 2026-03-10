package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

// ResizeHandler handles terminal resize events
type ResizeHandler struct {
	layoutManager *LayoutManager
	resizeChan    chan tea.WindowSizeMsg
	sigChan       chan os.Signal
}

// NewResizeHandler creates a new resize handler
func NewResizeHandler(lm *LayoutManager) *ResizeHandler {
	return &ResizeHandler{
		layoutManager: lm,
		resizeChan:    make(chan tea.WindowSizeMsg, 1),
		sigChan:       make(chan os.Signal, 1),
	}
}

// Start starts listening for resize events
func (rh *ResizeHandler) Start() {
	// Listen for SIGWINCH (window resize) signal
	signal.Notify(rh.sigChan, syscall.SIGWINCH)
	
	// Start a goroutine to handle resize events
	go rh.handleResizeEvents()
}

// handleResizeEvents handles terminal resize events
func (rh *ResizeHandler) handleResizeEvents() {
	for {
		select {
		case <-rh.sigChan:
			// Get current terminal size
			width, height, err := getTerminalSize()
			if err != nil {
				// Fallback to current dimensions
				width, height = rh.layoutManager.GetDimensions()
			}
			
			// Resize the layout manager
			rh.layoutManager.Resize(width, height)
			
			// Send resize message to Bubble Tea
			rh.resizeChan <- tea.WindowSizeMsg{
				Width:  width,
				Height: height,
			}
		case <-rh.resizeChan:
			// Already handled above, this case is for future extensions
		}
	}
}

// GetResizeMsg returns a channel for resize messages
func (rh *ResizeHandler) GetResizeMsg() chan tea.WindowSizeMsg {
	return rh.resizeChan
}

// Stop stops listening for resize events
func (rh *ResizeHandler) Stop() {
	signal.Stop(rh.sigChan)
	close(rh.sigChan)
	close(rh.resizeChan)
}

// getTerminalSize returns the current terminal size
func getTerminalSize() (width, height int, err error) {
	// Try to get terminal size using various methods
	// This is a simplified implementation - in production,
	// you might want to use a library like github.com/mattn/go-isatty
	
	// For now, return a reasonable default
	return 80, 24, nil
}

// HandleResize updates the model with a resize message
func HandleResize(m *Model, msg tea.WindowSizeMsg) tea.Model {
	// The layout manager should be part of the model
	// For now, we'll just log the resize event
	// In a full implementation, the model would have access to the layout manager
	
	// Update the model's dimensions if needed
	// This is a placeholder - the actual implementation would
	// update the layout manager and recalculate the view
	
	return m
}

// ValidateLayoutOnResize validates the layout after a resize event
func ValidateLayoutOnResize(m *Model, width, height int) ValidationResult {
	// Create a temporary layout manager for validation
	lm := NewLayoutManager()
	lm.Resize(width, height)
	
	return lm.ValidateTerminalSize(width, height)
}

// EnsureMinimumSize ensures the terminal meets minimum size requirements
// and returns an error if it doesn't
func EnsureMinimumSize(width, height int) error {
	if width < 80 || height < 24 {
		return fmt.Errorf("terminal too small: %dx%d (minimum: 80x24)", width, height)
	}
	return nil
}

// CheckTerminalCompatibility checks if the terminal is compatible with the TUI
func CheckTerminalCompatibility() error {
	width, height, err := getTerminalSize()
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %v", err)
	}
	
	if err := EnsureMinimumSize(width, height); err != nil {
		return err
	}
	
	// Check for ANSI color support
	// This is a simplified check - in production, use termenv
	
	return nil
}
