package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/DisplaceTech/tss-ceremony/tui"
	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// Config holds the CLI configuration
type Config struct {
	Fixed   bool
	Message string
	Speed   string
	NoColor bool

	// Verify subcommand flags
	Verify   bool
	PubKey   string
	SigR     string
	SigS     string
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

	// Verify subcommand flags
	flag.BoolVar(&config.Verify, "verify", false, "Verify a signature (subcommand)")
	flag.StringVar(&config.PubKey, "pubkey", "", "Public key for verification (hex encoded)")
	flag.StringVar(&config.SigR, "sig-r", "", "R component of signature (hex encoded)")
	flag.StringVar(&config.SigS, "sig-s", "", "S component of signature (hex encoded)")

	flag.Parse()

	// Validate speed flag
	if !validSpeeds[config.Speed] {
		return nil, fmt.Errorf("invalid speed '%s': must be one of slow, normal, or fast", config.Speed)
	}

	// Validate verify subcommand flags
	if config.Verify {
		if config.PubKey == "" {
			return nil, fmt.Errorf("--verify requires --pubkey")
		}
		if config.SigR == "" {
			return nil, fmt.Errorf("--verify requires --sig-r")
		}
		if config.SigS == "" {
			return nil, fmt.Errorf("--verify requires --sig-s")
		}
		if config.Message == "" {
			return nil, fmt.Errorf("--verify requires --message")
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
		valid, err := protocol.VerifySignature(config.PubKey, config.SigR, config.SigS, config.Message)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Verification error: %v\n", err)
			os.Exit(1)
		}

		if valid {
			fmt.Println("Signature verification: VALID")
		} else {
			fmt.Println("Signature verification: INVALID")
			os.Exit(1)
		}
		return
	}

	// Print configuration for debugging
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Fixed: %v\n", config.Fixed)
	fmt.Printf("  Message: %s\n", config.Message)
	fmt.Printf("  Speed: %s\n", config.Speed)
	fmt.Printf("  NoColor: %v\n", config.NoColor)

	// Initialize ceremony
	ceremony, err := protocol.NewCeremony(config.Fixed, config.Message, config.Speed, config.NoColor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing ceremony: %v\n", err)
		os.Exit(1)
	}

	// Print fixed mode banner if applicable
	if config.Fixed {
		fmt.Println("\n=== FIXED MODE ===")
	}

	// Initialize TUI model
	tuiConfig := &scenes.Config{
		FixedMode: config.Fixed,
		Message:   config.Message,
		Speed:     config.Speed,
		NoColor:   config.NoColor,
	}
	model := tui.NewModel(tuiConfig)

	// Create and run bubbletea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// Perform signing ceremony (for demonstration)
	if err := ceremony.SignMessage(); err != nil {
		fmt.Fprintf(os.Stderr, "Error signing message: %v\n", err)
		os.Exit(1)
	}

	// Output results
	fmt.Println("\n=== Ceremony Complete ===")
	fmt.Printf("Party A Public Key: %s\n", ceremony.GetPartyAPubKeyHex())
	fmt.Printf("Party B Public Key: %s\n", ceremony.GetPartyBPubKeyHex())
	fmt.Printf("Phantom Public Key: %s\n", ceremony.GetPhantomPubKeyHex())
	r, s := ceremony.GetSignatureHex()
	fmt.Printf("Signature R: %s\n", r)
	fmt.Printf("Signature S: %s\n", s)

	// Demonstrate ComputeNoncePublic with a sample nonce
	fmt.Println("\n=== ComputeNoncePublic Demo ===")
	// Use GenerateNonceShare to generate a cryptographically secure random nonce
	nonce, err := protocol.GenerateNonceShare()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating nonce: %v\n", err)
		os.Exit(1)
	}
	noncePublic, err := protocol.ComputeNoncePublic(nonce)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing nonce public: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated nonce (k): %s\n", nonce.String())
	fmt.Printf("Nonce public point (R = k * G):\n")
	fmt.Printf("  X: %s\n", noncePublic.X().String())
	fmt.Printf("  Y: %s\n", noncePublic.Y().String())

	// Demonstrate SimulateOT with sample inputs
	fmt.Println("\n=== SimulateOT Demo ===")
	// Generate sample OT inputs
	senderInputs, err := protocol.GenerateOTInputs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating OT inputs: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Sender inputs: [input0=%s, input1=%s]\n", senderInputs[0].String(), senderInputs[1].String())

	// Demonstrate with choice = 0
	result0, err := protocol.SimulateOT(senderInputs, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running SimulateOT with choice=0: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Receiver choice=0 -> receives: %s\n", result0.String())

	// Demonstrate with choice = 1
	result1, err := protocol.SimulateOT(senderInputs, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running SimulateOT with choice=1: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Receiver choice=1 -> receives: %s\n", result1.String())

	// Demonstrate error handling with invalid choice
	_, err = protocol.SimulateOT(senderInputs, 2)
	if err != nil {
		fmt.Printf("Invalid choice (2) correctly returns error: %v\n", err)
	}
}
