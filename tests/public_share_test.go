package tests

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestComputePublicShareValidSecret tests that ComputePublicShare produces a valid
// secp256k1 point for a valid 32-byte secret scalar.
func TestComputePublicShareValidSecret(t *testing.T) {
	// Generate a random 32-byte secret
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		t.Fatalf("Failed to generate random secret: %v", err)
	}

	// Compute public share
	publicKey, err := protocol.ComputePublicShare(secret)
	if err != nil {
		t.Fatalf("ComputePublicShare() returned error: %v", err)
	}

	// Verify the public key is not nil
	if publicKey == nil {
		t.Fatal("ComputePublicShare() returned nil public key")
	}

	// Verify the public key has valid coordinates
	x := publicKey.X()
	y := publicKey.Y()
	if x == nil || y == nil {
		t.Fatal("Public key has nil coordinates")
	}

	// Verify the point is on the secp256k1 curve: y^2 = x^3 + 7 (mod p)
	p := secp256k1.S256().P
	b := big.NewInt(7)

	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, p)

	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Mod(x3, p)
	x3.Add(x3, b)
	x3.Mod(x3, p)

	if y2.Cmp(x3) != 0 {
		t.Error("Computed public key point is not on the secp256k1 curve")
	}
}

// TestComputePublicShareDeterministic tests that ComputePublicShare produces
// the same output for the same input (deterministic behavior).
func TestComputePublicShareDeterministic(t *testing.T) {
	// Use a fixed seed for deterministic testing
	secret := protocol.GenerateSecretShareFixed(42)

	// Compute public share twice
	publicKey1, err := protocol.ComputePublicShare(secret)
	if err != nil {
		t.Fatalf("First ComputePublicShare() call returned error: %v", err)
	}

	publicKey2, err := protocol.ComputePublicShare(secret)
	if err != nil {
		t.Fatalf("Second ComputePublicShare() call returned error: %v", err)
	}

	// Compare serialized forms
	serialized1 := publicKey1.SerializeUncompressed()
	serialized2 := publicKey2.SerializeUncompressed()

	if len(serialized1) != len(serialized2) {
		t.Fatalf("Serialized lengths differ: %d vs %d", len(serialized1), len(serialized2))
	}

	for i := range serialized1 {
		if serialized1[i] != serialized2[i] {
			t.Errorf("ComputePublicShare() is not deterministic at byte %d: %02x vs %02x", i, serialized1[i], serialized2[i])
			break
		}
	}
}

// TestComputePublicShareZeroSecret tests edge case with zero secret scalar.
func TestComputePublicShareZeroSecret(t *testing.T) {
	// Create a zero secret (all bytes are 0)
	secret := make([]byte, 32)

	// Compute public share - the library may or may not accept this
	// depending on implementation. We just verify it doesn't panic.
	_, err := protocol.ComputePublicShare(secret)
	// Note: Some implementations accept 0, some reject it.
	// We don't fail the test either way - just log the behavior.
	if err != nil {
		t.Logf("ComputePublicShare() rejected zero secret scalar: %v", err)
	}
}

// TestComputePublicShareInvalidLength tests that ComputePublicShare rejects
// secrets that are not exactly 32 bytes.
func TestComputePublicShareInvalidLength(t *testing.T) {
	tests := []struct {
		name   string
		secret []byte
	}{
		{"empty secret", make([]byte, 0)},
		{"31-byte secret", make([]byte, 31)},
		{"33-byte secret", make([]byte, 33)},
		{"16-byte secret", make([]byte, 16)},
		{"64-byte secret", make([]byte, 64)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protocol.ComputePublicShare(tt.secret)
			if err == nil {
				t.Errorf("ComputePublicShare() should reject %d-byte secret", len(tt.secret))
			}
		})
	}
}

// TestComputePublicShareKnownValue tests ComputePublicShare against a known
// secp256k1 generator point multiplication.
func TestComputePublicShareKnownValue(t *testing.T) {
	// Use secret = 1, which should produce the generator point G
	secret := make([]byte, 32)
	secret[31] = 0x01 // Set the last byte to 1 (little-endian representation)

	publicKey, err := protocol.ComputePublicShare(secret)
	if err != nil {
		t.Fatalf("ComputePublicShare() returned error for secret=1: %v", err)
	}

	// The generator point G for secp256k1 is:
	// Gx = 0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798
	// Gy = 0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8
	expectedGx := "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
	expectedGy := "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"

	xHex := hex.EncodeToString(publicKey.X().Bytes())
	yHex := hex.EncodeToString(publicKey.Y().Bytes())

	// Compare with expected generator point (case-insensitive)
	if xHex != expectedGx {
		t.Errorf("X coordinate mismatch:\n  got:  %s\n  want: %s", xHex, expectedGx)
	}
	if yHex != expectedGy {
		t.Errorf("Y coordinate mismatch:\n  got:  %s\n  want: %s", yHex, expectedGy)
	}
}

// TestComputePublicShareSerializationFormat tests that the returned public key
// can be properly serialized in both compressed and uncompressed formats.
func TestComputePublicShareSerializationFormat(t *testing.T) {
	secret := protocol.GenerateSecretShareFixed(12345)

	publicKey, err := protocol.ComputePublicShare(secret)
	if err != nil {
		t.Fatalf("ComputePublicShare() returned error: %v", err)
	}

	// Test uncompressed serialization (should be 65 bytes: 0x04 || x || y)
	uncompressed := publicKey.SerializeUncompressed()
	if len(uncompressed) != 65 {
		t.Errorf("Uncompressed serialization is %d bytes, want 65", len(uncompressed))
	}
	if uncompressed[0] != 0x04 {
		t.Errorf("Uncompressed serialization prefix is %02x, want 04", uncompressed[0])
	}

	// Test compressed serialization (should be 33 bytes: 0x02/0x03 || x)
	compressed := publicKey.SerializeCompressed()
	if len(compressed) != 33 {
		t.Errorf("Compressed serialization is %d bytes, want 33", len(compressed))
	}
	if compressed[0] != 0x02 && compressed[0] != 0x03 {
		t.Errorf("Compressed serialization prefix is %02x, want 02 or 03", compressed[0])
	}
}

// TestComputePublicShareMultipleSecrets tests that different secrets produce
// different public keys.
func TestComputePublicShareMultipleSecrets(t *testing.T) {
	publicKeys := make(map[string]bool)

	for i := int64(0); i < 100; i++ {
		secret := protocol.GenerateSecretShareFixed(i)
		publicKey, err := protocol.ComputePublicShare(secret)
		if err != nil {
			t.Fatalf("ComputePublicShare() returned error for seed %d: %v", i, err)
		}

		serialized := hex.EncodeToString(publicKey.SerializeUncompressed())
		if publicKeys[serialized] {
			t.Errorf("ComputePublicShare() produced duplicate public key for different secrets")
			break
		}
		publicKeys[serialized] = true
	}
}

// TestComputePublicShareMaxValue tests edge case with maximum valid secret scalar.
func TestComputePublicShareMaxValue(t *testing.T) {
	// Use a large secret value (close to the curve order)
	// The secp256k1 curve order is approximately 2^256 - 2^128 - ...
	// We'll use a value that's definitely valid but large
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = 0xFF
	}

	// This might fail if the value exceeds the curve order, which is expected behavior
	_, err := protocol.ComputePublicShare(secret)
	// Note: The library may or may not accept this value depending on implementation
	// We just verify it doesn't panic
	if err != nil {
		t.Logf("ComputePublicShare() correctly rejected max value: %v", err)
	}
}

// TestComputePublicShareRoundTrip tests that we can serialize and deserialize
// the public key without losing information.
func TestComputePublicShareRoundTrip(t *testing.T) {
	secret := protocol.GenerateSecretShareFixed(999)

	publicKey, err := protocol.ComputePublicShare(secret)
	if err != nil {
		t.Fatalf("ComputePublicShare() returned error: %v", err)
	}

	// Serialize and deserialize
	serialized := publicKey.SerializeUncompressed()
	deserialized, err := secp256k1.ParsePubKey(serialized)
	if err != nil {
		t.Fatalf("Failed to parse serialized public key: %v", err)
	}

	// Compare coordinates
	x1, y1 := publicKey.X(), publicKey.Y()
	x2, y2 := deserialized.X(), deserialized.Y()

	if x1.Cmp(x2) != 0 {
		t.Error("X coordinate mismatch after round-trip serialization")
	}
	if y1.Cmp(y2) != 0 {
		t.Error("Y coordinate mismatch after round-trip serialization")
	}
}
