package zfs

import (
	zfsModels "github.com/alchemillahq/sylve/internal/db/models/zfs"
	zfsServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/zfs"
)

func (s *Service) GetS3Backups() ([]zfsServiceInterfaces.ZFSS3Backup, error) {
	var backups []zfsModels.ZFSS3Backup
	if err := s.DB.Find(&backups).Error; err != nil {
		return nil, err
	}

	if len(backups) == 0 {
		return []zfsServiceInterfaces.ZFSS3Backup{}, nil
	}

	ids := make([]uint, 0, len(backups))
	for _, b := range backups {
		ids = append(ids, b.ID)
	}

	var runs []zfsModels.ZFSS3BackupRun
	if err := s.DB.
		Where("job_id IN ?", ids).
		Order("started_at DESC").
		Find(&runs).
		Error; err != nil {
		return nil, err
	}

	runMap := make(map[uint][]zfsModels.ZFSS3BackupRun, len(ids))
	for _, r := range runs {
		runMap[r.JobID] = append(runMap[r.JobID], r)
	}

	out := make([]zfsServiceInterfaces.ZFSS3Backup, 0, len(backups))
	for _, b := range backups {
		out = append(out, zfsServiceInterfaces.ZFSS3Backup{
			Job:  b,
			Runs: runMap[b.ID],
		})
	}

	return out, nil
}
