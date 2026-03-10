package protocol

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ECDSASignature holds the R and S components of an ECDSA signature.
type ECDSASignature struct {
	R *big.Int
	S *big.Int
}

// SignECDSA signs message (SHA-256 hashed) with privKey and returns the signature.
func SignECDSA(privKey *secp256k1.PrivateKey, message []byte) (*ECDSASignature, error) {
	if privKey == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}
	hash := sha256.Sum256(message)

	privScalar := new(big.Int).SetBytes(privKey.Serialize())
	pubKey := privKey.PubKey()
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
		return nil, fmt.Errorf("ECDSA sign failed: %w", err)
	}
	return &ECDSASignature{R: r, S: s}, nil
}

// VerifyECDSA verifies sig against the SHA-256 hash of message using pubKey.
func VerifyECDSA(pubKey *secp256k1.PublicKey, message []byte, sig *ECDSASignature) bool {
	if pubKey == nil || sig == nil || sig.R == nil || sig.S == nil {
		return false
	}
	hash := sha256.Sum256(message)
	pubKeyECDSA := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}
	return ecdsa.Verify(pubKeyECDSA, hash[:], sig.R, sig.S)
}

// EncodeDER encodes an ECDSA signature into DER (ASN.1) format compatible with OpenSSL.
func (sig *ECDSASignature) EncodeDER() ([]byte, error) {
	if sig.R == nil || sig.S == nil {
		return nil, fmt.Errorf("signature R and S must not be nil")
	}
	type asn1Sig struct {
		R, S *big.Int
	}
	return asn1.Marshal(asn1Sig{R: sig.R, S: sig.S})
}

// DecodeDER parses a DER-encoded ECDSA signature.
func DecodeDER(derBytes []byte) (*ECDSASignature, error) {
	type asn1Sig struct {
		R, S *big.Int
	}
	var decoded asn1Sig
	if _, err := asn1.Unmarshal(derBytes, &decoded); err != nil {
		return nil, fmt.Errorf("failed to decode DER signature: %w", err)
	}
	return &ECDSASignature{R: decoded.R, S: decoded.S}, nil
}

// HexR returns the R component as a hex string.
func (sig *ECDSASignature) HexR() string {
	if sig.R == nil {
		return ""
	}
	return hex.EncodeToString(sig.R.Bytes())
}

// HexS returns the S component as a hex string.
func (sig *ECDSASignature) HexS() string {
	if sig.S == nil {
		return ""
	}
	return hex.EncodeToString(sig.S.Bytes())
}


