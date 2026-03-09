package tests

import (
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestCombinePublicKeysDistinctPoints tests adding two distinct valid points
func TestCombinePublicKeysDistinctPoints(t *testing.T) {
	// Generate two distinct secret shares
	secretA := protocol.GenerateSecretShareFixed(1)
	secretB := protocol.GenerateSecretShareFixed(2)

	// Compute public shares
	pubA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) error: %v", err)
	}

	pubB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) error: %v", err)
	}

	// Combine the public keys
	combined, err := protocol.CombinePublicKeys(pubA, pubB)
	if err != nil {
		t.Fatalf("CombinePublicKeys() error: %v", err)
	}

	// Verify the result is not nil
	if combined == nil {
		t.Fatal("CombinePublicKeys() returned nil")
	}

	// Verify the result is on the curve
	if !isPointOnCurve(combined) {
		t.Error("Combined point is not on the secp256k1 curve")
	}

	// Verify the combined point is different from both inputs
	if pointsEqual(combined, pubA) {
		t.Error("Combined point equals pubA, which is unexpected for distinct inputs")
	}
	if pointsEqual(combined, pubB) {
		t.Error("Combined point equals pubB, which is unexpected for distinct inputs")
	}
}

// TestCombinePublicKeysPointWithItself tests adding a point to itself (point doubling)
func TestCombinePublicKeysPointWithItself(t *testing.T) {
	// Generate a secret share
	secret := protocol.GenerateSecretShareFixed(42)

	// Compute public share
	pub, err := protocol.ComputePublicShare(secret)
	if err != nil {
		t.Fatalf("ComputePublicShare() error: %v", err)
	}

	// Combine the point with itself (point doubling)
	combined, err := protocol.CombinePublicKeys(pub, pub)
	if err != nil {
		t.Fatalf("CombinePublicKeys() error: %v", err)
	}

	// Verify the result is not nil
	if combined == nil {
		t.Fatal("CombinePublicKeys() returned nil")
	}

	// Verify the result is on the curve
	if !isPointOnCurve(combined) {
		t.Error("Combined point is not on the secp256k1 curve")
	}

	// The doubled point should be different from the original (unless it's the point at infinity)
	if pointsEqual(combined, pub) {
		t.Error("Doubled point equals original point, which is unexpected")
	}
}

// TestCombinePublicKeysIdempotent tests that the function is idempotent
// (calling it multiple times with the same inputs produces the same result)
func TestCombinePublicKeysIdempotent(t *testing.T) {
	// Generate two secret shares
	secretA := protocol.GenerateSecretShareFixed(100)
	secretB := protocol.GenerateSecretShareFixed(200)

	// Compute public shares
	pubA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) error: %v", err)
	}

	pubB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) error: %v", err)
	}

	// Call CombinePublicKeys multiple times
	combined1, err := protocol.CombinePublicKeys(pubA, pubB)
	if err != nil {
		t.Fatalf("First CombinePublicKeys() call error: %v", err)
	}

	combined2, err := protocol.CombinePublicKeys(pubA, pubB)
	if err != nil {
		t.Fatalf("Second CombinePublicKeys() call error: %v", err)
	}

	combined3, err := protocol.CombinePublicKeys(pubA, pubB)
	if err != nil {
		t.Fatalf("Third CombinePublicKeys() call error: %v", err)
	}

	// All results should be identical
	if !pointsEqual(combined1, combined2) {
		t.Error("CombinePublicKeys() is not idempotent: first and second calls differ")
	}
	if !pointsEqual(combined2, combined3) {
		t.Error("CombinePublicKeys() is not idempotent: second and third calls differ")
	}
}

// TestCombinePublicKeysCommutative tests that A + B = B + A
func TestCombinePublicKeysCommutative(t *testing.T) {
	// Generate two secret shares
	secretA := protocol.GenerateSecretShareFixed(10)
	secretB := protocol.GenerateSecretShareFixed(20)

	// Compute public shares
	pubA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) error: %v", err)
	}

	pubB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) error: %v", err)
	}

	// Compute A + B and B + A
	combinedAB, err := protocol.CombinePublicKeys(pubA, pubB)
	if err != nil {
		t.Fatalf("CombinePublicKeys(A, B) error: %v", err)
	}

	combinedBA, err := protocol.CombinePublicKeys(pubB, pubA)
	if err != nil {
		t.Fatalf("CombinePublicKeys(B, A) error: %v", err)
	}

	// Results should be identical (commutativity)
	if !pointsEqual(combinedAB, combinedBA) {
		t.Error("CombinePublicKeys() is not commutative: A + B != B + A")
	}
}

// TestCombinePublicKeysMatchesScalarAddition verifies that combining public keys
// matches adding the corresponding private scalars
func TestCombinePublicKeysMatchesScalarAddition(t *testing.T) {
	// Generate two secret shares
	secretA := protocol.GenerateSecretShareFixed(5)
	secretB := protocol.GenerateSecretShareFixed(7)

	// Compute public shares
	pubA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) error: %v", err)
	}

	pubB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) error: %v", err)
	}

	// Combine public keys
	combinedPub, err := protocol.CombinePublicKeys(pubA, pubB)
	if err != nil {
		t.Fatalf("CombinePublicKeys() error: %v", err)
	}

	// Compute the sum of private scalars
	// For secp256k1, the order is approximately 2^256, so we need to handle overflow
	order := secp256k1.S256().N
	secretAScalar := new(big.Int).SetBytes(secretA)
	secretBScalar := new(big.Int).SetBytes(secretB)
	sumScalar := new(big.Int).Add(secretAScalar, secretBScalar)
	sumScalar.Mod(sumScalar, order)

	// Create a combined secret from the sum
	combinedSecret := sumScalar.Bytes()
	// Pad to 32 bytes if necessary
	if len(combinedSecret) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(combinedSecret):], combinedSecret)
		combinedSecret = padded
	}

	// Compute public key from combined secret
	expectedPub, err := protocol.ComputePublicShare(combinedSecret)
	if err != nil {
		t.Fatalf("ComputePublicShare(combined) error: %v", err)
	}

	// The combined public key should match the public key of the sum of scalars
	if !pointsEqual(combinedPub, expectedPub) {
		t.Error("CombinePublicKeys() result does not match scalar addition")
	}
}

// TestCombinePublicKeysNilInputs tests error handling for nil inputs
func TestCombinePublicKeysNilInputs(t *testing.T) {
	tests := []struct {
		name    string
		publicA *secp256k1.PublicKey
		publicB *secp256k1.PublicKey
		wantErr bool
	}{
		{"nil publicA", nil, nil, true},
		{"nil publicB", nil, nil, true},
	}

	// Generate a valid public key for testing
	secret := protocol.GenerateSecretShareFixed(1)
	validPub, _ := protocol.ComputePublicShare(secret)

	tests[0].publicB = validPub
	tests[1].publicA = validPub

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protocol.CombinePublicKeys(tt.publicA, tt.publicB)
			if (err != nil) != tt.wantErr {
				t.Errorf("CombinePublicKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a point is on the secp256k1 curve
func isPointOnCurve(pub *secp256k1.PublicKey) bool {
	if pub == nil {
		return false
	}

	x := pub.X()
	y := pub.Y()
	if x == nil || y == nil {
		return false
	}

	p := secp256k1.S256().P
	b := big.NewInt(7) // secp256k1 has b=7

	// Check y^2 = x^3 + 7 (mod p)
	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, p)

	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Mod(x3, p)
	x3.Add(x3, b)
	x3.Mod(x3, p)

	return y2.Cmp(x3) == 0
}

// Helper function to check if two points are equal
func pointsEqual(a, b *secp256k1.PublicKey) bool {
	if a == nil || b == nil {
		return a == b
	}

	xa := a.X()
	ya := a.Y()
	xb := b.X()
	yb := b.Y()

	if xa == nil || ya == nil || xb == nil || yb == nil {
		return false
	}

	return xa.Cmp(xb) == 0 && ya.Cmp(yb) == 0
}
