package jail

import (
	"fmt"
	"os"
	"strings"
	"sylve/internal/config"
	utilitiesModels "sylve/internal/db/models/utilities"
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
