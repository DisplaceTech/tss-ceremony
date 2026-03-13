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
	flag.StringVar(&config.Message, "message", "", "Message to sign (default: Hello, threshold signatures!)")
	flag.StringVar(&config.Speed, "speed", DefaultSpeed, "Animation speed: slow, normal, or fast")
	flag.BoolVar(&config.NoColor, "no-color", false, "Disable ANSI color output")

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

	// Initialize ceremony and run signing protocol BEFORE the TUI starts.
	// This computes all real cryptographic values that scenes will display.
	ceremony, err := protocol.NewCeremony(config.Fixed, config.Message, config.Speed, config.NoColor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing ceremony: %v\n", err)
		os.Exit(1)
	}
	if err := ceremony.SignMessage(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running signing ceremony: %v\n", err)
		os.Exit(1)
	}

	// Build ceremony data for TUI display
	ceremonyData := buildCeremonyData(ceremony)

	// Determine message for display
	displayMessage := config.Message
	if displayMessage == "" {
		displayMessage = string(ceremony.Message)
	}

	// Initialize TUI model with ceremony results
	tuiConfig := &scenes.Config{
		FixedMode: config.Fixed,
		Message:   displayMessage,
		Speed:     config.Speed,
		NoColor:   config.NoColor,
		Ceremony:  ceremonyData,
	}
	model := tui.NewModel(tuiConfig, ceremony)

	// Create and run bubbletea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// Print summary after TUI exits
	r, s := ceremony.GetSignatureHex()
	fmt.Println("\n=== Ceremony Complete ===")
	fmt.Printf("Public Key:  %s\n", ceremony.GetPhantomPubKeyHex())
	fmt.Printf("Signature R: %s\n", r)
	fmt.Printf("Signature S: %s\n", s)
	fmt.Printf("Message:     %s\n", string(ceremony.Message))

	if cmd := ceremony.GetOpenSSLVerifyCmd(); cmd != "" {
		fmt.Println("\n=== Verify with OpenSSL ===")
		fmt.Println(cmd)
	}
}

// buildCeremonyData converts protocol.Ceremony results into TUI-displayable data.
func buildCeremonyData(c *protocol.Ceremony) *scenes.CeremonyData {
	data := &scenes.CeremonyData{
		MessageText:    string(c.Message),
		CombinedPubHex: c.GetPhantomPubKeyHex(),
		PartyAPubHex:   c.GetPartyAPubKeyHex(),
		PartyBPubHex:   c.GetPartyBPubKeyHex(),
		Valid:          true,
	}

	sigR, sigS := c.GetSignatureHex()
	data.SignatureRHex = sigR
	data.SignatureSHex = sigS

	if c.SigningResult != nil {
		sr := c.SigningResult
		data.MessageHash = fmt.Sprintf("%064x", sr.Hash)
		data.NonceAHex = fmt.Sprintf("%064x", sr.NonceA)
		data.NonceBHex = fmt.Sprintf("%064x", sr.NonceB)
		if sr.NonceAPub != nil {
			data.NonceAPubHex = fmt.Sprintf("%x", sr.NonceAPub.SerializeCompressed()[1:])
		}
		if sr.NonceBPub != nil {
			data.NonceBPubHex = fmt.Sprintf("%x", sr.NonceBPub.SerializeCompressed()[1:])
		}
		if sr.CombinedR != nil {
			data.CombinedRPubHex = fmt.Sprintf("%x", sr.CombinedR.SerializeCompressed()[1:])
		}
		if sr.R != nil {
			data.RHex = fmt.Sprintf("%064x", sr.R)
		}
		if sr.OTInputs[0] != nil {
			data.OTInput0Hex = fmt.Sprintf("%064x", sr.OTInputs[0])
		}
		if sr.OTInputs[1] != nil {
			data.OTInput1Hex = fmt.Sprintf("%064x", sr.OTInputs[1])
		}
		data.OTChoiceBit = sr.OTChoice
		if sr.OTOutput != nil {
			data.OTOutputHex = fmt.Sprintf("%064x", sr.OTOutput)
		}
		if sr.Alpha != nil {
			data.AlphaHex = fmt.Sprintf("%064x", sr.Alpha)
		}
		if sr.Beta != nil {
			data.BetaHex = fmt.Sprintf("%064x", sr.Beta)
		}
		if sr.PartialSigA != nil {
			data.PartialSigAHex = fmt.Sprintf("%064x", sr.PartialSigA)
		}
		if sr.PartialSigB != nil {
			data.PartialSigBHex = fmt.Sprintf("%064x", sr.PartialSigB)
		}
	}

	data.PartyASecretHex = fmt.Sprintf("%064x", c.PartyAKey.Serialize())
	data.PartyBSecretHex = fmt.Sprintf("%064x", c.PartyBKey.Serialize())
	data.OpenSSLVerify = c.GetOpenSSLVerifyCmd()

	return data
}
