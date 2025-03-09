// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"regexp"
	"strconv"
)

func ParseZfsTimeUnit(value string) int64 {
	if value == "-" {
		return 0
	}
	re := regexp.MustCompile(`([\d.]+)([a-zA-Z]*)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return 0
	}
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}
	unit := matches[2]
	switch unit {
	case "us":
		return int64(num)
	case "ms":
		return int64(num * 1000)
	case "s":
		return int64(num * 1000000)
	default:
		return int64(num)
	}
}
