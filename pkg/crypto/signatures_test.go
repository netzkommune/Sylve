package crypto_test

import (
	"sylve/pkg/crypto"
	"sylve/pkg/utils"
	"testing"
	"time"
)

func TestGenerateSignature(t *testing.T) {
	secretKey := []byte("supersecretkey")
	input := "freebsd"
	expires := time.Now().Unix()

	expected := crypto.GenerateSignature(input, expires, secretKey)
	actual := crypto.GenerateSignature(input, expires, secretKey)

	if expected != actual {
		t.Errorf("Expected signature %s, got %s", expected, actual)
	}
}

func TestGenerateSignatureDifferentInputs(t *testing.T) {
	secretKey := []byte("supersecretkey")

	sig1 := crypto.GenerateSignature("input1", 1234567890, secretKey)
	sig2 := crypto.GenerateSignature("input2", 1234567890, secretKey)

	if sig1 == sig2 {
		t.Errorf("Expected different signatures for different inputs, but got same: %s", sig1)
	}
}

func TestGenerateSignatureKeyEffect(t *testing.T) {
	secretKey1 := []byte("key1")
	secretKey2 := []byte("key2")

	sig1 := crypto.GenerateSignature("input", 1234567890, secretKey1)
	sig2 := crypto.GenerateSignature("input", 1234567890, secretKey2)

	if sig1 == sig2 {
		t.Errorf("Expected different signatures with different keys, got identical: %s", sig1)
	}
}

func TestGenerateSignatureFormat(t *testing.T) {
	secretKey := []byte("key")
	sig := crypto.GenerateSignature("input", 9999999999, secretKey)

	if !utils.IsHex(sig) {
		t.Errorf("Signature is not a valid hex string: %s", sig)
	}
}
