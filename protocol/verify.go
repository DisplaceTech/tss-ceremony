package protocol

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// VerifySignature verifies an ECDSA signature against a message and public key
// using secp256k1 curve.
//
// Parameters:
//   - pubkey: Hex-encoded uncompressed public key (64 bytes for x||y)
//   - sigR: Hex-encoded R component of the signature (32 bytes)
//   - sigS: Hex-encoded S component of the signature (32 bytes)
//   - message: Hex-encoded message to verify
//
// Returns true if the signature is valid, false otherwise.
func VerifySignature(pubkey, sigR, sigS, message string) (bool, error) {
	// Decode public key (expecting 64 hex chars = 32 bytes for x||y compressed form)
	pubKeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return false, fmt.Errorf("invalid public key hex: %w", err)
	}

	// Parse public key - secp256k1 expects uncompressed (65 bytes) or compressed (33 bytes)
	// If we have 64 bytes, we need to construct uncompressed form
	var pubKey *secp256k1.PublicKey
	if len(pubKeyBytes) == 64 {
		// Construct uncompressed public key from x||y
		uncompressed := make([]byte, 65)
		uncompressed[0] = 0x04 // Uncompressed prefix
		copy(uncompressed[1:], pubKeyBytes)
		pubKey, err = secp256k1.ParsePubKey(uncompressed)
		if err != nil {
			return false, fmt.Errorf("invalid public key: %w", err)
		}
	} else {
		pubKey, err = secp256k1.ParsePubKey(pubKeyBytes)
		if err != nil {
			return false, fmt.Errorf("invalid public key: %w", err)
		}
	}

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

	// Create signature from R and S
	r := new(secp256k1.ModNScalar).SetByteSlice(rBytes)
	s := new(secp256k1.ModNScalar).SetByteSlice(sBytes)
	_ = r
	_ = s

	// Decode message
	msgBytes, err := hex.DecodeString(message)
	if err != nil {
		return false, fmt.Errorf("invalid message hex: %w", err)
	}

	// Hash the message (ECDSA signs the hash of the message)
	hash := sha256.Sum256(msgBytes)

	// Verify the signature using secp256k1 library
	// Convert to big.Int for ecdsa verification
	rBig := new(big.Int).SetBytes(rBytes)
	sBig := new(big.Int).SetBytes(sBytes)
	_ = rBig
	_ = sBig

	// Convert secp256k1.PublicKey to crypto/ecdsa.PublicKey
	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	// Verify the signature
	valid := ecdsa.Verify(pubKeyECDSA, hash[:], rBig, sBig)

	return valid, nil
}
