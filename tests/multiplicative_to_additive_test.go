package tests

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestMultiplicativeToAdditive_Basic verifies that alpha + beta = k_a * k_b mod n
func TestMultiplicativeToAdditive_Basic(t *testing.T) {
	n := secp256k1.S256().N

	// Generate random k_a and k_b in [1, n-1]
	ka, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
	if err != nil {
		t.Fatalf("failed to generate k_a: %v", err)
	}
	ka.Add(ka, big.NewInt(1))

	kb, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
	if err != nil {
		t.Fatalf("failed to generate k_b: %v", err)
	}
	kb.Add(kb, big.NewInt(1))

	// Call the function
	alpha, beta, err := protocol.MultiplicativeToAdditive(ka, kb)
	if err != nil {
		t.Fatalf("MultiplicativeToAdditive failed: %v", err)
	}

	// Verify alpha + beta = k_a * k_b mod n
	expectedProduct := new(big.Int).Mul(ka, kb)
	expectedProduct.Mod(expectedProduct, n)

	actualSum := new(big.Int).Add(alpha, beta)
	actualSum.Mod(actualSum, n)

	if actualSum.Cmp(expectedProduct) != 0 {
		t.Errorf("alpha + beta != k_a * k_b mod n\n  expected: %s\n  got:      %s",
			expectedProduct.Text(16), actualSum.Text(16))
	}
}

// TestMultiplicativeToAdditive_RandomInputs runs multiple iterations with random inputs
func TestMultiplicativeToAdditive_RandomInputs(t *testing.T) {
	n := secp256k1.S256().N

	// Run multiple iterations with random inputs
	for i := 0; i < 100; i++ {
		ka, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
		if err != nil {
			t.Fatalf("iteration %d: failed to generate k_a: %v", i, err)
		}
		ka.Add(ka, big.NewInt(1))

		kb, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
		if err != nil {
			t.Fatalf("iteration %d: failed to generate k_b: %v", i, err)
		}
		kb.Add(kb, big.NewInt(1))

		alpha, beta, err := protocol.MultiplicativeToAdditive(ka, kb)
		if err != nil {
			t.Fatalf("iteration %d: MultiplicativeToAdditive failed: %v", i, err)
		}

		expectedProduct := new(big.Int).Mul(ka, kb)
		expectedProduct.Mod(expectedProduct, n)

		actualSum := new(big.Int).Add(alpha, beta)
		actualSum.Mod(actualSum, n)

		if actualSum.Cmp(expectedProduct) != 0 {
			t.Errorf("iteration %d: alpha + beta != k_a * k_b mod n", i)
		}
	}
}

// TestMultiplicativeToAdditive_EdgeCases tests edge cases
func TestMultiplicativeToAdditive_EdgeCases(t *testing.T) {
	n := secp256k1.S256().N

	// Test with k_a = 1, k_b = 1
	ka := big.NewInt(1)
	kb := big.NewInt(1)
	alpha, beta, err := protocol.MultiplicativeToAdditive(ka, kb)
	if err != nil {
		t.Fatalf("MultiplicativeToAdditive failed: %v", err)
	}

	// Verify alpha + beta = 1 mod n
	sum := new(big.Int).Add(alpha, beta)
	sum.Mod(sum, n)
	if sum.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("alpha + beta != 1 mod n, got %s", sum.Text(16))
	}

	// Test with k_a = n-1, k_b = n-1
	ka = new(big.Int).Sub(n, big.NewInt(1))
	kb = new(big.Int).Sub(n, big.NewInt(1))
	alpha, beta, err = protocol.MultiplicativeToAdditive(ka, kb)
	if err != nil {
		t.Fatalf("MultiplicativeToAdditive failed: %v", err)
	}

	// Verify alpha + beta = (n-1)*(n-1) mod n = 1 mod n
	expectedProduct := new(big.Int).Mul(ka, kb)
	expectedProduct.Mod(expectedProduct, n)
	sum = new(big.Int).Add(alpha, beta)
	sum.Mod(sum, n)
	if sum.Cmp(expectedProduct) != 0 {
		t.Errorf("alpha + beta != (n-1)*(n-1) mod n\n  expected: %s\n  got:      %s",
			expectedProduct.Text(16), sum.Text(16))
	}
}

// TestMultiplicativeToAdditive_InvalidInputs tests that invalid inputs are rejected
func TestMultiplicativeToAdditive_InvalidInputs(t *testing.T) {
	n := secp256k1.S256().N

	// Test with k_a = 0 (should fail)
	_, _, err := protocol.MultiplicativeToAdditive(big.NewInt(0), big.NewInt(1))
	if err == nil {
		t.Error("expected error for k_a = 0, got nil")
	}

	// Test with k_b = 0 (should fail)
	_, _, err = protocol.MultiplicativeToAdditive(big.NewInt(1), big.NewInt(0))
	if err == nil {
		t.Error("expected error for k_b = 0, got nil")
	}

	// Test with k_a >= n (should fail)
	_, _, err = protocol.MultiplicativeToAdditive(n, big.NewInt(1))
	if err == nil {
		t.Error("expected error for k_a >= n, got nil")
	}

	// Test with k_b >= n (should fail)
	_, _, err = protocol.MultiplicativeToAdditive(big.NewInt(1), n)
	if err == nil {
		t.Error("expected error for k_b >= n, got nil")
	}
}

// TestMultiplicativeToAdditive_ShareRange verifies that alpha and beta are in valid range
func TestMultiplicativeToAdditive_ShareRange(t *testing.T) {
	n := secp256k1.S256().N

	// Generate random k_a and k_b in [1, n-1]
	ka, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
	if err != nil {
		t.Fatalf("failed to generate k_a: %v", err)
	}
	ka.Add(ka, big.NewInt(1))

	kb, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
	if err != nil {
		t.Fatalf("failed to generate k_b: %v", err)
	}
	kb.Add(kb, big.NewInt(1))

	// Call the function
	alpha, beta, err := protocol.MultiplicativeToAdditive(ka, kb)
	if err != nil {
		t.Fatalf("MultiplicativeToAdditive failed: %v", err)
	}

	// Verify alpha and beta are in valid range [0, n-1]
	if alpha.Cmp(big.NewInt(0)) < 0 || alpha.Cmp(n) >= 0 {
		t.Errorf("alpha %s is out of range [0, n-1]", alpha.Text(16))
	}
	if beta.Cmp(big.NewInt(0)) < 0 || beta.Cmp(n) >= 0 {
		t.Errorf("beta %s is out of range [0, n-1]", beta.Text(16))
	}
}
