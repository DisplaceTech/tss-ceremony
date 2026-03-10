package protocol

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// SchnorrSigSize is the byte length of a serialised secp256k1 Schnorr
// signature: 32-byte r (x-coord of R) || 32-byte s.
const SchnorrSigSize = 64

// SerializeSchnorrSignature encodes a FROST/Schnorr (R, S) pair into the
// standard 64-byte secp256k1 Schnorr wire format:
//
//	bytes  0-31 : x-coordinate of R (big-endian, zero-padded to 32 bytes)
//	bytes 32-63 : s scalar           (big-endian, zero-padded to 32 bytes)
//
// This matches the format used by BIP-340 / libsecp256k1.
func SerializeSchnorrSignature(sigR *secp256k1.PublicKey, sigS *big.Int) ([]byte, error) {
	if sigR == nil {
		return nil, fmt.Errorf("schnorr: R point is nil")
	}
	if sigS == nil {
		return nil, fmt.Errorf("schnorr: S scalar is nil")
	}

	// r = x-coordinate of R, padded to exactly 32 bytes.
	rBytes := make([]byte, 32)
	sigR.X().FillBytes(rBytes)

	// s scalar, padded to exactly 32 bytes.
	sBytes := make([]byte, 32)
	sigS.FillBytes(sBytes)

	// Sanity-check encoding lengths (both must be exactly 32 bytes).
	if len(rBytes) != 32 {
		return nil, fmt.Errorf("schnorr: r encoding has unexpected length %d (want 32)", len(rBytes))
	}
	if len(sBytes) != 32 {
		return nil, fmt.Errorf("schnorr: s encoding has unexpected length %d (want 32)", len(sBytes))
	}

	sig := make([]byte, SchnorrSigSize)
	copy(sig[0:32], rBytes)
	copy(sig[32:64], sBytes)
	return sig, nil
}

// ParseSchnorrSignature decodes a 64-byte secp256k1 Schnorr wire-format
// signature produced by SerializeSchnorrSignature.  It returns the r
// x-coordinate (as a *big.Int) and the s scalar.  The caller can recover the
// full R point from r using the curve if needed; for Schnorr verification
// under our scheme the full uncompressed R point is carried separately.
func ParseSchnorrSignature(sig []byte) (r, s *big.Int, err error) {
	if len(sig) != SchnorrSigSize {
		return nil, nil, fmt.Errorf("schnorr: invalid signature length %d (want %d)", len(sig), SchnorrSigSize)
	}
	r = new(big.Int).SetBytes(sig[0:32])
	s = new(big.Int).SetBytes(sig[32:64])
	return r, s, nil
}

// ValidateSchnorrEncoding checks that a serialised Schnorr signature has the
// correct 64-byte length and that both the r and s components are within the
// valid scalar range [1, n-1].  It returns a non-nil error describing the
// first violation found.
func ValidateSchnorrEncoding(sig []byte) error {
	if len(sig) != SchnorrSigSize {
		return fmt.Errorf("schnorr: invalid signature byte length %d (want %d)", len(sig), SchnorrSigSize)
	}

	n := secp256k1.S256().N

	r := new(big.Int).SetBytes(sig[0:32])
	if r.Sign() == 0 {
		return fmt.Errorf("schnorr: r is zero")
	}
	if r.Cmp(n) >= 0 {
		return fmt.Errorf("schnorr: r (%x…) is >= curve order", sig[0:4])
	}

	s := new(big.Int).SetBytes(sig[32:64])
	if s.Sign() == 0 {
		return fmt.Errorf("schnorr: s is zero")
	}
	if s.Cmp(n) >= 0 {
		return fmt.Errorf("schnorr: s (%x…) is >= curve order", sig[32:36])
	}

	return nil
}

// VerifySchnorrSignature verifies a Schnorr signature (R, S) against a public key and message.
//
// Schnorr verification equation: s*G = R + e*P
// where:
//   - R is the nonce point (public key)
//   - S is the signature scalar
//   - P is the public key
//   - e = H(R || message || P) is the challenge
//   - G is the generator point
//
// Parameters:
//   - publicKey: The public key to verify against (combined public key P)
//   - sigR: The R component of the signature (nonce point)
//   - sigS: The S component of the signature (scalar)
//   - message: The message that was signed
//
// Returns true if the signature is valid, false otherwise.
func VerifySchnorrSignature(publicKey *secp256k1.PublicKey, sigR *secp256k1.PublicKey, sigS *big.Int, message []byte) (bool, error) {
	if publicKey == nil {
		return false, fmt.Errorf("public key cannot be nil")
	}
	if sigR == nil {
		return false, fmt.Errorf("signature R cannot be nil")
	}
	if sigS == nil {
		return false, fmt.Errorf("signature S cannot be nil")
	}

	// Compute the challenge e = H(R || message || P)
	RBytes := sigR.SerializeUncompressed()
	PBytes := publicKey.SerializeUncompressed()

	hash := sha256.New()
	hash.Write(RBytes)
	hash.Write(message)
	hash.Write(PBytes)
	digest := hash.Sum(nil)

	// Convert hash to scalar (mod n)
	e := new(big.Int).SetBytes(digest)
	e.Mod(e, secp256k1.S256().N)

	// Verify: s*G = R + e*P
	// Left side: s*G
	leftSide := scalarBaseMult(sigS)
	if leftSide == nil {
		return false, fmt.Errorf("failed to compute s*G")
	}

	// Right side: R + e*P
	// First compute e*P
	eTimesP := scalarMult(publicKey, e)
	if eTimesP == nil {
		return false, fmt.Errorf("failed to compute e*P")
	}

	// Then add R + e*P
	rightSide, err := CombinePublicKeys(sigR, eTimesP)
	if err != nil {
		return false, fmt.Errorf("failed to compute R + e*P: %w", err)
	}

	// Compare left side and right side
	leftX := leftSide.X()
	leftY := leftSide.Y()
	rightX := rightSide.X()
	rightY := rightSide.Y()

	valid := leftX.Cmp(rightX) == 0 && leftY.Cmp(rightY) == 0

	return valid, nil
}

// scalarBaseMult computes k*G where k is a scalar and G is the generator point.
// Returns the resulting public key point.
func scalarBaseMult(k *big.Int) *secp256k1.PublicKey {
	if k == nil || k.Sign() == 0 {
		return nil
	}

	// Convert scalar to bytes for private key construction
	kBytes := k.Bytes()
	// Pad to 32 bytes if necessary
	if len(kBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(kBytes):], kBytes)
		kBytes = padded
	}

	// Create a temporary private key and derive its public key
	// This effectively computes k*G
	privKey := secp256k1.PrivKeyFromBytes(kBytes)
	if privKey == nil {
		return nil
	}

	return privKey.PubKey()
}

// scalarMult computes k*P where k is a scalar and P is a public key point.
// Returns the resulting public key point.
func scalarMult(P *secp256k1.PublicKey, k *big.Int) *secp256k1.PublicKey {
	if P == nil || k == nil || k.Sign() == 0 {
		return nil
	}

	// For scalar multiplication k*P, we need to use the curve's scalar multiplication.
	// The secp256k1 library doesn't expose direct scalar multiplication for public keys,
	// so we implement it using the double-and-add algorithm.

	n := secp256k1.S256().N
	p := secp256k1.S256().P

	// Reduce k mod n
	k = new(big.Int).Mod(k, n)

	// Get the coordinates of P
	x := new(big.Int).Set(P.X())
	y := new(big.Int).Set(P.Y())

	// Result starts at point at infinity (represented as nil)
	var resultX, resultY *big.Int

	// Double-and-add algorithm
	for k.Sign() > 0 {
		// If the current bit is 1, add P to the result
		if k.Bit(0) == 1 {
			if resultX == nil {
				// First addition, result = P
				resultX = new(big.Int).Set(x)
				resultY = new(big.Int).Set(y)
			} else {
				// Add P to result
				resultX, resultY = addPoints(resultX, resultY, x, y, p)
			}
		}

		// Double P for the next bit
		x, y = doublePoint(x, y, p)

		// Shift k right by 1 bit
		k.Rsh(k, 1)
	}

	// If result is still nil, return nil (point at infinity)
	if resultX == nil {
		return nil
	}

	// Create public key from the resulting coordinates
	serialized := make([]byte, 65)
	serialized[0] = 0x04
	resultX.FillBytes(serialized[1:33])
	resultY.FillBytes(serialized[33:65])

	pubKey, _ := secp256k1.ParsePubKey(serialized)
	return pubKey
}

// addPoints performs elliptic curve point addition: R = P1 + P2
// Returns the x and y coordinates of the result.
func addPoints(x1, y1, x2, y2, p *big.Int) (*big.Int, *big.Int) {
	// Check if points are the same (would need point doubling)
	xEqual := x1.Cmp(x2) == 0
	yEqual := y1.Cmp(y2) == 0

	var x3, y3 big.Int

	if xEqual && yEqual {
		// Point doubling
		x1, y1 = doublePoint(x1, y1, p)
		x3.Set(x1)
		y3.Set(y1)
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

	return &x3, &y3
}

// doublePoint performs elliptic curve point doubling: R = 2*P
// Returns the x and y coordinates of the result.
func doublePoint(x, y, p *big.Int) (*big.Int, *big.Int) {
	// Point doubling: lambda = (3*x^2 + a) / (2*y)
	// For secp256k1, a = 0, so lambda = (3*x^2) / (2*y)
	xSquared := new(big.Int).Mul(x, x)
	three := big.NewInt(3)
	numerator := new(big.Int).Mul(three, xSquared)
	numerator.Mod(numerator, p)

	two := big.NewInt(2)
	denominator := new(big.Int).Mul(two, y)
	denominator.Mod(denominator, p)

	denominatorInv := new(big.Int).ModInverse(denominator, p)
	lambda := new(big.Int).Mul(numerator, denominatorInv)
	lambda.Mod(lambda, p)

	// x3 = lambda^2 - 2*x
	lambdaSquared := new(big.Int).Mul(lambda, lambda)
	lambdaSquared.Mod(lambdaSquared, p)
	twoX := new(big.Int).Mul(two, x)
	twoX.Mod(twoX, p)
	x3 := new(big.Int).Sub(lambdaSquared, twoX)
	x3.Mod(x3, p)

	// y3 = lambda * (x - x3) - y
	xMinusX3 := new(big.Int).Sub(x, x3)
	xMinusX3.Mod(xMinusX3, p)
	y3 := new(big.Int).Mul(lambda, xMinusX3)
	y3.Mod(y3, p)
	y3.Sub(y3, y)
	y3.Mod(y3, p)

	return x3, y3
}
