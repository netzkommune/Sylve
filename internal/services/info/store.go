// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package info

import (
	"sylve/internal/db"
	infoModels "sylve/internal/db/models/info"
	"time"
)

func (s *Service) StoreStats() {
	c, err := s.GetCPUInfo(true)
	if err == nil {
		db.StoreAndTrimRecords(s.DB, &infoModels.CPU{Usage: c.Usage}, 128)
	}
}

func (s *Service) Cron() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	s.StoreStats()

	for range ticker.C {
		s.StoreStats()
	}
}
