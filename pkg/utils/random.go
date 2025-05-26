// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

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
