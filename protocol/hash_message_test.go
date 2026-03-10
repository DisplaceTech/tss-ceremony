package protocol

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestHashMessage(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
	}{
		{
			name:    "empty message",
			message: []byte{},
		},
		{
			name:    "simple message",
			message: []byte("hello"),
		},
		{
			name:    "longer message",
			message: []byte("This is a longer test message for hashing"),
		},
		{
			name:    "binary data",
			message: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := NewHashMessage(tt.message)

			// Verify the hash matches standard library output
			expectedHash := sha256.Sum256(tt.message)
			if len(hm.Hash) != 32 {
				t.Errorf("expected hash length 32, got %d", len(hm.Hash))
			}
			for i := range expectedHash {
				if hm.Hash[i] != expectedHash[i] {
					t.Errorf("hash mismatch at byte %d: expected %x, got %x", i, expectedHash[i], hm.Hash[i])
				}
			}

			// Verify the message is preserved
			if len(hm.Message) != len(tt.message) {
				t.Errorf("message length mismatch: expected %d, got %d", len(tt.message), len(hm.Message))
			}
			for i := range tt.message {
				if hm.Message[i] != tt.message[i] {
					t.Errorf("message mismatch at byte %d: expected %x, got %x", i, tt.message[i], hm.Message[i])
				}
			}
		})
	}
}

func TestHashMessageKnownValues(t *testing.T) {
	// Test with known SHA-256 hash values
	tests := []struct {
		name     string
		message  []byte
		expected string
	}{
		{
			name:     "empty string",
			message:  []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello",
			message:  []byte("hello"),
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "hello world",
			message:  []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := NewHashMessage(tt.message)

			// Convert hash to hex string
			var hex string
			for _, b := range hm.Hash {
				hex += fmt.Sprintf("%02x", b)
			}

			if hex != tt.expected {
				t.Errorf("hash mismatch: expected %s, got %s", tt.expected, hex)
			}
		})
	}
}
