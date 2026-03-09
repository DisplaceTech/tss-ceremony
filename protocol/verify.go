package protocol

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// VerifySignature verifies an ECDSA signature against a message and public key
// using secp256k1 curve.
//
// Parameters:
//   - pubkey: Hex-encoded uncompressed public key (64 bytes for x||y)
//   - sigR: Hex-encoded R component of the signature (32 bytes)
//   - sigS: Hex-encoded S component of the signature (32 bytes)
//   - message: Hex-encoded message to verify
//
// Returns true if the signature is valid, false otherwise.
func VerifySignature(pubkey, sigR, sigS, message string) (bool, error) {
	// Decode public key (expecting 64 hex chars = 32 bytes for x||y compressed form)
	pubKeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return false, fmt.Errorf("invalid public key hex: %w", err)
	}

	// Parse public key - secp256k1 expects uncompressed (65 bytes) or compressed (33 bytes)
	// If we have 64 bytes, we need to construct uncompressed form
	var pubKey *secp256k1.PublicKey
	if len(pubKeyBytes) == 64 {
		// Construct uncompressed public key from x||y
		uncompressed := make([]byte, 65)
		uncompressed[0] = 0x04 // Uncompressed prefix
		copy(uncompressed[1:], pubKeyBytes)
		pubKey, err = secp256k1.ParsePubKey(uncompressed)
		if err != nil {
			return false, fmt.Errorf("invalid public key: %w", err)
		}
	} else {
		pubKey, err = secp256k1.ParsePubKey(pubKeyBytes)
		if err != nil {
			return false, fmt.Errorf("invalid public key: %w", err)
		}
	}

	// Decode R component
	rBytes, err := hex.DecodeString(sigR)
	if err != nil {
		return false, fmt.Errorf("invalid signature R hex: %w", err)
	}

	// Decode S component
	sBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return false, fmt.Errorf("invalid signature S hex: %w", err)
	}

	// Create signature from R and S
	r := new(secp256k1.ModNScalar).SetByteSlice(rBytes)
	s := new(secp256k1.ModNScalar).SetByteSlice(sBytes)
	_ = r
	_ = s

	// Decode message
	msgBytes, err := hex.DecodeString(message)
	if err != nil {
		return false, fmt.Errorf("invalid message hex: %w", err)
	}

	// Hash the message (ECDSA signs the hash of the message)
	hash := sha256.Sum256(msgBytes)

	// Verify the signature using secp256k1 library
	// Convert to big.Int for ecdsa verification
	rBig := new(big.Int).SetBytes(rBytes)
	sBig := new(big.Int).SetBytes(sBytes)
	_ = rBig
	_ = sBig

	// Convert secp256k1.PublicKey to crypto/ecdsa.PublicKey
	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	// Verify the signature
	valid := ecdsa.Verify(pubKeyECDSA, hash[:], rBig, sBig)

	return valid, nil
}

// GenerateOpenSSLVerifyCommand generates an OpenSSL command string that can be used
// to verify the ECDSA signature externally.
//
// Parameters:
//   - pubkey: Hex-encoded uncompressed public key (64 bytes for x||y)
//   - sigR: Hex-encoded R component of the signature (32 bytes)
//   - sigS: Hex-encoded S component of the signature (32 bytes)
//   - message: Hex-encoded message to verify
//
// Returns the complete openssl dgst command string that can be executed to verify the signature.
func GenerateOpenSSLVerifyCommand(pubkey, sigR, sigS, message string) (string, error) {
	// Decode public key (expecting 64 hex chars = 32 bytes for x||y compressed form)
	pubKeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return "", fmt.Errorf("invalid public key hex: %w", err)
	}

	// Decode R component
	rBytes, err := hex.DecodeString(sigR)
	if err != nil {
		return "", fmt.Errorf("invalid signature R hex: %w", err)
	}

	// Decode S component
	sBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return "", fmt.Errorf("invalid signature S hex: %w", err)
	}

	// Decode message
	msgBytes, err := hex.DecodeString(message)
	if err != nil {
		return "", fmt.Errorf("invalid message hex: %w", err)
	}

	// Construct uncompressed public key from x||y
	uncompressed := make([]byte, 65)
	uncompressed[0] = 0x04 // Uncompressed prefix
	copy(uncompressed[1:], pubKeyBytes)

	// Create the public key in PEM format for OpenSSL
	// We need to construct a proper EC public key in SEC1/DER format
	// For simplicity, we'll use a workaround: create a command that reads from stdin

	// Format the signature as DER for OpenSSL
	// DER format: 30 <total_len> 02 <r_len> <r> 02 <s_len> <s>
	derSig, err := createDERSignature(rBytes, sBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create DER signature: %w", err)
	}

	// Create the public key in a format OpenSSL can use
	// We'll use the EC public key format with the curve specified
	// Format: 04 || x || y (uncompressed)
	pubKeyHex := hex.EncodeToString(uncompressed)

	// Create the message in hex format for the command
	msgHex := hex.EncodeToString(msgBytes)

	// Build the OpenSSL command
	// The command will:
	// 1. Create a temporary public key file in PEM format
	// 2. Create a temporary signature file in DER format
	// 3. Verify the signature

	// For the public key, we need to construct a proper ASN.1 DER structure
	// and then convert it to PEM. We'll use a multi-step approach.

	// Step 1: Create the public key file
	// We'll use openssl to create the key from the raw bytes
	// Format: EC public key in SEC1 format

	// Step 2: Create the signature file from DER bytes

	// Step 3: Verify using openssl dgst

	// Build a shell command that does all of this
	cmd := buildOpenSSLVerifyCommand(pubKeyHex, derSig, msgHex)

	return cmd, nil
}

// createDERSignature creates a DER-encoded signature from R and S components
func createDERSignature(rBytes, sBytes []byte) ([]byte, error) {
	// Remove leading zeros from R and S
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	rBytes = r.Bytes()
	sBytes = s.Bytes()

	// Calculate DER encoding lengths
	rLen := len(rBytes)
	sLen := len(sBytes)
	contentLen := 2 + rLen + 2 + sLen // 2 bytes for each INTEGER tag + length
	totalLen := 1 + contentLen        // 1 byte for SEQUENCE tag

	// Build DER structure
	der := make([]byte, 1+totalLen)
	der[0] = 0x30 // SEQUENCE tag
	der[1] = byte(totalLen)

	// R INTEGER
	der[2] = 0x02 // INTEGER tag
	der[3] = byte(rLen)
	copy(der[4:], rBytes)

	// S INTEGER
	offset := 4 + rLen
	der[offset] = 0x02 // INTEGER tag
	der[offset+1] = byte(sLen)
	copy(der[offset+2:], sBytes)

	return der, nil
}

// buildOpenSSLVerifyCommand builds the complete OpenSSL verification command
func buildOpenSSLVerifyCommand(pubKeyHex string, derSig []byte, msgHex string) string {
	// Convert DER signature to hex string
	sigHex := hex.EncodeToString(derSig)

	// Build the command using a shell script approach
	// We'll create a command that:
	// 1. Creates the message file from hex
	// 2. Creates the signature file from DER hex
	// 3. Creates the public key file in DER format for secp256k1
	// 4. Runs openssl dgst to verify

	// For secp256k1, we need to construct the proper ASN.1 DER structure for the public key
	// Format: SEQUENCE { SEQUENCE { OID (secp256k1), OID (secp256k1) }, BIT STRING (public key) }
	// OID for secp256k1: 1.3.132.0.10 = 2a 86 48 ce 3d 02 01 06 08 2a 86 48 ce 3d 03 01 07

	// Construct the public key DER structure:
	// 30 59 - SEQUENCE, length 89
	//   30 13 - SEQUENCE (algorithm identifier), length 19
	//     06 07 2a 86 48 ce 3d 02 01 - OID 1.2.840.10045.2.1 (ecPublicKey)
	//     06 08 2a 86 48 ce 3d 03 01 07 - OID 1.3.132.0.10 (secp256k1)
	//   03 42 00 - BIT STRING, length 66, 0 unused bits
	//     <65 bytes of public key>

	// The prefix for the public key DER structure
	pubKeyDERPrefix := "3059301306072a8648ce3d020106082a8648ce3d030107034200"

	// Build the command as a shell one-liner
	cmd := fmt.Sprintf(`echo "%s" | xxd -r -p > /tmp/msg.bin && echo "%s" | xxd -r -p > /tmp/sig.der && printf '%s%s' | xxd -r -p > /tmp/pubkey.der && openssl dgst -sha256 -verify /tmp/pubkey.der -signature /tmp/sig.der /tmp/msg.bin`, msgHex, sigHex, pubKeyDERPrefix, pubKeyHex)

	return cmd
}
