package tui

import (
	"fmt"
	"strings"
)

// PartyLabel returns a labelled string identifying a party, with optional
// bracketed prefix in no-color mode so parties remain visually distinct.
//
//	color mode  → ANSI-colored name (handled by Styles)
//	no-color    → "[A] Party A" / "[B] Party B" prefix style
func PartyLabel(party string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("[%s] Party %s", party, party)
	}
	return "Party " + party
}

// SectionHeader renders a section header. In no-color mode the header is
// wrapped in ASCII dashes for visual separation.
func SectionHeader(title string, width int, noColor bool) string {
	if width <= 0 {
		width = len(title) + 4
	}
	if noColor {
		bar := strings.Repeat("=", width)
		return fmt.Sprintf("%s\n%s\n%s", bar, title, bar)
	}
	return title
}

// StatusLine renders a status message. In no-color mode a prefix marker is
// added so different status classes remain distinguishable.
func StatusLine(msg string, statusType string, noColor bool) string {
	if !noColor {
		return msg
	}
	switch statusType {
	case "success":
		return "[OK] " + msg
	case "warning":
		return "[!!] " + msg
	case "error":
		return "[ERR] " + msg
	case "info":
		return "[>>] " + msg
	default:
		return msg
	}
}

// HexBlock formats a hex string in 8-char groups with spaces, matching the
// project style convention (all hex displayed in 8-char groups).
func HexBlock(hexStr string) string {
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

// TwoColumnLayout renders two named columns side by side separated by a
// vertical bar. Each column has content lines. In no-color mode the column
// headers are prefixed with [A] / [B] markers instead of relying on color.
func TwoColumnLayout(leftTitle, rightTitle string, leftLines, rightLines []string, colWidth int, noColor bool) string {
	if colWidth <= 0 {
		colWidth = 30
	}

	padRight := func(s string, w int) string {
		if len(s) >= w {
			return s[:w]
		}
		return s + strings.Repeat(" ", w-len(s))
	}

	// Headers
	var leftHdr, rightHdr string
	if noColor {
		leftHdr = "[A] " + leftTitle
		rightHdr = "[B] " + rightTitle
	} else {
		leftHdr = leftTitle
		rightHdr = rightTitle
	}

	var sb strings.Builder
	sb.WriteString(padRight(leftHdr, colWidth) + " │ " + rightHdr + "\n")
	sb.WriteString(strings.Repeat("─", colWidth) + "─┼─" + strings.Repeat("─", colWidth) + "\n")

	maxLen := len(leftLines)
	if len(rightLines) > maxLen {
		maxLen = len(rightLines)
	}
	for i := 0; i < maxLen; i++ {
		var l, r string
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		sb.WriteString(padRight(l, colWidth) + " │ " + r + "\n")
	}

	return sb.String()
}
