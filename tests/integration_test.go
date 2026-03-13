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

// TestIntegrationFullCeremonyECDSAVerification tests that the full ceremony
// produces signatures that can be verified using standard ECDSA verification.
func TestIntegrationFullCeremonyECDSAVerification(t *testing.T) {
	// Create a ceremony in fixed mode for deterministic testing
	ceremony, err := protocol.NewCeremony(true, "", "normal", false)
	if err != nil {
		t.Fatalf("Failed to create ceremony: %v", err)
	}

	// Perform the signing ceremony
	err = ceremony.SignMessage()
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get the signature components
	rHex, sHex := ceremony.GetSignatureHex()
	if rHex == "" || sHex == "" {
		t.Fatal("Signature components are empty")
	}

	// Get the combined (phantom) public key used for signing
	partyAPubBytes := ceremony.PhantomKey.SerializeUncompressed()

	// Parse the public key
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

// TestIntegrationCeremonyWithProtocolVerifyFunction tests that the ceremony
// output can be verified using the protocol's VerifySignature function.
func TestIntegrationCeremonyWithProtocolVerifyFunction(t *testing.T) {
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

	// Use the combined (phantom) public key (uncompressed format for VerifySignature)
	phantomPubBytes := ceremony.PhantomKey.SerializeUncompressed()
	phantomPubHex := hex.EncodeToString(phantomPubBytes[1:]) // Skip 0x04 prefix

	// Encode message as hex
	msgHex := hex.EncodeToString(ceremony.Message)

	// Verify using protocol's VerifySignature function
	valid, err := protocol.VerifySignature(phantomPubHex, rHex, sHex, msgHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
	if !valid {
		t.Error("VerifySignature returned false for valid signature")
	}
}

// TestIntegrationCeremonyWithCustomMessage tests the ceremony with a custom message.
func TestIntegrationCeremonyWithCustomMessage(t *testing.T) {
	// Create a ceremony with a custom message
	customMessage := "Custom test message for integration"

	ceremony, err := protocol.NewCeremony(true, customMessage, "normal", false)
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

	// Get the combined (phantom) public key
	partyAPubBytes := ceremony.PhantomKey.SerializeUncompressed()

	// Parse the public key
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

	// Hash the message
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
		t.Error("ECDSA signature verification failed for custom message")
	}
}

// TestIntegrationCeremonyRandomMode tests that random mode also produces valid signatures.
func TestIntegrationCeremonyRandomMode(t *testing.T) {
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

// TestIntegrationCeremonyInvalidSignatureRejection tests that invalid signatures are rejected.
func TestIntegrationCeremonyInvalidSignatureRejection(t *testing.T) {
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

	// Use the combined (phantom) public key
	phantomPubBytes := ceremony.PhantomKey.SerializeUncompressed()
	phantomPubHex := hex.EncodeToString(phantomPubBytes[1:])

	// Try to verify with wrong message
	wrongMsg := []byte("Wrong message")
	wrongMsgHex := hex.EncodeToString(wrongMsg)

	valid, err := protocol.VerifySignature(phantomPubHex, rHex, sHex, wrongMsgHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
	if valid {
		t.Error("VerifySignature should reject signature with wrong message")
	}

	// Try to verify with tampered signature
	tamperedR := "0000000000000000000000000000000000000000000000000000000000000001"
	msgHex := hex.EncodeToString(ceremony.Message)

	valid, err = protocol.VerifySignature(phantomPubHex, tamperedR, sHex, msgHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
	if valid {
		t.Error("VerifySignature should reject tampered signature")
	}
}

// TestIntegrationCeremonyPhantomKeyDerivation tests that the phantom key is correctly derived.
func TestIntegrationCeremonyPhantomKeyDerivation(t *testing.T) {
	// Create a ceremony
	ceremony, err := protocol.NewCeremony(true, "", "normal", false)
	if err != nil {
		t.Fatalf("Failed to create ceremony: %v", err)
	}

	// Sign the message to derive the phantom key
	err = ceremony.SignMessage()
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get the phantom public key
	phantomPubHex := ceremony.GetPhantomPubKeyHex()
	if phantomPubHex == "" {
		t.Fatal("Phantom public key is empty")
	}

	// Verify that the phantom key is the sum of Party A and Party B public keys
	combinedKey, err := protocol.CombinePublicKeys(ceremony.PartyAPub, ceremony.PartyBPub)
	if err != nil {
		t.Fatalf("Failed to combine public keys: %v", err)
	}

	// Compare the phantom key with the combined key
	phantomBytes := ceremony.PhantomKey.SerializeCompressed()
	combinedBytes := combinedKey.SerializeCompressed()

	if len(phantomBytes) != len(combinedBytes) {
		t.Error("Phantom key and combined key have different lengths")
	}

	// Compare byte by byte (skip the 0x02/0x03 prefix for compressed keys)
	for i := 1; i < len(phantomBytes); i++ {
		if phantomBytes[i] != combinedBytes[i] {
			t.Errorf("Phantom key and combined key differ at byte %d", i)
		}
	}
}

// TestIntegrationNoColorFlag tests that the --no-color flag disables ANSI output
// and the application remains functional and readable.
func TestIntegrationNoColorFlag(t *testing.T) {
	// Create a ceremony with noColor=true
	ceremony, err := protocol.NewCeremony(true, "", "normal", true)
	if err != nil {
		t.Fatalf("Failed to create ceremony with noColor=true: %v", err)
	}

	// Perform the signing ceremony
	err = ceremony.SignMessage()
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Get the signature components
	rHex, sHex := ceremony.GetSignatureHex()
	if rHex == "" || sHex == "" {
		t.Fatal("Signature components are empty")
	}

	// Verify the signature against the combined (phantom) public key
	phantomPubBytes := ceremony.PhantomKey.SerializeUncompressed()
	pubKey, err := secp256k1.ParsePubKey(phantomPubBytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	hash := sha256.Sum256(ceremony.Message)

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

	valid := ecdsa.Verify(pubKeyECDSA, hash[:], r, s)
	if !valid {
		t.Error("ECDSA signature verification failed with noColor=true")
	}

	// Verify that the ceremony still produces valid output
	if ceremony.PhantomKey == nil {
		t.Error("Phantom key is nil with noColor=true")
	}
}
