package tests

import (
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
)

// TestGenerateSecretShareLength tests that the returned slice is exactly 32 bytes
func TestGenerateSecretShareLength(t *testing.T) {
	secret, err := protocol.GenerateSecretShare()
	if err != nil {
		t.Fatalf("GenerateSecretShare() error: %v", err)
	}

	if len(secret) != 32 {
		t.Errorf("GenerateSecretShare() returned %d bytes, expected 32", len(secret))
	}
}

// TestGenerateSecretShareNoPanic tests that the function does not panic
func TestGenerateSecretShareNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GenerateSecretShare() panicked: %v", r)
		}
	}()

	_, err := protocol.GenerateSecretShare()
	if err != nil {
		t.Fatalf("GenerateSecretShare() error: %v", err)
	}
}

// TestGenerateSecretShareRandomness tests that multiple calls produce different values
func TestGenerateSecretShareRandomness(t *testing.T) {
	secrets := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		secret, err := protocol.GenerateSecretShare()
		if err != nil {
			t.Fatalf("GenerateSecretShare() call %d error: %v", i, err)
		}
		secrets[i] = secret
	}

	// Check that all secrets are unique
	for i := 0; i < len(secrets); i++ {
		for j := i + 1; j < len(secrets); j++ {
			if slicesEqual(secrets[i], secrets[j]) {
				t.Errorf("GenerateSecretShare() produced duplicate values at indices %d and %d", i, j)
			}
		}
	}
}

// TestGenerateSecretShareNonZero tests that the generated secret is not all zeros
func TestGenerateSecretShareNonZero(t *testing.T) {
	secret, err := protocol.GenerateSecretShare()
	if err != nil {
		t.Fatalf("GenerateSecretShare() error: %v", err)
	}

	allZeros := true
	for _, b := range secret {
		if b != 0 {
			allZeros = false
			break
		}
	}

	if allZeros {
		t.Error("GenerateSecretShare() returned all zeros, which is highly unlikely and indicates a problem")
	}
}

// TestGenerateSecretShareValidBytes tests that all bytes are valid (0-255)
func TestGenerateSecretShareValidBytes(t *testing.T) {
	secret, err := protocol.GenerateSecretShare()
	if err != nil {
		t.Fatalf("GenerateSecretShare() error: %v", err)
	}

	for i, b := range secret {
		if b < 0 || b > 255 {
			t.Errorf("GenerateSecretShare() returned invalid byte %d at index %d", b, i)
		}
	}
}

// Helper function to compare two byte slices
func slicesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
