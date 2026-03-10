package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestSchnorrSignatureCompatibility verifies that FROST signatures can be verified
// using the standard Schnorr verification logic.
func TestSchnorrSignatureCompatibility(t *testing.T) {
	// Create a FROST signer in fixed mode for deterministic testing
	config := protocol.FROSTConfig{
		Fixed:   true,
		Message: []byte("Hello, FROST!"),
	}

	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create FROST signer: %v", err)
	}

	// Sign the message
	message := []byte("Hello, FROST!")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get the combined public key
	publicKey := signer.GetCombinedPublicKey()
	if publicKey == nil {
		t.Fatal("Combined public key is nil")
	}

	// Verify the signature using the standard Schnorr verification function
	valid, err := protocol.VerifySchnorrSignature(publicKey, sig.R, sig.S, message)
	if err != nil {
		t.Fatalf("VerifySchnorrSignature returned error: %v", err)
	}
	if !valid {
		t.Error("Schnorr signature verification failed")
	}
}

// TestSchnorrSignatureInvalidRejects verifies that invalid Schnorr signatures are rejected.
func TestSchnorrSignatureInvalidRejects(t *testing.T) {
	// Create a FROST signer
	config := protocol.FROSTConfig{
		Fixed:   true,
		Message: []byte("Hello, FROST!"),
	}

	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create FROST signer: %v", err)
	}

	// Sign the message
	message := []byte("Hello, FROST!")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	publicKey := signer.GetCombinedPublicKey()

	// Test 1: Verify with wrong message
	wrongMessage := []byte("Wrong message")
	valid, err := protocol.VerifySchnorrSignature(publicKey, sig.R, sig.S, wrongMessage)
	if err != nil {
		t.Fatalf("VerifySchnorrSignature returned error: %v", err)
	}
	if valid {
		t.Error("Schnorr verification should reject signature with wrong message")
	}

	// Test 2: Verify with tampered R
	tamperedR := secp256k1.PrivKeyFromBytes([]byte{1}).PubKey()
	valid, err = protocol.VerifySchnorrSignature(publicKey, tamperedR, sig.S, message)
	if err != nil {
		t.Fatalf("VerifySchnorrSignature returned error: %v", err)
	}
	if valid {
		t.Error("Schnorr verification should reject signature with tampered R")
	}

	// Test 3: Verify with tampered S
	tamperedS := big.NewInt(12345)
	valid, err = protocol.VerifySchnorrSignature(publicKey, sig.R, tamperedS, message)
	if err != nil {
		t.Fatalf("VerifySchnorrSignature returned error: %v", err)
	}
	if valid {
		t.Error("Schnorr verification should reject signature with tampered S")
	}
}

// TestSchnorrSignatureRandomMode verifies that random mode also produces valid signatures.
func TestSchnorrSignatureRandomMode(t *testing.T) {
	// Create a FROST signer in random mode
	config := protocol.FROSTConfig{
		Fixed:   false,
		Message: []byte("Hello, FROST!"),
	}

	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create FROST signer: %v", err)
	}

	// Sign the message
	message := []byte("Hello, FROST!")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	publicKey := signer.GetCombinedPublicKey()

	// Verify the signature
	valid, err := protocol.VerifySchnorrSignature(publicKey, sig.R, sig.S, message)
	if err != nil {
		t.Fatalf("VerifySchnorrSignature returned error: %v", err)
	}
	if !valid {
		t.Error("Schnorr signature verification failed for random mode")
	}
}

// TestSchnorrSignatureDeterminism verifies that fixed mode produces deterministic signatures.
func TestSchnorrSignatureDeterminism(t *testing.T) {
	// Create two FROST signers in fixed mode
	config := protocol.FROSTConfig{
		Fixed:   true,
		Message: []byte("Hello, FROST!"),
	}

	signer1, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create first FROST signer: %v", err)
	}

	signer2, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create second FROST signer: %v", err)
	}

	// Sign the same message
	message := []byte("Hello, FROST!")
	sig1, err := signer1.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign with first signer: %v", err)
	}

	sig2, err := signer2.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign with second signer: %v", err)
	}

	// Compare public keys
	pub1 := signer1.GetCombinedPublicKey()
	pub2 := signer2.GetCombinedPublicKey()

	if pub1.X().Cmp(pub2.X()) != 0 || pub1.Y().Cmp(pub2.Y()) != 0 {
		t.Error("Combined public keys should be identical in fixed mode")
	}

	// Compare signature R points
	if sig1.R.X().Cmp(sig2.R.X()) != 0 || sig1.R.Y().Cmp(sig2.R.Y()) != 0 {
		t.Error("Signature R points should be identical in fixed mode")
	}

	// Compare signature S values
	if sig1.S.Cmp(sig2.S) != 0 {
		t.Error("Signature S values should be identical in fixed mode")
	}
}

// TestSchnorrSignatureHexEncoding verifies that signature components can be properly
// encoded and decoded as hex strings.
func TestSchnorrSignatureHexEncoding(t *testing.T) {
	config := protocol.FROSTConfig{
		Fixed:   true,
		Message: []byte("Hello, FROST!"),
	}

	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create FROST signer: %v", err)
	}

	message := []byte("Hello, FROST!")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	publicKey := signer.GetCombinedPublicKey()

	// Encode R as hex
	RBytes := sig.R.SerializeUncompressed()
	RHex := hex.EncodeToString(RBytes)

	// Encode S as hex
	SBytes := sig.S.Bytes()
	SHex := hex.EncodeToString(SBytes)

	// Decode R from hex
	RBytesDecoded, err := hex.DecodeString(RHex)
	if err != nil {
		t.Fatalf("Failed to decode R from hex: %v", err)
	}
	RDecoded, err := secp256k1.ParsePubKey(RBytesDecoded)
	if err != nil {
		t.Fatalf("Failed to parse R from decoded bytes: %v", err)
	}

	// Decode S from hex
	SBytesDecoded, err := hex.DecodeString(SHex)
	if err != nil {
		t.Fatalf("Failed to decode S from hex: %v", err)
	}
	SDecoded := new(big.Int).SetBytes(SBytesDecoded)

	// Verify with decoded values
	valid, err := protocol.VerifySchnorrSignature(publicKey, RDecoded, SDecoded, message)
	if err != nil {
		t.Fatalf("VerifySchnorrSignature returned error: %v", err)
	}
	if !valid {
		t.Error("Schnorr signature verification failed after hex encoding/decoding")
	}
}

// TestSchnorrSignatureWithHashedMessage verifies that signatures work with hashed messages.
func TestSchnorrSignatureWithHashedMessage(t *testing.T) {
	config := protocol.FROSTConfig{
		Fixed:   true,
		Message: []byte("Hello, FROST!"),
	}

	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create FROST signer: %v", err)
	}

	// Use a hashed message
	originalMessage := []byte("Hello, FROST!")
	hashedMessage := sha256.Sum256(originalMessage)

	sig, err := signer.Sign(hashedMessage[:])
	if err != nil {
		t.Fatalf("Failed to sign hashed message: %v", err)
	}

	publicKey := signer.GetCombinedPublicKey()

	// Verify the signature
	valid, err := protocol.VerifySchnorrSignature(publicKey, sig.R, sig.S, hashedMessage[:])
	if err != nil {
		t.Fatalf("VerifySchnorrSignature returned error: %v", err)
	}
	if !valid {
		t.Error("Schnorr signature verification failed for hashed message")
	}
}

// TestSchnorrSignatureFormat verifies that the signature format matches the
// secp256k1 Schnorr standard: 32-byte r (x-coordinate of R) || 32-byte s.
func TestSchnorrSignatureFormat(t *testing.T) {
	config := protocol.FROSTConfig{
		Fixed:   true,
		Message: []byte("Hello, FROST!"),
	}

	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("Failed to create FROST signer: %v", err)
	}

	message := []byte("Hello, FROST!")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Serialize the signature to standard 64-byte format
	sigBytes, err := protocol.SerializeSchnorrSignature(sig.R, sig.S)
	if err != nil {
		t.Fatalf("SerializeSchnorrSignature failed: %v", err)
	}

	// Verify the serialized signature is exactly 64 bytes
	if len(sigBytes) != protocol.SchnorrSigSize {
		t.Errorf("Serialized signature has length %d, expected %d", len(sigBytes), protocol.SchnorrSigSize)
	}

	// Validate the encoding
	if err := protocol.ValidateSchnorrEncoding(sigBytes); err != nil {
		t.Errorf("ValidateSchnorrEncoding failed: %v", err)
	}

	// Parse the signature back
	r, s, err := protocol.ParseSchnorrSignature(sigBytes)
	if err != nil {
		t.Fatalf("ParseSchnorrSignature failed: %v", err)
	}

	// Verify parsed values match original
	if r.Cmp(sig.R.X()) != 0 {
		t.Error("Parsed r does not match original R x-coordinate")
	}
	if s.Cmp(sig.S) != 0 {
		t.Error("Parsed s does not match original S")
	}

	// Verify the signature still validates after round-trip
	publicKey := signer.GetCombinedPublicKey()
	valid, err := protocol.VerifySchnorrSignature(publicKey, sig.R, sig.S, message)
	if err != nil {
		t.Fatalf("VerifySchnorrSignature after round-trip failed: %v", err)
	}
	if !valid {
		t.Error("Schnorr signature verification failed after serialize/parse round-trip")
	}
}

// TestSchnorrSignatureInvalidEncoding tests that invalid encodings are rejected.
func TestSchnorrSignatureInvalidEncoding(t *testing.T) {
	n := secp256k1.S256().N

	// Test 1: Wrong length
	invalidSig := make([]byte, 63)
	if err := protocol.ValidateSchnorrEncoding(invalidSig); err == nil {
		t.Error("Expected error for 63-byte signature")
	}

	// Test 2: r = 0
	zeroRSig := make([]byte, 64)
	copy(zeroRSig[32:], n.Bytes())
	if err := protocol.ValidateSchnorrEncoding(zeroRSig); err == nil {
		t.Error("Expected error for r=0")
	}

	// Test 3: s = 0
	zeroSSig := make([]byte, 64)
	copy(zeroSSig[0:32], n.Bytes())
	if err := protocol.ValidateSchnorrEncoding(zeroSSig); err == nil {
		t.Error("Expected error for s=0")
	}

	// Test 4: r >= n
	bigRSig := make([]byte, 64)
	copy(bigRSig[0:32], n.Bytes())
	copy(bigRSig[32:], big.NewInt(1).Bytes())
	if err := protocol.ValidateSchnorrEncoding(bigRSig); err == nil {
		t.Error("Expected error for r >= n")
	}

	// Test 5: s >= n
	bigSSig := make([]byte, 64)
	copy(bigSSig[0:32], big.NewInt(1).Bytes())
	copy(bigSSig[32:], n.Bytes())
	if err := protocol.ValidateSchnorrEncoding(bigSSig); err == nil {
		t.Error("Expected error for s >= n")
	}
}
