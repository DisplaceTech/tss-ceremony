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
