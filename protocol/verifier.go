package protocol

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ECDSAVerifier provides ECDSA signature verification functionality.
type ECDSAVerifier struct {
	publicKey *secp256k1.PublicKey
}

// NewECDSAVerifier creates a new ECDSAVerifier with the given public key.
func NewECDSAVerifier(pubKey *secp256k1.PublicKey) *ECDSAVerifier {
	return &ECDSAVerifier{
		publicKey: pubKey,
	}
}

// NewECDSAVerifierFromHex creates a new ECDSAVerifier from a hex-encoded public key.
// The public key can be:
//   - 64 bytes (x||y without 0x04 prefix)
//   - 33 bytes (compressed format)
//   - 65 bytes (uncompressed format with 0x04 prefix)
func NewECDSAVerifierFromHex(pubKeyHex string) (*ECDSAVerifier, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid public key hex: %w", err)
	}

	var pubKey *secp256k1.PublicKey
	switch len(pubKeyBytes) {
	case 64:
		// Construct uncompressed public key from x||y
		uncompressed := make([]byte, 65)
		uncompressed[0] = 0x04
		copy(uncompressed[1:], pubKeyBytes)
		pubKey, err = secp256k1.ParsePubKey(uncompressed)
	case 33:
		// Compressed public key
		pubKey, err = secp256k1.ParsePubKey(pubKeyBytes)
	case 65:
		// Uncompressed public key with 0x04 prefix
		pubKey, err = secp256k1.ParsePubKey(pubKeyBytes)
	default:
		return nil, fmt.Errorf("invalid public key length: expected 64, 33, or 65 bytes, got %d", len(pubKeyBytes))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &ECDSAVerifier{publicKey: pubKey}, nil
}

// Verify verifies an ECDSA signature against a message.
// Returns true if the signature is valid, false otherwise.
func (v *ECDSAVerifier) Verify(message []byte, r, s *big.Int) bool {
	if v.publicKey == nil || r == nil || s == nil {
		return false
	}

	hash := sha256.Sum256(message)
	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     v.publicKey.X(),
		Y:     v.publicKey.Y(),
	}

	return ecdsa.Verify(pubKeyECDSA, hash[:], r, s)
}

// VerifyHex verifies an ECDSA signature given hex-encoded components.
// Parameters:
//   - sigR: Hex-encoded R component (32 bytes)
//   - sigS: Hex-encoded S component (32 bytes)
//   - message: Hex-encoded message to verify
//
// Returns true if the signature is valid, false otherwise.
func (v *ECDSAVerifier) VerifyHex(sigR, sigS, message string) (bool, error) {
	// Decode R component
	rBytes, err := hex.DecodeString(sigR)
	if err != nil {
		return false, fmt.Errorf("invalid signature R hex: %w", err)
	}

	// Decode S component
	sBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return false, fmt.Errorf("invalid signature S hex: %w", err)
	}

	// Decode message
	msgBytes, err := hex.DecodeString(message)
	if err != nil {
		return false, fmt.Errorf("invalid message hex: %w", err)
	}

	// Convert to big.Int
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	return v.Verify(msgBytes, r, s), nil
}

// VerifyWithSignature verifies an ECDSA signature given an ECDSASignature struct.
func (v *ECDSAVerifier) VerifyWithSignature(message []byte, sig *ECDSASignature) bool {
	if sig == nil || sig.R == nil || sig.S == nil {
		return false
	}
	return v.Verify(message, sig.R, sig.S)
}
