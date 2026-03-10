package tui

import (
	"strings"
)

// LayoutSpec defines the expected TUI layout grid and styling
// This specification ensures visual consistency across all scenes
type LayoutSpec struct {
	// Terminal requirements
	MinWidth  int `json:"min_width"`
	MinHeight int `json:"min_height"`

	// Layout grid dimensions (in characters)
	HeaderHeight int `json:"header_height"`
	FooterHeight int `json:"footer_height"`
	NarratorHeight int `json:"narrator_height"`
	ContentHeight int `json:"content_height"`

	// Column layout for Party A / Shared / Party B
	LeftColumnWidth  int `json:"left_column_width"`  // Party A
	SharedColumnWidth int `json:"shared_column_width"` // Shared state
	RightColumnWidth int `json:"right_column_width"` // Party B

	// Border specifications
	BorderType       string `json:"border_type"`       // e.g., "rounded", "single", "double"
	BorderForeground string `json:"border_foreground"` // Lipgloss color code

	// Color scheme
	PartyAColor   string `json:"party_a_color"`   // Cyan
	PartyBColor   string `json:"party_b_color"`   // Magenta
	SharedColor   string `json:"shared_color"`    // Yellow
	PhantomColor  string `json:"phantom_color"`   // Red
	NarratorColor string `json:"narrator_color"`  // Dim/Gray

	// Alignment
	ContentAlignment string `json:"content_alignment"` // "left", "center", "right"
	HexGroupSize     int    `json:"hex_group_size"`    // Hex display grouping (8 chars)
}

// DefaultLayoutSpec returns the standard layout specification
func DefaultLayoutSpec() LayoutSpec {
	return LayoutSpec{
		MinWidth:       80,
		MinHeight:      24,
		HeaderHeight:   1,
		FooterHeight:   1,
		NarratorHeight: 2,
		ContentHeight:  20,
		LeftColumnWidth:  24,
		SharedColumnWidth: 32,
		RightColumnWidth: 24,
		BorderType:       "rounded",
		BorderForeground: "241",
		PartyAColor:      "6",    // Cyan
		PartyBColor:      "5",    // Magenta
		SharedColor:      "3",    // Yellow
		PhantomColor:     "1",    // Red
		NarratorColor:    "243",  // Gray
		ContentAlignment: "left",
		HexGroupSize:     8,
	}
}

// ASCIIArtSpec defines the expected ASCII art layout for each scene
// This is used for visual verification against the specification
type ASCIIArtSpec struct {
	SceneName string `json:"scene_name"`
	Layout    string `json:"layout"` // ASCII art representation
}

// SceneLayouts contains the ASCII art specifications for all scenes
var SceneLayouts = []ASCIIArtSpec{
	{
		SceneName: "Scene 0: Title Screen",
		Layout: `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │
│                    Threshold Signature Schemes                 │
│                                                                │
│  ┌──────────────┐              ┌──────────────┐               │
│  │   PARTY A    │              │   PARTY B    │               │
│  │   (Cyan)     │              │   (Magenta)  │               │
│  └──────────────┘              └──────────────┘               │
│                                                                │
│              ┌──────────────┐                                 │
│              │   SHARED     │                                 │
│              │   (Yellow)   │                                 │
│              └──────────────┘                                 │
│                                                                │
│  Press Enter to begin the ceremony...                         │
└────────────────────────────────────────────────────────────────┘
[0/20] [Key bindings: Enter/←/→/q]`,
	},
	{
		SceneName: "Scene 1: Protocol Parameters",
		Layout: `┌────────────────────────────────────────────────────────────────┐
│ Scene 1/20 · Protocol Parameters                               │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌────────┐ │
│  │   PARTY A           │  │   SHARED STATE      │  │ PARTY B│ │
│  │   (Cyan)            │  │   (Yellow)          │  │(Magenta)││
│  ├─────────────────────┤  ├─────────────────────┤  ├────────┤ │
│  │ Curve: secp256k1    │  │ Field Order: n      │  │        │ │
│  │ Field Order: n      │  │ Generator: G        │  │        │ │
│  │ Generator: G        │  │ Public Key: P       │  │        │ │
│  │ Public Key: P       │  │ Nonce: k            │  │        │ │
│  │ Nonce: k            │  │ Share: s_i          │  │        │ │
│  │ Share: s_i          │  │ Signature: (r, s)   │  │        │ │
│  └─────────────────────┘  └─────────────────────┘  └────────┘ │
│                                                                │
│  Mode: FIXED MODE                                              │
│  Message: "Hello, World!"                                      │
│  Speed: normal                                                 │
│                                                                │
│  Narrator: These are the protocol parameters for our ceremony │
└────────────────────────────────────────────────────────────────┘
[1/20] [Key bindings: Enter/←/→/q]`,
	},
	{
		SceneName: "Scene 4: Combined Public Key",
		Layout: `┌────────────────────────────────────────────────────────────────┐
│ Scene 4/20 · Combined Public Key                               │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌────────┐ │
│  │   PARTY A           │  │   SHARED STATE      │  │ PARTY B│ │
│  │   (Cyan)            │  │   (Yellow)          │  │(Magenta)││
│  ├─────────────────────┤  ├─────────────────────┤  ├────────┤ │
│  │ Private: k_A        │  │ P_A = k_A * G       │  │        │ │
│  │ Public: P_A         │  │ P_B = k_B * G       │  │        │ │
│  │ Share: s_A          │  │ P = P_A + P_B       │  │        │ │
│  │                     │  │                     │  │        │ │
│  │                     │  │ P = 04...ABC123...  │  │        │ │
│  │                     │  │ (65 bytes)          │  │        │ │
│  └─────────────────────┘  └─────────────────────┘  └────────┘ │
│                                                                │
│  Narrator: The combined public key is the sum of both parties'│
│            public keys. This is the key that will verify      │
│            signatures produced by the threshold scheme.       │
└────────────────────────────────────────────────────────────────┘
[4/20] [Key bindings: Enter/←/→/q]`,
	},
}

// NormalizeLayout converts a raw ASCII art string into a structured grid
// This is used for verification tests
func NormalizeLayout(asciiArt string) [][]string {
	lines := strings.Split(asciiArt, "\n")
	grid := make([][]string, len(lines))

	for i, line := range lines {
		grid[i] = strings.Split(line, "")
	}

	return grid
}

// ValidateLayout checks if the layout meets minimum requirements
func (s LayoutSpec) ValidateLayout(width, height int) bool {
	return width >= s.MinWidth && height >= s.MinHeight
}

// GetColumnWidths returns the widths of the three columns
func (s LayoutSpec) GetColumnWidths() (left, shared, right int) {
	return s.LeftColumnWidth, s.SharedColumnWidth, s.RightColumnWidth
}

// GetTotalWidth returns the total layout width including borders
func (s LayoutSpec) GetTotalWidth() int {
	left, shared, right := s.GetColumnWidths()
	// +2 for left/right borders, +2 for column separators
	return left + shared + right + 4
}

// GetTotalHeight returns the total layout height
func (s LayoutSpec) GetTotalHeight() int {
	return s.HeaderHeight + s.ContentHeight + s.NarratorHeight + s.FooterHeight
}
