package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/DisplaceTech/tss-ceremony/tui"
	"github.com/DisplaceTech/tss-ceremony/tui/commands"
	"github.com/DisplaceTech/tss-ceremony/tui/scenes"
)

// Config holds the CLI configuration
type Config struct {
	Fixed    bool
	Message  string
	Speed    string
	NoColor  bool
	AutoQuit bool

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
	flag.StringVar(&config.Message, "message", "", "Message to sign (default: Hello, threshold signatures)")
	flag.StringVar(&config.Speed, "speed", DefaultSpeed, "Animation speed: slow, normal, or fast")
	flag.BoolVar(&config.NoColor, "no-color", false, "Disable ANSI color output")
	flag.BoolVar(&config.AutoQuit, "auto-quit", false, "Quit automatically after animation completes")

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
	buildFrostData(ceremonyData, ceremony)

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
		AutoQuit:  config.AutoQuit,
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

// buildFrostData runs FROST signing with the exact same keys and nonces as
// the DKLS ceremony for a direct apples-to-apples comparison.
func buildFrostData(data *scenes.CeremonyData, c *protocol.Ceremony) {
	sr := c.SigningResult
	if sr == nil {
		return
	}

	// Reuse the DKLS keys and nonces so the viewer sees identical inputs
	aSecret := new(big.Int).SetBytes(c.PartyAKey.Serialize())
	bSecret := new(big.Int).SetBytes(c.PartyBKey.Serialize())

	signer := &protocol.FROSTSigner{
		Parties: []*protocol.FROSTParty{
			{
				ID:         0,
				Secret:     aSecret,
				Public:     c.PartyAPub,
				Nonce:      sr.NonceA,
				NoncePoint: sr.NonceAPub,
			},
			{
				ID:         1,
				Secret:     bSecret,
				Public:     c.PartyBPub,
				Nonce:      sr.NonceB,
				NoncePoint: sr.NonceBPub,
			},
		},
		P: c.PhantomKey,
		R: sr.CombinedR,
	}

	// Run only the FROST-specific steps (challenge, partials, aggregate)
	if err := signer.ComputeChallenge(c.Message); err != nil {
		return
	}
	if err := signer.ComputePartialSignatures(); err != nil {
		return
	}
	if err := signer.AggregateSignatures(); err != nil {
		return
	}

	// Keys and nonces match DKLS; copy the already-formatted hex
	data.FrostPartyASecretHex = data.PartyASecretHex
	data.FrostPartyBSecretHex = data.PartyBSecretHex
	data.FrostPartyAPubHex = data.PartyAPubHex
	data.FrostPartyBPubHex = data.PartyBPubHex
	data.FrostCombinedPubHex = data.CombinedPubHex
	data.FrostNonceAHex = data.NonceAHex
	data.FrostNonceBHex = data.NonceBHex

	// FROST-specific values (different from DKLS)
	data.FrostChallengeHex = fmt.Sprintf("%064x", signer.E)
	data.FrostPartialSigAHex = fmt.Sprintf("%064x", signer.Parties[0].PartialSig)
	data.FrostPartialSigBHex = fmt.Sprintf("%064x", signer.Parties[1].PartialSig)

	rBytes := make([]byte, 32)
	signer.R.X().FillBytes(rBytes)
	data.FrostSignatureRHex = fmt.Sprintf("%x", rBytes)
	data.FrostSignatureSHex = fmt.Sprintf("%064x", signer.S)

	valid, verr := protocol.VerifySchnorrSignature(signer.P, signer.R, signer.S, c.Message)
	data.FrostValid = verr == nil && valid
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
	data.PubKeyDERHex = c.GetPubKeyDERHex()
	data.SigDERHex = c.GetSignatureDERHex()
	if c.PhantomKey != nil {
		uncompressed := c.PhantomKey.SerializeUncompressed()
		data.PubKeyYHex = fmt.Sprintf("%x", uncompressed[33:]) // Y coordinate
	}

	return data
}
