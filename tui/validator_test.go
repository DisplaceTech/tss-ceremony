package tui

import (
	"strings"
	"testing"
)

func TestValidateLayout_Match(t *testing.T) {
	rendered := `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │
└────────────────────────────────────────────────────────────────┘`

	spec := `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │
└────────────────────────────────────────────────────────────────┘`

	result := ValidateLayout(rendered, spec, "Test Scene")

	if !result.IsValid {
		t.Errorf("Expected layout to match, but got mismatches: %v", result.Mismatches)
	}

	if len(result.Mismatches) != 0 {
		t.Errorf("Expected 0 mismatches, got %d", len(result.Mismatches))
	}
}

func TestValidateLayout_Mismatch(t *testing.T) {
	rendered := `┌────────────────────────────────────────────────────────────────┐
│                    DIFFERENT TEXT                              │
└────────────────────────────────────────────────────────────────┘`

	spec := `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │
└────────────────────────────────────────────────────────────────┘`

	result := ValidateLayout(rendered, spec, "Test Scene")

	if result.IsValid {
		t.Error("Expected layout to have mismatches, but validation passed")
	}

	if len(result.Mismatches) == 0 {
		t.Error("Expected at least one mismatch, got none")
	}

	// Check that the mismatch is on line 1 (second line)
	foundLine1Mismatch := false
	for _, m := range result.Mismatches {
		if m.Line == 1 {
			foundLine1Mismatch = true
			break
		}
	}

	if !foundLine1Mismatch {
		t.Error("Expected mismatch on line 1, but none found")
	}
}

func TestValidateLayout_MissingLine(t *testing.T) {
	rendered := `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │`

	spec := `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │
└────────────────────────────────────────────────────────────────┘`

	result := ValidateLayout(rendered, spec, "Test Scene")

	if result.IsValid {
		t.Error("Expected validation to fail due to missing line")
	}

	if len(result.Mismatches) == 0 {
		t.Error("Expected mismatch for missing line")
	}
}

func TestValidateLayout_ExtraLine(t *testing.T) {
	rendered := `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │
└────────────────────────────────────────────────────────────────┘
Extra line`

	spec := `┌────────────────────────────────────────────────────────────────┐
│                    TSS CEREMONY DEMO                           │
└────────────────────────────────────────────────────────────────┘`

	result := ValidateLayout(rendered, spec, "Test Scene")

	// Extra lines in rendered output are not considered mismatches
	// (we only validate against the spec)
	if !result.IsValid {
		t.Errorf("Expected validation to pass (extra lines ignored), got mismatches: %v", result.Mismatches)
	}
}

func TestValidateTerminalSize_Valid(t *testing.T) {
	spec := DefaultLayoutSpec()

	result := ValidateTerminalSize(80, 24, spec)

	if !result.IsValid {
		t.Errorf("Expected valid terminal size, got mismatches: %v", result.Mismatches)
	}
}

func TestValidateTerminalSize_TooNarrow(t *testing.T) {
	spec := DefaultLayoutSpec()

	result := ValidateTerminalSize(79, 24, spec)

	if result.IsValid {
		t.Error("Expected validation to fail for width < 80")
	}

	if len(result.Mismatches) == 0 {
		t.Error("Expected mismatch for width validation")
	}
}

func TestValidateTerminalSize_TooShort(t *testing.T) {
	spec := DefaultLayoutSpec()

	result := ValidateTerminalSize(80, 23, spec)

	if result.IsValid {
		t.Error("Expected validation to fail for height < 24")
	}

	if len(result.Mismatches) == 0 {
		t.Error("Expected mismatch for height validation")
	}
}

func TestValidateTerminalSize_BothInvalid(t *testing.T) {
	spec := DefaultLayoutSpec()

	result := ValidateTerminalSize(70, 20, spec)

	if result.IsValid {
		t.Error("Expected validation to fail for both width and height")
	}

	if len(result.Mismatches) < 2 {
		t.Errorf("Expected at least 2 mismatches, got %d", len(result.Mismatches))
	}
}

func TestValidateLayoutStructure_Valid(t *testing.T) {
	rendered := `┌────────────────────────────────────────────────────────────────┐
│                                                                │
│  ┌──────────────┐              ┌──────────────┐               │
│  │   PARTY A    │              │   PARTY B    │               │
│  └──────────────┘              └──────────────┘               │
│                                                                │
│              ┌──────────────┐                                 │
│              │   SHARED     │                                 │
│              └──────────────┘                                 │
│                                                                │
└────────────────────────────────────────────────────────────────┘`

	spec := DefaultLayoutSpec()

	result := ValidateLayoutStructure(rendered, spec)

	if !result.IsValid {
		t.Errorf("Expected valid layout structure, got mismatches: %v", result.Mismatches)
	}
}

func TestValidateLayoutStructure_Invalid(t *testing.T) {
	rendered := `┌────────────────────────────────────────────────────────────────┐
│                                                                │
│  PARTY A only - no columns                                     │
│                                                                │
└────────────────────────────────────────────────────────────────┘`

	spec := DefaultLayoutSpec()

	result := ValidateLayoutStructure(rendered, spec)

	if result.IsValid {
		t.Error("Expected validation to fail for missing three-column structure")
	}

	if len(result.Mismatches) == 0 {
		t.Error("Expected mismatch for missing column structure")
	}
}

func TestValidateLayoutStructure_ShortHeader(t *testing.T) {
	rendered := `┌──────────┐
│ Short    │
└──────────┘`

	spec := DefaultLayoutSpec()

	result := ValidateLayoutStructure(rendered, spec)

	if result.IsValid {
		t.Error("Expected validation to fail for short header")
	}

	if len(result.Mismatches) == 0 {
		t.Error("Expected mismatch for short header")
	}
}

func TestFormatMismatchReport_Passed(t *testing.T) {
	result := ValidationResult{
		IsValid:    true,
		TotalLines: 10,
		TotalChars: 100,
		Mismatches: []Mismatch{},
		SpecName:   "Test Spec",
	}

	report := FormatMismatchReport(result)

	if !strings.Contains(report, "PASSED") {
		t.Error("Expected report to contain 'PASSED'")
	}

	if !strings.Contains(report, "Test Spec") {
		t.Error("Expected report to contain spec name")
	}
}

func TestFormatMismatchReport_Failed(t *testing.T) {
	result := ValidationResult{
		IsValid:    false,
		TotalLines: 10,
		TotalChars: 100,
		Mismatches: []Mismatch{
			{
				Line:     0,
				Column:   5,
				Expected: 'A',
				Actual:   'B',
				Context:  "test context",
			},
		},
		SpecName: "Test Spec",
	}

	report := FormatMismatchReport(result)

	if !strings.Contains(report, "FAILED") {
		t.Error("Expected report to contain 'FAILED'")
	}

	if !strings.Contains(report, "Line 0, Column 5") {
		t.Error("Expected report to contain mismatch coordinates")
	}

	if !strings.Contains(report, "U+0041") {
		t.Error("Expected report to contain expected character (U+0041)")
	}

	if !strings.Contains(report, "U+0042") {
		t.Error("Expected report to contain actual character (U+0042)")
	}
}

func TestValidateLayout_EmptySpec(t *testing.T) {
	rendered := "Some text"
	spec := ""

	result := ValidateLayout(rendered, spec, "Empty Spec")

	// Empty spec should result in no mismatches (nothing to compare against)
	// The function only validates against spec lines, so empty spec = no validation
	if len(result.Mismatches) != 0 {
		t.Errorf("Expected empty spec to pass validation (no spec lines to compare), got mismatches: %v", result.Mismatches)
	}
}

func TestValidateLayout_EmptyRendered(t *testing.T) {
	rendered := ""
	spec := "Some text"

	result := ValidateLayout(rendered, spec, "Non-empty Spec")

	if result.IsValid {
		t.Error("Expected validation to fail for empty rendered output")
	}
}

func TestValidateLayout_SingleCharMismatch(t *testing.T) {
	rendered := "Hello World"
	spec := "Hello world"

	result := ValidateLayout(rendered, spec, "Single Char Test")

	if result.IsValid {
		t.Error("Expected validation to fail for case mismatch")
	}

	if len(result.Mismatches) != 1 {
		t.Errorf("Expected exactly 1 mismatch, got %d", len(result.Mismatches))
	}

	mismatch := result.Mismatches[0]
	if mismatch.Line != 0 || mismatch.Column != 6 {
		t.Errorf("Expected mismatch at line 0, column 6, got line %d, column %d", mismatch.Line, mismatch.Column)
	}
}

func TestValidateLayout_MultilineMismatch(t *testing.T) {
	rendered := `Line 1
Line 2
Line 3`

	spec := `Line 1
Line X
Line 3`

	result := ValidateLayout(rendered, spec, "Multiline Test")

	if result.IsValid {
		t.Error("Expected validation to fail for multiline mismatch")
	}

	if len(result.Mismatches) != 1 {
		t.Errorf("Expected exactly 1 mismatch, got %d", len(result.Mismatches))
	}

	mismatch := result.Mismatches[0]
	if mismatch.Line != 1 || mismatch.Column != 5 {
		t.Errorf("Expected mismatch at line 1, column 5, got line %d, column %d", mismatch.Line, mismatch.Column)
	}
}
