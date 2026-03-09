package protocol

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func TestGenerateSecretShare(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "generates 32-byte random scalar",
			wantErr: false,
		},
		{
			name:    "multiple calls produce different values",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSecretShare()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecretShare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 32 {
				t.Errorf("GenerateSecretShare() returned %d bytes, want 32", len(got))
			}
		})
	}

	// Test that multiple calls produce different values
	t.Run("randomness distribution", func(t *testing.T) {
		values := make(map[string]bool)
		for i := 0; i < 100; i++ {
			secret, err := GenerateSecretShare()
			if err != nil {
				t.Fatalf("GenerateSecretShare() error: %v", err)
			}
			key := string(secret)
			if values[key] {
				t.Errorf("GenerateSecretShare() produced duplicate value after %d iterations", i)
				break
			}
			values[key] = true
		}
	})
}

func TestGenerateSecretShareFixed(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{"seed 1", 1},
		{"seed 2", 2},
		{"seed 0", 0},
		{"seed large", 123456789012345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSecretShareFixed(tt.seed)
			if len(got) != 32 {
				t.Errorf("GenerateSecretShareFixed() returned %d bytes, want 32", len(got))
			}
			// Same seed should produce same result
			got2 := GenerateSecretShareFixed(tt.seed)
			for i := range got {
				if got[i] != got2[i] {
					t.Errorf("GenerateSecretShareFixed() not deterministic for seed %d", tt.seed)
					break
				}
			}
		})
	}
}

func TestComputePublicShare(t *testing.T) {
	tests := []struct {
		name    string
		secret  []byte
		wantErr bool
	}{
		{
			name:    "valid 32-byte secret",
			secret:  make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "fixed seed secret",
			secret:  GenerateSecretShareFixed(1),
			wantErr: false,
		},
		{
			name:    "invalid 31-byte secret",
			secret:  make([]byte, 31),
			wantErr: true,
		},
		{
			name:    "invalid 33-byte secret",
			secret:  make([]byte, 33),
			wantErr: true,
		},
		{
			name:    "invalid empty secret",
			secret:  make([]byte, 0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid 32-byte secret" {
				_, err := rand.Read(tt.secret)
				if err != nil {
					t.Fatalf("failed to generate random secret: %v", err)
				}
			}

			got, err := ComputePublicShare(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputePublicShare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ComputePublicShare() returned nil public key")
			}
		})
	}

	// Test determinism
	t.Run("deterministic for same input", func(t *testing.T) {
		secret := GenerateSecretShareFixed(42)
		pub1, err := ComputePublicShare(secret)
		if err != nil {
			t.Fatalf("ComputePublicShare() error: %v", err)
		}
		pub2, err := ComputePublicShare(secret)
		if err != nil {
			t.Fatalf("ComputePublicShare() error: %v", err)
		}

		// Compare serialized forms
		serialized1 := pub1.SerializeUncompressed()
		serialized2 := pub2.SerializeUncompressed()
		for i := range serialized1 {
			if serialized1[i] != serialized2[i] {
				t.Errorf("ComputePublicShare() not deterministic")
				break
			}
		}
	})
}

func TestCombinePublicKeys(t *testing.T) {
	tests := []struct {
		name      string
		publicA   *secp256k1.PublicKey
		publicB   *secp256k1.PublicKey
		wantErr   bool
	}{
		{
			name:    "two valid public keys",
			publicA: nil, // Will be set in test
			publicB: nil, // Will be set in test
			wantErr: false,
		},
		{
			name:    "nil publicA",
			publicA: nil,
			publicB: nil, // Will be set in test
			wantErr: true,
		},
		{
			name:    "nil publicB",
			publicA: nil, // Will be set in test
			publicB: nil,
			wantErr: true,
		},
	}

	// Generate test keys
	secretA := GenerateSecretShareFixed(1)
	secretB := GenerateSecretShareFixed(2)
	pubA, _ := ComputePublicShare(secretA)
	pubB, _ := ComputePublicShare(secretB)

	tests[0].publicA = pubA
	tests[0].publicB = pubB
	tests[1].publicB = pubB
	tests[2].publicA = pubA

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CombinePublicKeys(tt.publicA, tt.publicB)
			if (err != nil) != tt.wantErr {
				t.Errorf("CombinePublicKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("CombinePublicKeys() returned nil public key")
			}
		})
	}

	// Test idempotency (commutativity)
	t.Run("commutative", func(t *testing.T) {
		combinedAB, err := CombinePublicKeys(pubA, pubB)
		if err != nil {
			t.Fatalf("CombinePublicKeys(A, B) error: %v", err)
		}
		combinedBA, err := CombinePublicKeys(pubB, pubA)
		if err != nil {
			t.Fatalf("CombinePublicKeys(B, A) error: %v", err)
		}

		serializedAB := combinedAB.SerializeUncompressed()
		serializedBA := combinedBA.SerializeUncompressed()
		for i := range serializedAB {
			if serializedAB[i] != serializedBA[i] {
				t.Errorf("CombinePublicKeys() is not commutative")
				break
			}
		}
	})

	// Test that combined key is valid secp256k1 point
	t.Run("result is valid secp256k1 point", func(t *testing.T) {
		combined, err := CombinePublicKeys(pubA, pubB)
		if err != nil {
			t.Fatalf("CombinePublicKeys() error: %v", err)
		}

		// Verify the point is on the curve
		x := combined.X()
		y := combined.Y()
		if x == nil || y == nil {
			t.Errorf("Combined key has nil coordinates")
		}

		// Verify it's on the curve: y^2 = x^3 + 7 (mod p)
		p := secp256k1.S256().P
		b := big.NewInt(7) // secp256k1 has b=7

		y2 := new(big.Int).Mul(y, y)
		y2.Mod(y2, p)

		x3 := new(big.Int).Mul(x, x)
		x3.Mul(x3, x)
		x3.Mod(x3, p)
		x3.Add(x3, b)
		x3.Mod(x3, p)

		if y2.Cmp(x3) != 0 {
			t.Errorf("Combined key point is not on the curve")
		}
	})
}

func TestKeyGenerationRoundTrip(t *testing.T) {
	// Test that we can generate keys and verify they are valid secp256k1 points
	t.Run("round-trip key generation", func(t *testing.T) {
		// Generate secret shares
		secretA, err := GenerateSecretShare()
		if err != nil {
			t.Fatalf("GenerateSecretShare() error: %v", err)
		}

		secretB, err := GenerateSecretShare()
		if err != nil {
			t.Fatalf("GenerateSecretShare() error: %v", err)
		}

		// Compute public shares
		pubA, err := ComputePublicShare(secretA)
		if err != nil {
			t.Fatalf("ComputePublicShare(A) error: %v", err)
		}

		pubB, err := ComputePublicShare(secretB)
		if err != nil {
			t.Fatalf("ComputePublicShare(B) error: %v", err)
		}

		// Combine public keys
		combined, err := CombinePublicKeys(pubA, pubB)
		if err != nil {
			t.Fatalf("CombinePublicKeys() error: %v", err)
		}

		// Verify all points are valid
		if pubA.X() == nil || pubA.Y() == nil {
			t.Error("Public key A has invalid coordinates")
		}
		if pubB.X() == nil || pubB.Y() == nil {
			t.Error("Public key B has invalid coordinates")
		}
		if combined.X() == nil || combined.Y() == nil {
			t.Error("Combined public key has invalid coordinates")
		}

		// Verify serialization format
		serializedA := pubA.SerializeUncompressed()
		serializedB := pubB.SerializeUncompressed()
		serializedCombined := combined.SerializeUncompressed()

		if len(serializedA) != 65 {
			t.Errorf("Public key A serialization is %d bytes, want 65", len(serializedA))
		}
		if len(serializedB) != 65 {
			t.Errorf("Public key B serialization is %d bytes, want 65", len(serializedB))
		}
		if len(serializedCombined) != 65 {
			t.Errorf("Combined public key serialization is %d bytes, want 65", len(serializedCombined))
		}
	})
}
