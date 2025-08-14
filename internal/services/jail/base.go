// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package jail

import (
	"fmt"
	"os"
	"strings"
	"sylve/internal/config"
	utilitiesModels "sylve/internal/db/models/utilities"
	"sylve/pkg/utils"
)

func (s *Service) FindBaseByUUID(uuid string) (string, error) {
	if uuid == "" {
		return "", fmt.Errorf("base_download_uuid_required")
	}

	var download utilitiesModels.Downloads
	if err := s.DB.
		Preload("Files").
		Where("uuid = ?", uuid).
		First(&download).Error; err != nil {
		return "", fmt.Errorf("failed_to_find_download: %w", err)
	}

	var bPath string

	switch download.Type {
	case "http":
		downloadsDir := config.GetDownloadsPath("http")
		extractsDir := config.GetDownloadsPath("extracted")
		bPath = fmt.Sprintf("%s/%s", downloadsDir, download.Name)

		if strings.HasSuffix(bPath, ".txz") {
			bPath = fmt.Sprintf("%s/%s", extractsDir, download.UUID)
		}
	case "torrent":
		torrentsDir := config.GetDownloadsPath("torrents")
		for _, file := range download.Files {
			if strings.HasSuffix(file.Name, ".txz") {
				bPath = fmt.Sprintf("%s/%s/%s", torrentsDir, uuid, file.Name)
			}
		}
	}

	if bPath == "" {
		return "", fmt.Errorf("base_file_not_found_in_download: %s", uuid)
	}

	if _, err := os.Stat(bPath); os.IsNotExist(err) {
		return "", fmt.Errorf("base_file_not_found: %s", bPath)
	}

	return bPath, nil
}

func (s *Service) ExtractBase(mountPoint, baseTxz string) (string, error) {
	args := []string{"-C", mountPoint, "-xf", baseTxz}
	return utils.RunCommand("tar", args...)
}
