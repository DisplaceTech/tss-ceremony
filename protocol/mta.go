package protocol

import (
	"crypto/rand"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// MultiplicativeToAdditive converts the product of two multiplicative shares
// into additive shares (alpha, beta) such that alpha + beta = k_a * k_b mod n.
//
// Given multiplicative shares k_a and k_b, this function computes additive shares
// alpha and beta where:
//   - alpha + beta = k_a * k_b (mod n)
//   - alpha is computed as k_a * k_b - beta
//   - beta is a random value in [0, n)
//
// This is used in the DKLS protocol to convert multiplicative shares of the
// nonce product into additive shares that can be combined to produce the final
// signature.
func MultiplicativeToAdditive(ka, kb *big.Int) (alpha, beta *big.Int, err error) {
	n := secp256k1.S256().N

	// Validate inputs are within field order
	if ka.Cmp(big.NewInt(0)) <= 0 || ka.Cmp(n) >= 0 {
		return nil, nil, ErrInvalidScalar
	}
	if kb.Cmp(big.NewInt(0)) <= 0 || kb.Cmp(n) >= 0 {
		return nil, nil, ErrInvalidScalar
	}

	// Compute the product: product = k_a * k_b mod n
	product := new(big.Int).Mul(ka, kb)
	product.Mod(product, n)

	// Generate a random beta in [1, n-1]
	beta, err = rand.Int(rand.Reader, new(big.Int).Sub(n, big.NewInt(1)))
	if err != nil {
		return nil, nil, err
	}
	// Ensure beta is in [1, n-1]
	beta.Add(beta, big.NewInt(1))

	// Compute alpha = product - beta mod n
	alpha = new(big.Int).Sub(product, beta)
	alpha.Mod(alpha, n)

	return alpha, beta, nil
}

// ErrInvalidScalar is returned when a scalar is not within the valid range
var ErrInvalidScalar = &ScalarError{msg: "scalar must be in range [1, n-1]"}

// ScalarError represents an error with scalar values
type ScalarError struct {
	msg string
}

func (e *ScalarError) Error() string {
	return e.msg
}
