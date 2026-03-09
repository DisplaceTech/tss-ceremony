package tests

import (
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/DisplaceTech/tss-ceremony/protocol"
)

// TestGenerateNonceShareValid tests that GenerateNonceShare returns a valid scalar
func TestGenerateNonceShareValid(t *testing.T) {
	n := secp256k1.S256().N

	// Generate multiple nonce shares and verify they are valid
	for i := 0; i < 100; i++ {
		k, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		// Verify k is not nil
		if k == nil {
			t.Errorf("GenerateNonceShare() returned nil scalar")
		}

		// Verify k is in valid range [1, n-1]
		if k.Cmp(big.NewInt(1)) < 0 {
			t.Errorf("k is less than 1: k=%v", k)
		}

		if k.Cmp(n) >= 0 {
			t.Errorf("k >= n: k=%v, n=%v", k, n)
		}
	}
}

// TestGenerateNonceShareRandomness tests that multiple calls produce different values
func TestGenerateNonceShareRandomness(t *testing.T) {
	// Generate 1000 nonce shares and verify they are all unique
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		k, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		kStr := k.String()
		if seen[kStr] {
			t.Errorf("GenerateNonceShare() produced duplicate value: %v", kStr)
		}
		seen[kStr] = true
	}

	// Verify we got 1000 unique values
	if len(seen) != 1000 {
		t.Errorf("Expected 1000 unique values, got %d", len(seen))
	}
}

// TestGenerateNonceShareNeverZero tests that GenerateNonceShare never returns zero
func TestGenerateNonceShareNeverZero(t *testing.T) {
	for i := 0; i < 100; i++ {
		k, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		if k.Cmp(big.NewInt(0)) <= 0 {
			t.Errorf("GenerateNonceShare() returned zero or negative value: %v", k)
		}
	}
}

// TestGenerateNonceShareNeverOrder tests that GenerateNonceShare never returns the order
func TestGenerateNonceShareNeverOrder(t *testing.T) {
	n := secp256k1.S256().N

	for i := 0; i < 100; i++ {
		k, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		if k.Cmp(n) >= 0 {
			t.Errorf("GenerateNonceShare() returned value >= order: k=%v, n=%v", k, n)
		}
	}
}

// TestGenerateNonceShareCanComputePublic tests that generated nonce can be used to compute public point
func TestGenerateNonceShareCanComputePublic(t *testing.T) {
	for i := 0; i < 100; i++ {
		k, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		// Verify the generated nonce can be used to compute a public point
		R, err := protocol.ComputeNoncePublic(k)
		if err != nil {
			t.Errorf("ComputeNoncePublic() failed with generated nonce: %v", err)
		}

		if R == nil {
			t.Errorf("ComputeNoncePublic() returned nil for valid nonce")
		}

		// Verify the point is on the curve
		if !R.IsOnCurve() {
			t.Errorf("Computed public point is not on the curve")
		}
	}
}
