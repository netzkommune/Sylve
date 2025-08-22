// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfs

import (
	"context"
	"fmt"
	"time"

	zfsModels "github.com/alchemillahq/sylve/internal/db/models/zfs"
	"github.com/alchemillahq/sylve/internal/logger"
	"github.com/alchemillahq/sylve/pkg/zfs"

	"github.com/robfig/cron/v3"
)

func (s *Service) DeleteSnapshot(guid string, recursive bool) error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	datasets, err := zfs.Snapshots("")

	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		for _, v := range properties {
			if v == guid {
				var err error

				if recursive {
					err = dataset.Destroy(zfs.DestroyRecursive)
				} else {
					err = dataset.Destroy(zfs.DestroyDefault)
				}

				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	return fmt.Errorf("snapshot with guid %s not found", guid)
}

func (s *Service) CreateSnapshot(guid string, name string, recursive bool) error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	datasets, err := zfs.Datasets("")
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		if dataset.Name == dataset.Name+"@"+name {
			return fmt.Errorf("snapshot with name %s already exists", name)
		}

		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		for k, v := range properties {
			if k == "guid" {
				if v == guid {
					shot, err := dataset.Snapshot(name, recursive)
					if err != nil {
						return err
					}

					if shot.Name == dataset.Name+"@"+name {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("dataset with guid %s not found", guid)
}

func (s *Service) GetPeriodicSnapshots() ([]zfsModels.PeriodicSnapshot, error) {
	var snapshots []zfsModels.PeriodicSnapshot

	if err := s.DB.Find(&snapshots).Error; err != nil {
		return nil, err
	}

	return snapshots, nil
}

func (s *Service) AddPeriodicSnapshot(guid string, prefix string, recursive bool, interval int, cronExpr string) error {
	dataset, err := s.GetDatasetByGUID(guid)
	if err != nil {
		return err
	}

	properties, err := dataset.GetAllProperties()
	if err != nil {
		return err
	}

	for k, v := range properties {
		if k == "guid" && v == guid {
			snapshot := zfsModels.PeriodicSnapshot{
				GUID:      guid,
				Prefix:    prefix,
				Recursive: recursive,
				Interval:  interval,
				CronExpr:  cronExpr,
			}

			if err := s.DB.Create(&snapshot).Error; err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("dataset with guid %s not found", guid)
}

func (s *Service) DeletePeriodicSnapshot(guid string) error {
	var snapshot zfsModels.PeriodicSnapshot

	if err := s.DB.Where("guid = ?", guid).First(&snapshot).Error; err != nil {
		return err
	}

	if err := s.DB.Delete(&snapshot).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) StartSnapshotScheduler(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				var snapshotJobs []zfsModels.PeriodicSnapshot
				if err := s.DB.Find(&snapshotJobs).Error; err != nil {
					logger.L.Debug().Err(err).Msg("Failed to load snapshotJobs")
					continue
				}

				now := time.Now()

				for _, job := range snapshotJobs {
					shouldRun := false

					if job.CronExpr != "" {
						sched, err := cron.ParseStandard(job.CronExpr)
						if err != nil {
							logger.L.Debug().Err(err).Msgf("Invalid cron expression for job %s", job.GUID)
							continue
						}

						nextRun := sched.Next(job.LastRunAt)
						if job.LastRunAt.IsZero() || now.After(nextRun) {
							shouldRun = true
						}
					} else if job.Interval > 0 {
						if job.LastRunAt.IsZero() || now.Sub(job.LastRunAt).Seconds() >= float64(job.Interval) {
							shouldRun = true
						}
					} else {
						logger.L.Debug().Msgf("Skipping job %s: no valid interval or cronExpr", job.GUID)
						continue
					}

					if !shouldRun {
						continue
					}

					allSets, err := zfs.Snapshots("")
					if err != nil {
						logger.L.Debug().Err(err).Msgf("Failed to get snapshots for %s", job.GUID)
						continue
					}

					name := job.Prefix + "-" + now.Format("2006-01-02-15-04")
					dataset, err := s.GetDatasetByGUID(job.GUID)
					if err != nil {
						logger.L.Debug().Err(err).Msgf("Failed to get dataset for %s", job.GUID)
						if err := s.DB.Delete(&job).Error; err != nil {
							logger.L.Debug().Err(err).Msgf("Failed to delete job %s", job.GUID)
						}

						logger.L.Debug().Msgf("Deleted job %s due to missing dataset", job.GUID)
						continue
					}

					snapshotExists := false
					for _, v := range allSets {
						if v.Name == dataset.Name+"@"+name {
							snapshotExists = true
							break
						}
					}

					if snapshotExists {
						logger.L.Debug().Msgf("Snapshot %s already exists", name)
						continue
					}

					if err := s.CreateSnapshot(job.GUID, name, job.Recursive); err != nil {
						logger.L.Debug().Err(err).Msgf("Failed to create snapshot for %s", job.GUID)
						continue
					}

					if err := s.DB.Model(&job).Update("LastRunAt", now).Error; err != nil {
						logger.L.Debug().Err(err).Msgf("Failed to update LastRunAt for %d", job.ID)
					}

					logger.L.Debug().Msgf("Snapshot %s created successfully", name)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Service) RollbackSnapshot(guid string, destroyMoreRecent bool) error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	datasets, err := zfs.Snapshots("")
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		for _, v := range properties {
			if v == guid {
				err := dataset.Rollback(destroyMoreRecent)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	return fmt.Errorf("snapshot with guid %s not found", guid)
}
