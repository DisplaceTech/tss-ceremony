package tests

import (
	"math/big"
	"testing"

	"github.com/DisplaceTech/tss-ceremony/protocol"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// helper: create a valid Schnorr signature for testing
func makeSchnorrSig(t *testing.T) (*secp256k1.PublicKey, *secp256k1.PublicKey, *big.Int, []byte) {
	t.Helper()
	config := protocol.FROSTConfig{Fixed: true, Message: []byte("test")}
	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("NewFROSTSigner: %v", err)
	}
	message := []byte("test message")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return signer.GetCombinedPublicKey(), sig.R, sig.S, message
}

func TestVerifySchnorrSignature_Valid(t *testing.T) {
	pubKey, sigR, sigS, message := makeSchnorrSig(t)
	valid, err := protocol.VerifySchnorrSignature(pubKey, sigR, sigS, message)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid signature to pass verification")
	}
}

func TestVerifySchnorrSignature_WrongMessage(t *testing.T) {
	pubKey, sigR, sigS, _ := makeSchnorrSig(t)
	valid, err := protocol.VerifySchnorrSignature(pubKey, sigR, sigS, []byte("wrong message"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected wrong message to fail verification")
	}
}

func TestVerifySchnorrSignature_WrongR(t *testing.T) {
	pubKey, _, sigS, message := makeSchnorrSig(t)
	fakeR := secp256k1.PrivKeyFromBytes([]byte{0x42}).PubKey()
	valid, err := protocol.VerifySchnorrSignature(pubKey, fakeR, sigS, message)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected wrong R to fail verification")
	}
}

func TestVerifySchnorrSignature_WrongS(t *testing.T) {
	pubKey, sigR, _, message := makeSchnorrSig(t)
	fakeS := big.NewInt(99999)
	valid, err := protocol.VerifySchnorrSignature(pubKey, sigR, fakeS, message)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected wrong S to fail verification")
	}
}

func TestVerifySchnorrSignature_WrongPublicKey(t *testing.T) {
	_, sigR, sigS, message := makeSchnorrSig(t)
	wrongKey := secp256k1.PrivKeyFromBytes([]byte{0x77}).PubKey()
	valid, err := protocol.VerifySchnorrSignature(wrongKey, sigR, sigS, message)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected wrong public key to fail verification")
	}
}

func TestVerifySchnorrSignature_NilPublicKey(t *testing.T) {
	_, sigR, sigS, message := makeSchnorrSig(t)
	_, err := protocol.VerifySchnorrSignature(nil, sigR, sigS, message)
	if err == nil {
		t.Error("expected error for nil public key")
	}
}

func TestVerifySchnorrSignature_NilR(t *testing.T) {
	pubKey, _, sigS, message := makeSchnorrSig(t)
	_, err := protocol.VerifySchnorrSignature(pubKey, nil, sigS, message)
	if err == nil {
		t.Error("expected error for nil R")
	}
}

func TestVerifySchnorrSignature_NilS(t *testing.T) {
	pubKey, sigR, _, message := makeSchnorrSig(t)
	_, err := protocol.VerifySchnorrSignature(pubKey, sigR, nil, message)
	if err == nil {
		t.Error("expected error for nil S")
	}
}

func TestVerifySchnorrSignature_EmptyMessage(t *testing.T) {
	config := protocol.FROSTConfig{Fixed: true, Message: []byte("test")}
	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("NewFROSTSigner: %v", err)
	}
	message := []byte{}
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	pubKey := signer.GetCombinedPublicKey()
	valid, err := protocol.VerifySchnorrSignature(pubKey, sig.R, sig.S, message)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected empty message signature to pass verification")
	}
}

func TestVerifySchnorrSignature_RandomMode(t *testing.T) {
	config := protocol.FROSTConfig{Fixed: false, Message: []byte("test")}
	signer, err := protocol.NewFROSTSigner(config)
	if err != nil {
		t.Fatalf("NewFROSTSigner: %v", err)
	}
	message := []byte("random mode test")
	sig, err := signer.Sign(message)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	pubKey := signer.GetCombinedPublicKey()
	valid, err := protocol.VerifySchnorrSignature(pubKey, sig.R, sig.S, message)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected random mode signature to pass verification")
	}
}
