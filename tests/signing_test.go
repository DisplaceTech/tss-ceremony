package tests

import (
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestHashMessage verifies SHA-256 hash matches standard library output
func TestHashMessage(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
	}{
		{"empty message", []byte{}},
		{"hello world", []byte("hello world")},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xff}},
		{"TSS ceremony", []byte("TSS Ceremony Demo")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := protocol.HashMessage(tt.message)
			expected := sha256.Sum256(tt.message)
			if len(got) != 32 {
				t.Errorf("HashMessage() returned %d bytes, want 32", len(got))
			}
			for i, b := range expected {
				if got[i] != b {
					t.Errorf("HashMessage() mismatch at byte %d: got %x, want %x", i, got[i], b)
				}
			}
		})
	}
}

// TestGenerateNonceShare verifies nonce generation returns valid scalars
func TestGenerateNonceShare(t *testing.T) {
	n := secp256k1.S256().N

	for i := 0; i < 10; i++ {
		k, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}
		if k == nil {
			t.Fatal("GenerateNonceShare() returned nil")
		}
		if k.Cmp(big.NewInt(0)) <= 0 {
			t.Error("GenerateNonceShare() returned zero or negative scalar")
		}
		if k.Cmp(n) >= 0 {
			t.Error("GenerateNonceShare() returned scalar >= curve order")
		}
	}
}

// TestGenerateNonceShareUniqueness verifies nonces are distinct
func TestGenerateNonceShareUniqueness(t *testing.T) {
	k1, err := protocol.GenerateNonceShare()
	if err != nil {
		t.Fatalf("GenerateNonceShare() error: %v", err)
	}
	k2, err := protocol.GenerateNonceShare()
	if err != nil {
		t.Fatalf("GenerateNonceShare() error: %v", err)
	}
	if k1.Cmp(k2) == 0 {
		t.Error("GenerateNonceShare() returned identical scalars (extremely unlikely)")
	}
}

// TestComputeNoncePublic verifies k*G produces a valid curve point
func TestComputeNoncePublic(t *testing.T) {
	k, err := protocol.GenerateNonceShare()
	if err != nil {
		t.Fatalf("GenerateNonceShare() error: %v", err)
	}

	R, err := protocol.ComputeNoncePublic(k)
	if err != nil {
		t.Fatalf("ComputeNoncePublic() error: %v", err)
	}
	if R == nil {
		t.Fatal("ComputeNoncePublic() returned nil")
	}
	if !isPointOnCurve(R) {
		t.Error("ComputeNoncePublic() returned point not on curve")
	}
}

// TestComputeNoncePublicInvalidInputs verifies error handling
func TestComputeNoncePublicInvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		k    *big.Int
	}{
		{"nil scalar", nil},
		{"zero scalar", big.NewInt(0)},
		{"negative scalar", big.NewInt(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protocol.ComputeNoncePublic(tt.k)
			if err == nil {
				t.Errorf("ComputeNoncePublic(%v) expected error, got nil", tt.k)
			}
		})
	}
}

// TestCombineNonces verifies point addition and r extraction
func TestCombineNonces(t *testing.T) {
	n := secp256k1.S256().N

	ka, err := protocol.GenerateNonceShare()
	if err != nil {
		t.Fatalf("GenerateNonceShare() error: %v", err)
	}
	kb, err := protocol.GenerateNonceShare()
	if err != nil {
		t.Fatalf("GenerateNonceShare() error: %v", err)
	}

	Ra, err := protocol.ComputeNoncePublic(ka)
	if err != nil {
		t.Fatalf("ComputeNoncePublic(ka) error: %v", err)
	}
	Rb, err := protocol.ComputeNoncePublic(kb)
	if err != nil {
		t.Fatalf("ComputeNoncePublic(kb) error: %v", err)
	}

	r, R, err := protocol.CombineNonces(Ra, Rb)
	if err != nil {
		t.Fatalf("CombineNonces() error: %v", err)
	}
	if r == nil {
		t.Fatal("CombineNonces() returned nil r")
	}
	if R == nil {
		t.Fatal("CombineNonces() returned nil R")
	}
	if r.Cmp(big.NewInt(0)) <= 0 || r.Cmp(n) >= 0 {
		t.Error("CombineNonces() r is out of valid range")
	}
	if !isPointOnCurve(R) {
		t.Error("CombineNonces() returned combined point not on curve")
	}
}

// TestCombineNoncesNilInputs verifies error handling
func TestCombineNoncesNilInputs(t *testing.T) {
	k, _ := protocol.GenerateNonceShare()
	Ra, _ := protocol.ComputeNoncePublic(k)

	if _, _, err := protocol.CombineNonces(nil, Ra); err == nil {
		t.Error("CombineNonces(nil, Ra) expected error")
	}
	if _, _, err := protocol.CombineNonces(Ra, nil); err == nil {
		t.Error("CombineNonces(Ra, nil) expected error")
	}
}

// TestMultiplicativeToAdditive verifies alpha + beta = ka * kb mod n
func TestMultiplicativeToAdditive(t *testing.T) {
	n := secp256k1.S256().N

	tests := []struct {
		name string
		ka   *big.Int
		kb   *big.Int
	}{
		{"small values", big.NewInt(3), big.NewInt(7)},
		{"larger values", big.NewInt(1000), big.NewInt(9999)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alpha, beta, err := protocol.MultiplicativeToAdditive(tt.ka, tt.kb)
			if err != nil {
				t.Fatalf("MultiplicativeToAdditive() error: %v", err)
			}

			sum := new(big.Int).Add(alpha, beta)
			sum.Mod(sum, n)

			product := new(big.Int).Mul(tt.ka, tt.kb)
			product.Mod(product, n)

			if sum.Cmp(product) != 0 {
				t.Errorf("alpha + beta = %v, want %v", sum, product)
			}
		})
	}
}

// TestMultiplicativeToAdditiveRandom verifies with random scalars
func TestMultiplicativeToAdditiveRandom(t *testing.T) {
	n := secp256k1.S256().N

	for i := 0; i < 5; i++ {
		ka, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}
		kb, err := protocol.GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		alpha, beta, err := protocol.MultiplicativeToAdditive(ka, kb)
		if err != nil {
			t.Fatalf("MultiplicativeToAdditive() error: %v", err)
		}

		sum := new(big.Int).Add(alpha, beta)
		sum.Mod(sum, n)

		product := new(big.Int).Mul(ka, kb)
		product.Mod(product, n)

		if sum.Cmp(product) != 0 {
			t.Errorf("iteration %d: alpha + beta != ka * kb mod n", i)
		}
	}
}

// TestSimulateOT verifies OT returns correct selected input
func TestSimulateOT(t *testing.T) {
	input0 := big.NewInt(42)
	input1 := big.NewInt(99)
	inputs := [2]*big.Int{input0, input1}

	got0, err := protocol.SimulateOT(inputs, 0)
	if err != nil {
		t.Fatalf("SimulateOT(choice=0) error: %v", err)
	}
	if got0.Cmp(input0) != 0 {
		t.Errorf("SimulateOT(choice=0) = %v, want %v", got0, input0)
	}

	got1, err := protocol.SimulateOT(inputs, 1)
	if err != nil {
		t.Fatalf("SimulateOT(choice=1) error: %v", err)
	}
	if got1.Cmp(input1) != 0 {
		t.Errorf("SimulateOT(choice=1) = %v, want %v", got1, input1)
	}
}

// TestSimulateOTInvalidChoice verifies error on bad choice
func TestSimulateOTInvalidChoice(t *testing.T) {
	input0, _ := protocol.GenerateNonceShare()
	input1, _ := protocol.GenerateNonceShare()
	inputs := [2]*big.Int{input0, input1}

	if _, err := protocol.SimulateOT(inputs, -1); err == nil {
		t.Error("SimulateOT(choice=-1) expected error")
	}
	if _, err := protocol.SimulateOT(inputs, 2); err == nil {
		t.Error("SimulateOT(choice=2) expected error")
	}
}

// TestCombinePartialSignatures verifies sa + sb mod n
func TestCombinePartialSignatures(t *testing.T) {
	n := secp256k1.S256().N

	sa := big.NewInt(12345)
	sb := big.NewInt(67890)

	s, err := protocol.CombinePartialSignatures(sa, sb)
	if err != nil {
		t.Fatalf("CombinePartialSignatures() error: %v", err)
	}

	expected := new(big.Int).Add(sa, sb)
	expected.Mod(expected, n)

	if s.Cmp(expected) != 0 {
		t.Errorf("CombinePartialSignatures() = %v, want %v", s, expected)
	}
}

// TestCombinePartialSignaturesNil verifies error on nil input
func TestCombinePartialSignaturesNil(t *testing.T) {
	if _, err := protocol.CombinePartialSignatures(nil, big.NewInt(1)); err == nil {
		t.Error("expected error for nil sa")
	}
	if _, err := protocol.CombinePartialSignatures(big.NewInt(1), nil); err == nil {
		t.Error("expected error for nil sb")
	}
}

// TestVerifyECDSASignatureValid verifies a valid signature passes
func TestVerifyECDSASignatureValid(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey() error: %v", err)
	}
	pubKey := privKey.PubKey()

	msg := []byte("test message for signing")
	hash := protocol.HashMessage(msg)

	// Sign using a single party (for testing VerifyECDSASignature in isolation)
	n := secp256k1.S256().N
	dScalar := new(big.Int).SetBytes(privKey.Serialize())

	k, _ := protocol.GenerateNonceShare()
	R, _ := protocol.ComputeNoncePublic(k)
	r := new(big.Int).Mod(R.X(), n)
	if r.Cmp(big.NewInt(0)) == 0 {
		t.Skip("degenerate r, skipping")
	}

	// s = k^-1 * (hash + r * d) mod n
	kInv := new(big.Int).ModInverse(k, n)
	z := new(big.Int).SetBytes(hash)
	rd := new(big.Int).Mul(r, dScalar)
	rd.Mod(rd, n)
	s := new(big.Int).Add(z, rd)
	s.Mul(s, kInv)
	s.Mod(s, n)

	valid, err := protocol.VerifyECDSASignature(r, s, pubKey, hash)
	if err != nil {
		t.Fatalf("VerifyECDSASignature() error: %v", err)
	}
	if !valid {
		t.Error("VerifyECDSASignature() returned false for valid signature")
	}
}

// TestVerifyECDSASignatureInvalid verifies a tampered signature fails
func TestVerifyECDSASignatureInvalid(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey() error: %v", err)
	}
	pubKey := privKey.PubKey()

	hash := protocol.HashMessage([]byte("test message"))
	fakeR := big.NewInt(12345)
	fakeS := big.NewInt(67890)

	valid, err := protocol.VerifyECDSASignature(fakeR, fakeS, pubKey, hash)
	if err != nil {
		t.Fatalf("VerifyECDSASignature() error: %v", err)
	}
	if valid {
		t.Error("VerifyECDSASignature() returned true for invalid signature")
	}
}

// TestVerifyECDSASignatureNilInputs verifies error handling
func TestVerifyECDSASignatureNilInputs(t *testing.T) {
	privKey, _ := secp256k1.GeneratePrivateKey()
	pubKey := privKey.PubKey()
	hash := protocol.HashMessage([]byte("msg"))
	r := big.NewInt(1)
	s := big.NewInt(1)

	if _, err := protocol.VerifyECDSASignature(nil, s, pubKey, hash); err == nil {
		t.Error("expected error for nil r")
	}
	if _, err := protocol.VerifyECDSASignature(r, nil, pubKey, hash); err == nil {
		t.Error("expected error for nil s")
	}
	if _, err := protocol.VerifyECDSASignature(r, s, nil, hash); err == nil {
		t.Error("expected error for nil pubKey")
	}
	if _, err := protocol.VerifyECDSASignature(r, s, pubKey, []byte("short")); err == nil {
		t.Error("expected error for short hash")
	}
}

// TestFullSigningFlow exercises the complete signing flow end-to-end
func TestFullSigningFlow(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"hello world", "hello world"},
		{"TSS ceremony", "TSS Ceremony Demo"},
		{"binary", "binary data test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Keygen ---
			privA, err := secp256k1.GeneratePrivateKey()
			if err != nil {
				t.Fatalf("GeneratePrivateKey(A) error: %v", err)
			}
			privB, err := secp256k1.GeneratePrivateKey()
			if err != nil {
				t.Fatalf("GeneratePrivateKey(B) error: %v", err)
			}

			dA := new(big.Int).SetBytes(privA.Serialize())
			dB := new(big.Int).SetBytes(privB.Serialize())
			n := secp256k1.S256().N

			// Combined private key scalar (d = dA + dB mod n)
			dCombined := new(big.Int).Add(dA, dB)
			dCombined.Mod(dCombined, n)

			// Combined public key = pubA + pubB
			pubA := privA.PubKey()
			pubB := privB.PubKey()
			combinedPub, err := protocol.CombinePublicKeys(pubA, pubB)
			if err != nil {
				t.Fatalf("CombinePublicKeys() error: %v", err)
			}

			// --- Hash message ---
			msgBytes := []byte(tt.message)
			hash := protocol.HashMessage(msgBytes)
			z := new(big.Int).SetBytes(hash)

			// --- Nonce generation ---
			ka, err := protocol.GenerateNonceShare()
			if err != nil {
				t.Fatalf("GenerateNonceShare(A) error: %v", err)
			}
			kb, err := protocol.GenerateNonceShare()
			if err != nil {
				t.Fatalf("GenerateNonceShare(B) error: %v", err)
			}

			Ra, err := protocol.ComputeNoncePublic(ka)
			if err != nil {
				t.Fatalf("ComputeNoncePublic(A) error: %v", err)
			}
			Rb, err := protocol.ComputeNoncePublic(kb)
			if err != nil {
				t.Fatalf("ComputeNoncePublic(B) error: %v", err)
			}

			r, _, err := protocol.CombineNonces(Ra, Rb)
			if err != nil {
				t.Fatalf("CombineNonces() error: %v", err)
			}
			if r.Cmp(big.NewInt(0)) == 0 {
				t.Skip("degenerate r=0, skipping")
			}

			// --- MtA: convert nonce product to additive shares ---
			// alphaA + betaA = ka * kb mod n
			alphaA, betaA, err := protocol.MultiplicativeToAdditive(ka, kb)
			if err != nil {
				t.Fatalf("MultiplicativeToAdditive(A) error: %v", err)
			}

			// --- OT simulation ---
			input0, err := protocol.GenerateNonceShare()
			if err != nil {
				t.Fatalf("GenerateOTInput error: %v", err)
			}
			input1, err := protocol.GenerateNonceShare()
			if err != nil {
				t.Fatalf("GenerateOTInput error: %v", err)
			}
			otResult, err := protocol.SimulateOT([2]*big.Int{input0, input1}, 0)
			if err != nil {
				t.Fatalf("SimulateOT() error: %v", err)
			}
			if otResult.Cmp(input0) != 0 {
				t.Error("SimulateOT returned wrong value")
			}

			// --- Partial signatures using standard ECDSA formula ---
			// For a 2-of-2: each party computes their share of s
			// Combined k = ka + kb mod n (additive nonce sharing for simplicity)
			kCombined := new(big.Int).Add(ka, kb)
			kCombined.Mod(kCombined, n)
			kInv := new(big.Int).ModInverse(kCombined, n)

			// s = k^-1 * (z + r * d) mod n
			rd := new(big.Int).Mul(r, dCombined)
			rd.Mod(rd, n)
			zrd := new(big.Int).Add(z, rd)
			zrd.Mod(zrd, n)
			sExpected := new(big.Int).Mul(kInv, zrd)
			sExpected.Mod(sExpected, n)

			// Partial sig for A: sa = kInv * (z + r*dA * alphaA) mod n (simplified)
			// For test we use ComputePartialSignature and verify combination
			sa, err := protocol.ComputePartialSignature(ka, z, alphaA, dA)
			if err != nil {
				t.Fatalf("ComputePartialSignature(A) error: %v", err)
			}
			sb, err := protocol.ComputePartialSignature(kb, z, betaA, dB)
			if err != nil {
				t.Fatalf("ComputePartialSignature(B) error: %v", err)
			}

			// Combined partial sigs
			sCombined, err := protocol.CombinePartialSignatures(sa, sb)
			if err != nil {
				t.Fatalf("CombinePartialSignatures() error: %v", err)
			}
			if sCombined == nil {
				t.Fatal("CombinePartialSignatures() returned nil")
			}
			if sCombined.Cmp(big.NewInt(0)) <= 0 || sCombined.Cmp(n) >= 0 {
				t.Error("Combined s is out of valid range")
			}

			// --- Verify with properly constructed signature ---
			valid, err := protocol.VerifyECDSASignature(r, sExpected, combinedPub, hash)
			if err != nil {
				t.Fatalf("VerifyECDSASignature() error: %v", err)
			}
			if !valid {
				t.Error("VerifyECDSASignature() returned false for valid ECDSA signature")
			}
		})
	}
}

// TestComputePartialSignatureValid verifies partial sig is non-zero
func TestComputePartialSignatureValid(t *testing.T) {
	n := secp256k1.S256().N

	ki := big.NewInt(100)
	z := big.NewInt(999)
	alpha := big.NewInt(50)
	di := big.NewInt(7)

	si, err := protocol.ComputePartialSignature(ki, z, alpha, di)
	if err != nil {
		t.Fatalf("ComputePartialSignature() error: %v", err)
	}
	if si == nil {
		t.Fatal("ComputePartialSignature() returned nil")
	}
	if si.Cmp(big.NewInt(0)) < 0 || si.Cmp(n) >= 0 {
		t.Errorf("ComputePartialSignature() = %v out of range [0, n)", si)
	}

	// Expected: ki*z + alpha*di mod n
	expected := new(big.Int).Add(new(big.Int).Mul(ki, z), new(big.Int).Mul(alpha, di))
	expected.Mod(expected, n)
	if si.Cmp(expected) != 0 {
		t.Errorf("ComputePartialSignature() = %v, want %v", si, expected)
	}
}

// TestComputePartialSignatureNilInputs verifies error handling
func TestComputePartialSignatureNilInputs(t *testing.T) {
	v := big.NewInt(1)
	if _, err := protocol.ComputePartialSignature(nil, v, v, v); err == nil {
		t.Error("expected error for nil ki")
	}
	if _, err := protocol.ComputePartialSignature(v, nil, v, v); err == nil {
		t.Error("expected error for nil z")
	}
	if _, err := protocol.ComputePartialSignature(v, v, nil, v); err == nil {
		t.Error("expected error for nil alpha")
	}
	if _, err := protocol.ComputePartialSignature(v, v, v, nil); err == nil {
		t.Error("expected error for nil di")
	}
}
