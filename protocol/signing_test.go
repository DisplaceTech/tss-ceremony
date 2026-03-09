package protocol

import (
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestCombineNonces tests the CombineNonces function
func TestCombineNonces(t *testing.T) {
	tests := []struct {
		name    string
		Ra      *secp256k1.PublicKey
		Rb      *secp256k1.PublicKey
		wantErr bool
	}{
		{
			name:    "two valid nonce public points",
			Ra:      nil, // Will be set in test
			Rb:      nil, // Will be set in test
			wantErr: false,
		},
		{
			name:    "nil Ra",
			Ra:      nil,
			Rb:      nil, // Will be set in test
			wantErr: true,
		},
		{
			name:    "nil Rb",
			Ra:      nil, // Will be set in test
			Rb:      nil,
			wantErr: true,
		},
	}

	// Generate test nonce public points
	kA, _ := GenerateNonceShare()
	kB, _ := GenerateNonceShare()
	Ra, _ := ComputeNoncePublic(kA)
	Rb, _ := ComputeNoncePublic(kB)

	tests[0].Ra = Ra
	tests[0].Rb = Rb
	tests[1].Rb = Rb
	tests[2].Ra = Ra

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, R, err := CombineNonces(tt.Ra, tt.Rb)
			if (err != nil) != tt.wantErr {
				t.Errorf("CombineNonces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if r == nil {
					t.Errorf("CombineNonces() returned nil r")
				}
				if R == nil {
					t.Errorf("CombineNonces() returned nil R")
				}
			}
		})
	}
}

// TestCombineNoncesRValueInRange tests that r is within field order
func TestCombineNoncesRValueInRange(t *testing.T) {
	// Generate multiple nonce pairs and verify r is always in valid range
	n := secp256k1.S256().N

	for i := 0; i < 100; i++ {
		kA, err := GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}
		kB, err := GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		Ra, err := ComputeNoncePublic(kA)
		if err != nil {
			t.Fatalf("ComputeNoncePublic() error: %v", err)
		}
		Rb, err := ComputeNoncePublic(kB)
		if err != nil {
			t.Fatalf("ComputeNoncePublic() error: %v", err)
		}

		r, R, err := CombineNonces(Ra, Rb)
		if err != nil {
			t.Fatalf("CombineNonces() error: %v", err)
		}

		// Verify r is in valid range [0, n-1]
		if r.Cmp(big.NewInt(0)) < 0 {
			t.Errorf("r is negative: %v", r)
		}
		if r.Cmp(n) >= 0 {
			t.Errorf("r >= n: r=%v, n=%v", r, n)
		}

		// Verify R is a valid point on the curve
		if R.X() == nil || R.Y() == nil {
			t.Errorf("Combined R has nil coordinates")
		}

		// Verify the point is on the curve: y^2 = x^3 + 7 (mod p)
		p := secp256k1.S256().P
		b := big.NewInt(7)

		y2 := new(big.Int).Mul(R.Y(), R.Y())
		y2.Mod(y2, p)

		x3 := new(big.Int).Mul(R.X(), R.X())
		x3.Mul(x3, R.X())
		x3.Mod(x3, p)
		x3.Add(x3, b)
		x3.Mod(x3, p)

		if y2.Cmp(x3) != 0 {
			t.Errorf("Combined R point is not on the curve")
		}
	}
}

// TestCombineNoncesCommutative tests that point addition is commutative
func TestCombineNoncesCommutative(t *testing.T) {
	kA, _ := GenerateNonceShare()
	kB, _ := GenerateNonceShare()
	Ra, _ := ComputeNoncePublic(kA)
	Rb, _ := ComputeNoncePublic(kB)

	r1, R1, err := CombineNonces(Ra, Rb)
	if err != nil {
		t.Fatalf("CombineNonces(Ra, Rb) error: %v", err)
	}

	r2, R2, err := CombineNonces(Rb, Ra)
	if err != nil {
		t.Fatalf("CombineNonces(Rb, Ra) error: %v", err)
	}

	// r values should be equal
	if r1.Cmp(r2) != 0 {
		t.Errorf("CombineNonces is not commutative: r1=%v, r2=%v", r1, r2)
	}

	// R points should be equal
	if R1.X().Cmp(R2.X()) != 0 || R1.Y().Cmp(R2.Y()) != 0 {
		t.Errorf("CombineNonces is not commutative: R1 != R2")
	}
}

// TestCombineNoncesMatchesScalarAddition tests that R = (kA + kB) * G
func TestCombineNoncesMatchesScalarAddition(t *testing.T) {
	kA, _ := GenerateNonceShare()
	kB, _ := GenerateNonceShare()
	Ra, _ := ComputeNoncePublic(kA)
	Rb, _ := ComputeNoncePublic(kB)

	// Get r and R from CombineNonces
	r, R, err := CombineNonces(Ra, Rb)
	if err != nil {
		t.Fatalf("CombineNonces() error: %v", err)
	}

	// Compute expected R by adding scalars first: k = kA + kB mod n
	n := secp256k1.S256().N
	k := new(big.Int).Add(kA, kB)
	k.Mod(k, n)

	// Compute expected R = k * G
	expectedR, err := ComputeNoncePublic(k)
	if err != nil {
		t.Fatalf("ComputeNoncePublic() error: %v", err)
	}

	// Verify R matches expected R
	if R.X().Cmp(expectedR.X()) != 0 || R.Y().Cmp(expectedR.Y()) != 0 {
		t.Errorf("CombineNonces R does not match scalar addition: R=%v, expected=%v",
			R.SerializeUncompressed(), expectedR.SerializeUncompressed())
	}

	// Verify r = x mod n
	expectedRValue := new(big.Int).Mod(expectedR.X(), n)
	if r.Cmp(expectedRValue) != 0 {
		t.Errorf("r does not match x mod n: r=%v, expected=%v", r, expectedRValue)
	}
}

// TestCombineNoncesDeterministic tests deterministic runs with fixed seeds
func TestCombineNoncesDeterministic(t *testing.T) {
	// Use fixed seeds for deterministic testing
	seedA := big.NewInt(12345)
	seedB := big.NewInt(67890)

	// Create nonce public points from fixed seeds
	privA := secp256k1.PrivKeyFromBytes(seedA.Bytes())
	privB := secp256k1.PrivKeyFromBytes(seedB.Bytes())
	Ra := privA.PubKey()
	Rb := privB.PubKey()

	// Run CombineNonces twice
	r1, R1, err := CombineNonces(Ra, Rb)
	if err != nil {
		t.Fatalf("First CombineNonces() error: %v", err)
	}

	r2, R2, err := CombineNonces(Ra, Rb)
	if err != nil {
		t.Fatalf("Second CombineNonces() error: %v", err)
	}

	// Results should be identical
	if r1.Cmp(r2) != 0 {
		t.Errorf("CombineNonces is not deterministic: r1=%v, r2=%v", r1, r2)
	}

	if R1.X().Cmp(R2.X()) != 0 || R1.Y().Cmp(R2.Y()) != 0 {
		t.Errorf("CombineNonces is not deterministic: R1 != R2")
	}
}
