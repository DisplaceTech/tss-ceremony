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

// TestExternalECDSAVerification verifies that signatures produced by the ceremony
// can be verified using the standard crypto/ecdsa library.
func TestExternalECDSAVerification(t *testing.T) {
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

	// The signature is signed with combined (phantom) key, so we verify against the phantom public key
	partyAPubBytes := ceremony.PhantomKey.SerializeUncompressed()

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

// TestExternalECDSAVerificationRandomMode verifies that random mode also produces valid signatures.
func TestExternalECDSAVerificationRandomMode(t *testing.T) {
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

	// The signature is signed with combined (phantom) key, so we verify against the phantom public key
	partyAPubBytes := ceremony.PhantomKey.SerializeUncompressed()

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

// TestExternalECDSAVerificationInvalidSignature verifies that invalid signatures are rejected.
func TestExternalECDSAVerificationInvalidSignature(t *testing.T) {
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

	// Get the combined (phantom) public key
	partyAPubBytes := ceremony.PhantomKey.SerializeUncompressed()
	pubKey, err := secp256k1.ParsePubKey(partyAPubBytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	// Decode signature components
	rBytes, _ := hex.DecodeString(rHex)
	sBytes, _ := hex.DecodeString(sHex)
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	// Try to verify with wrong message
	wrongHash := sha256.Sum256([]byte("Wrong message"))
	valid := ecdsa.Verify(pubKeyECDSA, wrongHash[:], r, s)
	if valid {
		t.Error("ECDSA verification should reject signature with wrong message")
	}

	// Try to verify with tampered signature
	tamperedR := big.NewInt(12345)
	msgHash := sha256.Sum256(ceremony.Message)
	valid = ecdsa.Verify(pubKeyECDSA, msgHash[:], tamperedR, s)
	if valid {
		t.Error("ECDSA verification should reject tampered signature")
	}
}

// TestOpenSSLCommandGeneration verifies that OpenSSL verification commands are generated correctly.
func TestOpenSSLCommandGeneration(t *testing.T) {
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

	// Get the combined (phantom) public key (without 0x04 prefix)
	partyAPubBytes := ceremony.PhantomKey.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:]) // Skip 0x04 prefix

	// Encode message as hex
	msgHex := hex.EncodeToString(ceremony.Message)

	// Generate OpenSSL command
	cmd := protocol.GenerateOpenSSLVerifyCommand(rHex, sHex, msgHex, partyAPubHex)

	// Verify the command contains expected components
	if len(cmd) == 0 {
		t.Error("Generated OpenSSL command is empty")
	}
	if !containsSubstring(cmd, "openssl") {
		t.Error("Generated command does not contain 'openssl'")
	}
	if !containsSubstring(cmd, "dgst") {
		t.Error("Generated command does not contain 'dgst'")
	}
	if !containsSubstring(cmd, "verify") {
		t.Error("Generated command does not contain 'verify'")
	}
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestRoundTripVerification verifies that a signature can be verified both internally and externally.
func TestRoundTripVerification(t *testing.T) {
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

	// Get the combined (phantom) public key (without 0x04 prefix)
	partyAPubBytes := ceremony.PhantomKey.SerializeUncompressed()
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

	// Also verify using standard library
	partyAPubBytesFull := ceremony.PhantomKey.SerializeUncompressed()
	pubKey, err := secp256k1.ParsePubKey(partyAPubBytesFull)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	hash := sha256.Sum256(ceremony.Message)
	rBytes, _ := hex.DecodeString(rHex)
	sBytes, _ := hex.DecodeString(sHex)
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	valid = ecdsa.Verify(pubKeyECDSA, hash[:], r, s)
	if !valid {
		t.Error("Standard library verification failed")
	}
}
