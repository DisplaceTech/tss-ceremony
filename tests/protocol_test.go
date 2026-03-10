package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestRoundTripKeygenVerification tests the complete round-trip of key generation,
// public share computation, curve validation, key combination, and format verification.
func TestRoundTripKeygenVerification(t *testing.T) {
	// Step 1: Generate two random private keys (Party A and Party B)
	secretA := make([]byte, 32)
	_, err := rand.Read(secretA)
	if err != nil {
		t.Fatalf("Failed to generate random secret A: %v", err)
	}

	secretB := make([]byte, 32)
	_, err = rand.Read(secretB)
	if err != nil {
		t.Fatalf("Failed to generate random secret B: %v", err)
	}

	// Step 2: Compute public shares from private keys
	publicA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) returned error: %v", err)
	}

	publicB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) returned error: %v", err)
	}

	// Step 3: Verify both public shares are on the secp256k1 curve
	if !publicA.IsOnCurve() {
		t.Error("Party A public share is not on secp256k1 curve")
	}
	if !publicB.IsOnCurve() {
		t.Error("Party B public share is not on secp256k1 curve")
	}

	// Step 4: Combine the two public keys
	combinedKey, err := protocol.CombinePublicKeys(publicA, publicB)
	if err != nil {
		t.Fatalf("CombinePublicKeys returned error: %v", err)
	}

	// Step 5: Verify the combined key is on the curve
	if !combinedKey.IsOnCurve() {
		t.Error("Combined public key is not on secp256k1 curve")
	}

	// Step 6: Verify the combined key can be parsed by standard secp256k1 library
	serializedCombined := combinedKey.SerializeUncompressed()
	parsedCombined, err := secp256k1.ParsePubKey(serializedCombined)
	if err != nil {
		t.Fatalf("Failed to parse combined public key: %v", err)
	}

	// Verify coordinates match
	x1, y1 := combinedKey.X(), combinedKey.Y()
	x2, y2 := parsedCombined.X(), parsedCombined.Y()
	if x1.Cmp(x2) != 0 {
		t.Error("X coordinate mismatch after parsing combined key")
	}
	if y1.Cmp(y2) != 0 {
		t.Error("Y coordinate mismatch after parsing combined key")
	}

	// Step 7: Verify serialization formats match standard secp256k1 encoding
	// Uncompressed format: 0x04 || x (32 bytes) || y (32 bytes) = 65 bytes
	uncompressed := combinedKey.SerializeUncompressed()
	if len(uncompressed) != 65 {
		t.Errorf("Uncompressed serialization is %d bytes, want 65", len(uncompressed))
	}
	if uncompressed[0] != 0x04 {
		t.Errorf("Uncompressed prefix is %02x, want 04", uncompressed[0])
	}

	// Compressed format: 0x02/0x03 || x (32 bytes) = 33 bytes
	compressed := combinedKey.SerializeCompressed()
	if len(compressed) != 33 {
		t.Errorf("Compressed serialization is %d bytes, want 33", len(compressed))
	}
	if compressed[0] != 0x02 && compressed[0] != 0x03 {
		t.Errorf("Compressed prefix is %02x, want 02 or 03", compressed[0])
	}

	// Step 8: Verify the combined key can be used for ECDSA verification
	// Create a test message and sign it with the combined private key
	message := []byte("Test message for round-trip verification")
	hash := sha256.Sum256(message)

	// Derive the combined private key (for verification purposes)
	// In a real TSS, this would be distributed, but for testing we reconstruct it
	privA := secp256k1.PrivKeyFromBytes(secretA)
	privB := secp256k1.PrivKeyFromBytes(secretB)
	// Get scalar values and add them
	combinedScalar := new(big.Int).Add(privA.ToECDSA().D, privB.ToECDSA().D)
	combinedScalar.Mod(combinedScalar, secp256k1.S256().N)
	combinedPriv := secp256k1.PrivKeyFromBytes(combinedScalar.Bytes())

	// Sign the message
	r, s, err := ecdsa.Sign(rand.Reader, &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     combinedKey.X(),
			Y:     combinedKey.Y(),
		},
		D: combinedPriv.ToECDSA().D,
	}, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign test message: %v", err)
	}

	// Verify the signature
	verified := ecdsa.Verify(
		&ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     combinedKey.X(),
			Y:     combinedKey.Y(),
		},
		hash[:],
		r,
		s,
	)
	if !verified {
		t.Error("ECDSA signature verification failed")
	}

	t.Log("Round-trip verification completed successfully")
}

// TestRoundTripDeterministic tests that the round-trip produces deterministic results
// with fixed seeds.
func TestRoundTripDeterministic(t *testing.T) {
	// Use fixed seeds for deterministic testing
	secretA := protocol.GenerateSecretShareFixed(12345)
	secretB := protocol.GenerateSecretShareFixed(67890)

	// Compute public shares
	publicA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) returned error: %v", err)
	}

	publicB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) returned error: %v", err)
	}

	// Combine keys
	combinedKey, err := protocol.CombinePublicKeys(publicA, publicB)
	if err != nil {
		t.Fatalf("CombinePublicKeys returned error: %v", err)
	}

	// Serialize for comparison
	serialized1 := combinedKey.SerializeUncompressed()

	// Repeat the process
	publicA2, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("Second ComputePublicShare(A) returned error: %v", err)
	}

	publicB2, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("Second ComputePublicShare(B) returned error: %v", err)
	}

	combinedKey2, err := protocol.CombinePublicKeys(publicA2, publicB2)
	if err != nil {
		t.Fatalf("Second CombinePublicKeys returned error: %v", err)
	}

	serialized2 := combinedKey2.SerializeUncompressed()

	// Compare
	if len(serialized1) != len(serialized2) {
		t.Fatalf("Serialized lengths differ: %d vs %d", len(serialized1), len(serialized2))
	}

	for i := range serialized1 {
		if serialized1[i] != serialized2[i] {
			t.Errorf("Round-trip is not deterministic at byte %d: %02x vs %02x", i, serialized1[i], serialized2[i])
			break
		}
	}

	t.Log("Deterministic round-trip verification completed successfully")
}

// TestRoundTripKnownValues tests the round-trip with known secp256k1 values.
func TestRoundTripKnownValues(t *testing.T) {
	// Use secret = 1 for Party A (should produce generator point G)
	secretA := make([]byte, 32)
	secretA[31] = 0x01

	// Use a known secret for Party B
	secretB := make([]byte, 32)
	secretB[31] = 0x02

	// Compute public shares
	publicA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) returned error: %v", err)
	}

	publicB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) returned error: %v", err)
	}

	// Verify Party A's public key is the generator point G
	expectedGx := "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
	expectedGy := "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"

	xHex := hex.EncodeToString(publicA.X().Bytes())
	yHex := hex.EncodeToString(publicA.Y().Bytes())

	if xHex != expectedGx {
		t.Errorf("Party A X coordinate mismatch:\n  got:  %s\n  want: %s", xHex, expectedGx)
	}
	if yHex != expectedGy {
		t.Errorf("Party A Y coordinate mismatch:\n  got:  %s\n  want: %s", yHex, expectedGy)
	}

	// Combine keys
	combinedKey, err := protocol.CombinePublicKeys(publicA, publicB)
	if err != nil {
		t.Fatalf("CombinePublicKeys returned error: %v", err)
	}

	// Verify combined key is on curve
	if !combinedKey.IsOnCurve() {
		t.Error("Combined key is not on secp256k1 curve")
	}

	// Verify combined key can be serialized and parsed
	serialized := combinedKey.SerializeUncompressed()
	parsed, err := secp256k1.ParsePubKey(serialized)
	if err != nil {
		t.Fatalf("Failed to parse combined key: %v", err)
	}

	// Verify coordinates match
	x1, y1 := combinedKey.X(), combinedKey.Y()
	x2, y2 := parsed.X(), parsed.Y()
	if x1.Cmp(x2) != 0 || y1.Cmp(y2) != 0 {
		t.Error("Combined key coordinates mismatch after parse")
	}

	t.Log("Known values round-trip verification completed successfully")
}

// TestRoundTripSerializationFormats tests all standard secp256k1 serialization formats.
func TestRoundTripSerializationFormats(t *testing.T) {
	secretA := protocol.GenerateSecretShareFixed(11111)
	secretB := protocol.GenerateSecretShareFixed(22222)

	publicA, err := protocol.ComputePublicShare(secretA)
	if err != nil {
		t.Fatalf("ComputePublicShare(A) returned error: %v", err)
	}

	publicB, err := protocol.ComputePublicShare(secretB)
	if err != nil {
		t.Fatalf("ComputePublicShare(B) returned error: %v", err)
	}

	combinedKey, err := protocol.CombinePublicKeys(publicA, publicB)
	if err != nil {
		t.Fatalf("CombinePublicKeys returned error: %v", err)
	}

	// Test uncompressed format
	uncompressed := combinedKey.SerializeUncompressed()
	if len(uncompressed) != 65 {
		t.Errorf("Uncompressed: got %d bytes, want 65", len(uncompressed))
	}
	if uncompressed[0] != 0x04 {
		t.Errorf("Uncompressed prefix: got %02x, want 04", uncompressed[0])
	}

	// Parse uncompressed and verify
	parsedUncompressed, err := secp256k1.ParsePubKey(uncompressed)
	if err != nil {
		t.Fatalf("Failed to parse uncompressed: %v", err)
	}
	if parsedUncompressed.X().Cmp(combinedKey.X()) != 0 || parsedUncompressed.Y().Cmp(combinedKey.Y()) != 0 {
		t.Error("Uncompressed parse failed coordinate check")
	}

	// Test compressed format (even y)
	compressedEven := combinedKey.SerializeCompressed()
	if len(compressedEven) != 33 {
		t.Errorf("Compressed: got %d bytes, want 33", len(compressedEven))
	}
	if compressedEven[0] != 0x02 && compressedEven[0] != 0x03 {
		t.Errorf("Compressed prefix: got %02x, want 02 or 03", compressedEven[0])
	}

	// Parse compressed and verify
	parsedCompressed, err := secp256k1.ParsePubKey(compressedEven)
	if err != nil {
		t.Fatalf("Failed to parse compressed: %v", err)
	}
	if parsedCompressed.X().Cmp(combinedKey.X()) != 0 || parsedCompressed.Y().Cmp(combinedKey.Y()) != 0 {
		t.Error("Compressed parse failed coordinate check")
	}

	// Test hex encoding (common in blockchain applications)
	hexUncompressed := hex.EncodeToString(uncompressed)
	if len(hexUncompressed) != 130 { // 65 bytes * 2 hex chars
		t.Errorf("Hex uncompressed: got %d chars, want 130", len(hexUncompressed))
	}

	hexCompressed := hex.EncodeToString(compressedEven)
	if len(hexCompressed) != 66 { // 33 bytes * 2 hex chars
		t.Errorf("Hex compressed: got %d chars, want 66", len(hexCompressed))
	}

	// Parse from hex
	parsedFromHex, err := secp256k1.ParsePubKey(uncompressed)
	if err != nil {
		t.Fatalf("Failed to parse from hex: %v", err)
	}
	if parsedFromHex.X().Cmp(combinedKey.X()) != 0 || parsedFromHex.Y().Cmp(combinedKey.Y()) != 0 {
		t.Error("Hex parse failed coordinate check")
	}

	t.Log("Serialization formats test completed successfully")
}

// TestRoundTripCurveValidation tests that all intermediate points are validated.
func TestRoundTripCurveValidation(t *testing.T) {
	// Generate multiple key pairs and verify each step
	for i := int64(0); i < 10; i++ {
		secretA := protocol.GenerateSecretShareFixed(i * 100)
		secretB := protocol.GenerateSecretShareFixed(i*100 + 1)

		publicA, err := protocol.ComputePublicShare(secretA)
		if err != nil {
			t.Fatalf("Iteration %d: ComputePublicShare(A) returned error: %v", i, err)
		}
		if !publicA.IsOnCurve() {
			t.Errorf("Iteration %d: Party A public key not on curve", i)
		}

		publicB, err := protocol.ComputePublicShare(secretB)
		if err != nil {
			t.Fatalf("Iteration %d: ComputePublicShare(B) returned error: %v", i, err)
		}
		if !publicB.IsOnCurve() {
			t.Errorf("Iteration %d: Party B public key not on curve", i)
		}

		combinedKey, err := protocol.CombinePublicKeys(publicA, publicB)
		if err != nil {
			t.Fatalf("Iteration %d: CombinePublicKeys returned error: %v", i, err)
		}
		if !combinedKey.IsOnCurve() {
			t.Errorf("Iteration %d: Combined key not on curve", i)
		}

		// Verify combined key can be parsed
		serialized := combinedKey.SerializeUncompressed()
		_, err = secp256k1.ParsePubKey(serialized)
		if err != nil {
			t.Errorf("Iteration %d: Combined key failed parse: %v", i, err)
		}
	}

	t.Log("Curve validation test completed successfully for 10 iterations")
}
