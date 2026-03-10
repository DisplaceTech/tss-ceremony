package tests

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestECDSARoundTripVerification verifies that a signature produced by the ceremony
// can be verified using OpenSSL, ensuring round-trip compatibility with standard ECDSA tools.
func TestECDSARoundTripVerification(t *testing.T) {
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

	// Get signature components
	rHex, sHex := ceremony.GetSignatureHex()
	if rHex == "" || sHex == "" {
		t.Fatal("Signature components are empty")
	}

	// Get Party A's public key (uncompressed format without 0x04 prefix)
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:]) // Skip 0x04 prefix

	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "ecdsa-roundtrip-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create temporary files
	msgFile := filepath.Join(tmpDir, "message.txt")
	sigFile := filepath.Join(tmpDir, "signature.der")
	pubFile := filepath.Join(tmpDir, "pubkey.pem")

	// Write message to file
	err = os.WriteFile(msgFile, ceremony.Message, 0644)
	if err != nil {
		t.Fatalf("Failed to write message file: %v", err)
	}

	// Create DER-encoded signature
	derSig := createDERSignature(rHex, sHex)
	err = os.WriteFile(sigFile, derSig, 0644)
	if err != nil {
		t.Fatalf("Failed to write signature file: %v", err)
	}

	// Create PEM-encoded public key
	pubKeyDER := createPublicKeyDER(partyAPubHex)
	err = os.WriteFile(pubFile, pubKeyDER, 0644)
	if err != nil {
		t.Fatalf("Failed to write public key file: %v", err)
	}

	// Run OpenSSL verification command
	cmd := exec.Command("openssl", "dgst", "-sha256", "-verify", pubFile, "-signature", sigFile, msgFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("OpenSSL stderr: %s", stderr.String())
		t.Fatalf("OpenSSL verification failed: %v", err)
	}

	// Check output
	output := stdout.String()
	if !bytes.Contains(stdout.Bytes(), []byte("Verified OK")) {
		t.Errorf("OpenSSL verification did not return 'Verified OK'. Output: %s", output)
	}

	t.Log("Round-trip verification successful: Ceremony signature accepted by OpenSSL")
}

// TestECDSARoundTripVerificationRandomMode verifies round-trip compatibility in random mode.
func TestECDSARoundTripVerificationRandomMode(t *testing.T) {
	// Skip if OpenSSL is not available
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("OpenSSL not available, skipping round-trip test")
	}

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

	// Get Party A's public key
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:])

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "ecdsa-roundtrip-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create temporary files
	msgFile := filepath.Join(tmpDir, "message.txt")
	sigFile := filepath.Join(tmpDir, "signature.der")
	pubFile := filepath.Join(tmpDir, "pubkey.pem")

	// Write files
	err = os.WriteFile(msgFile, ceremony.Message, 0644)
	if err != nil {
		t.Fatalf("Failed to write message file: %v", err)
	}

	derSig := createDERSignature(rHex, sHex)
	err = os.WriteFile(sigFile, derSig, 0644)
	if err != nil {
		t.Fatalf("Failed to write signature file: %v", err)
	}

	pubKeyDER := createPublicKeyDER(partyAPubHex)
	err = os.WriteFile(pubFile, pubKeyDER, 0644)
	if err != nil {
		t.Fatalf("Failed to write public key file: %v", err)
	}

	// Run OpenSSL verification
	cmd := exec.Command("openssl", "dgst", "-sha256", "-verify", pubFile, "-signature", sigFile, msgFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("OpenSSL stderr: %s", stderr.String())
		t.Fatalf("OpenSSL verification failed: %v", err)
	}

	if !bytes.Contains(stdout.Bytes(), []byte("Verified OK")) {
		t.Errorf("OpenSSL verification did not return 'Verified OK'. Output: %s", stdout.String())
	}

	t.Log("Random mode round-trip verification successful")
}

// TestECDSARoundTripInvalidSignature verifies that invalid signatures are rejected by OpenSSL.
func TestECDSARoundTripInvalidSignature(t *testing.T) {
	// Skip if OpenSSL is not available
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("OpenSSL not available, skipping round-trip test")
	}

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
	_, sHex := ceremony.GetSignatureHex()

	// Tamper with the signature (change R component)
	tamperedR := "0000000000000000000000000000000000000000000000000000000000000001"

	// Get Party A's public key
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:])

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "ecdsa-roundtrip-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create temporary files
	msgFile := filepath.Join(tmpDir, "message.txt")
	sigFile := filepath.Join(tmpDir, "signature.der")
	pubFile := filepath.Join(tmpDir, "pubkey.pem")

	// Write files
	err = os.WriteFile(msgFile, ceremony.Message, 0644)
	if err != nil {
		t.Fatalf("Failed to write message file: %v", err)
	}

	// Create DER signature with tampered R
	derSig := createDERSignature(tamperedR, sHex)
	err = os.WriteFile(sigFile, derSig, 0644)
	if err != nil {
		t.Fatalf("Failed to write signature file: %v", err)
	}

	pubKeyDER := createPublicKeyDER(partyAPubHex)
	err = os.WriteFile(pubFile, pubKeyDER, 0644)
	if err != nil {
		t.Fatalf("Failed to write public key file: %v", err)
	}

	// Run OpenSSL verification - should fail
	cmd := exec.Command("openssl", "dgst", "-sha256", "-verify", pubFile, "-signature", sigFile, msgFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		t.Error("OpenSSL should have rejected tampered signature")
	}

	t.Log("Invalid signature correctly rejected by OpenSSL")
}

// TestECDSARoundTripWrongMessage verifies that signatures are rejected for wrong messages.
func TestECDSARoundTripWrongMessage(t *testing.T) {
	// Skip if OpenSSL is not available
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("OpenSSL not available, skipping round-trip test")
	}

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

	// Get Party A's public key
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:])

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "ecdsa-roundtrip-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create temporary files
	msgFile := filepath.Join(tmpDir, "message.txt")
	sigFile := filepath.Join(tmpDir, "signature.der")
	pubFile := filepath.Join(tmpDir, "pubkey.pem")

	// Write message with wrong content
	wrongMessage := []byte("Wrong message")
	err = os.WriteFile(msgFile, wrongMessage, 0644)
	if err != nil {
		t.Fatalf("Failed to write message file: %v", err)
	}

	// Create DER signature
	derSig := createDERSignature(rHex, sHex)
	err = os.WriteFile(sigFile, derSig, 0644)
	if err != nil {
		t.Fatalf("Failed to write signature file: %v", err)
	}

	pubKeyDER := createPublicKeyDER(partyAPubHex)
	err = os.WriteFile(pubFile, pubKeyDER, 0644)
	if err != nil {
		t.Fatalf("Failed to write public key file: %v", err)
	}

	// Run OpenSSL verification - should fail
	cmd := exec.Command("openssl", "dgst", "-sha256", "-verify", pubFile, "-signature", sigFile, msgFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		t.Error("OpenSSL should have rejected signature for wrong message")
	}

	t.Log("Wrong message correctly rejected by OpenSSL")
}

// TestECDSARoundTripInternalExternalConsistency verifies that internal and OpenSSL verification agree.
func TestECDSARoundTripInternalExternalConsistency(t *testing.T) {
	// Skip if OpenSSL is not available
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("OpenSSL not available, skipping round-trip test")
	}

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

	// Get Party A's public key
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	partyAPubHex := hex.EncodeToString(partyAPubBytes[1:])

	// Encode message as hex
	msgHex := hex.EncodeToString(ceremony.Message)

	// Verify using protocol's VerifySignature function
	valid, err := protocol.VerifySignature(partyAPubHex, rHex, sHex, msgHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
	if !valid {
		t.Error("Internal verification failed for valid signature")
	}

	// Also verify using OpenSSL
	tmpDir, err := os.MkdirTemp("", "ecdsa-roundtrip-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	msgFile := filepath.Join(tmpDir, "message.txt")
	sigFile := filepath.Join(tmpDir, "signature.der")
	pubFile := filepath.Join(tmpDir, "pubkey.pem")

	err = os.WriteFile(msgFile, ceremony.Message, 0644)
	if err != nil {
		t.Fatalf("Failed to write message file: %v", err)
	}

	derSig := createDERSignature(rHex, sHex)
	err = os.WriteFile(sigFile, derSig, 0644)
	if err != nil {
		t.Fatalf("Failed to write signature file: %v", err)
	}

	pubKeyDER := createPublicKeyDER(partyAPubHex)
	err = os.WriteFile(pubFile, pubKeyDER, 0644)
	if err != nil {
		t.Fatalf("Failed to write public key file: %v", err)
	}

	cmd := exec.Command("openssl", "dgst", "-sha256", "-verify", pubFile, "-signature", sigFile, msgFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("OpenSSL stderr: %s", stderr.String())
		t.Fatalf("OpenSSL verification failed: %v", err)
	}

	opensslValid := bytes.Contains(stdout.Bytes(), []byte("Verified OK"))
	if !opensslValid {
		t.Error("OpenSSL verification failed for valid signature")
	}

	// Both should agree
	if valid != opensslValid {
		t.Error("Internal and OpenSSL verification disagree")
	}

	t.Log("Internal and OpenSSL verification are consistent")
}

// TestECDSARoundTripStandardLibraryConsistency verifies that OpenSSL and standard library agree.
func TestECDSARoundTripStandardLibraryConsistency(t *testing.T) {
	// Skip if OpenSSL is not available
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("OpenSSL not available, skipping round-trip test")
	}

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

	// Get Party A's public key
	partyAPubBytes := ceremony.PartyAPub.SerializeUncompressed()
	pubKey, err := secp256k1.ParsePubKey(partyAPubBytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	// Verify using standard library
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
	standardLibValid := ecdsa.Verify(pubKeyECDSA, hash[:], r, s)
	if !standardLibValid {
		t.Error("Standard library verification failed for valid signature")
	}

	// Also verify using OpenSSL
	tmpDir, err := os.MkdirTemp("", "ecdsa-roundtrip-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	msgFile := filepath.Join(tmpDir, "message.txt")
	sigFile := filepath.Join(tmpDir, "signature.der")
	pubFile := filepath.Join(tmpDir, "pubkey.pem")

	err = os.WriteFile(msgFile, ceremony.Message, 0644)
	if err != nil {
		t.Fatalf("Failed to write message file: %v", err)
	}

	derSig := createDERSignature(rHex, sHex)
	err = os.WriteFile(sigFile, derSig, 0644)
	if err != nil {
		t.Fatalf("Failed to write signature file: %v", err)
	}

	pubKeyDER := createPublicKeyDER(hex.EncodeToString(partyAPubBytes[1:]))
	err = os.WriteFile(pubFile, pubKeyDER, 0644)
	if err != nil {
		t.Fatalf("Failed to write public key file: %v", err)
	}

	cmd := exec.Command("openssl", "dgst", "-sha256", "-verify", pubFile, "-signature", sigFile, msgFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("OpenSSL stderr: %s", stderr.String())
		t.Fatalf("OpenSSL verification failed: %v", err)
	}

	opensslValid := bytes.Contains(stdout.Bytes(), []byte("Verified OK"))
	if !opensslValid {
		t.Error("OpenSSL verification failed for valid signature")
	}

	// Both should agree
	if standardLibValid != opensslValid {
		t.Error("Standard library and OpenSSL verification disagree")
	}

	t.Log("Standard library and OpenSSL verification are consistent")
}

// createDERSignature creates a DER-encoded ECDSA signature from hex R and S components.
func createDERSignature(rHex, sHex string) []byte {
	rBytes, _ := hex.DecodeString(rHex)
	sBytes, _ := hex.DecodeString(sHex)

	// Remove leading zeros for DER encoding (but keep at least one byte if all zeros)
	rTrimmed := removeLeadingZeros(rBytes)
	sTrimmed := removeLeadingZeros(sBytes)

	// Add high bit if needed to prevent negative interpretation
	if len(rTrimmed) > 0 && rTrimmed[0] >= 0x80 {
		rTrimmed = append([]byte{0x00}, rTrimmed...)
	}
	if len(sTrimmed) > 0 && sTrimmed[0] >= 0x80 {
		sTrimmed = append([]byte{0x00}, sTrimmed...)
	}

	// Build DER encoding: 0x30 <total_len> 0x02 <r_len> <r> 0x02 <s_len> <s>
	derSig := make([]byte, 0, 8+len(rTrimmed)+len(sTrimmed))
	derSig = append(derSig, 0x30) // SEQUENCE tag
	derSig = append(derSig, byte(len(rTrimmed)+len(sTrimmed)+4)) // Total length
	derSig = append(derSig, 0x02) // INTEGER tag for R
	derSig = append(derSig, byte(len(rTrimmed))) // R length
	derSig = append(derSig, rTrimmed...)
	derSig = append(derSig, 0x02) // INTEGER tag for S
	derSig = append(derSig, byte(len(sTrimmed))) // S length
	derSig = append(derSig, sTrimmed...)

	return derSig
}

// createPublicKeyDER creates a PEM-encoded SubjectPublicKeyInfo for secp256k1.
func createPublicKeyDER(pubKeyHex string) []byte {
	// secp256k1 SubjectPublicKeyInfo DER structure:
	// SEQUENCE {
	//   SEQUENCE {
	//     OID 1.2.840.10045.2.1 (ecPublicKey)
	//     OID 1.3.132.0.10       (secp256k1)
	//   }
	//   BIT STRING { 0x04 <x> <y> }
	// }
	//
	// DER prefix for secp256k1 public key (ecPublicKey + secp256k1 OID):
	// 3056 3010 0607 2a86 48ce 3d02 01 0605 2b81 0400 0a 0342 00
	pubKeyDERPrefix := []byte{
		0x30, 0x56, // SEQUENCE, length 86
		0x30, 0x10, // SEQUENCE, length 16
		0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02, 0x01, // OID ecPublicKey
		0x06, 0x05, 0x2b, 0x81, 0x04, 0x00, 0x0a, // OID secp256k1
		0x03, 0x42, 0x00, // BIT STRING, length 66, no unused bits
	}

	// Decode public key hex (x||y without 0x04 prefix)
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		// Fallback: try with 0x04 prefix
		pubKeyBytes, _ = hex.DecodeString("04" + pubKeyHex)
	}

	// Prepend 0x04 if not present
	if len(pubKeyBytes) > 0 && pubKeyBytes[0] != 0x04 {
		pubKeyBytes = append([]byte{0x04}, pubKeyBytes...)
	}

	// Combine prefix and public key bytes into DER
	derBytes := make([]byte, len(pubKeyDERPrefix)+len(pubKeyBytes))
	copy(derBytes, pubKeyDERPrefix)
	copy(derBytes[len(pubKeyDERPrefix):], pubKeyBytes)

	// Encode as PEM
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}
	return pem.EncodeToMemory(pemBlock)
}

// removeLeadingZeros removes leading zero bytes from a byte slice.
func removeLeadingZeros(b []byte) []byte {
	for len(b) > 0 && b[0] == 0 {
		b = b[1:]
	}
	if len(b) == 0 {
		return []byte{0}
	}
	return b
}
