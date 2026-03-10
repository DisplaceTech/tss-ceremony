package tui

import (
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text unchanged",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "strips foreground color",
			input: "\033[36mParty A\033[0m",
			want:  "Party A",
		},
		{
			name:  "strips bold",
			input: "\033[1mBold Text\033[0m",
			want:  "Bold Text",
		},
		{
			name:  "strips dim",
			input: "\033[2mDim Text\033[0m",
			want:  "Dim Text",
		},
		{
			name:  "strips multiple codes",
			input: "\033[1m\033[36mParty A\033[0m \033[35mParty B\033[0m",
			want:  "Party A Party B",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "256-color code",
			input: "\033[38;5;226mYellow\033[0m",
			want:  "Yellow",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StripANSI(tc.input)
			if got != tc.want {
				t.Errorf("StripANSI(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRendererNoColor(t *testing.T) {
	r := NewRenderer(true)
	input := "\033[36mParty A\033[0m"
	got := r.Render(input)
	if strings.Contains(got, "\033[") {
		t.Errorf("Renderer.Render with noColor=true returned ANSI codes: %q", got)
	}
	if got != "Party A" {
		t.Errorf("expected %q, got %q", "Party A", got)
	}
}

func TestRendererWithColor(t *testing.T) {
	r := NewRenderer(false)
	input := "\033[36mParty A\033[0m"
	got := r.Render(input)
	if got != input {
		t.Errorf("Renderer.Render with noColor=false modified string: got %q, want %q", got, input)
	}
}

func TestRendererRenderLines(t *testing.T) {
	r := NewRenderer(true)
	input := "\033[36mLine1\033[0m\n\033[35mLine2\033[0m"
	got := r.RenderLines(input)
	if strings.Contains(got, "\033[") {
		t.Errorf("RenderLines left ANSI codes: %q", got)
	}
	if got != "Line1\nLine2" {
		t.Errorf("RenderLines: got %q, want %q", got, "Line1\nLine2")
	}
}
