// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

//go:build freebsd

package sysctl

import (
	"strings"
	"testing"
)

func TestGetString(t *testing.T) {
	val, err := GetString("kern.ostype")
	if err != nil {
		t.Fatalf("GetString failed: %v", err)
	}
	if !strings.HasPrefix(val, "FreeBSD") && val != "Darwin" {
		t.Errorf("unexpected kern.ostype value: %q", val)
	}
}

func TestGetInt64(t *testing.T) {
	val, err := GetInt64("vm.swap_idle_enabled")
	if err != nil {
		t.Errorf("GetInt64 failed: %v", err)
	}
	if val != 0 && val != 1 {
		t.Errorf("Unexpected value for vm.swap_idle_enabled: %d", val)
	}
}

func TestGetBytes(t *testing.T) {
	bytes, err := GetBytes("kern.hostname")
	if err != nil {
		t.Fatalf("GetBytes failed: %v", err)
	}
	if len(bytes) == 0 {
		t.Errorf("GetBytes returned empty data")
	}
}

func TestGetStringNullTermination(t *testing.T) {
	val, err := GetString("kern.hostname")
	if err != nil {
		t.Fatalf("GetString failed: %v", err)
	}
	if strings.ContainsRune(val, '\x00') {
		t.Errorf("value contains unexpected null byte: %q", val)
	}
}
