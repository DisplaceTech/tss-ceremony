package protocol

import (
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func TestSimulateOT_ChoiceZero(t *testing.T) {
	input0 := big.NewInt(12345)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	result, err := SimulateOT(senderInputs, 0)
	if err != nil {
		t.Fatalf("SimulateOT returned unexpected error: %v", err)
	}

	if result.Cmp(input0) != 0 {
		t.Errorf("Expected %v, got %v", input0, result)
	}
}

func TestSimulateOT_ChoiceOne(t *testing.T) {
	input0 := big.NewInt(12345)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	result, err := SimulateOT(senderInputs, 1)
	if err != nil {
		t.Fatalf("SimulateOT returned unexpected error: %v", err)
	}

	if result.Cmp(input1) != 0 {
		t.Errorf("Expected %v, got %v", input1, result)
	}
}

func TestSimulateOT_InvalidChoiceNegative(t *testing.T) {
	input0 := big.NewInt(12345)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, -1)
	if err == nil {
		t.Fatal("Expected error for negative choice, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_InvalidChoiceTwo(t *testing.T) {
	input0 := big.NewInt(12345)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 2)
	if err == nil {
		t.Fatal("Expected error for choice=2, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_InvalidChoiceLarge(t *testing.T) {
	input0 := big.NewInt(12345)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 100)
	if err == nil {
		t.Fatal("Expected error for large choice, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_InvalidInputNil(t *testing.T) {
	senderInputs := [2]*big.Int{nil, big.NewInt(67890)}

	_, err := SimulateOT(senderInputs, 0)
	if err == nil {
		t.Fatal("Expected error for nil input, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_InvalidInputZero(t *testing.T) {
	input0 := big.NewInt(0)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 0)
	if err == nil {
		t.Fatal("Expected error for zero input, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_InvalidInputNegative(t *testing.T) {
	input0 := big.NewInt(-1)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 0)
	if err == nil {
		t.Fatal("Expected error for negative input, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_InvalidInputEqualToN(t *testing.T) {
	n := secp256k1.S256().N
	input0 := new(big.Int).Set(n)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 0)
	if err == nil {
		t.Fatal("Expected error for input equal to n, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_InvalidInputGreaterThanN(t *testing.T) {
	n := secp256k1.S256().N
	input0 := new(big.Int).Add(n, big.NewInt(1))
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 0)
	if err == nil {
		t.Fatal("Expected error for input greater than n, got nil")
	}

	if _, ok := err.(*OTError); !ok {
		t.Errorf("Expected OTError, got %T", err)
	}
}

func TestSimulateOT_LargeValidInputs(t *testing.T) {
	n := secp256k1.S256().N
	// Use values close to n but still valid
	input0 := new(big.Int).Sub(n, big.NewInt(100))
	input1 := new(big.Int).Sub(n, big.NewInt(200))
	senderInputs := [2]*big.Int{input0, input1}

	result0, err := SimulateOT(senderInputs, 0)
	if err != nil {
		t.Fatalf("SimulateOT returned unexpected error: %v", err)
	}

	if result0.Cmp(input0) != 0 {
		t.Errorf("Expected %v, got %v", input0, result0)
	}

	result1, err := SimulateOT(senderInputs, 1)
	if err != nil {
		t.Fatalf("SimulateOT returned unexpected error: %v", err)
	}

	if result1.Cmp(input1) != 0 {
		t.Errorf("Expected %v, got %v", input1, result1)
	}
}

func TestSimulateOT_MinimalValidInputs(t *testing.T) {
	// Minimal valid input is 1
	input0 := big.NewInt(1)
	input1 := big.NewInt(1)
	senderInputs := [2]*big.Int{input0, input1}

	result0, err := SimulateOT(senderInputs, 0)
	if err != nil {
		t.Fatalf("SimulateOT returned unexpected error: %v", err)
	}

	if result0.Cmp(input0) != 0 {
		t.Errorf("Expected %v, got %v", input0, result0)
	}

	result1, err := SimulateOT(senderInputs, 1)
	if err != nil {
		t.Fatalf("SimulateOT returned unexpected error: %v", err)
	}

	if result1.Cmp(input1) != 0 {
		t.Errorf("Expected %v, got %v", input1, result1)
	}
}

func TestSimulateOT_ReturnsCopy(t *testing.T) {
	input0 := big.NewInt(12345)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	result, err := SimulateOT(senderInputs, 0)
	if err != nil {
		t.Fatalf("SimulateOT returned unexpected error: %v", err)
	}

	// Modify the result and ensure original is unchanged
	result.Add(result, big.NewInt(1))
	if input0.Cmp(big.NewInt(12345)) != 0 {
		t.Error("Original input was modified, expected SimulateOT to return a copy")
	}
}

func TestSimulateOT_ErrorMessage(t *testing.T) {
	input0 := big.NewInt(12345)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 2)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedMsg := "receiver choice must be 0 or 1"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestSimulateOT_InvalidInputErrorMessage(t *testing.T) {
	input0 := big.NewInt(0)
	input1 := big.NewInt(67890)
	senderInputs := [2]*big.Int{input0, input1}

	_, err := SimulateOT(senderInputs, 0)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedMsg := "sender inputs must be valid scalars in range [1, n-1]"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}
