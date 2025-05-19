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
	"os"
	zfsModels "sylve/internal/db/models/zfs"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/logger"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"
	"time"
)

func (s *Service) GetDatasets() ([]zfsServiceInterfaces.Dataset, error) {
	var results []zfsServiceInterfaces.Dataset

	datasets, err := zfs.Datasets("")
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		props, err := dataset.GetAllProperties()
		if err != nil {
			return nil, err
		}

		propMap := make(map[string]string, len(props))
		for k, v := range props {
			propMap[k] = v
		}

		results = append(results, zfsServiceInterfaces.Dataset{
			Dataset:    *dataset,
			Properties: propMap,
		})
	}

	return results, nil
}

func (s *Service) GetDatasetByGUID(guid string) (*zfsServiceInterfaces.Dataset, error) {
	datasets, err := zfs.Datasets("")
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return nil, err
		}

		for _, v := range properties {
			if v == guid {
				propMap := make(map[string]string, len(properties))
				for k, v := range properties {
					propMap[k] = v
				}

				return &zfsServiceInterfaces.Dataset{
					Dataset:    *dataset,
					Properties: propMap,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("dataset with guid %s not found", guid)
}

func (s *Service) DeleteSnapshot(guid string, recursive bool) error {
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

func (s *Service) AddPeriodicSnapshot(guid string, prefix string, recursive bool, interval int) error {
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
					if job.LastRunAt.IsZero() || now.Sub(job.LastRunAt).Seconds() >= float64(job.Interval) {
						allSets, err := zfs.Snapshots("")
						if err != nil {
							logger.L.Debug().Err(err).Msgf("Failed to get snapshots for %s", job.GUID)
							continue
						}

						name := job.Prefix + "-" + now.Format("2006-01-02-15-04")
						dataset, err := s.GetDatasetByGUID(job.GUID)

						if err != nil {
							logger.L.Debug().Err(err).Msgf("Failed to get dataset for %s", job.GUID)
							continue
						}

						for _, v := range allSets {
							if v.Name == dataset.Name+"@"+name {
								logger.L.Debug().Msgf("Snapshot %s already exists", name)
								continue
							}
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
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Service) CreateFilesystem(name string, props map[string]string) error {
	parent := ""

	for k, v := range props {
		if k == "parent" {
			parent = v
			continue
		}
	}

	if parent == "" {
		return fmt.Errorf("parent_not_found")
	}

	name = fmt.Sprintf("%s/%s", parent, name)
	delete(props, "parent")

	_, err := zfs.CreateFilesystem(name, props)

	if err != nil {
		return err
	}

	datasets, err := zfs.Datasets(name)
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		if dataset.Name == name {
			return nil
		}
	}

	return fmt.Errorf("failed to create filesystem %s", name)
}

func (s *Service) DeleteFilesystem(guid string) error {
	datasets, err := zfs.Datasets("")
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		var keylocation string
		found := false

		for k, v := range properties {
			if v == guid {
				found = true
			}
			if k == "keylocation" {
				keylocation = v
			}
		}

		if found {
			if err := dataset.Destroy(zfs.DestroyDefault); err != nil {
				return err
			}

			if keylocation != "" {
				keylocation = keylocation[7:]
				if _, err := os.Stat(keylocation); err == nil {
					if err := os.Remove(keylocation); err != nil {
						return err
					}
				} else {
					fmt.Println("Keylocation file not found", keylocation)
				}
			}

			return nil
		}
	}

	return fmt.Errorf("filesystem with guid %s not found", guid)
}

func (s *Service) RollbackSnapshot(guid string, destroyMoreRecent bool) error {
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

func (s *Service) CreateVolume(name string, parent string, props map[string]string) error {
	datasets, err := zfs.Datasets("")
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		if dataset.Name == fmt.Sprintf("%s/%s", parent, name) && dataset.Type == "volume" {
			return fmt.Errorf("volume with name %s already exists", name)
		}
	}

	name = fmt.Sprintf("%s/%s", parent, name)

	if _, ok := props["size"]; !ok {
		return fmt.Errorf("size property not found")
	}

	pSize := utils.HumanFormatToSize(props["size"])

	_, err = zfs.CreateVolume(name, pSize, props)

	return err
}

func (s *Service) DeleteVolume(guid string) error {
	datasets, err := zfs.Datasets("")
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
				err := dataset.Destroy(zfs.DestroyDefault)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	return fmt.Errorf("volume with guid %s not found", guid)
}

func (s *Service) BulkDeleteDataset(guids []string) error {
	guidsMap := make(map[string]struct{})
	for _, guid := range guids {
		guidsMap[guid] = struct{}{}
	}

	datasets, err := zfs.Datasets("")
	if err != nil {
		return err
	}

	matched := make(map[string]*zfs.Dataset)

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		for _, v := range properties {
			if _, ok := guidsMap[v]; ok {
				matched[v] = dataset
				delete(guidsMap, v)
				break
			}
		}
	}

	if len(guidsMap) > 0 {
		return fmt.Errorf("datasets with guids %v not found", utils.MapKeys(guidsMap))
	}

	for guid, dataset := range matched {
		if err := dataset.Destroy(zfs.DestroyDefault); err != nil {
			return fmt.Errorf("failed to delete dataset with guid %s: %w", guid, err)
		}
	}

	return nil
}
