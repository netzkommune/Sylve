package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

func GenerateSignature(input string, expires int64, secretKey []byte) string {
	h := hmac.New(sha256.New, secretKey)
	io.WriteString(h, fmt.Sprintf("%s:%d", input, expires))
	return hex.EncodeToString(h.Sum(nil))
}
