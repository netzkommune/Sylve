package utilitiesModels

import "time"

type DownloadedFile struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	DownloadID int       `json:"downloadId" gorm:"not null"`
	Download   Downloads `json:"download" gorm:"foreignKey:DownloadID;constraint:OnDelete:CASCADE"`
	Name       string    `json:"name" gorm:"not null"`
	Size       int64     `json:"size" gorm:"not null"`
}

type Downloads struct {
	ID        int              `json:"id" gorm:"primaryKey"`
	UUID      string           `json:"uuid" gorm:"unique;not null"`
	Path      string           `json:"path" gorm:"unique;not null"`
	Name      string           `json:"name" gorm:"not null"`
	Type      string           `json:"type" gorm:"not null"`
	URL       string           `json:"url" gorm:"unique;not null"`
	Progress  int              `json:"progress" gorm:"not null"`
	Size      int64            `json:"size" gorm:"not null"`
	Files     []DownloadedFile `json:"files" gorm:"foreignKey:DownloadID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updatedAt" gorm:"autoUpdateTime"`
}
