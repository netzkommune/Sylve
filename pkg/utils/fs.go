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
	"io"
	"os"
	"path/filepath"
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

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed_to_open_source: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed_to_create_dest: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed_to_copy_file: %w", err)
	}

	return nil
}

func FindFileInDirectoryByPrefix(dir, prefix string) (string, error) {
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if len(d.Name()) >= len(prefix) && d.Name()[:len(prefix)] == prefix {
			return fmt.Errorf("FOUND:%s", path)
		}

		return nil
	})

	if err != nil && len(err.Error()) > 6 && err.Error()[:6] == "FOUND:" {
		return err.Error()[6:], nil
	}

	if err != nil {
		return "", fmt.Errorf("walk_error: %w", err)
	}

	return "", fmt.Errorf("file_with_prefix_not_found: %s in %s", prefix, dir)
}

func IsAbsPath(path string) bool {
	return len(path) > 0 && os.IsPathSeparator(path[0])
}

func CreateOrTruncateFile(path string, size int64) error {
	if !IsAbsPath(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := f.Truncate(size); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	return nil
}

func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("stat %q: %w", path, err)
	}

	if info.IsDir() {
		return false, fmt.Errorf("%q is a directory, not a file", path)
	}

	return true, nil
}

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	return data, nil
}
