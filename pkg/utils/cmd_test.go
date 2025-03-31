// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"os/exec"
	"testing"
)

func TestRunCommand_Success(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("echo", "Hello, world!")
	}

	output, err := RunCommand("dummy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "Hello, world!\n" {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestRunCommand_Failure(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	output, err := RunCommand("false")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}
