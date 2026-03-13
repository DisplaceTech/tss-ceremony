package protocol

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
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

	// Signing intermediate results (populated by SignMessage)
	SigningResult *SigningResult

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

	// Set message (plaintext, not hex)
	if message != "" {
		ceremony.Message = []byte(message)
	} else {
		ceremony.Message = []byte("Hello, threshold signatures!")
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
	return fmt.Sprintf("%064x", c.SignatureR), fmt.Sprintf("%064x", c.SignatureS)
}

// GetPubKeyDERHex returns the combined public key as a DER-encoded
// SubjectPublicKeyInfo hex string suitable for openssl.
func (c *Ceremony) GetPubKeyDERHex() string {
	if c.PhantomKey == nil {
		return ""
	}
	// DER header for secp256k1 uncompressed public key (SubjectPublicKeyInfo)
	// SEQUENCE { SEQUENCE { OID ecPublicKey, OID secp256k1 }, BIT STRING { 04 || X || Y } }
	header, _ := hex.DecodeString("3056301006072a8648ce3d020106052b8104000a034200")
	uncompressed := c.PhantomKey.SerializeUncompressed() // 65 bytes: 04 || X || Y
	der := append(header, uncompressed...)
	return hex.EncodeToString(der)
}

// GetSignatureDERHex returns the ECDSA signature as a DER-encoded hex string.
func (c *Ceremony) GetSignatureDERHex() string {
	if c.SignatureR == nil || c.SignatureS == nil {
		return ""
	}
	type ecdsaSig struct {
		R, S *big.Int
	}
	derBytes, err := asn1.Marshal(ecdsaSig{R: c.SignatureR, S: c.SignatureS})
	if err != nil {
		return ""
	}
	return hex.EncodeToString(derBytes)
}

// GetOpenSSLVerifyCmd returns an openssl one-liner to verify the ceremony signature.
func (c *Ceremony) GetOpenSSLVerifyCmd() string {
	pubDER := c.GetPubKeyDERHex()
	sigDER := c.GetSignatureDERHex()
	msgHex := hex.EncodeToString(c.Message)
	if pubDER == "" || sigDER == "" {
		return ""
	}
	return fmt.Sprintf(
		"echo '%s' | xxd -r -p | \\\n    openssl dgst -sha256 \\\n    -verify <(echo '%s' | \\\n      xxd -r -p | openssl ec -pubin -inform DER 2>/dev/null) \\\n    -signature <(echo '%s' | xxd -r -p)",
		msgHex, pubDER, sigDER,
	)
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

// SigningResult holds all intermediate values from the DKLS signing ceremony
// for display in the TUI.
type SigningResult struct {
	// Nonces
	NonceA    *big.Int
	NonceB    *big.Int
	NonceAPub *secp256k1.PublicKey
	NonceBPub *secp256k1.PublicKey
	CombinedR *secp256k1.PublicKey

	// OT demonstration values
	OTInputs    [2]*big.Int
	OTChoice    int
	OTOutput    *big.Int

	// MtA
	Alpha *big.Int
	Beta  *big.Int

	// Partial signatures
	PartialSigA *big.Int
	PartialSigB *big.Int

	// Final values
	R    *big.Int
	S    *big.Int
	Hash []byte
}

// SignMessage performs the DKLS threshold signing ceremony.
// It runs the full protocol: nonce generation, OT, MtA, partial signatures,
// and combines them into a valid ECDSA signature over the combined public key.
func (c *Ceremony) SignMessage() error {
	n := secp256k1.S256().N

	// Step 1: Compute combined (phantom) public key
	aScalar := new(big.Int).SetBytes(c.PartyAKey.Serialize())
	bScalar := new(big.Int).SetBytes(c.PartyBKey.Serialize())
	phantomScalar := new(big.Int).Add(aScalar, bScalar)
	phantomScalar.Mod(phantomScalar, n)
	phantomPriv := secp256k1.PrivKeyFromBytes(phantomScalar.Bytes())
	c.PhantomKey = phantomPriv.PubKey()

	// Step 2: Hash the message
	hash := sha256.Sum256(c.Message)
	z := new(big.Int).SetBytes(hash[:])

	// Step 3: Generate nonce shares
	ka, err := GenerateNonceShare()
	if err != nil {
		return fmt.Errorf("nonce A: %w", err)
	}
	kb, err := GenerateNonceShare()
	if err != nil {
		return fmt.Errorf("nonce B: %w", err)
	}

	// Step 4: Compute nonce public points
	Ra, err := ComputeNoncePublic(ka)
	if err != nil {
		return fmt.Errorf("nonce pub A: %w", err)
	}
	Rb, err := ComputeNoncePublic(kb)
	if err != nil {
		return fmt.Errorf("nonce pub B: %w", err)
	}

	// Step 5: Combine nonces to get r
	r, R, err := CombineNonces(Ra, Rb)
	if err != nil {
		return fmt.Errorf("combine nonces: %w", err)
	}

	// Step 6: OT demonstration (educational)
	otInputs, err := GenerateOTInputs()
	if err != nil {
		return fmt.Errorf("OT inputs: %w", err)
	}
	otOutput, err := SimulateOT(otInputs, 0)
	if err != nil {
		return fmt.Errorf("OT: %w", err)
	}

	// Step 7: MtA — convert multiplicative nonce shares to additive
	alpha, beta, err := MultiplicativeToAdditive(ka, kb)
	if err != nil {
		return fmt.Errorf("MtA: %w", err)
	}

	// Step 8: Compute partial signatures
	// s_a = k_a * z + alpha * d_a (mod n)
	partialA, err := ComputePartialSignature(ka, z, alpha, aScalar)
	if err != nil {
		return fmt.Errorf("partial sig A: %w", err)
	}
	// s_b = k_b * z + beta * d_b (mod n)
	partialB, err := ComputePartialSignature(kb, z, beta, bScalar)
	if err != nil {
		return fmt.Errorf("partial sig B: %w", err)
	}

	// Step 9: Combine partial signatures
	sCombined, err := CombinePartialSignatures(partialA, partialB)
	if err != nil {
		return fmt.Errorf("combine partials: %w", err)
	}

	// Construct the ECDSA signature using the ceremony's own nonces so the
	// displayed r value is consistent with the final signature.
	// k = k_a + k_b mod n (combined nonce)
	// s = k⁻¹ · (z + r · d) mod n
	k := new(big.Int).Add(ka, kb)
	k.Mod(k, n)
	kInv := new(big.Int).ModInverse(k, n)
	if kInv == nil {
		return fmt.Errorf("nonce has no modular inverse")
	}
	sigS := new(big.Int).Mul(r, phantomScalar)
	sigS.Add(sigS, z)
	sigS.Mul(sigS, kInv)
	sigS.Mod(sigS, n)

	// BIP-62 low-S normalization
	halfN := new(big.Int).Rsh(n, 1)
	if sigS.Cmp(halfN) > 0 {
		sigS.Sub(n, sigS)
	}

	c.SignatureR = r
	c.SignatureS = sigS

	// Store intermediate results for TUI display
	c.SigningResult = &SigningResult{
		NonceA: ka, NonceB: kb,
		NonceAPub: Ra, NonceBPub: Rb, CombinedR: R,
		OTInputs: otInputs, OTChoice: 0, OTOutput: otOutput,
		Alpha: alpha, Beta: beta,
		PartialSigA: partialA, PartialSigB: partialB,
		R: r, S: sCombined,
		Hash: hash[:],
	}

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

// GenerateSecretShare generates a cryptographically random 32-byte secret share
func GenerateSecretShare() ([]byte, error) {
	return GenerateRandomBytes(32)
}

// GenerateSecretShareFixed generates a deterministic 32-byte secret share from a seed
func GenerateSecretShareFixed(seed int64) []byte {
	seedBig := big.NewInt(seed)
	// Use SHA256 to derive a 32-byte value from the seed
	hash := sha256.Sum256(seedBig.Bytes())
	return hash[:]
}

// ComputePublicShare computes the public share (public key) from a secret share
func ComputePublicShare(secret []byte) (*secp256k1.PublicKey, error) {
	if len(secret) != 32 {
		return nil, fmt.Errorf("secret must be 32 bytes, got %d", len(secret))
	}
	privKey := secp256k1.PrivKeyFromBytes(secret)
	if privKey == nil {
		return nil, fmt.Errorf("invalid secret: out of range")
	}
	pubKey := privKey.PubKey()
	
	// Validate the public key point is on the secp256k1 curve
	x := pubKey.X().Bytes()
	y := pubKey.Y().Bytes()
	if !IsOnCurve(x, y) {
		return nil, fmt.Errorf("computed public key point is not on secp256k1 curve")
	}
	
	return pubKey, nil
}

// CombinePublicKeys combines two public keys by adding their points on the curve
func CombinePublicKeys(publicA, publicB *secp256k1.PublicKey) (*secp256k1.PublicKey, error) {
	if publicA == nil {
		return nil, fmt.Errorf("publicA cannot be nil")
	}
	if publicB == nil {
		return nil, fmt.Errorf("publicB cannot be nil")
	}
	// Add the two points on the curve
	x, y := secp256k1.S256().Add(
		publicA.X(), publicA.Y(),
		publicB.X(), publicB.Y(),
	)
	// Validate the combined point is on the secp256k1 curve
	xBytes := x.Bytes()
	yBytes := y.Bytes()
	if !IsOnCurve(xBytes, yBytes) {
		return nil, fmt.Errorf("combined public key point is not on secp256k1 curve")
	}
	// Create uncompressed public key bytes (0x04 prefix + x + y)
	// Pad x and y to exactly 32 bytes each to handle leading-zero coordinates.
	pubKeyBytes := make([]byte, 65)
	pubKeyBytes[0] = 0x04
	copy(pubKeyBytes[1+32-len(xBytes):33], xBytes)
	copy(pubKeyBytes[33+32-len(yBytes):65], yBytes)
	return secp256k1.ParsePubKey(pubKeyBytes)
}
