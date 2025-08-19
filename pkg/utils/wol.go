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
)

func IsWOLPacket(payload []byte) bool {
	if len(payload) < 102 {
		return false
	}
	if !bytes.Equal(payload[:6], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
		return false
	}
	mac := payload[6:12]
	for i := 1; i < 16; i++ {
		start := 6 + i*6
		if !bytes.Equal(payload[start:start+6], mac) {
			return false
		}
	}
	return true
}
