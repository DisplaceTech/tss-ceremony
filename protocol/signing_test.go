package protocol

import (
	"crypto/rand"
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

// TestComputePartialSignature tests the ComputePartialSignature function
func TestComputePartialSignature(t *testing.T) {
	n := secp256k1.S256().N

	tests := []struct {
		name    string
		ki      *big.Int
		z       *big.Int
		alpha_i *big.Int
		di      *big.Int
		wantErr bool
	}{
		{
			name:    "valid inputs",
			ki:      big.NewInt(12345),
			z:       big.NewInt(67890),
			alpha_i: big.NewInt(11111),
			di:      big.NewInt(22222),
			wantErr: false,
		},
		{
			name:    "nil ki",
			ki:      nil,
			z:       big.NewInt(1),
			alpha_i: big.NewInt(1),
			di:      big.NewInt(1),
			wantErr: true,
		},
		{
			name:    "nil z",
			ki:      big.NewInt(1),
			z:       nil,
			alpha_i: big.NewInt(1),
			di:      big.NewInt(1),
			wantErr: true,
		},
		{
			name:    "nil alpha_i",
			ki:      big.NewInt(1),
			z:       big.NewInt(1),
			alpha_i: nil,
			di:      big.NewInt(1),
			wantErr: true,
		},
		{
			name:    "nil di",
			ki:      big.NewInt(1),
			z:       big.NewInt(1),
			alpha_i: big.NewInt(1),
			di:      nil,
			wantErr: true,
		},
		{
			name:    "large values near field order",
			ki:      new(big.Int).Sub(n, big.NewInt(100)),
			z:       new(big.Int).Sub(n, big.NewInt(200)),
			alpha_i: new(big.Int).Sub(n, big.NewInt(300)),
			di:      new(big.Int).Sub(n, big.NewInt(400)),
			wantErr: false,
		},
		{
			name:    "zero values",
			ki:      big.NewInt(0),
			z:       big.NewInt(0),
			alpha_i: big.NewInt(0),
			di:      big.NewInt(0),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s_i, err := ComputePartialSignature(tt.ki, tt.z, tt.alpha_i, tt.di)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputePartialSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if s_i == nil {
					t.Errorf("ComputePartialSignature() returned nil s_i")
				}
			}
		})
	}
}

// TestComputePartialSignatureValidScalar tests that output is a valid scalar
func TestComputePartialSignatureValidScalar(t *testing.T) {
	n := secp256k1.S256().N

	// Test with many random inputs
	for i := 0; i < 100; i++ {
		ki, err := GenerateNonceShare()
		if err != nil {
			t.Fatalf("GenerateNonceShare() error: %v", err)
		}

		// Generate random z (message hash as scalar)
		z, err := rand.Int(rand.Reader, n)
		if err != nil {
			t.Fatalf("rand.Int() error: %v", err)
		}

		// Generate random alpha_i
		alpha_i, err := rand.Int(rand.Reader, n)
		if err != nil {
			t.Fatalf("rand.Int() error: %v", err)
		}

		// Generate random d_i (private key scalar)
		di, err := rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
		if err != nil {
			t.Fatalf("rand.Int() error: %v", err)
		}
		di.Add(di, big.NewInt(1))

		s_i, err := ComputePartialSignature(ki, z, alpha_i, di)
		if err != nil {
			t.Fatalf("ComputePartialSignature() error: %v", err)
		}

		// Verify s_i is in valid range [0, n-1]
		if s_i.Cmp(big.NewInt(0)) < 0 {
			t.Errorf("s_i is negative: %v", s_i)
		}
		if s_i.Cmp(n) >= 0 {
			t.Errorf("s_i >= n: s_i=%v, n=%v", s_i, n)
		}
	}
}

// TestComputePartialSignatureDeterministic tests deterministic computation
func TestComputePartialSignatureDeterministic(t *testing.T) {
	// Use fixed inputs
	ki := big.NewInt(12345)
	z := big.NewInt(67890)
	alpha_i := big.NewInt(11111)
	di := big.NewInt(22222)

	// Run twice
	s1, err := ComputePartialSignature(ki, z, alpha_i, di)
	if err != nil {
		t.Fatalf("First ComputePartialSignature() error: %v", err)
	}

	s2, err := ComputePartialSignature(ki, z, alpha_i, di)
	if err != nil {
		t.Fatalf("Second ComputePartialSignature() error: %v", err)
	}

	// Results should be identical
	if s1.Cmp(s2) != 0 {
		t.Errorf("ComputePartialSignature is not deterministic: s1=%v, s2=%v", s1, s2)
	}
}

// TestComputePartialSignatureFormula tests the formula s_i = k_i * z + alpha_i * d_i mod n
func TestComputePartialSignatureFormula(t *testing.T) {
	// Use known values
	ki := big.NewInt(10)
	z := big.NewInt(20)
	alpha_i := big.NewInt(30)
	di := big.NewInt(40)

	// Expected: s_i = 10 * 20 + 30 * 40 = 200 + 1200 = 1400
	expected := big.NewInt(1400)

	s_i, err := ComputePartialSignature(ki, z, alpha_i, di)
	if err != nil {
		t.Fatalf("ComputePartialSignature() error: %v", err)
	}

	if s_i.Cmp(expected) != 0 {
		t.Errorf("ComputePartialSignature() = %v, expected %v", s_i, expected)
	}
}

// TestComputePartialSignatureModulo tests that result is properly reduced mod n
func TestComputePartialSignatureModulo(t *testing.T) {
	n := secp256k1.S256().N

	// Use values that would overflow n without modulo
	ki := new(big.Int).Sub(n, big.NewInt(1))
	z := new(big.Int).Sub(n, big.NewInt(1))
	alpha_i := new(big.Int).Sub(n, big.NewInt(1))
	di := new(big.Int).Sub(n, big.NewInt(1))

	s_i, err := ComputePartialSignature(ki, z, alpha_i, di)
	if err != nil {
		t.Fatalf("ComputePartialSignature() error: %v", err)
	}

	// Verify result is in valid range
	if s_i.Cmp(big.NewInt(0)) < 0 {
		t.Errorf("s_i is negative: %v", s_i)
	}
	if s_i.Cmp(n) >= 0 {
		t.Errorf("s_i >= n: s_i=%v, n=%v", s_i, n)
	}

	// Manually compute expected value
	term1 := new(big.Int).Mul(ki, z)
	term2 := new(big.Int).Mul(alpha_i, di)
	expected := new(big.Int).Add(term1, term2)
	expected.Mod(expected, n)

	if s_i.Cmp(expected) != 0 {
		t.Errorf("ComputePartialSignature() = %v, expected %v", s_i, expected)
	}
}

// TestComputePartialSignatureEdgeCases tests edge cases
func TestComputePartialSignatureEdgeCases(t *testing.T) {
	n := secp256k1.S256().N

	tests := []struct {
		name    string
		ki      *big.Int
		z       *big.Int
		alpha_i *big.Int
		di      *big.Int
	}{
		{
			name:    "ki = 1, z = 1, alpha_i = 1, di = 1",
			ki:      big.NewInt(1),
			z:       big.NewInt(1),
			alpha_i: big.NewInt(1),
			di:      big.NewInt(1),
		},
		{
			name:    "ki = n-1, z = 1, alpha_i = 0, di = 0",
			ki:      new(big.Int).Sub(n, big.NewInt(1)),
			z:       big.NewInt(1),
			alpha_i: big.NewInt(0),
			di:      big.NewInt(0),
		},
		{
			name:    "ki = 0, z = n-1, alpha_i = 0, di = 0",
			ki:      big.NewInt(0),
			z:       new(big.Int).Sub(n, big.NewInt(1)),
			alpha_i: big.NewInt(0),
			di:      big.NewInt(0),
		},
		{
			name:    "ki = 0, z = 0, alpha_i = n-1, di = n-1",
			ki:      big.NewInt(0),
			z:       big.NewInt(0),
			alpha_i: new(big.Int).Sub(n, big.NewInt(1)),
			di:      new(big.Int).Sub(n, big.NewInt(1)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s_i, err := ComputePartialSignature(tt.ki, tt.z, tt.alpha_i, tt.di)
			if err != nil {
				t.Fatalf("ComputePartialSignature() error: %v", err)
			}

			// Verify result is in valid range
			if s_i.Cmp(big.NewInt(0)) < 0 {
				t.Errorf("s_i is negative: %v", s_i)
			}
			if s_i.Cmp(n) >= 0 {
				t.Errorf("s_i >= n: s_i=%v, n=%v", s_i, n)
			}
		})
	}
}
