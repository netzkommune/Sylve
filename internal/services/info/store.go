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
	"sync"
	"time"
)

func (s *Service) StoreStats() {
	type task struct {
		get func() (float64, error)
		ptr func(float64) interface{}
	}

	jobs := []task{
		{
			get: func() (float64, error) { c, err := s.GetCPUInfo(true); return c.Usage, err },
			ptr: func(v float64) interface{} { return &infoModels.CPU{Usage: v} },
		},
		{
			get: func() (float64, error) { r, err := s.GetRAMInfo(); return r.UsedPercent, err },
			ptr: func(v float64) interface{} { return &infoModels.RAM{Usage: v} },
		},
		{
			get: func() (float64, error) { sw, err := s.GetSwapInfo(); return sw.UsedPercent, err },
			ptr: func(v float64) interface{} { return &infoModels.Swap{Usage: v} },
		},
	}

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(j task) {
			defer wg.Done()
			if v, err := j.get(); err == nil {
				db.StoreAndTrimRecords(s.DB, j.ptr(v), 128)
			}
		}(job)
	}
	wg.Wait()
}

func (s *Service) Cron() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	s.StoreStats()

	for range ticker.C {
		s.StoreStats()
	}
}
