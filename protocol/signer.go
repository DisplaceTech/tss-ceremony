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

// Signer holds a secp256k1 private key and provides ECDSA signing functionality.
type Signer struct {
	privateKey *secp256k1.PrivateKey
	publicKey  *secp256k1.PublicKey
}

// NewSigner generates a new random secp256k1 private key and derives the public key.
func NewSigner() (*Signer, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return &Signer{
		privateKey: privKey,
		publicKey:  privKey.PubKey(),
	}, nil
}

// NewSignerFromBytes creates a Signer from raw 32-byte private key scalar.
func NewSignerFromBytes(keyBytes []byte) (*Signer, error) {
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("private key must be 32 bytes, got %d", len(keyBytes))
	}
	privKey := secp256k1.PrivKeyFromBytes(keyBytes)
	if privKey == nil {
		return nil, fmt.Errorf("invalid private key bytes")
	}
	return &Signer{
		privateKey: privKey,
		publicKey:  privKey.PubKey(),
	}, nil
}

// PublicKey returns the signer's secp256k1 public key.
func (s *Signer) PublicKey() *secp256k1.PublicKey {
	return s.publicKey
}

// PublicKeyHex returns the uncompressed public key (without 0x04 prefix) as a hex string.
func (s *Signer) PublicKeyHex() string {
	uncompressed := s.publicKey.SerializeUncompressed()
	return hex.EncodeToString(uncompressed[1:]) // strip 0x04 prefix
}

// PublicKeyCompressedHex returns the compressed public key as a hex string.
func (s *Signer) PublicKeyCompressedHex() string {
	return hex.EncodeToString(s.publicKey.SerializeCompressed())
}

// Sign hashes message with SHA-256 and produces a standard ECDSA signature.
// Returns the (R, S) components as *big.Int values.
func (s *Signer) Sign(message []byte) (*big.Int, *big.Int, error) {
	hash := sha256.Sum256(message)

	privScalar := new(big.Int).SetBytes(s.privateKey.Serialize())
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     s.publicKey.X(),
			Y:     s.publicKey.Y(),
		},
		D: privScalar,
	}

	r, sig, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		return nil, nil, fmt.Errorf("ECDSA sign failed: %w", err)
	}
	return r, sig, nil
}

// SignHex hashes message with SHA-256, signs it, and returns (R, S) as hex strings.
func (s *Signer) SignHex(message []byte) (rHex, sHex string, err error) {
	r, sig, err := s.Sign(message)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(r.Bytes()), hex.EncodeToString(sig.Bytes()), nil
}

// Verify checks an ECDSA (R, S) signature against the SHA-256 hash of message
// using the signer's own public key.
func (s *Signer) Verify(message []byte, r, sigS *big.Int) bool {
	hash := sha256.Sum256(message)
	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     s.publicKey.X(),
		Y:     s.publicKey.Y(),
	}
	return ecdsa.Verify(pubKeyECDSA, hash[:], r, sigS)
}
