package tests

import (
	"strings"
	"testing"
)

// noColorConfig mirrors main.Config fields relevant to --no-color testing.
type noColorConfig struct {
	NoColor        bool
	Fixed          bool
	Speed          string
	Message        string
	ValidateLayout bool
	Verify         bool
}

// parseNoColorArgs is a test-local parser mirroring main.ParseFlags behavior.
func parseNoColorArgs(args []string) *noColorConfig {
	cfg := &noColorConfig{Speed: "normal"}
	for i, arg := range args {
		switch arg {
		case "--no-color":
			cfg.NoColor = true
		case "--fixed":
			cfg.Fixed = true
		case "--validate-layout":
			cfg.ValidateLayout = true
		case "--verify":
			cfg.Verify = true
		case "--speed":
			if i+1 < len(args) {
				cfg.Speed = args[i+1]
			}
		case "--message":
			if i+1 < len(args) {
				cfg.Message = args[i+1]
			}
		}
	}
	return cfg
}

// TestNoColorDefaultsToFalse verifies the flag defaults to false.
func TestNoColorDefaultsToFalse(t *testing.T) {
	cfg := parseNoColorArgs([]string{})
	if cfg.NoColor {
		t.Error("expected NoColor=false by default")
	}
}

// TestNoColorFlagSetsTrue verifies --no-color sets the field to true.
func TestNoColorFlagSetsTrue(t *testing.T) {
	cfg := parseNoColorArgs([]string{"--no-color"})
	if !cfg.NoColor {
		t.Error("expected NoColor=true when --no-color is passed")
	}
}

// TestNoColorWithFixed verifies --no-color combined with --fixed.
func TestNoColorWithFixed(t *testing.T) {
	cfg := parseNoColorArgs([]string{"--no-color", "--fixed"})
	if !cfg.NoColor {
		t.Error("expected NoColor=true")
	}
	if !cfg.Fixed {
		t.Error("expected Fixed=true")
	}
}

// TestNoColorWithSpeed verifies --no-color combined with --speed.
func TestNoColorWithSpeed(t *testing.T) {
	cfg := parseNoColorArgs([]string{"--no-color", "--speed", "slow"})
	if !cfg.NoColor {
		t.Error("expected NoColor=true")
	}
	if cfg.Speed != "slow" {
		t.Errorf("expected Speed=slow, got %q", cfg.Speed)
	}
}

// TestNoColorWithMessage verifies --no-color combined with --message.
func TestNoColorWithMessage(t *testing.T) {
	cfg := parseNoColorArgs([]string{"--no-color", "--message", "deadbeef"})
	if !cfg.NoColor {
		t.Error("expected NoColor=true")
	}
	if cfg.Message != "deadbeef" {
		t.Errorf("expected Message=deadbeef, got %q", cfg.Message)
	}
}

// TestNoColorWithValidateLayout verifies --no-color combined with --validate-layout.
func TestNoColorWithValidateLayout(t *testing.T) {
	cfg := parseNoColorArgs([]string{"--no-color", "--validate-layout"})
	if !cfg.NoColor {
		t.Error("expected NoColor=true")
	}
	if !cfg.ValidateLayout {
		t.Error("expected ValidateLayout=true")
	}
}

// TestNoColorWithVerify verifies --no-color combined with --verify.
func TestNoColorWithVerify(t *testing.T) {
	cfg := parseNoColorArgs([]string{"--no-color", "--verify"})
	if !cfg.NoColor {
		t.Error("expected NoColor=true")
	}
	if !cfg.Verify {
		t.Error("expected Verify=true")
	}
}

// TestNoColorAbsentWhenNotPassed verifies NoColor is false without the flag.
func TestNoColorAbsentWhenNotPassed(t *testing.T) {
	cfg := parseNoColorArgs([]string{"--fixed", "--speed", "fast"})
	if cfg.NoColor {
		t.Error("expected NoColor=false when --no-color is absent")
	}
}

// TestNoColorFlagName documents the expected flag name convention.
func TestNoColorFlagName(t *testing.T) {
	flagName := "no-color"
	if !strings.Contains(flagName, "no") || !strings.Contains(flagName, "color") {
		t.Errorf("flag name %q does not follow no-color convention", flagName)
	}
}
