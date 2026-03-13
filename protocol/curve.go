// Package protocol implements threshold signature schemes including DKLS23 ECDSA and FROST Schnorr.
package protocol

import (
	"crypto/elliptic"
	"math/big"
)

// secp256k1 curve parameters (Bitcoin/Ethereum curve)
// These are the standard parameters for the secp256k1 elliptic curve
const (
	// p is the prime field modulus: 2^256 - 2^32 - 977
	p = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"

	// b is the curve parameter b (for secp256k1: y^2 = x^3 + 7, so b = 7)
	b = "7"

	// n is the order of the base point G (number of points on the curve)
	n = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"

	// Gx is the x-coordinate of the base point G
	Gx = "79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"

	// Gy is the y-coordinate of the base point G
	Gy = "483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8"
)

// secp256k1Curve implements elliptic.Curve for secp256k1
type secp256k1Curve struct{}

// NewSecp256k1Curve returns a new secp256k1 curve instance
func NewSecp256k1Curve() elliptic.Curve {
	return &secp256k1Curve{}
}

// Params returns the curve parameters
func (c *secp256k1Curve) Params() *elliptic.CurveParams {
	return &elliptic.CurveParams{
		Name:    "secp256k1",
		P:       decodeHex(p),
		N:       decodeHex(n),
		B:       decodeHex(b),
		Gx:      decodeHex(Gx),
		Gy:      decodeHex(Gy),
		BitSize: 256,
	}
}

// Add returns the sum of two points (x1, y1) and (x2, y2)
func (c *secp256k1Curve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	return c.Params().Add(x1, y1, x2, y2)
}

// Double returns 2*(x, y)
func (c *secp256k1Curve) Double(x, y *big.Int) (*big.Int, *big.Int) {
	return c.Params().Double(x, y)
}

// ScalarMult returns k*(x, y)
func (c *secp256k1Curve) ScalarMult(x, y *big.Int, k []byte) (*big.Int, *big.Int) {
	return c.Params().ScalarMult(x, y, k)
}

// ScalarBaseMult returns k*G
func (c *secp256k1Curve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return c.Params().ScalarBaseMult(k)
}

// IsOnCurve reports whether the point (x, y) is on the curve
func (c *secp256k1Curve) IsOnCurve(x, y *big.Int) bool {
	return c.Params().IsOnCurve(x, y)
}

// DecodePoint converts a byte slice to a point
func (c *secp256k1Curve) DecodePoint(b []byte) (x, y *big.Int) {
	// secp256k1 uses uncompressed format: 0x04 || x || y
	if len(b) != 65 || b[0] != 0x04 {
		return nil, nil
	}
	x = new(big.Int).SetBytes(b[1:33])
	y = new(big.Int).SetBytes(b[33:65])
	return x, y
}

// decodeHex converts a hex string to a big.Int
func decodeHex(hex string) *big.Int {
	val, ok := new(big.Int).SetString(hex, 16)
	if !ok {
		panic("failed to decode hex")
	}
	return val
}

// IsOnCurve verifies that a given (x, y) coordinate pair satisfies the secp256k1 equation
// y^2 = x^3 + 7 mod p. It handles affine coordinates and edge cases.
// x and y are expected to be 32-byte slices representing the coordinates.
func IsOnCurve(x, y []byte) bool {
	// Handle nil or empty inputs
	if len(x) == 0 || len(y) == 0 {
		return false
	}

	// Convert byte slices to big.Int
	xBig := new(big.Int).SetBytes(x)
	yBig := new(big.Int).SetBytes(y)

	// Get curve parameters
	pBig := decodeHex(p)
	bBig := decodeHex(b)

	// Calculate y^2 mod p
	y2 := new(big.Int).Exp(yBig, big.NewInt(2), pBig)

	// Calculate x^3 mod p
	x3 := new(big.Int).Exp(xBig, big.NewInt(3), pBig)

	// Calculate x^3 + 7 mod p
	x3Plus7 := new(big.Int).Add(x3, bBig)
	x3Plus7.Mod(x3Plus7, pBig)

	// Check if y^2 = x^3 + 7 mod p
	return y2.Cmp(x3Plus7) == 0
}
