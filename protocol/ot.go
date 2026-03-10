package protocol

import (
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// SimulateOT simulates oblivious transfer semantics for educational demonstration.
//
// In a 1-out-of-2 OT protocol:
//   - The sender has two inputs: input0 and input1
//   - The receiver has a choice bit: 0 or 1
//   - The receiver learns only the input corresponding to their choice
//   - The sender learns nothing about the receiver's choice
//
// This function simulates the OT outcome without implementing full OT crypto,
// ensuring educational correctness for the DKLS ceremony demonstration.
//
// Parameters:
//   - senderInputs: array of two scalars [input0, input1]
//   - receiverChoice: the receiver's choice bit (0 or 1)
//
// Returns:
//   - The selected input based on receiver's choice
//   - Error if inputs are invalid
func SimulateOT(senderInputs [2]*big.Int, receiverChoice int) (*big.Int, error) {
	n := secp256k1.S256().N

	// Validate receiver choice
	if receiverChoice < 0 || receiverChoice > 1 {
		return nil, ErrInvalidOTChoice
	}

	// Validate sender inputs are within field order
	for _, input := range senderInputs {
		if input == nil || input.Cmp(big.NewInt(0)) <= 0 || input.Cmp(n) >= 0 {
			return nil, ErrInvalidOTInput
		}
	}

	// Return the selected input based on receiver's choice
	// In a real OT protocol, this would be computed cryptographically
	// but for educational purposes, we directly return the selected value
	return new(big.Int).Set(senderInputs[receiverChoice]), nil
}

// ErrInvalidOTChoice is returned when the receiver's choice is not 0 or 1
var ErrInvalidOTChoice = &OTError{msg: "receiver choice must be 0 or 1"}

// ErrInvalidOTInput is returned when sender inputs are invalid
var ErrInvalidOTInput = &OTError{msg: "sender inputs must be valid scalars in range [1, n-1]"}

// OTError represents an error with OT protocol parameters
type OTError struct {
	msg string
}

func (e *OTError) Error() string {
	return e.msg
}

// GenerateOTInputs generates two random scalar inputs for the OT sender
func GenerateOTInputs() ([2]*big.Int, error) {
	n := secp256k1.S256().N
	inputs := [2]*big.Int{}

	for i := 0; i < 2; i++ {
		val, err := randInt(new(big.Int).Sub(n, big.NewInt(1)))
		if err != nil {
			return [2]*big.Int{}, err
		}
		val.Add(val, big.NewInt(1)) // Ensure in [1, n-1]
		inputs[i] = val
	}

	return inputs, nil
}
