package libvirt

import (
	"fmt"
	"os"
	"strings"
	"sylve/internal/config"
	utilitiesModels "sylve/internal/db/models/utilities"
)

func (s *Service) FindISOByUUID(uuid string) (string, error) {
	var download utilitiesModels.Downloads
	if err := s.DB.
		Preload("Files").
		Where("uuid = ?", uuid).
		First(&download).Error; err != nil {
		return "", fmt.Errorf("failed_to_find_download: %w", err)
	}

	switch download.Type {
	case "http":
		downloadsDir := config.GetDownloadsPath("http")
		isoPath := fmt.Sprintf("%s/%s", downloadsDir, download.Name)
		if _, err := os.Stat(isoPath); os.IsNotExist(err) {
			return "", fmt.Errorf("iso_not_found: %s", isoPath)
		}
		return isoPath, nil

	case "torrent":
		torrentsDir := config.GetDownloadsPath("torrents")
		for _, file := range download.Files {
			if strings.HasSuffix(file.Name, ".iso") {
				isoPath := fmt.Sprintf("%s/%s/%s", torrentsDir, uuid, file.Name)
				if _, err := os.Stat(isoPath); os.IsNotExist(err) {
					return "", fmt.Errorf("iso_not_found: %s", isoPath)
				}
				return isoPath, nil
			}
		}

		return "", fmt.Errorf("iso_not_found_in_torrent: %s", uuid)

	default:
		return "", fmt.Errorf("unsupported_download_type: %s", download.Type)
	}
}
