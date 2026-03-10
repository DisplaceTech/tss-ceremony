package tests

import (
	"strings"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/tui"
)

// TestRendererStripANSI_PlainTextUnchanged verifies plain text passes through unchanged.
func TestRendererStripANSI_PlainTextUnchanged(t *testing.T) {
	input := "Hello, World!"
	got := tui.StripANSI(input)
	if got != input {
		t.Errorf("StripANSI altered plain text: got %q, want %q", got, input)
	}
}

// TestRendererStripANSI_RemovesForegroundColor verifies cyan color codes are stripped.
func TestRendererStripANSI_RemovesForegroundColor(t *testing.T) {
	// Cyan = Party A color
	input := "\033[36mParty A\033[0m"
	got := tui.StripANSI(input)
	if got != "Party A" {
		t.Errorf("StripANSI: got %q, want %q", got, "Party A")
	}
	if strings.Contains(got, "\033[") {
		t.Error("StripANSI left ANSI escape codes in output")
	}
}

// TestRendererStripANSI_RemovesMagentaColor verifies magenta color codes are stripped.
func TestRendererStripANSI_RemovesMagentaColor(t *testing.T) {
	// Magenta = Party B color
	input := "\033[35mParty B\033[0m"
	got := tui.StripANSI(input)
	if got != "Party B" {
		t.Errorf("StripANSI: got %q, want %q", got, "Party B")
	}
}

// TestRendererStripANSI_RemovesYellowColor verifies yellow (shared) color codes are stripped.
func TestRendererStripANSI_RemovesYellowColor(t *testing.T) {
	// Yellow = Shared value color
	input := "\033[33mShared Key\033[0m"
	got := tui.StripANSI(input)
	if got != "Shared Key" {
		t.Errorf("StripANSI: got %q, want %q", got, "Shared Key")
	}
}

// TestRendererStripANSI_RemovesRedColor verifies red (phantom key) color codes are stripped.
func TestRendererStripANSI_RemovesRedColor(t *testing.T) {
	// Red = Phantom key color
	input := "\033[31mPhantom Key\033[0m"
	got := tui.StripANSI(input)
	if got != "Phantom Key" {
		t.Errorf("StripANSI: got %q, want %q", got, "Phantom Key")
	}
}

// TestRendererStripANSI_RemovesBold verifies bold codes are stripped.
func TestRendererStripANSI_RemovesBold(t *testing.T) {
	input := "\033[1mBold Title\033[0m"
	got := tui.StripANSI(input)
	if got != "Bold Title" {
		t.Errorf("StripANSI: got %q, want %q", got, "Bold Title")
	}
}

// TestRendererStripANSI_Remove256Color verifies 256-color codes are stripped.
func TestRendererStripANSI_Remove256Color(t *testing.T) {
	input := "\033[38;5;226mYellow256\033[0m"
	got := tui.StripANSI(input)
	if got != "Yellow256" {
		t.Errorf("StripANSI: got %q, want %q", got, "Yellow256")
	}
}

// TestRendererStripANSI_EmptyString verifies empty string is handled.
func TestRendererStripANSI_EmptyString(t *testing.T) {
	got := tui.StripANSI("")
	if got != "" {
		t.Errorf("StripANSI(\"\") = %q, want %q", got, "")
	}
}

// TestRendererStripANSI_MultipleCodesInLine verifies multiple codes on one line.
func TestRendererStripANSI_MultipleCodesInLine(t *testing.T) {
	input := "\033[1m\033[36mParty A\033[0m \033[35mParty B\033[0m"
	got := tui.StripANSI(input)
	if got != "Party A Party B" {
		t.Errorf("StripANSI: got %q, want %q", got, "Party A Party B")
	}
}

// TestRendererNoColor_DisablesANSI verifies Renderer with noColor=true strips ANSI.
func TestRendererNoColor_DisablesANSI(t *testing.T) {
	r := tui.NewRenderer(true)
	input := "\033[36mParty A\033[0m"
	got := r.Render(input)
	if strings.Contains(got, "\033[") {
		t.Errorf("Renderer(noColor=true).Render left ANSI codes: %q", got)
	}
	if got != "Party A" {
		t.Errorf("Renderer(noColor=true).Render: got %q, want %q", got, "Party A")
	}
}

// TestRendererWithColor_PreservesANSI verifies Renderer with noColor=false preserves ANSI.
func TestRendererWithColor_PreservesANSI(t *testing.T) {
	r := tui.NewRenderer(false)
	input := "\033[36mParty A\033[0m"
	got := r.Render(input)
	if got != input {
		t.Errorf("Renderer(noColor=false).Render modified string: got %q, want %q", got, input)
	}
}

// TestRendererNoColor_RenderLines_StripsAllLines verifies RenderLines strips ANSI from all lines.
func TestRendererNoColor_RenderLines_StripsAllLines(t *testing.T) {
	r := tui.NewRenderer(true)
	input := "\033[36mLine1\033[0m\n\033[35mLine2\033[0m\n\033[33mLine3\033[0m"
	got := r.RenderLines(input)
	if strings.Contains(got, "\033[") {
		t.Errorf("RenderLines left ANSI codes: %q", got)
	}
	want := "Line1\nLine2\nLine3"
	if got != want {
		t.Errorf("RenderLines: got %q, want %q", got, want)
	}
}

// TestRendererWithColor_RenderLines_PreservesANSI verifies RenderLines preserves ANSI when color enabled.
func TestRendererWithColor_RenderLines_PreservesANSI(t *testing.T) {
	r := tui.NewRenderer(false)
	input := "\033[36mLine1\033[0m\n\033[35mLine2\033[0m"
	got := r.RenderLines(input)
	if got != input {
		t.Errorf("RenderLines(noColor=false) modified string: got %q, want %q", got, input)
	}
}

// TestRendererNoColor_PlainTextUnchanged verifies plain text is unaffected by stripping.
func TestRendererNoColor_PlainTextUnchanged(t *testing.T) {
	r := tui.NewRenderer(true)
	input := "Party A | Party B | Shared"
	got := r.Render(input)
	if got != input {
		t.Errorf("Renderer(noColor=true).Render altered plain text: got %q, want %q", got, input)
	}
}

// TestRendererPartyDistinction_PlainText verifies party labels remain readable after strip.
func TestRendererPartyDistinction_PlainText(t *testing.T) {
	r := tui.NewRenderer(true)
	// Simulate a TUI output that uses colors to distinguish parties
	partyALine := "\033[36m[Party A] sk_a = 0xABCD\033[0m"
	partyBLine := "\033[35m[Party B] sk_b = 0xEF01\033[0m"

	gotA := r.Render(partyALine)
	gotB := r.Render(partyBLine)

	if !strings.Contains(gotA, "[Party A]") {
		t.Errorf("Party A label missing after strip: %q", gotA)
	}
	if !strings.Contains(gotB, "[Party B]") {
		t.Errorf("Party B label missing after strip: %q", gotB)
	}
	// No ANSI in either
	if strings.Contains(gotA, "\033[") || strings.Contains(gotB, "\033[") {
		t.Error("ANSI escape codes remain after strip")
	}
}
