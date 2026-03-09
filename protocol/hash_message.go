package protocol

import (
	"crypto/sha256"
)

// HashMessage represents a message and its SHA-256 hash.
type HashMessage struct {
	// Message is the original message bytes.
	Message []byte
	// Hash is the SHA-256 hash of the message.
	Hash []byte
}

// NewHashMessage creates a new HashMessage by computing the SHA-256 hash of the input message.
//
// Parameters:
//   - message: the message bytes to hash
//
// Returns:
//   - A pointer to a HashMessage struct containing the original message and its hash
func NewHashMessage(message []byte) *HashMessage {
	hash := sha256.Sum256(message)
	return &HashMessage{
		Message: message,
		Hash:    hash[:],
	}
}
