package utilities

import (
	"sylve/internal/config"
	utilitiesServiceInterfaces "sylve/internal/interfaces/services/utilities"
	"sylve/internal/logger"

	"github.com/cenkalti/rain/torrent"
	"gorm.io/gorm"
)

var _ utilitiesServiceInterfaces.UtilitiesServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	BTTClient *torrent.Session
}

func NewUtilitiesService(db *gorm.DB) utilitiesServiceInterfaces.UtilitiesServiceInterface {
	torrent.DisableLogging()
	cfg := torrent.DefaultConfig
	cfg.Database = config.GetDownloadsPath("torrent.db")
	cfg.DataDir = config.GetDownloadsPath("torrents")

	session, err := torrent.NewSession(cfg)

	if err != nil {
		logger.L.Fatal().Msgf("Failed to create torrent downloader %v", err)
	}

	return &Service{
		DB:        db,
		BTTClient: session,
	}
}
