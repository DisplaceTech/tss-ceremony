package protocol

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Ceremony represents the TSS signing ceremony state, aggregating all protocol phases
type Ceremony struct {
	// Configuration
	FixedMode bool
	Message   []byte
	Speed     string
	NoColor   bool

	// Party A state
	PartyAKey *secp256k1.PrivateKey
	PartyAPub *secp256k1.PublicKey

	// Party B state
	PartyBKey *secp256k1.PrivateKey
	PartyBPub *secp256k1.PublicKey

	// Shared/derived state
	PhantomKey *secp256k1.PublicKey
	SignatureR *big.Int
	SignatureS *big.Int

	// Current scene
	CurrentScene int

	// Keygen phase state
	KeygenState *KeygenState

	// Signing phase state
	SigningState *SigningState

	// OT (Oblivious Transfer) phase state
	OTState *OTState

	// MTA (Message Transfer Agreement) phase state
	MTAState *MTAState

	// Verify phase state
	VerifyState *VerifyState

	// Progress tracking
	CurrentPhase int // 0=keygen, 1=signing, 2=OT, 3=MTA, 4=verify
	PhaseComplete map[int]bool

	// Ceremony lifecycle state
	complete bool
}

// NewCeremony creates a new ceremony instance
func NewCeremony(fixedMode bool, message string, speed string, noColor bool) (*Ceremony, error) {
	ceremony := &Ceremony{
		FixedMode: fixedMode,
		Speed:     speed,
		NoColor:   noColor,
		KeygenState: &KeygenState{},
		SigningState: &SigningState{},
		OTState: &OTState{},
		MTAState: &MTAState{},
		VerifyState: &VerifyState{},
		CurrentPhase: 0,
		PhaseComplete: make(map[int]bool),
	}

	// Parse message if provided
	if message != "" {
		msgBytes, err := hex.DecodeString(message)
		if err != nil {
			return nil, fmt.Errorf("invalid message hex: %w", err)
		}
		ceremony.Message = msgBytes
	} else {
		// Default message for demonstration
		ceremony.Message = []byte("TSS Ceremony Demo")
	}

	// Generate keys
	if fixedMode {
		// Use fixed seeds for deterministic runs
		ceremony.PartyAKey = generateFixedKey(1)
		ceremony.PartyBKey = generateFixedKey(2)
	} else {
		// Generate random keys
		var err error
		ceremony.PartyAKey, err = secp256k1.GeneratePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate party A key: %w", err)
		}
		ceremony.PartyBKey, err = secp256k1.GeneratePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate party B key: %w", err)
		}
	}

	// Derive public keys
	ceremony.PartyAPub = ceremony.PartyAKey.PubKey()
	ceremony.PartyBPub = ceremony.PartyBKey.PubKey()

	return ceremony, nil
}

// generateFixedKey generates a deterministic private key from a seed
func generateFixedKey(seed int) *secp256k1.PrivateKey {
	// Use a simple deterministic derivation for fixed mode
	seedBig := big.NewInt(int64(seed))
	privateKey := secp256k1.PrivKeyFromBytes(seedBig.Bytes())
	return privateKey
}

// GetPartyAPubKeyHex returns Party A's public key as hex string
func (c *Ceremony) GetPartyAPubKeyHex() string {
	return hex.EncodeToString(c.PartyAPub.SerializeCompressed()[1:])
}

// GetPartyBPubKeyHex returns Party B's public key as hex string
func (c *Ceremony) GetPartyBPubKeyHex() string {
	return hex.EncodeToString(c.PartyBPub.SerializeCompressed()[1:])
}

// GetPhantomPubKeyHex returns the phantom public key as hex string
func (c *Ceremony) GetPhantomPubKeyHex() string {
	if c.PhantomKey == nil {
		return ""
	}
	return hex.EncodeToString(c.PhantomKey.SerializeCompressed()[1:])
}

// GetSignatureHex returns the signature as hex strings (R, S)
func (c *Ceremony) GetSignatureHex() (string, string) {
	if c.SignatureR == nil || c.SignatureS == nil {
		return "", ""
	}
	return hex.EncodeToString(c.SignatureR.Bytes()), hex.EncodeToString(c.SignatureS.Bytes())
}

// GetSpeedDelay returns the delay multiplier based on speed setting
func (c *Ceremony) GetSpeedDelay() float64 {
	switch c.Speed {
	case "slow":
		return 2.0
	case "fast":
		return 0.5
	default:
		return 1.0
	}
}

// SignMessage performs the TSS signing ceremony
func (c *Ceremony) SignMessage() error {
	// This is a simplified signing for demonstration
	// In a real TSS ceremony, this would involve multiple rounds of communication
	
	// For now, just use Party A's key to sign (simplified)
	hash := sha256.Sum256(c.Message)
	
	// Get private key scalar as big.Int
	privScalar := new(big.Int).SetBytes(c.PartyAKey.Serialize())
	
	// Convert secp256k1 private key to ecdsa.PrivateKey for signing
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     c.PartyAPub.X(),
			Y:     c.PartyAPub.Y(),
		},
		D: privScalar,
	}
	
	r, s, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		return err
	}
	c.SignatureR = r
	c.SignatureS = s

	// Derive phantom key (simplified - in real TSS this would be more complex)
	// Add the private key scalars using the scalar field
	phantomScalar := new(big.Int).Add(new(big.Int).SetBytes(c.PartyAKey.Serialize()), new(big.Int).SetBytes(c.PartyBKey.Serialize()))
	phantomScalar.Mod(phantomScalar, secp256k1.S256().N)
	phantomKey := secp256k1.PrivKeyFromBytes(phantomScalar.Bytes())
	c.PhantomKey = phantomKey.PubKey()

	return nil
}

// GetCurrentPhase returns the current phase index
func (c *Ceremony) GetCurrentPhase() int {
	return c.CurrentPhase
}

// SetCurrentPhase sets the current phase index
func (c *Ceremony) SetCurrentPhase(phase int) {
	c.CurrentPhase = phase
}

// IsPhaseComplete returns whether a phase is complete
func (c *Ceremony) IsPhaseComplete(phase int) bool {
	return c.PhaseComplete[phase]
}

// MarkPhaseComplete marks a phase as complete
func (c *Ceremony) MarkPhaseComplete(phase int) {
	c.PhaseComplete[phase] = true
}

// GetPhaseName returns the name of a phase
func (c *Ceremony) GetPhaseName(phase int) string {
	phaseNames := []string{"Keygen", "Signing", "OT", "MTA", "Verify"}
	if phase < 0 || phase >= len(phaseNames) {
		return "Unknown"
	}
	return phaseNames[phase]
}

// CeremonyState represents the overall state of the ceremony
type CeremonyState struct {
	CurrentPhase    int
	PhaseComplete   map[int]bool
	KeygenState     *KeygenState
	SigningState    *SigningState
	OTState         *OTState
	MTAState        *MTAState
	VerifyState     *VerifyState
	IsInitialized   bool
	IsComplete      bool
}

// Init initializes the ceremony state and prepares it for execution
func (c *Ceremony) Init() error {
	// Reset all state
	c.Reset()
	
	// Initialize phase completion tracking
	c.PhaseComplete = make(map[int]bool)
	c.CurrentPhase = 0
	
	// Initialize all phase states
	c.KeygenState = &KeygenState{}
	c.SigningState = &SigningState{}
	c.OTState = &OTState{}
	c.MTAState = &MTAState{}
	c.VerifyState = &VerifyState{}
	
	// Generate keys based on mode
	if c.FixedMode {
		if err := c.initializeFixedMode(0); err != nil {
			return fmt.Errorf("failed to initialize fixed mode: %w", err)
		}
	} else {
		if err := c.initializeRandomMode(); err != nil {
			return fmt.Errorf("failed to initialize random mode: %w", err)
		}
	}
	
	return nil
}

// StartPhase begins execution of a specific phase
func (c *Ceremony) StartPhase(phase int) error {
	// Validate phase number
	if phase < 0 || phase > 4 {
		return fmt.Errorf("invalid phase %d: must be between 0 and 4", phase)
	}
	
	// Check if previous phase is complete (except for phase 0)
	if phase > 0 && !c.PhaseComplete[phase-1] {
		return fmt.Errorf("cannot start phase %d: previous phase %d is not complete", phase, phase-1)
	}
	
	// Set current phase
	c.CurrentPhase = phase
	
	// Initialize phase-specific state
	switch phase {
	case 0: // Keygen
		c.KeygenState = &KeygenState{}
	case 1: // Signing
		c.SigningState = &SigningState{
			Message: c.Message,
		}
	case 2: // OT
		c.OTState = &OTState{}
	case 3: // MTA
		c.MTAState = &MTAState{}
	case 4: // Verify
		hash := sha256.Sum256(c.Message)
		c.VerifyState = &VerifyState{
			MessageHash: hash[:],
		}
	}
	
	return nil
}

// CompletePhase marks the current phase as complete and advances to the next phase
func (c *Ceremony) CompletePhase() error {
	// Mark current phase as complete
	c.PhaseComplete[c.CurrentPhase] = true
	
	// Check if all phases are complete
	allComplete := true
	for i := 0; i <= 4; i++ {
		if !c.PhaseComplete[i] {
			allComplete = false
			break
		}
	}
	
	if allComplete {
		c.complete = true
		return nil
	}
	
	// Advance to next phase
	nextPhase := c.CurrentPhase + 1
	if nextPhase > 4 {
		return nil // No more phases
	}
	
	return c.StartPhase(nextPhase)
}

// GetState returns the current ceremony state
func (c *Ceremony) GetState() *CeremonyState {
	return &CeremonyState{
		CurrentPhase:  c.CurrentPhase,
		PhaseComplete: c.PhaseComplete,
		KeygenState:   c.KeygenState,
		SigningState:  c.SigningState,
		OTState:       c.OTState,
		MTAState:      c.MTAState,
		VerifyState:   c.VerifyState,
		IsInitialized: c.PartyAKey != nil && c.PartyBKey != nil,
		IsComplete:    c.complete,
	}
}

// IsComplete returns true if all phases have been completed
func (c *Ceremony) IsComplete() bool {
	return c.complete
}

// GetCurrentPhaseName returns the name of the current phase
func (c *Ceremony) GetCurrentPhaseName() string {
	return c.GetPhaseName(c.CurrentPhase)
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
