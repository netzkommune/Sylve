package utilities

import (
	"fmt"
	utilitiesModels "sylve/internal/db/models/utilities"
	"sylve/internal/logger"
	"sylve/pkg/utils"

	valid "github.com/asaskevich/govalidator"
	"github.com/cenkalti/rain/torrent"
)

func (s *Service) ListDownloads() ([]utilitiesModels.Downloads, error) {
	var downloads []utilitiesModels.Downloads

	if err := s.DB.Preload("Files").Find(&downloads).Error; err != nil {
		logger.L.Error().Msgf("Failed to list downloads: %v", err)
		return nil, err
	}

	return downloads, nil
}

func (s *Service) DownloadFile(url string) error {
	var existing utilitiesModels.Downloads

	if s.DB.Where("url = ?", url).First(&existing).RowsAffected > 0 {
		logger.L.Info().Msgf("Download already exists: %s", url)
		return nil
	}

	if valid.IsMagnetURI(url) {
		torrentOpts := torrent.AddTorrentOptions{
			ID:                utils.GenerateDeterministicUUID(url),
			StopAfterDownload: false,
		}

		t, err := s.BTTClient.AddURI(url, &torrentOpts)

		if err != nil {
			logger.L.Error().Msgf("Failed to add torrent: %v", err)
			return err
		}

		download := utilitiesModels.Downloads{
			URL:      url,
			UUID:     t.ID(),
			Path:     t.Dir(),
			Type:     "torrent",
			Name:     t.Name(),
			Size:     0,
			Progress: 0,
			Files:    []utilitiesModels.DownloadedFile{},
		}

		if err := s.DB.Create(&download).Error; err != nil {
			logger.L.Error().Msgf("Failed to create download record: %v", err)
			return err
		}

		return nil
	} else {
		return fmt.Errorf("invalid_magnet_link")
	}

	// return nil
}

func (s *Service) SyncDownloadProgress() error {
	var downloads []utilitiesModels.Downloads
	if err := s.DB.Where("progress < 100").Find(&downloads).Error; err != nil {
		return err
	}

	for _, download := range downloads {
		if download.Type == "torrent" {
			torrent := s.BTTClient.GetTorrent(download.UUID)
			if torrent == nil {
				logger.L.Error().Msgf("Torrent %s not found", download.UUID)
				continue
			}

			piecesHave := torrent.Stats().Pieces.Have
			piecesTotal := torrent.Stats().Pieces.Total

			if piecesHave == 0 {
				download.Progress = 0
			} else {
				download.Progress = int((piecesHave * 100) / piecesTotal)
			}

			download.Size = torrent.Stats().Bytes.Total
			download.Name = torrent.Stats().Name
			files, err := torrent.Files()

			if err != nil {
				continue
			}

			fileList := make([]string, len(files))
			for i, file := range files {
				fileList[i] = file.Path()
				downloadedFile := utilitiesModels.DownloadedFile{
					DownloadID: download.ID,
					Download:   download,
					Name:       file.Path(),
					Size:       file.Length(),
				}

				var existingFile utilitiesModels.DownloadedFile
				if s.DB.Where("download_id = ? AND name = ?", download.ID, file.Path()).First(&existingFile).RowsAffected > 0 {
					continue
				}

				if err := s.DB.Create(&downloadedFile).Error; err != nil {
					logger.L.Error().Msgf("Failed to create downloaded file record: %v", err)
					continue
				}

				download.Files = append(download.Files, downloadedFile)
			}

			if err := s.DB.Save(&download).Error; err != nil {
				logger.L.Error().Msgf("Failed to update download record: %v", err)
				return err
			}
		}
	}

	return nil
}

func (s *Service) DeleteDownload(id int) error {
	var download utilitiesModels.Downloads
	if err := s.DB.Where("id = ?", id).First(&download).Error; err != nil {
		logger.L.Debug().Msgf("Failed to find download: %v", err)
		return err
	}

	if download.Type == "torrent" {
		torrent := s.BTTClient.GetTorrent(download.UUID)
		if torrent != nil {
			if err := s.BTTClient.RemoveTorrent(download.UUID); err != nil {
				logger.L.Debug().Msgf("Failed to remove torrent: %v", err)
				return err
			}
		}
	}

	for _, file := range download.Files {
		if err := s.DB.Delete(&file).Error; err != nil {
			logger.L.Debug().Msgf("Failed to delete downloaded file: %v", err)
			return err
		}
	}

	if err := s.DB.Delete(&download).Error; err != nil {
		logger.L.Debug().Msgf("Failed to delete download: %v", err)
		return err
	}

	return nil
}

func (s *Service) BulkDeleteDownload(ids []int) error {
	var downloads []utilitiesModels.Downloads
	if err := s.DB.Where("id IN ?", ids).Find(&downloads).Error; err != nil {
		return err
	}

	for _, download := range downloads {
		if download.Type == "torrent" {
			torrent := s.BTTClient.GetTorrent(download.UUID)
			if torrent != nil {
				if err := s.BTTClient.RemoveTorrent(download.UUID); err != nil {
					logger.L.Debug().Msgf("Failed to remove torrent: %v", err)
					return err
				}
			}
		}

		for _, file := range download.Files {
			if err := s.DB.Delete(&file).Error; err != nil {
				logger.L.Debug().Msgf("Failed to delete downloaded file: %v", err)
				return err
			}
		}

		if err := s.DB.Delete(&download).Error; err != nil {
			logger.L.Debug().Msgf("Failed to delete download: %v", err)
			return err
		}
	}

	return nil
}
