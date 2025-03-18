// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

//go:build linux

package sysctl

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetInt64(name string) (int64, error) {
	data, err := readProcSys(name)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(data), 10, 64)
}

func GetString(name string) (string, error) {
	data, err := readProcSys(name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(data), nil
}

func GetBytes(name string) ([]byte, error) {
	data, err := readProcSys(name)
	if err != nil {
		return nil, err
	}
	return []byte(strings.TrimSpace(data)), nil
}

func readProcSys(name string) (string, error) {
	path := "/proc/sys/" + strings.ReplaceAll(name, ".", "/")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %v", path, err)
	}
	return string(data), nil
}
