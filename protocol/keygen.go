package protocol

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// CeremonyConfig holds configuration for the TSS ceremony
type CeremonyConfig struct {
	// Fixed enables deterministic key generation for testing
	Fixed bool
	// Message is the message to sign (hex-encoded or plain text)
	Message string
}

// GenerateSecretShare generates a cryptographically secure random 256-bit scalar
// using crypto/rand for use as a secret share in the TSS ceremony.
// Returns a 32-byte random scalar suitable for use as a secp256k1 private key.
func GenerateSecretShare() ([]byte, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// ComputePublicShare computes the public share by multiplying the secret scalar
// with the generator point G on secp256k1.
// The secret must be a valid 32-byte scalar (private key).
// Returns the corresponding public key point.
func ComputePublicShare(secret []byte) (*secp256k1.PublicKey, error) {
	if len(secret) != 32 {
		return nil, fmt.Errorf("secret must be 32 bytes, got %d", len(secret))
	}

	// Create private key from bytes
	privateKey := secp256k1.PrivKeyFromBytes(secret)
	if privateKey == nil {
		return nil, fmt.Errorf("invalid private key bytes")
	}

	// Compute public key by scalar multiplication with generator point G
	publicKey := privateKey.PubKey()
	if publicKey == nil {
		return nil, fmt.Errorf("failed to compute public key")
	}

	return publicKey, nil
}

// CombinePublicKeys adds two public key points to produce the combined public key.
// This performs elliptic curve point addition on secp256k1.
// Both publicA and publicB must be valid secp256k1 public key points.
// Returns the sum of the two points.
//
// Note: Since we only have public keys (not private keys), we perform point addition
// directly on the curve. The Add method computes R = P + G, so we use a workaround:
// we compute the combined key by adding the private key scalars and deriving the public key.
func CombinePublicKeys(publicA, publicB *secp256k1.PublicKey) (*secp256k1.PublicKey, error) {
	if publicA == nil {
		return nil, fmt.Errorf("publicA cannot be nil")
	}
	if publicB == nil {
		return nil, fmt.Errorf("publicB cannot be nil")
	}

	// Get the affine coordinates from both public keys
	x1 := publicA.X()
	y1 := publicA.Y()
	x2 := publicB.X()
	y2 := publicB.Y()

	// The secp256k1 library's Add method computes R = P + G (adds to generator)
	// For general point addition P + Q, we need to implement it manually
	// or use the curve's Jacobian coordinate operations.
	//
	// We'll implement point addition directly using the curve equation.
	// For secp256k1: y^2 = x^3 + 7 (mod p)
	//
	// Point addition formula for P1 != P2:
	//   lambda = (y2 - y1) / (x2 - x1) (mod p)
	//   x3 = lambda^2 - x1 - x2 (mod p)
	//   y3 = lambda * (x1 - x3) - y1 (mod p)

	p := secp256k1.S256().P

	// Check if points are the same (would need point doubling)
	xEqual := x1.Cmp(x2) == 0
	yEqual := y1.Cmp(y2) == 0

	var x3, y3 big.Int

	if xEqual && yEqual {
		// Point doubling: lambda = (3*x1^2 + a) / (2*y1)
		// For secp256k1, a = 0, so lambda = (3*x1^2) / (2*y1)
		x1Squared := new(big.Int).Mul(x1, x1)
		three := big.NewInt(3)
		numerator := new(big.Int).Mul(three, x1Squared)
		numerator.Mod(numerator, p)

		two := big.NewInt(2)
		denominator := new(big.Int).Mul(two, y1)
		denominator.Mod(denominator, p)

		denominatorInv := new(big.Int).ModInverse(denominator, p)
		lambda := new(big.Int).Mul(numerator, denominatorInv)
		lambda.Mod(lambda, p)

		// x3 = lambda^2 - 2*x1
		lambdaSquared := new(big.Int).Mul(lambda, lambda)
		lambdaSquared.Mod(lambdaSquared, p)
		twoX1 := new(big.Int).Mul(two, x1)
		twoX1.Mod(twoX1, p)
		x3.Sub(lambdaSquared, twoX1)
		x3.Mod(&x3, p)

		// y3 = lambda * (x1 - x3) - y1
		x1MinusX3 := new(big.Int).Sub(x1, &x3)
		x1MinusX3.Mod(x1MinusX3, p)
		y3.Mul(lambda, x1MinusX3)
		y3.Mod(&y3, p)
		y3.Sub(&y3, y1)
		y3.Mod(&y3, p)
	} else {
		// Point addition: lambda = (y2 - y1) / (x2 - x1)
		numerator := new(big.Int).Sub(y2, y1)
		numerator.Mod(numerator, p)

		denominator := new(big.Int).Sub(x2, x1)
		denominator.Mod(denominator, p)

		denominatorInv := new(big.Int).ModInverse(denominator, p)
		lambda := new(big.Int).Mul(numerator, denominatorInv)
		lambda.Mod(lambda, p)

		// x3 = lambda^2 - x1 - x2
		lambdaSquared := new(big.Int).Mul(lambda, lambda)
		lambdaSquared.Mod(lambdaSquared, p)
		x3.Sub(lambdaSquared, x1)
		x3.Sub(&x3, x2)
		x3.Mod(&x3, p)

		// y3 = lambda * (x1 - x3) - y1
		x1MinusX3 := new(big.Int).Sub(x1, &x3)
		x1MinusX3.Mod(x1MinusX3, p)
		y3.Mul(lambda, x1MinusX3)
		y3.Mod(&y3, p)
		y3.Sub(&y3, y1)
		y3.Mod(&y3, p)
	}

	// Create public key from the resulting coordinates
	// Serialize in uncompressed format: 0x04 || x || y
	serialized := make([]byte, 65)
	serialized[0] = 0x04
	x3.FillBytes(serialized[1:33])
	y3.FillBytes(serialized[33:65])

	// Parse the serialized public key
	combinedPubKey, _ := secp256k1.ParsePubKey(serialized)

	if combinedPubKey == nil {
		return nil, fmt.Errorf("failed to create combined public key")
	}

	return combinedPubKey, nil
}

// GenerateSecretShareFixed generates a deterministic secret share from a seed
// for testing purposes.
func GenerateSecretShareFixed(seed int64) []byte {
	seedBig := big.NewInt(seed)
	bytes := seedBig.Bytes()
	// Pad to 32 bytes
	result := make([]byte, 32)
	copy(result[32-len(bytes):], bytes)
	return result
}
