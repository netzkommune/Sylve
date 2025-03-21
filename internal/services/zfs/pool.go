// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfs

import (
	"strconv"
	"strings"
	"sylve/internal/db"
	infoModels "sylve/internal/db/models/info"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/logger"
	"sylve/pkg/utils"
)

func (s *Service) GetPoolNames() ([]string, error) {
	output, err := utils.RunCommand("zpool", "list", "-H", "-o", "name")
	if err != nil {
		return nil, err
	}

	poolNames := strings.Fields(output)

	return poolNames, nil
}

func (s *Service) GetPool(name string) (zfsServiceInterfaces.Zpool, error) {
	pools, err := utils.RunCommand("zpool", "list", "-H", "-p", "-o", "name,health,alloc,size,free,readonly,freeing,leaked,dedupratio", name)
	if err != nil {
		return zfsServiceInterfaces.Zpool{}, err
	}

	vdevs, err := utils.RunCommand("zpool", "iostat", "-v", "-H", "-P", "-p", name)
	if err != nil {
		return zfsServiceInterfaces.Zpool{}, err
	}

	zpool, err := utils.ParseZpoolListOutput(pools, vdevs)

	return *zpool, err
}

func (s *Service) GetPools() ([]zfsServiceInterfaces.Zpool, error) {
	names, err := s.GetPoolNames()
	if err != nil {
		return []zfsServiceInterfaces.Zpool{}, err
	}

	var pools []zfsServiceInterfaces.Zpool

	for _, name := range names {
		pool, err := s.GetPool(name)
		if err != nil {
			return []zfsServiceInterfaces.Zpool{}, err
		}
		pools = append(pools, pool)
	}

	return pools, nil
}

func (s *Service) GetPoolIODelay(poolName string) float64 {
	names, err := s.GetPoolNames()

	if err != nil {
		logger.L.Debug().Msgf("Error getting pool names: %v", err)
		return 0.0
	}

	if !utils.StringInSlice(poolName, names) {
		logger.L.Debug().Msgf("Pool %s not found", poolName)
		return 0.0
	}

	output, err := utils.RunCommand("zpool", "iostat", "-l", "-H", "-v", poolName, "1", "2")
	if err != nil {
		return 0.0
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")

	var samples [][]string
	var currentSample []string
	seenPools := make(map[string]bool)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(currentSample) > 0 {
				samples = append(samples, currentSample)
				currentSample = nil
				seenPools = make(map[string]bool)
			}
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pool := fields[0]
		if seenPools[pool] {
			samples = append(samples, currentSample)
			currentSample = nil
			seenPools = make(map[string]bool)
		}
		seenPools[pool] = true
		currentSample = append(currentSample, line)
	}

	if len(currentSample) > 0 {
		samples = append(samples, currentSample)
	}

	if len(samples) < 2 {
		return 0.0
	}

	secondSample := samples[1]
	sampleInterval := int64(1000000)

	for _, line := range secondSample {
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 9 || fields[0] != poolName {
			continue
		}

		readOps, err1 := strconv.ParseInt(fields[3], 10, 64)
		writeOps, err2 := strconv.ParseInt(fields[4], 10, 64)
		if err1 != nil || err2 != nil || (readOps+writeOps) == 0 {
			return 0.0
		}

		totalReadWait := utils.ParseZfsTimeUnit(fields[7])
		totalWriteWait := utils.ParseZfsTimeUnit(fields[8])
		totalWaitAccumulated := (readOps * totalReadWait) + (writeOps * totalWriteWait)
		averageWait := totalWaitAccumulated / (readOps + writeOps)

		return (float64(averageWait) / float64(sampleInterval)) * 100
	}

	return 0.0
}

func (s *Service) GetTotalIODelay() float64 {
	names, err := s.GetPoolNames()
	if err != nil {
		logger.L.Debug().Msgf("Error getting pool names: %v", err)
		return 0.0
	}

	var totalDelay float64
	count := 0

	for _, name := range names {
		delay := s.GetPoolIODelay(name)
		if delay > 0 {
			totalDelay += delay
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalDelay / float64(count)
}

func (s *Service) GetTotalIODelayHisorical() ([]infoModels.IODelay, error) {
	historicalData, err := db.GetHistorical[infoModels.IODelay](s.DB, 128)

	if err != nil {
		return nil, err
	}

	return historicalData, nil
}
