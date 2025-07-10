// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfs

import (
	"encoding/json"
	"sylve/internal/db"
	infoModels "sylve/internal/db/models/info"
	"sylve/internal/logger"
	"sylve/pkg/zfs"
	"time"
)

func (s *Service) StoreStats(interval int) {
	if interval == 10 || interval == 0 {
		d := zfs.GetTotalIODelay()
		db.StoreAndTrimRecords(s.DB, &infoModels.IODelay{Delay: d}, 128)
	}

	if interval == 60 || interval == 0 {
		pools, err := zfs.ListZpools()
		if err != nil {
			logger.L.Debug().Err(err).Msg("zfs_cron: Failed to list zpools")
			return
		}

		for _, pool := range pools {
			newStat := infoModels.ZPoolHistorical{
				Pools: infoModels.ZpoolJSON(*pool),
			}
			if err := s.DB.Create(&newStat).Error; err != nil {
				logger.L.Debug().Err(err).Msg("zfs_cron: Failed to insert zpool data")
			}
		}

		if time.Now().Minute()%10 == 0 {
			s.trimZPoolHistoricalData()
		}
	}
}

func (s *Service) RemoveNonExistentPools() {
	pools, err := zfs.ListZpools()
	if err != nil {
		logger.L.Error().Err(err).Msg("zfs_cron: Failed to list zpools")
		return
	}

	existingPools := make(map[string]struct{}, len(pools))
	for _, pool := range pools {
		existingPools[pool.Name] = struct{}{}
	}

	rows, err := s.DB.Model(&infoModels.ZPoolHistorical{}).Select("id, pools").Rows()
	if err != nil {
		logger.L.Error().Err(err).Msg("zfs_cron: Failed to stream zpool records from DB")
		return
	}
	defer rows.Close()

	var idsToDelete []int64
	for rows.Next() {
		var id int64
		var raw json.RawMessage

		if err := rows.Scan(&id, &raw); err != nil {
			logger.L.Warn().Err(err).Msg("zfs_cron: Failed to scan row")
			continue
		}

		var pool zfs.Zpool
		if err := json.Unmarshal(raw, &pool); err != nil {
			logger.L.Warn().Err(err).Int64("id", id).Msg("zfs_cron: Failed to unmarshal pool")
			continue
		}

		if _, exists := existingPools[pool.Name]; !exists {
			idsToDelete = append(idsToDelete, id)
		}
	}

	if len(idsToDelete) > 0 {
		if err := s.DB.Where("id IN ?", idsToDelete).Delete(&infoModels.ZPoolHistorical{}).Error; err != nil {
			logger.L.Error().Err(err).Msg("zfs_cron: Failed to delete old pool entries")
		} else {
			logger.L.Info().Int("count", len(idsToDelete)).Msg("zfs_cron: Deleted non-existent pool entries")
		}
	} else {
		logger.L.Debug().Msg("zfs_cron: No non-existent pools to delete")
	}
}

func (s *Service) trimZPoolHistoricalData() {
	now := time.Now()
	cutoff24h := now.Add(-24 * time.Hour).UnixMilli()

	var oldRecords []infoModels.ZPoolHistorical
	err := s.DB.Where("created_at < ?", cutoff24h).
		Order("created_at ASC").
		Find(&oldRecords).Error
	if err != nil {
		logger.L.Debug().Err(err).Msg("zfs_cron: Failed to fetch old zpool records")
		return
	}

	if len(oldRecords) == 0 {
		return
	}

	hourlyRecords := make(map[int64]infoModels.ZPoolHistorical)
	recordsToDelete := make([]int64, 0)

	for _, record := range oldRecords {
		recordTime := time.UnixMilli(record.CreatedAt)
		hourKey := recordTime.Truncate(time.Hour).UnixMilli()

		if _, exists := hourlyRecords[hourKey]; !exists {
			hourlyRecords[hourKey] = record
		} else {
			recordsToDelete = append(recordsToDelete, record.ID)
		}
	}

	if len(recordsToDelete) > 0 {
		err = s.DB.Where("id IN ?", recordsToDelete).Delete(&infoModels.ZPoolHistorical{}).Error
		if err != nil {
			logger.L.Debug().Err(err).Msg("ZFS Cron: Failed to delete old zpool records")
		} else {
			logger.L.Debug().
				Int("deleted_count", len(recordsToDelete)).
				Msg("ZFS Cron: Trimmed old zpool historical data")
		}
	}

	maxAge := now.Add(-365 * 24 * time.Hour).UnixMilli()
	result := s.DB.Where("created_at < ?", maxAge).Delete(&infoModels.ZPoolHistorical{})
	deletedCount := result.RowsAffected
	if result.Error == nil && deletedCount > 0 {
		logger.L.Debug().
			Int64("deleted_count", deletedCount).
			Msg("ZFS Cron: Deleted very old zpool records (>1 year)")
	}

	if err := s.DB.Exec("VACUUM").Error; err != nil {
		logger.L.Warn().Msgf("ZFS Cron: VACUUM failed: %v", err)
	}
}

func (s *Service) Cron() {
	tickerFast := time.NewTicker(10 * time.Second)
	tickerSlow := time.NewTicker(60 * time.Second)
	defer tickerFast.Stop()
	defer tickerSlow.Stop()

	s.StoreStats(0)
	s.RemoveNonExistentPools()

	for {
		select {
		case <-tickerFast.C:
			s.StoreStats(10)
		case <-tickerSlow.C:
			s.StoreStats(60)
			s.RemoveNonExistentPools()
		}
	}
}
