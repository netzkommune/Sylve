// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"fmt"
	"os"
)

func DeleteFile(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("stat %q: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("%q is a directory, not a file", path)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove %q: %w", path, err)
	}

	return nil
}
