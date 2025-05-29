package system

import (
	systemServiceInterfaces "sylve/internal/interfaces/services/system"
	"sync"

	"gorm.io/gorm"
)

var _ systemServiceInterfaces.SystemServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	syncMutex sync.Mutex
	achMutex  sync.Mutex
}

func NewSystemService(db *gorm.DB) systemServiceInterfaces.SystemServiceInterface {
	return &Service{
		DB: db,
	}
}
