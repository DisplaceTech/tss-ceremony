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

// Ceremony represents the TSS signing ceremony state
type Ceremony struct {
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
}

// NewCeremony creates a new ceremony instance
func NewCeremony(fixedMode bool, message string, speed string, noColor bool) (*Ceremony, error) {
	ceremony := &Ceremony{
		FixedMode: fixedMode,
		Speed:     speed,
		NoColor:   noColor,
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

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
