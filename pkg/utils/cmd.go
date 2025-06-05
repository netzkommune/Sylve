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
	"context"
	"fmt"
	"os/exec"
)

var execCommand = exec.Command

func RunCommand(command string, args ...string) (string, error) {
	cmd := execCommand(command, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		return output, fmt.Errorf("command execution failed: %v, output: %s", err, output)
	}

	return output, nil
}

func RunCommandWithContext(ctx context.Context, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		return output, fmt.Errorf("command execution failed: %v, output: %s", err, output)
	}

	return output, nil
}
