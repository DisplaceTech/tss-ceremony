package protocol

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// CeremonyConfig holds configuration for the TSS signing ceremony
type CeremonyConfig struct {
	// FixedMode enables deterministic runs with fixed seeds for testing
	FixedMode bool

	// Message is the plaintext message to sign (optional, defaults to demo message)
	Message string

	// Speed controls animation speed: "slow", "normal", or "fast"
	Speed string

	// NoColor disables ANSI color output
	NoColor bool

	// FixedSeed is used for deterministic key generation when FixedMode is true
	FixedSeed uint64

	// ParticipantCount is the number of participants in the ceremony (default: 2)
	ParticipantCount int

	// Threshold is the minimum number of participants required to sign (default: 2)
	Threshold int

	// UseFROST enables FROST (Schnorr) signing instead of DKLS (ECDSA)
	UseFROST bool

	// ShowSecurityProof enables display of security proof details
	ShowSecurityProof bool

	// SkipScenes is a list of scene numbers to skip during the ceremony
	SkipScenes []int
}

// DefaultCeremonyConfig returns a CeremonyConfig with default values
func DefaultCeremonyConfig() *CeremonyConfig {
	return &CeremonyConfig{
		FixedMode:          false,
		Message:            "",
		Speed:              "normal",
		NoColor:            false,
		FixedSeed:          0,
		ParticipantCount:   2,
		Threshold:          2,
		UseFROST:           false,
		ShowSecurityProof:  false,
		SkipScenes:         []int{},
	}
}

// Validate checks that the configuration is valid
func (c *CeremonyConfig) Validate() error {
	validSpeeds := map[string]bool{
		"slow":   true,
		"normal": true,
		"fast":   true,
	}

	if !validSpeeds[c.Speed] {
		return fmt.Errorf("invalid speed '%s': must be one of slow, normal, or fast", c.Speed)
	}

	if c.ParticipantCount < 2 {
		return fmt.Errorf("participant count must be at least 2, got %d", c.ParticipantCount)
	}

	if c.Threshold < 1 || c.Threshold > c.ParticipantCount {
		return fmt.Errorf("threshold must be between 1 and participant count (%d), got %d", c.ParticipantCount, c.Threshold)
	}

	return nil
}

// KeygenState holds state for the key generation phase
type KeygenState struct {
	// Party A key material
	PartyAKey []byte
	PartyAPub []byte

	// Party B key material
	PartyBKey []byte
	PartyBPub []byte

	// Shared key material
	PhantomPub []byte

	// Keygen progress
	KeysGenerated bool
	PublicKeysDerived bool
}

// SigningState holds state for the signing phase
type SigningState struct {
	// Message to sign
	Message []byte

	// Partial signatures
	PartyAPartialSig []byte
	PartyBPartialSig []byte

	// Combined signature components
	SignatureR []byte
	SignatureS []byte

	// Signing progress
	MessageHashed bool
	PartialSigsGenerated bool
	SignatureCombined bool
}

// OTState holds state for the Oblivious Transfer phase
type OTState struct {
	// Party A OT inputs
	PartyAOTInput []byte
	PartyAOTOutput []byte

	// Party B OT inputs
	PartyBOTInput []byte
	PartyBOTOutput []byte

	// OT progress
	OTInitiated bool
	OTCompleted bool
}

// MTAState holds state for the Message Transfer Agreement phase
type MTAState struct {
	// Party A MTA data
	PartyAMTAData []byte
	PartyAMTASig []byte

	// Party B MTA data
	PartyBMTAData []byte
	PartyBMTASig []byte

	// MTA progress
	MTAInitiated bool
	MTAVerified bool
}

// VerifyState holds state for the verification phase
type VerifyState struct {
	// Signature to verify
	SignatureR []byte
	SignatureS []byte

	// Public key for verification
	PublicKey []byte

	// Message hash
	MessageHash []byte

	// Verification result
	IsValid bool
	Verified bool
}

// NewCeremonyFromConfig creates a new Ceremony instance from a CeremonyConfig
func NewCeremonyFromConfig(config *CeremonyConfig) (*Ceremony, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ceremony config: %w", err)
	}

	ceremony := &Ceremony{
		FixedMode:    config.FixedMode,
		Speed:        config.Speed,
		NoColor:      config.NoColor,
		KeygenState:  &KeygenState{},
		SigningState: &SigningState{},
		OTState:      &OTState{},
		MTAState:     &MTAState{},
		VerifyState:  &VerifyState{},
		CurrentPhase: 0,
		PhaseComplete: make(map[int]bool),
	}

	// Initialize message (plaintext, not hex)
	if config.Message != "" {
		ceremony.Message = []byte(config.Message)
	} else {
		ceremony.Message = []byte("Hello, threshold signatures!")
	}

	// Initialize based on mode
	if config.FixedMode {
		if err := ceremony.initializeFixedMode(config.FixedSeed); err != nil {
			return nil, fmt.Errorf("failed to initialize fixed mode: %w", err)
		}
	} else {
		if err := ceremony.initializeRandomMode(); err != nil {
			return nil, fmt.Errorf("failed to initialize random mode: %w", err)
		}
	}

	return ceremony, nil
}

// initializeFixedMode initializes the ceremony with deterministic keys for testing
func (c *Ceremony) initializeFixedMode(seed uint64) error {
	// Use fixed seeds for deterministic runs
	if seed == 0 {
		seed = 42 // Default seed if not specified
	}

	// Generate Party A key from seed
	seedA := big.NewInt(int64(seed))
	c.PartyAKey = secp256k1.PrivKeyFromBytes(seedA.Bytes())
	if c.PartyAKey == nil {
		return fmt.Errorf("failed to generate party A key from seed")
	}

	// Generate Party B key from seed + 1
	seedB := big.NewInt(int64(seed + 1))
	c.PartyBKey = secp256k1.PrivKeyFromBytes(seedB.Bytes())
	if c.PartyBKey == nil {
		return fmt.Errorf("failed to generate party B key from seed")
	}

	// Derive public keys
	c.PartyAPub = c.PartyAKey.PubKey()
	c.PartyBPub = c.PartyBKey.PubKey()

	// Mark keygen state as initialized
	c.KeygenState.KeysGenerated = true
	c.KeygenState.PublicKeysDerived = true

	return nil
}

// initializeRandomMode initializes the ceremony with cryptographically random keys
func (c *Ceremony) initializeRandomMode() error {
	// Generate random keys for Party A
	var err error
	c.PartyAKey, err = secp256k1.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate party A key: %w", err)
	}

	// Generate random keys for Party B
	c.PartyBKey, err = secp256k1.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate party B key: %w", err)
	}

	// Derive public keys
	c.PartyAPub = c.PartyAKey.PubKey()
	c.PartyBPub = c.PartyBKey.PubKey()

	// Mark keygen state as initialized
	c.KeygenState.KeysGenerated = true
	c.KeygenState.PublicKeysDerived = true

	return nil
}

// IsFixedMode returns true if the ceremony is running in fixed/deterministic mode
func (c *Ceremony) IsFixedMode() bool {
	return c.FixedMode
}

// Reset resets the ceremony state while preserving configuration
func (c *Ceremony) Reset() {
	c.KeygenState = &KeygenState{}
	c.SigningState = &SigningState{}
	c.OTState = &OTState{}
	c.MTAState = &MTAState{}
	c.VerifyState = &VerifyState{}
	c.CurrentPhase = 0
	c.PhaseComplete = make(map[int]bool)
	c.SignatureR = nil
	c.SignatureS = nil
	c.PhantomKey = nil
}

// GetConfig returns a CeremonyConfig based on the current ceremony state
func (c *Ceremony) GetConfig() *CeremonyConfig {
	return &CeremonyConfig{
		FixedMode:      c.FixedMode,
		Message:        hex.EncodeToString(c.Message),
		Speed:          c.Speed,
		NoColor:        c.NoColor,
		ParticipantCount: 2,
		Threshold:      2,
	}
}
