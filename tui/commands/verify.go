package commands

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// VerifyFromFile verifies an ECDSA signature from file inputs.
// Parameters:
//   - pubKeyFile: Path to file containing hex-encoded public key
//   - sigRFile: Path to file containing hex-encoded R component
//   - sigSFile: Path to file containing hex-encoded S component
//   - messageFile: Path to file containing hex-encoded message
//
// Returns true if the signature is valid, false otherwise.
func VerifyFromFile(pubKeyFile, sigRFile, sigSFile, messageFile string) (bool, error) {
	// Read public key file
	pubKeyBytes, err := readFileAsHex(pubKeyFile)
	if err != nil {
		return false, fmt.Errorf("failed to read public key file: %w", err)
	}

	// Read R component file
	rBytes, err := readFileAsHex(sigRFile)
	if err != nil {
		return false, fmt.Errorf("failed to read R component file: %w", err)
	}

	// Read S component file
	sBytes, err := readFileAsHex(sigSFile)
	if err != nil {
		return false, fmt.Errorf("failed to read S component file: %w", err)
	}

	// Read message file
	msgBytes, err := readFileAsHex(messageFile)
	if err != nil {
		return false, fmt.Errorf("failed to read message file: %w", err)
	}

	// Parse public key
	var pubKey *secp256k1.PublicKey
	switch len(pubKeyBytes) {
	case 64:
		// Construct uncompressed public key from x||y
		uncompressed := make([]byte, 65)
		uncompressed[0] = 0x04
		copy(uncompressed[1:], pubKeyBytes)
		pubKey, err = secp256k1.ParsePubKey(uncompressed)
	case 33:
		// Compressed public key
		pubKey, err = secp256k1.ParsePubKey(pubKeyBytes)
	case 65:
		// Uncompressed public key with 0x04 prefix
		pubKey, err = secp256k1.ParsePubKey(pubKeyBytes)
	default:
		return false, fmt.Errorf("invalid public key length: expected 64, 33, or 65 bytes, got %d", len(pubKeyBytes))
	}

	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Convert R and S to big.Int
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	// Verify signature
	hash := sha256.Sum256(msgBytes)
	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	return ecdsa.Verify(pubKeyECDSA, hash[:], r, s), nil
}

// readFileAsHex reads a file and returns its contents as a hex string.
// The file can contain either raw bytes or hex-encoded data.
func readFileAsHex(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Trim whitespace
	content = trimHexWhitespace(content)

	// Try to decode as hex first
	hexBytes, err := hex.DecodeString(string(content))
	if err == nil {
		return hexBytes, nil
	}

	// If not valid hex, return raw bytes
	return content, nil
}

// trimHexWhitespace removes common whitespace characters from hex strings.
func trimHexWhitespace(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for _, b := range data {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			result = append(result, b)
		}
	}
	return result
}

// FormatSignatureHex formats signature components as hex strings.
func FormatSignatureHex(r, s *big.Int) (string, string) {
	rHex := hex.EncodeToString(r.Bytes())
	sHex := hex.EncodeToString(s.Bytes())
	return rHex, sHex
}

// FormatPublicKeyHex formats a public key as hex string.
func FormatPublicKeyHex(pubKey *secp256k1.PublicKey) string {
	pubKeyBytes := pubKey.SerializeUncompressed()
	return hex.EncodeToString(pubKeyBytes[1:]) // Remove 0x04 prefix
}
