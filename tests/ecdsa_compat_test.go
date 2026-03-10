package tests

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestECDSASignatureCompatibility verifies that signatures produced by the ceremony
// can be verified using standard ECDSA verification tools.
func TestECDSASignatureCompatibility(t *testing.T) {
	// Create a ceremony in fixed mode for deterministic testing
	ceremony, err := protocol.NewCeremony(true, "", "normal", false)
	if err != nil {
		t.Fatalf("Failed to create ceremony: %v", err)
	}

	// Sign the message
	err = ceremony.SignMessage()
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get the signature components
	rHex, sHex := ceremony.GetSignatureHex()
	if rHex == "" || sHex == "" {
		t.Fatal("Signature components are empty")
	}

	// The signature is signed with Party A's key, so we verify against Party A's public key
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()

	pubKey, err := secp256k1.ParsePubKey(partyAPubBytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	// Convert to ecdsa.PublicKey for standard verification
	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	// Hash the message (ECDSA signs the hash)
	hash := sha256.Sum256(ceremony.Message)

	// Decode the signature components
	rBytes, err := hex.DecodeString(rHex)
	if err != nil {
		t.Fatalf("Failed to decode R: %v", err)
	}
	sBytes, err := hex.DecodeString(sHex)
	if err != nil {
		t.Fatalf("Failed to decode S: %v", err)
	}

	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	// Verify using standard ecdsa.Verify
	valid := ecdsa.Verify(pubKeyECDSA, hash[:], r, s)
	if !valid {
		t.Error("ECDSA signature verification failed using standard library")
	}
}

// TestECDSASignatureCompatibilityWithVerifyFunction verifies that the protocol's
// VerifySignature function works correctly with ceremony-produced signatures.
func TestECDSASignatureCompatibilityWithVerifyFunction(t *testing.T) {
	// Create a ceremony in fixed mode
	ceremony, err := protocol.NewCeremony(true, "", "normal", false)
	if err != nil {
		t.Fatalf("Failed to create ceremony: %v", err)
	}

	// Sign the message
	err = ceremony.SignMessage()
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get signature components
	rHex, sHex := ceremony.GetSignatureHex()

	// Use Party A's public key (uncompressed format for VerifySignature)
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:]) // Skip 0x04 prefix

	// Encode message as hex
	msgHex := hex.EncodeToString(ceremony.Message)

	// Verify using protocol's VerifySignature function
	valid, err := protocol.VerifySignature(partyAPubHex, rHex, sHex, msgHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
	if !valid {
		t.Error("VerifySignature returned false for valid signature")
	}
}

// TestECDSASignatureInvalidRejects verifies that invalid signatures are rejected.
func TestECDSASignatureInvalidRejects(t *testing.T) {
	// Create a ceremony
	ceremony, err := protocol.NewCeremony(true, "", "normal", false)
	if err != nil {
		t.Fatalf("Failed to create ceremony: %v", err)
	}

	// Sign the message
	err = ceremony.SignMessage()
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get signature components
	rHex, sHex := ceremony.GetSignatureHex()

	// Use Party A's public key (uncompressed format for VerifySignature)
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:]) // Skip 0x04 prefix

	// Try to verify with wrong message
	wrongMsg := []byte("Wrong message")
	wrongMsgHex := hex.EncodeToString(wrongMsg)

	valid, err := protocol.VerifySignature(partyAPubHex, rHex, sHex, wrongMsgHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
	if valid {
		t.Error("VerifySignature should reject signature with wrong message")
	}

	// Try to verify with tampered signature
	tamperedR := "0000000000000000000000000000000000000000000000000000000000000001"
	msgHex := hex.EncodeToString(ceremony.Message)

	valid, err = protocol.VerifySignature(partyAPubHex, tamperedR, sHex, msgHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
	if valid {
		t.Error("VerifySignature should reject tampered signature")
	}
}

// TestECDSASignatureRandomMode verifies that random mode also produces valid signatures.
func TestECDSASignatureRandomMode(t *testing.T) {
	// Create a ceremony in random mode
	ceremony, err := protocol.NewCeremony(false, "", "normal", false)
	if err != nil {
		t.Fatalf("Failed to create ceremony: %v", err)
	}

	// Sign the message
	err = ceremony.SignMessage()
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get signature components
	rHex, sHex := ceremony.GetSignatureHex()
	if rHex == "" || sHex == "" {
		t.Fatal("Signature components are empty")
	}

	// The signature is signed with Party A's key, so we verify against Party A's public key
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()

	pubKey, err := secp256k1.ParsePubKey(partyAPubBytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	// Hash the message
	hash := sha256.Sum256(ceremony.Message)

	// Decode signature components
	rBytes, err := hex.DecodeString(rHex)
	if err != nil {
		t.Fatalf("Failed to decode R: %v", err)
	}
	sBytes, err := hex.DecodeString(sHex)
	if err != nil {
		t.Fatalf("Failed to decode S: %v", err)
	}

	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	// Verify using standard ecdsa.Verify
	valid := ecdsa.Verify(pubKeyECDSA, hash[:], r, s)
	if !valid {
		t.Error("ECDSA signature verification failed for random mode")
	}
}
