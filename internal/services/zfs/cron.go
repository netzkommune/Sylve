package zfs

import (
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
	}
}

func (s *Service) Cron() {
	tickerFast := time.NewTicker(10 * time.Second)
	tickerSlow := time.NewTicker(60 * time.Second)
	defer tickerFast.Stop()
	defer tickerSlow.Stop()

	s.StoreStats(0)

	for {
		select {
		case <-tickerFast.C:
			s.StoreStats(10)
		case <-tickerSlow.C:
			s.StoreStats(60)
		}
	}
}
