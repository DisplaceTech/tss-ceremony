package protocol

import (
	"crypto/rand"
	"crypto/sha256"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// GenerateNonceShare generates a cryptographically secure random scalar for the nonce.
//
// This function generates a random scalar k in the range [1, n-1] where n is
// the order of the secp256k1 curve. This scalar is used as a nonce share in
// the DKLS signing protocol.
//
// Returns:
//   - A valid secp256k1 scalar
//   - Error if random generation fails
func GenerateNonceShare() (*big.Int, error) {
	n := secp256k1.S256().N

	// Generate random scalar in [1, n-1]
	k, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	k.Add(k, big.NewInt(1)) // Ensure k is in [1, n-1]

	return k, nil
}

// ComputeNoncePublic computes the public nonce point R = k * G.
//
// This function multiplies the nonce scalar k by the generator point G on
// the secp256k1 curve, returning the resulting point R.
//
// Parameters:
//   - k: the nonce scalar
//
// Returns:
//   - The public nonce point R = k * G
//   - Error if k is invalid
func ComputeNoncePublic(k *big.Int) (*secp256k1.PublicKey, error) {
	if k == nil || k.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrInvalidScalar
	}

	n := secp256k1.S256().N
	if k.Cmp(n) >= 0 {
		return nil, ErrInvalidScalar
	}

	// Create a private key from the scalar to compute the public key
	privKey := secp256k1.PrivKeyFromBytes(k.Bytes())
	if privKey == nil {
		return nil, ErrInvalidScalar
	}

	return privKey.PubKey(), nil
}

// CombineNonces adds two nonce public points and extracts the r value.
//
// This function adds the two nonce public points R_a and R_b to get R = R_a + R_b,
// then extracts the x-coordinate modulo n to get the signature r value.
//
// Parameters:
//   - R_a: Party A's nonce public point
//   - R_b: Party B's nonce public point
//
// Returns:
//   - r: the x-coordinate of R modulo n
//   - R: the combined nonce public point
//   - Error if inputs are invalid
func CombineNonces(Ra, Rb *secp256k1.PublicKey) (*big.Int, *secp256k1.PublicKey, error) {
	if Ra == nil || Rb == nil {
		return nil, nil, ErrInvalidPublicKey
	}

	// Add the two points: R = R_a + R_b
	x, y := secp256k1.S256().Add(Ra.X(), Ra.Y(), Rb.X(), Rb.Y())

	// Create the combined public key
	xBytes := x.Bytes()
	yBytes := y.Bytes()
	pubKeyBytes := make([]byte, 65)
	pubKeyBytes[0] = 0x04
	copy(pubKeyBytes[1+32-len(xBytes):33], xBytes)
	copy(pubKeyBytes[33+32-len(yBytes):65], yBytes)
	R, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	// Extract r = x mod n
	n := secp256k1.S256().N
	r := new(big.Int).Mod(x, n)

	return r, R, nil
}

// HashMessage computes the SHA-256 hash of the input message.
//
// This function computes the SHA-256 hash of the message bytes, which is
// used as the input to the ECDSA signature computation.
//
// Parameters:
//   - message: the message bytes to hash
//
// Returns:
//   - The 32-byte SHA-256 hash
func HashMessage(message []byte) []byte {
	hash := sha256.Sum256(message)
	return hash[:]
}

// ComputePartialSignature computes a party's partial signature component.
//
// This function computes the partial signature s_i for party i using:
//   s_i = k_i * z + alpha_i * d_i
//
// where:
//   - k_i: the party's nonce share
//   - z: the message hash (as a scalar)
//   - alpha_i: the party's additive share of the nonce product
//   - d_i: the party's private key scalar
//
// Parameters:
//   - k_i: the party's nonce share
//   - z: the message hash as a big.Int
//   - alpha_i: the party's additive share
//   - d_i: the party's private key scalar
//
// Returns:
//   - The partial signature s_i
//   - Error if inputs are invalid
func ComputePartialSignature(ki, z, alpha_i, di *big.Int) (*big.Int, error) {
	n := secp256k1.S256().N

	// Validate inputs
	if ki == nil || z == nil || alpha_i == nil || di == nil {
		return nil, ErrInvalidScalar
	}

	// Compute s_i = k_i * z + alpha_i * d_i mod n
	term1 := new(big.Int).Mul(ki, z)
	term2 := new(big.Int).Mul(alpha_i, di)
	s_i := new(big.Int).Add(term1, term2)
	s_i.Mod(s_i, n)

	return s_i, nil
}

// CombinePartialSignatures adds two partial signatures modulo n.
//
// This function combines the partial signatures s_a and s_b from both parties
// to produce the final signature s = s_a + s_b mod n.
//
// Parameters:
//   - s_a: Party A's partial signature
//   - s_b: Party B's partial signature
//
// Returns:
//   - The combined signature s
//   - Error if inputs are invalid
func CombinePartialSignatures(sa, sb *big.Int) (*big.Int, error) {
	if sa == nil || sb == nil {
		return nil, ErrInvalidScalar
	}

	n := secp256k1.S256().N

	// Compute s = s_a + s_b mod n
	s := new(big.Int).Add(sa, sb)
	s.Mod(s, n)

	return s, nil
}

// VerifyECDSASignature verifies an ECDSA signature (r, s) against a public key.
//
// This function verifies that the signature (r, s) is valid for the given
// message hash and public key using standard ECDSA verification on secp256k1.
//
// Parameters:
//   - r: the r component of the signature
//   - s: the s component of the signature
//   - publicKey: the public key to verify against
//   - hash: the message hash
//
// Returns:
//   - true if the signature is valid
//   - false if the signature is invalid
//   - Error if inputs are invalid
func VerifyECDSASignature(r, s *big.Int, publicKey *secp256k1.PublicKey, hash []byte) (bool, error) {
	if r == nil || s == nil || publicKey == nil {
		return false, ErrInvalidSignature
	}

	if len(hash) != 32 {
		return false, ErrInvalidHash
	}

	n := secp256k1.S256().N

	// Validate r and s are in valid range [1, n-1]
	if r.Cmp(big.NewInt(1)) < 0 || r.Cmp(n) >= 0 {
		return false, nil
	}
	if s.Cmp(big.NewInt(1)) < 0 || s.Cmp(n) >= 0 {
		return false, nil
	}

	// ECDSA verification:
	// 1. Compute w = s^(-1) mod n
	w := new(big.Int).ModInverse(s, n)
	if w == nil {
		return false, ErrInvalidSignature
	}

	// 2. Compute u1 = hash * w mod n
	u1 := new(big.Int).Mul(new(big.Int).SetBytes(hash), w)
	u1.Mod(u1, n)

	// 3. Compute u2 = r * w mod n
	u2 := new(big.Int).Mul(r, w)
	u2.Mod(u2, n)

	// 4. Compute point (x, y) = u1 * G + u2 * Q
	// where G is the generator and Q is the public key

	// u1 * G (scalar base multiplication)
	u1Pub := scalarBaseMult(u1)
	if u1Pub == nil {
		return false, ErrInvalidSignature
	}

	// u2 * Q (scalar multiplication with public key)
	u2Pub := scalarMult(publicKey, u2)
	if u2Pub == nil {
		return false, ErrInvalidSignature
	}

	// (x, y) = u1 * G + u2 * Q
	x, _ := secp256k1.S256().Add(u1Pub.X(), u1Pub.Y(), u2Pub.X(), u2Pub.Y())

	// 5. Verify r = x mod n
	xModN := new(big.Int).Mod(x, n)

	return xModN.Cmp(r) == 0, nil
}

// ErrInvalidPublicKey is returned when a public key is invalid
var ErrInvalidPublicKey = &SigningError{msg: "public key is invalid"}

// ErrInvalidSignature is returned when a signature is invalid
var ErrInvalidSignature = &SigningError{msg: "signature is invalid"}

// ErrInvalidHash is returned when the hash is invalid
var ErrInvalidHash = &SigningError{msg: "hash must be 32 bytes"}

// SigningError represents an error with signing protocol parameters
type SigningError struct {
	msg string
}

func (e *SigningError) Error() string {
	return e.msg
}
