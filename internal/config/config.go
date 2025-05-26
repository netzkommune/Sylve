// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sylve/internal"
)

var ParsedConfig *internal.SylveConfig

func ParseConfig(path string) *internal.SylveConfig {
	file, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	decoder := json.NewDecoder(file)
	ParsedConfig = &internal.SylveConfig{}
	err = decoder.Decode(ParsedConfig)

	if err != nil {
		log.Fatal(err)
	}

	err = SetupDataPath()

	if err != nil {
		log.Fatal(err)
	}

	return ParsedConfig
}

func SetupDataPath() error {
	if ParsedConfig.DataPath == "" {
		ParsedConfig.DataPath = "./data"
	}

	dirs := []string{
		ParsedConfig.DataPath,
		filepath.Join(ParsedConfig.DataPath, "downloads"),
		filepath.Join(ParsedConfig.DataPath, "downloads", "torrents"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func GetDownloadsPath(dType string) string {
	if ParsedConfig == nil {
		return "./data/downloads"
	}

	if dType == "torrents" {
		return filepath.Join(ParsedConfig.DataPath, "downloads", "torrents")
	} else if dType == "torrent.db" {
		return filepath.Join(ParsedConfig.DataPath, "downloads", "torrents", "torrent.db")
	} else if dType == "http" {
		return filepath.Join(ParsedConfig.DataPath, "downloads", "http")
	}

	return filepath.Join(ParsedConfig.DataPath, "downloads")
}
