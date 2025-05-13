package utils

import (
	"bytes"
	"crypto/sha512"
)

func DeterministicEntropy(seed []byte) []byte {
	hash := sha512.Sum512(seed)
	var entropy bytes.Buffer
	for i := 0; i < 1024; i++ {
		h := sha512.Sum512(append(hash[:], byte(i)))
		entropy.Write(h[:])
	}
	return entropy.Bytes()
}
