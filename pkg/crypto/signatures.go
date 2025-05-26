// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

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
