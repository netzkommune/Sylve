package jail

import (
	"fmt"
	"os"
	"strings"
	"sylve/internal/config"
	utilitiesModels "sylve/internal/db/models/utilities"
	"sylve/internal/logger"
	"sylve/pkg/system"
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
		bPath = fmt.Sprintf("%s/%s", downloadsDir, download.Name)
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
	tryArgs := func(usePixz bool) (string, error) {
		args := []string{}
		if usePixz {
			args = append(args, "--use-compress-program=pixz")
		}
		args = append(args, "-C", mountPoint, "-xf", baseTxz)
		return utils.RunCommand("tar", args...)
	}

	if system.PixzExists() {
		output, err := tryArgs(true)
		if err == nil {
			return output, nil
		}
		logger.L.Warn().Err(err).Msg("pixz extraction failed, falling back to default tar")
	}

	return tryArgs(false)
}
