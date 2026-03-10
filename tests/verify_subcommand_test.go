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

// TestVerifySignatureValid tests verification of a valid signature
func TestVerifySignatureValid(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Sign a test message
	message := []byte("Test message for verification")
	hash := sha256.Sum256(message)

	// Convert to ecdsa.PrivateKey for signing
	privScalar := new(big.Int).SetBytes(privKey.Serialize())
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     pubKey.X(),
			Y:     pubKey.Y(),
		},
		D: privScalar,
	}

	r, s, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Prepare test data
	pubKeyHex := hex.EncodeToString(pubKey.SerializeUncompressed()[1:]) // Remove 0x04 prefix
	sigRHex := hex.EncodeToString(r.Bytes())
	sigSHex := hex.EncodeToString(s.Bytes())
	messageHex := hex.EncodeToString(message)

	// Verify the signature
	valid, err := protocol.VerifySignature(pubKeyHex, sigRHex, sigSHex, messageHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}

	if !valid {
		t.Error("Expected valid signature to return true")
	}
}

// TestVerifySignatureInvalid tests verification of an invalid signature
func TestVerifySignatureInvalid(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Sign a test message
	message := []byte("Test message for verification")
	hash := sha256.Sum256(message)

	// Convert to ecdsa.PrivateKey for signing
	privScalar := new(big.Int).SetBytes(privKey.Serialize())
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     pubKey.X(),
			Y:     pubKey.Y(),
		},
		D: privScalar,
	}

	r, _, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Prepare test data
	pubKeyHex := hex.EncodeToString(pubKey.SerializeUncompressed()[1:])
	sigRHex := hex.EncodeToString(r.Bytes())
	messageHex := hex.EncodeToString(message)

	// Modify the signature to make it invalid
	invalidSigSHex := "0000000000000000000000000000000000000000000000000000000000000001"

	// Verify the invalid signature
	valid, err := protocol.VerifySignature(pubKeyHex, sigRHex, invalidSigSHex, messageHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}

	if valid {
		t.Error("Expected invalid signature to return false")
	}
}

// TestVerifySignatureWrongMessage tests verification with wrong message
func TestVerifySignatureWrongMessage(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Sign a test message
	message := []byte("Test message for verification")
	hash := sha256.Sum256(message)

	// Convert to ecdsa.PrivateKey for signing
	privScalar := new(big.Int).SetBytes(privKey.Serialize())
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     pubKey.X(),
			Y:     pubKey.Y(),
		},
		D: privScalar,
	}

	r, s, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Prepare test data
	pubKeyHex := hex.EncodeToString(pubKey.SerializeUncompressed()[1:])
	sigRHex := hex.EncodeToString(r.Bytes())
	sigSHex := hex.EncodeToString(s.Bytes())
	wrongMessageHex := hex.EncodeToString([]byte("Wrong message"))

	// Verify with wrong message
	valid, err := protocol.VerifySignature(pubKeyHex, sigRHex, sigSHex, wrongMessageHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}

	if valid {
		t.Error("Expected signature with wrong message to return false")
	}
}

// TestVerifySignatureInvalidPubKeyHex tests verification with invalid public key hex
func TestVerifySignatureInvalidPubKeyHex(t *testing.T) {
	message := []byte("Test message")
	messageHex := hex.EncodeToString(message)

	// Test with invalid hex characters
	_, err := protocol.VerifySignature("invalid_hex", "0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001", messageHex)
	if err == nil {
		t.Error("Expected error for invalid public key hex")
	}

	// Test with too short hex
	_, err = protocol.VerifySignature("00", "0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001", messageHex)
	if err == nil {
		t.Error("Expected error for too short public key hex")
	}
}

// TestVerifySignatureInvalidSigRHex tests verification with invalid signature R hex
func TestVerifySignatureInvalidSigRHex(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	message := []byte("Test message")
	messageHex := hex.EncodeToString(message)
	pubKeyHex := hex.EncodeToString(pubKey.SerializeUncompressed()[1:])

	// Test with invalid hex characters
	_, err = protocol.VerifySignature(pubKeyHex, "invalid_hex", "0000000000000000000000000000000000000000000000000000000000000001", messageHex)
	if err == nil {
		t.Error("Expected error for invalid signature R hex")
	}
}

// TestVerifySignatureInvalidSigSHex tests verification with invalid signature S hex
func TestVerifySignatureInvalidSigSHex(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	message := []byte("Test message")
	messageHex := hex.EncodeToString(message)
	pubKeyHex := hex.EncodeToString(pubKey.SerializeUncompressed()[1:])

	// Test with invalid hex characters
	_, err = protocol.VerifySignature(pubKeyHex, "0000000000000000000000000000000000000000000000000000000000000000", "invalid_hex", messageHex)
	if err == nil {
		t.Error("Expected error for invalid signature S hex")
	}
}

// TestVerifySignatureInvalidMessageHex tests verification with invalid message hex
func TestVerifySignatureInvalidMessageHex(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	message := []byte("Test message")
	hash := sha256.Sum256(message)

	privScalar := new(big.Int).SetBytes(privKey.Serialize())
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     pubKey.X(),
			Y:     pubKey.Y(),
		},
		D: privScalar,
	}

	r, s, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	pubKeyHex := hex.EncodeToString(pubKey.SerializeUncompressed()[1:])
	sigRHex := hex.EncodeToString(r.Bytes())
	sigSHex := hex.EncodeToString(s.Bytes())

	// Test with invalid hex characters
	_, err = protocol.VerifySignature(pubKeyHex, sigRHex, sigSHex, "invalid_hex")
	if err == nil {
		t.Error("Expected error for invalid message hex")
	}
}

// TestVerifySignatureEmptyMessage tests verification with empty message
func TestVerifySignatureEmptyMessage(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Sign an empty message
	message := []byte("")
	hash := sha256.Sum256(message)

	privScalar := new(big.Int).SetBytes(privKey.Serialize())
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     pubKey.X(),
			Y:     pubKey.Y(),
		},
		D: privScalar,
	}

	r, s, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	pubKeyHex := hex.EncodeToString(pubKey.SerializeUncompressed()[1:])
	sigRHex := hex.EncodeToString(r.Bytes())
	sigSHex := hex.EncodeToString(s.Bytes())
	messageHex := hex.EncodeToString(message)

	// Verify the signature
	valid, err := protocol.VerifySignature(pubKeyHex, sigRHex, sigSHex, messageHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}

	if !valid {
		t.Error("Expected valid signature for empty message to return true")
	}
}

// TestVerifySignatureCompressedPubKey tests verification with compressed public key
func TestVerifySignatureCompressedPubKey(t *testing.T) {
	// Generate a valid key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Sign a test message
	message := []byte("Test message for verification")
	hash := sha256.Sum256(message)

	// Convert to ecdsa.PrivateKey for signing
	privScalar := new(big.Int).SetBytes(privKey.Serialize())
	privKeyECDSA := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     pubKey.X(),
			Y:     pubKey.Y(),
		},
		D: privScalar,
	}

	r, s, err := ecdsa.Sign(rand.Reader, privKeyECDSA, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Use compressed public key (33 bytes)
	pubKeyCompressed := hex.EncodeToString(pubKey.SerializeCompressed())
	sigRHex := hex.EncodeToString(r.Bytes())
	sigSHex := hex.EncodeToString(s.Bytes())
	messageHex := hex.EncodeToString(message)

	// Verify the signature with compressed public key
	valid, err := protocol.VerifySignature(pubKeyCompressed, sigRHex, sigSHex, messageHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}

	if !valid {
		t.Error("Expected valid signature with compressed public key to return true")
	}
}

// TestVerifySignatureDifferentKeys tests that signature doesn't verify with different key
func TestVerifySignatureDifferentKeys(t *testing.T) {
	// Generate two different key pairs
	privKey1, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key 1: %v", err)
	}
	pubKey1 := privKey1.PubKey()

	privKey2, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key 2: %v", err)
	}
	pubKey2 := privKey2.PubKey()

	// Sign with first key
	message := []byte("Test message for verification")
	hash := sha256.Sum256(message)

	privScalar1 := new(big.Int).SetBytes(privKey1.Serialize())
	privKeyECDSA1 := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1.S256(),
			X:     pubKey1.X(),
			Y:     pubKey1.Y(),
		},
		D: privScalar1,
	}

	r, s, err := ecdsa.Sign(rand.Reader, privKeyECDSA1, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	_ = pubKey1.SerializeUncompressed() // pubKey1 used for signing; verify with pubKey2
	pubKey2Hex := hex.EncodeToString(pubKey2.SerializeUncompressed()[1:])
	sigRHex := hex.EncodeToString(r.Bytes())
	sigSHex := hex.EncodeToString(s.Bytes())
	messageHex := hex.EncodeToString(message)

	// Try to verify with second key (should fail)
	valid, err := protocol.VerifySignature(pubKey2Hex, sigRHex, sigSHex, messageHex)
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}

	if valid {
		t.Error("Expected signature to be invalid when verified with different key")
	}
}

// TestVerifySignatureEdgeCases tests edge cases for signature verification
func TestVerifySignatureEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pubkey  string
		sigR    string
		sigS    string
		message string
		wantErr bool
	}{
		{
			name:    "valid inputs",
			pubkey:  "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			sigR:    "0000000000000000000000000000000000000000000000000000000000000001",
			sigS:    "0000000000000000000000000000000000000000000000000000000000000001",
			message: "68656c6c6f", // "hello" in hex
			wantErr: true,
		},
		{
			name:    "odd length pubkey",
			pubkey:  "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			sigR:    "0000000000000000000000000000000000000000000000000000000000000001",
			sigS:    "0000000000000000000000000000000000000000000000000000000000000001",
			message: "68656c6c6f",
			wantErr: true,
		},
		{
			name:    "odd length sigR",
			pubkey:  "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			sigR:    "000000000000000000000000000000000000000000000000000000000000000",
			sigS:    "0000000000000000000000000000000000000000000000000000000000000001",
			message: "68656c6c6f",
			wantErr: true,
		},
		{
			name:    "odd length sigS",
			pubkey:  "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			sigR:    "0000000000000000000000000000000000000000000000000000000000000001",
			sigS:    "000000000000000000000000000000000000000000000000000000000000000",
			message: "68656c6c6f",
			wantErr: true,
		},
		{
			name:    "odd length message",
			pubkey:  "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			sigR:    "0000000000000000000000000000000000000000000000000000000000000001",
			sigS:    "0000000000000000000000000000000000000000000000000000000000000001",
			message: "68656c6c",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protocol.VerifySignature(tt.pubkey, tt.sigR, tt.sigS, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
