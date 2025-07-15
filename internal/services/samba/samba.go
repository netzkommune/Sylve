package samba

import (
	sambaServiceInterfaces "sylve/internal/interfaces/services/samba"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"

	"gorm.io/gorm"
)

var _ sambaServiceInterfaces.SambaServiceInterface = (*Service)(nil)

type Service struct {
	DB  *gorm.DB
	ZFS zfsServiceInterfaces.ZfsServiceInterface
}

func NewSambaService(db *gorm.DB, zfs zfsServiceInterfaces.ZfsServiceInterface) sambaServiceInterfaces.SambaServiceInterface {
	return &Service{
		DB:  db,
		ZFS: zfs,
	}
}
