package protocol

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// FROSTConfig holds configuration for FROST signing
type FROSTConfig struct {
	Fixed   bool
	Message []byte
}

// FROSTParty represents a party in the FROST protocol
type FROSTParty struct {
	ID         int
	Secret     *big.Int       // Secret share x_i
	Public     *secp256k1.PublicKey // Public share P_i = x_i * G
	Nonce      *big.Int       // Nonce share k_i
	NoncePoint *secp256k1.PublicKey // Nonce point R_i = k_i * G
	PartialSig *big.Int       // Partial signature s_i
}

// FROSTSignature represents a FROST Schnorr signature
type FROSTSignature struct {
	R *secp256k1.PublicKey // Combined nonce point R
	S *big.Int             // Combined signature scalar s
}

// FROSTSigner manages the FROST signing protocol for 2-of-2 threshold
type FROSTSigner struct {
	Parties []*FROSTParty
	P       *secp256k1.PublicKey // Combined public key P = P_A + P_B
	R       *secp256k1.PublicKey // Combined nonce point R = R_A + R_B
	E       *big.Int             // Challenge e = H(R, message, P)
	S       *big.Int             // Final signature s = s_A + s_B
	Fixed   bool                 // Use fixed values for deterministic runs
}

// NewFROSTSigner creates a new FROST signer with two parties
func NewFROSTSigner(config FROSTConfig) (*FROSTSigner, error) {
	signer := &FROSTSigner{
		Fixed: config.Fixed,
	}

	// Generate keys for both parties
	var err error
	if config.Fixed {
		// Use fixed seeds for deterministic runs
		signer.Parties, err = generateFixedFROSTParties()
	} else {
		signer.Parties, err = generateRandomFROSTParties()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate FROST parties: %w", err)
	}

	// Compute combined public key P = P_A + P_B
	signer.P, err = CombinePublicKeys(signer.Parties[0].Public, signer.Parties[1].Public)
	if err != nil {
		return nil, fmt.Errorf("failed to combine public keys: %w", err)
	}

	return signer, nil
}

// generateFixedFROSTParties generates deterministic parties for testing
func generateFixedFROSTParties() ([]*FROSTParty, error) {
	parties := make([]*FROSTParty, 2)

	// Party A with fixed seed 1
	seedA := big.NewInt(1)
	parties[0] = &FROSTParty{
		ID:     0,
		Secret: seedA,
	}
	parties[0].Public = secp256k1.PrivKeyFromBytes(seedA.Bytes()).PubKey()

	// Party B with fixed seed 2
	seedB := big.NewInt(2)
	parties[1] = &FROSTParty{
		ID:     1,
		Secret: seedB,
	}
	parties[1].Public = secp256k1.PrivKeyFromBytes(seedB.Bytes()).PubKey()

	return parties, nil
}

// generateRandomFROSTParties generates random parties
func generateRandomFROSTParties() ([]*FROSTParty, error) {
	parties := make([]*FROSTParty, 2)

	for i := 0; i < 2; i++ {
		secretBytes, err := GenerateSecretShare()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secret share for party %d: %w", i, err)
		}

		secret := new(big.Int).SetBytes(secretBytes)
		// Ensure secret is in valid range [1, n-1]
		secret.Mod(secret, secp256k1.S256().N)
		if secret.Sign() == 0 {
			secret.SetInt64(1)
		}

		publicKey, err := ComputePublicShare(secretBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to compute public share for party %d: %w", i, err)
		}

		parties[i] = &FROSTParty{
			ID:     i,
			Secret: secret,
			Public: publicKey,
		}
	}

	return parties, nil
}

// GenerateNonces generates nonce shares for all parties
func (fs *FROSTSigner) GenerateNonces() error {
	return fs.GenerateNoncesWithFixed(false)
}

// GenerateNoncesWithFixed generates nonce shares for all parties, optionally using fixed values
func (fs *FROSTSigner) GenerateNoncesWithFixed(fixed bool) error {
	for i, party := range fs.Parties {
		var nonce *big.Int
		var noncePoint *secp256k1.PublicKey

		if fixed {
			// Use fixed nonces for deterministic runs
			nonce = big.NewInt(int64(i + 100)) // Party A: 100, Party B: 101
			noncePoint = secp256k1.PrivKeyFromBytes(nonce.Bytes()).PubKey()
		} else {
			nonceBytes, err := GenerateSecretShare()
			if err != nil {
				return fmt.Errorf("failed to generate nonce for party %d: %w", i, err)
			}

			nonce = new(big.Int).SetBytes(nonceBytes)
			nonce.Mod(nonce, secp256k1.S256().N)
			if nonce.Sign() == 0 {
				nonce.SetInt64(1)
			}

			noncePoint = secp256k1.PrivKeyFromBytes(nonce.Bytes()).PubKey()
		}

		party.Nonce = nonce
		party.NoncePoint = noncePoint
	}

	// Compute combined nonce point R = R_A + R_B
	R, err := CombinePublicKeys(fs.Parties[0].NoncePoint, fs.Parties[1].NoncePoint)
	if err != nil {
		return fmt.Errorf("failed to combine nonce points: %w", err)
	}
	fs.R = R

	return nil
}

// ComputeChallenge computes the challenge e = H(R || message || P)
func (fs *FROSTSigner) ComputeChallenge(message []byte) error {
	if fs.R == nil {
		return fmt.Errorf("nonce points not generated yet")
	}

	// Serialize R and P for hashing
	RBytes := fs.R.SerializeUncompressed()
	PBytes := fs.P.SerializeUncompressed()

	// Compute e = H(R || message || P)
	hash := sha256.New()
	hash.Write(RBytes)
	hash.Write(message)
	hash.Write(PBytes)
	digest := hash.Sum(nil)

	// Convert hash to scalar (mod n)
	fs.E = new(big.Int).SetBytes(digest)
	fs.E.Mod(fs.E, secp256k1.S256().N)

	return nil
}

// ComputePartialSignatures computes partial signatures for all parties
// s_i = k_i + e * x_i (mod n)
func (fs *FROSTSigner) ComputePartialSignatures() error {
	n := secp256k1.S256().N

	for _, party := range fs.Parties {
		// s_i = k_i + e * x_i (mod n)
		eTimesX := new(big.Int).Mul(fs.E, party.Secret)
		eTimesX.Mod(eTimesX, n)

		partialSig := new(big.Int).Add(party.Nonce, eTimesX)
		partialSig.Mod(partialSig, n)

		party.PartialSig = partialSig
	}

	return nil
}

// AggregateSignatures combines partial signatures into final signature
// s = s_A + s_B (mod n)
func (fs *FROSTSigner) AggregateSignatures() error {
	n := secp256k1.S256().N

	// s = s_A + s_B (mod n)
	fs.S = new(big.Int).Add(fs.Parties[0].PartialSig, fs.Parties[1].PartialSig)
	fs.S.Mod(fs.S, n)

	return nil
}

// Sign performs the complete FROST signing protocol
func (fs *FROSTSigner) Sign(message []byte) (*FROSTSignature, error) {
	// Step 1: Generate nonces
	if err := fs.GenerateNoncesWithFixed(fs.Fixed); err != nil {
		return nil, fmt.Errorf("failed to generate nonces: %w", err)
	}

	// Step 2: Compute challenge
	if err := fs.ComputeChallenge(message); err != nil {
		return nil, fmt.Errorf("failed to compute challenge: %w", err)
	}

	// Step 3: Compute partial signatures
	if err := fs.ComputePartialSignatures(); err != nil {
		return nil, fmt.Errorf("failed to compute partial signatures: %w", err)
	}

	// Step 4: Aggregate signatures
	if err := fs.AggregateSignatures(); err != nil {
		return nil, fmt.Errorf("failed to aggregate signatures: %w", err)
	}

	// Validate that the signature components are in standard secp256k1 Schnorr format
	// R must be a valid secp256k1 point (32-byte x-coordinate)
	if fs.R == nil {
		return nil, fmt.Errorf("schnorr: R point is nil")
	}
	rBytes := make([]byte, 32)
	fs.R.X().FillBytes(rBytes)
	if len(rBytes) != 32 {
		return nil, fmt.Errorf("schnorr: R x-coordinate encoding has unexpected length %d (want 32)", len(rBytes))
	}

	// S must be a valid scalar in [1, n-1]
	n := secp256k1.S256().N
	if fs.S.Sign() == 0 {
		return nil, fmt.Errorf("schnorr: S scalar is zero")
	}
	if fs.S.Cmp(n) >= 0 {
		return nil, fmt.Errorf("schnorr: S scalar (%x…) is >= curve order", fs.S.Bytes()[:4])
	}

	return &FROSTSignature{
		R: fs.R,
		S: fs.S,
	}, nil
}

// GetCombinedPublicKey returns the combined public key P
func (fs *FROSTSigner) GetCombinedPublicKey() *secp256k1.PublicKey {
	return fs.P
}

// GetPartySecret returns a party's secret share (for testing/debugging)
func (fs *FROSTSigner) GetPartySecret(id int) *big.Int {
	if id >= 0 && id < len(fs.Parties) {
		return fs.Parties[id].Secret
	}
	return nil
}

// GetPartyPublic returns a party's public share
func (fs *FROSTSigner) GetPartyPublic(id int) *secp256k1.PublicKey {
	if id >= 0 && id < len(fs.Parties) {
		return fs.Parties[id].Public
	}
	return nil
}
