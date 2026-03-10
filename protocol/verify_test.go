package protocol

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestVerifySignature tests the VerifySignature function with table-driven tests
func TestVerifySignature(t *testing.T) {
	// Generate a valid key pair for testing
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

	tests := []struct {
		name     string
		pubkey   string
		sigR     string
		sigS     string
		message  string
		want     bool
		wantErr  bool
	}{
		{
			name:     "valid signature",
			pubkey:   pubKeyHex,
			sigR:     sigRHex,
			sigS:     sigSHex,
			message:  messageHex,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "invalid signature - wrong R",
			pubkey:   pubKeyHex,
			sigR:     "0000000000000000000000000000000000000000000000000000000000000001",
			sigS:     sigSHex,
			message:  messageHex,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "invalid signature - wrong S",
			pubkey:   pubKeyHex,
			sigR:     sigRHex,
			sigS:     "0000000000000000000000000000000000000000000000000000000000000001",
			message:  messageHex,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "invalid signature - wrong message",
			pubkey:   pubKeyHex,
			sigR:     sigRHex,
			sigS:     sigSHex,
			message:  hex.EncodeToString([]byte("Different message")),
			want:     false,
			wantErr:  false,
		},
		{
			name:     "invalid public key hex",
			pubkey:   "invalid_hex",
			sigR:     sigRHex,
			sigS:     sigSHex,
			message:  messageHex,
			want:     false,
			wantErr:  true,
		},
		{
			name:     "invalid R hex",
			pubkey:   pubKeyHex,
			sigR:     "invalid_hex",
			sigS:     sigSHex,
			message:  messageHex,
			want:     false,
			wantErr:  true,
		},
		{
			name:     "invalid S hex",
			pubkey:   pubKeyHex,
			sigR:     sigRHex,
			sigS:     "invalid_hex",
			message:  messageHex,
			want:     false,
			wantErr:  true,
		},
		{
			name:     "invalid message hex",
			pubkey:   pubKeyHex,
			sigR:     sigRHex,
			sigS:     sigSHex,
			message:  "invalid_hex",
			want:     false,
			wantErr:  true,
		},
		{
			name:     "empty signature components",
			pubkey:   pubKeyHex,
			sigR:     "",
			sigS:     "",
			message:  messageHex,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "signature with wrong public key",
			pubkey:   "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			sigR:     sigRHex,
			sigS:     sigSHex,
			message:  messageHex,
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifySignature(tt.pubkey, tt.sigR, tt.sigS, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestVerifySignatureWithCompressedPublicKey tests verification with compressed public key format
func TestVerifySignatureWithCompressedPublicKey(t *testing.T) {
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

	// Use compressed public key (33 bytes with 0x02 or 0x03 prefix)
	compressedPubKey := pubKey.SerializeCompressed()
	pubKeyHex := hex.EncodeToString(compressedPubKey)
	sigRHex := hex.EncodeToString(r.Bytes())
	sigSHex := hex.EncodeToString(s.Bytes())
	messageHex := hex.EncodeToString(message)

	got, err := VerifySignature(pubKeyHex, sigRHex, sigSHex, messageHex)
	if err != nil {
		t.Errorf("VerifySignature() unexpected error = %v", err)
	}
	if !got {
		t.Errorf("VerifySignature() = false, want true for compressed public key")
	}
}

// TestVerifySignatureEdgeCases tests edge cases
func TestVerifySignatureEdgeCases(t *testing.T) {
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
	messageHex := hex.EncodeToString(message)

	tests := []struct {
		name    string
		sigR    string
		sigS    string
		wantErr bool
	}{
		{
			name:    "R with leading zeros",
			sigR:    "00000000000000000000000000000000" + hex.EncodeToString(r.Bytes()),
			sigS:    hex.EncodeToString(s.Bytes()),
			wantErr: false,
		},
		{
			name:    "S with leading zeros",
			sigR:    hex.EncodeToString(r.Bytes()),
			sigS:    "00000000000000000000000000000000" + hex.EncodeToString(s.Bytes()),
			wantErr: false,
		},
		{
			name:    "empty message",
			sigR:    hex.EncodeToString(r.Bytes()),
			sigS:    hex.EncodeToString(s.Bytes()),
			wantErr: false,
		},
		{
			name:    "very long message",
			sigR:    hex.EncodeToString(r.Bytes()),
			sigS:    hex.EncodeToString(s.Bytes()),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msgHex string
			if tt.name == "empty message" {
				msgHex = hex.EncodeToString([]byte(""))
			} else if tt.name == "very long message" {
				msgHex = hex.EncodeToString(make([]byte, 10000))
			} else {
				msgHex = messageHex
			}

			_, err := VerifySignature(pubKeyHex, tt.sigR, tt.sigS, msgHex)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
