package utilitiesServiceInterfaces

import utilitiesModels "sylve/internal/db/models/utilities"

type UtilitiesServiceInterface interface {
	DownloadFile(url string) error
	ListDownloads() ([]utilitiesModels.Downloads, error)
	SyncDownloadProgress() error
	DeleteDownload(id int) error
}
