package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/DisplaceTech/tss-ceremony/tui"
	"github.com/DisplaceTech/tss-ceremony/tui/commands"
	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// Config holds the CLI configuration
type Config struct {
	Fixed   bool
	Message string
	Speed   string
	NoColor bool

	// Verify subcommand flags
	Verify       bool
	PubKey       string
	SigR         string
	SigS         string
	PubKeyFile   string
	SigRFile     string
	SigSFile     string
	MessageFile  string

	// Layout validation flags
	ValidateLayout bool
}

// Default values
const (
	DefaultSpeed = "normal"
)

// Valid speed values
var validSpeeds = map[string]bool{
	"slow":   true,
	"normal": true,
	"fast":   true,
}

// ParseFlags parses command-line arguments and returns the configuration
func ParseFlags() (*Config, error) {
	config := &Config{
		Speed: DefaultSpeed,
	}

	flag.BoolVar(&config.Fixed, "fixed", false, "Use fixed seed for deterministic runs")
	flag.StringVar(&config.Message, "message", "", "Message to sign (hex encoded)")
	flag.StringVar(&config.Speed, "speed", DefaultSpeed, "Animation speed: slow, normal, or fast")
	flag.BoolVar(&config.NoColor, "no-color", false, "Disable ANSI color output")
	flag.BoolVar(&config.ValidateLayout, "validate-layout", false, "Run layout validation and exit")

	// Verify subcommand flags
	flag.BoolVar(&config.Verify, "verify", false, "Verify a signature (subcommand)")
	flag.StringVar(&config.PubKey, "pubkey", "", "Public key for verification (hex encoded)")
	flag.StringVar(&config.SigR, "sig-r", "", "R component of signature (hex encoded)")
	flag.StringVar(&config.SigS, "sig-s", "", "S component of signature (hex encoded)")
	flag.StringVar(&config.PubKeyFile, "pubkey-file", "", "File containing public key (hex encoded)")
	flag.StringVar(&config.SigRFile, "sig-r-file", "", "File containing R component (hex encoded)")
	flag.StringVar(&config.SigSFile, "sig-s-file", "", "File containing S component (hex encoded)")
	flag.StringVar(&config.MessageFile, "message-file", "", "File containing message (hex encoded)")

	flag.Parse()

	// Validate speed flag
	if !validSpeeds[config.Speed] {
		return nil, fmt.Errorf("invalid speed '%s': must be one of slow, normal, or fast", config.Speed)
	}

	// Validate verify subcommand flags
	if config.Verify {
		// Check if using file-based verification
		if config.PubKeyFile != "" || config.SigRFile != "" || config.SigSFile != "" || config.MessageFile != "" {
			// File-based verification requires all file flags
			if config.PubKeyFile == "" || config.SigRFile == "" || config.SigSFile == "" || config.MessageFile == "" {
				return nil, fmt.Errorf("--verify with file inputs requires all of: --pubkey-file, --sig-r-file, --sig-s-file, --message-file")
			}
		} else if config.PubKey == "" || config.SigR == "" || config.SigS == "" || config.Message == "" {
			// String-based verification requires all string flags
			if config.PubKey == "" {
				return nil, fmt.Errorf("--verify requires --pubkey or --pubkey-file")
			}
			if config.SigR == "" {
				return nil, fmt.Errorf("--verify requires --sig-r or --sig-r-file")
			}
			if config.SigS == "" {
				return nil, fmt.Errorf("--verify requires --sig-s or --sig-s-file")
			}
			if config.Message == "" && config.MessageFile == "" {
				return nil, fmt.Errorf("--verify requires --message or --message-file")
			}
		}
	}

	return config, nil
}

func main() {
	config, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle verify subcommand
	if config.Verify {
		var valid bool
		var err error

		// Check if using file-based verification
		if config.PubKeyFile != "" || config.SigRFile != "" || config.SigSFile != "" || config.MessageFile != "" {
			valid, err = commands.VerifyFromFile(config.PubKeyFile, config.SigRFile, config.SigSFile, config.MessageFile)
		} else {
			// Use string-based verification
			valid, err = protocol.VerifySignature(config.PubKey, config.SigR, config.SigS, config.Message)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Verification error: %v\n", err)
			os.Exit(1)
		}

		if valid {
			fmt.Println("Valid")
		} else {
			fmt.Println("Invalid")
			os.Exit(1)
		}
		return
	}

	// Handle layout validation mode
	if config.ValidateLayout {
		fmt.Println("\n=== Layout Validation ===")
		spec := tui.DefaultLayoutSpec()
		
		// Validate terminal size
		width, height := 80, 24 // Default terminal size for validation
		sizeResult := tui.ValidateTerminalSize(width, height, spec)
		fmt.Println(tui.FormatMismatchReport(sizeResult))
		
		// Validate layout structure
		// For now, we'll just validate the spec itself
		structResult := tui.ValidateLayoutStructure("", spec)
		fmt.Println(tui.FormatMismatchReport(structResult))
		
		// Check if all specs are valid
		allValid := sizeResult.IsValid && structResult.IsValid
		if !allValid {
			os.Exit(1)
		}
		fmt.Println("\n✓ All layout validations passed")
		return
	}

	// Initialize ceremony
	ceremony, err := protocol.NewCeremony(config.Fixed, config.Message, config.Speed, config.NoColor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing ceremony: %v\n", err)
		os.Exit(1)
	}

	// Initialize TUI model with ceremony reference
	tuiConfig := &scenes.Config{
		FixedMode: config.Fixed,
		Message:   config.Message,
		Speed:     config.Speed,
		NoColor:   config.NoColor,
	}
	model := tui.NewModel(tuiConfig, ceremony)

	// Create and run bubbletea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
