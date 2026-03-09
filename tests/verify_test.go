package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestGenerateOpenSSLVerifyCommand tests the OpenSSL command generation
func TestGenerateOpenSSLVerifyCommand(t *testing.T) {
	// Generate a valid key pair for testing
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	// Sign a test message
	message := []byte("Test message for OpenSSL command")
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
		name      string
		pubkey    string
		sigR      string
		sigS      string
		message   string
		wantErr   bool
		wantValid bool
	}{
		{
			name:      "valid signature generates command",
			pubkey:    pubKeyHex,
			sigR:      sigRHex,
			sigS:      sigSHex,
			message:   messageHex,
			wantErr:   false,
			wantValid: true,
		},
		{
			name:      "invalid pubkey hex",
			pubkey:    "invalid_hex",
			sigR:      sigRHex,
			sigS:      sigSHex,
			message:   messageHex,
			wantErr:   true,
			wantValid: false,
		},
		{
			name:      "invalid sigR hex",
			pubkey:    pubKeyHex,
			sigR:      "invalid_hex",
			sigS:      sigSHex,
			message:   messageHex,
			wantErr:   true,
			wantValid: false,
		},
		{
			name:      "invalid sigS hex",
			pubkey:    pubKeyHex,
			sigR:      sigRHex,
			sigS:      "invalid_hex",
			message:   messageHex,
			wantErr:   true,
			wantValid: false,
		},
		{
			name:      "invalid message hex",
			pubkey:    pubKeyHex,
			sigR:      sigRHex,
			sigS:      sigSHex,
			message:   "invalid_hex",
			wantErr:   true,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := protocol.GenerateOpenSSLVerifyCommand(tt.pubkey, tt.sigR, tt.sigS, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateOpenSSLVerifyCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Check that the command contains expected elements
				if cmd == "" {
					t.Error("GenerateOpenSSLVerifyCommand() returned empty command")
				}
				if !strings.Contains(cmd, "openssl") {
					t.Error("Generated command does not contain 'openssl'")
				}
				if !strings.Contains(cmd, "dgst") {
					t.Error("Generated command does not contain 'dgst'")
				}
				if !strings.Contains(cmd, "-sha256") {
					t.Error("Generated command does not contain '-sha256'")
				}
				if !strings.Contains(cmd, "-verify") {
					t.Error("Generated command does not contain '-verify'")
				}
				if !strings.Contains(cmd, "-signature") {
					t.Error("Generated command does not contain '-signature'")
				}
			}
		})
	}
}

// TestGenerateOpenSSLVerifyCommandSyntax tests that the generated command has correct syntax
func TestGenerateOpenSSLVerifyCommandSyntax(t *testing.T) {
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
	messageHex := hex.EncodeToString(message)

	cmd, err := protocol.GenerateOpenSSLVerifyCommand(pubKeyHex, sigRHex, sigSHex, messageHex)
	if err != nil {
		t.Fatalf("GenerateOpenSSLVerifyCommand() unexpected error = %v", err)
	}

	// Verify command structure
	// The command should contain:
	// 1. echo with message hex piped to xxd
	// 2. echo with signature hex piped to xxd
	// 3. printf with public key DER piped to xxd
	// 4. openssl dgst -sha256 -verify ...

	// Check for xxd commands
	if !strings.Contains(cmd, "xxd -r -p") {
		t.Error("Generated command does not contain 'xxd -r -p'")
	}

	// Check for temporary file references
	if !strings.Contains(cmd, "/tmp/") {
		t.Error("Generated command does not reference temporary files")
	}

	// Check for the public key DER prefix (secp256k1 curve identifier)
	if !strings.Contains(cmd, "3059301306072a8648ce3d020106082a8648ce3d030107034200") {
		t.Error("Generated command does not contain secp256k1 public key DER prefix")
	}
}

// TestGenerateOpenSSLVerifyCommandWithDifferentCurves tests command generation for different scenarios
func TestGenerateOpenSSLVerifyCommandWithDifferentCurves(t *testing.T) {
	// Note: This implementation only supports secp256k1, but we test that
	// the command structure is consistent regardless of input values

	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	tests := []struct {
		name    string
		message []byte
	}{
		{
			name:    "short message",
			message: []byte("Hi"),
		},
		{
			name:    "long message",
			message: make([]byte, 1000),
		},
		{
			name:    "binary message",
			message: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name:    "empty message",
			message: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := sha256.Sum256(tt.message)

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
			messageHex := hex.EncodeToString(tt.message)

			cmd, err := protocol.GenerateOpenSSLVerifyCommand(pubKeyHex, sigRHex, sigSHex, messageHex)
			if err != nil {
				t.Errorf("GenerateOpenSSLVerifyCommand() unexpected error = %v", err)
				return
			}

			// Verify command contains all required components
			requiredComponents := []string{
				"openssl",
				"dgst",
				"-sha256",
				"-verify",
				"-signature",
				"xxd -r -p",
			}

			for _, component := range requiredComponents {
				if !strings.Contains(cmd, component) {
					t.Errorf("Generated command missing required component: %s", component)
				}
			}
		})
	}
}

// TestGenerateOpenSSLVerifyCommandDEREncoding tests that the DER encoding is correct
func TestGenerateOpenSSLVerifyCommandDEREncoding(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	pubKey := privKey.PubKey()

	message := []byte("Test message for DER encoding")
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

	cmd, err := protocol.GenerateOpenSSLVerifyCommand(pubKeyHex, sigRHex, sigSHex, messageHex)
	if err != nil {
		t.Fatalf("GenerateOpenSSLVerifyCommand() unexpected error = %v", err)
	}

	// The command should contain the message hex
	if !strings.Contains(cmd, messageHex) {
		t.Error("Generated command does not contain the message hex")
	}

	// The command should contain the public key hex (without 0x04 prefix)
	if !strings.Contains(cmd, pubKeyHex) {
		t.Error("Generated command does not contain the public key hex")
	}
}
