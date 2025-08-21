// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package info

import (
	"encoding/json"
	"fmt"
	"time"

	infoModels "github.com/alchemillahq/sylve/internal/db/models/info"
	infoServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/info"
	"github.com/alchemillahq/sylve/pkg/utils"
)

func (s *Service) GetNetworkInterfacesInfo() ([]infoServiceInterfaces.NetworkInterface, error) {
	var tOutput struct {
		Statistics struct {
			Interfaces []infoServiceInterfaces.NetworkInterface `json:"interface"`
		}
	}

	output, err := utils.RunCommand("netstat", "-ibdn", "--libxo", "json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(output), &tOutput)
	if err != nil {
		return nil, err
	}

	if len(tOutput.Statistics.Interfaces) > 0 {
		return tOutput.Statistics.Interfaces, nil
	}

	return nil, nil
}

func (s *Service) GetNetworkInterfacesHistorical() ([]infoServiceInterfaces.HistoricalNetworkInterface, error) {
	type _niRow struct {
		CreatedAtStr  string `gorm:"column:created_at"`
		ReceivedBytes int64  `gorm:"column:received_bytes"`
		SentBytes     int64  `gorm:"column:sent_bytes"`
	}

	var rows []_niRow
	if err := s.DB.
		Model(&infoModels.NetworkInterface{}).
		Select(
			"strftime('%Y-%m-%d %H:%M:%S', created_at) AS created_at, " +
				"SUM(received_bytes)                         AS received_bytes, " +
				"SUM(sent_bytes)                             AS sent_bytes",
		).
		Group("strftime('%Y-%m-%d %H:%M:%S', created_at)").
		Order("strftime('%Y-%m-%d %H:%M:%S', created_at) ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	var prev *_niRow
	deltas := make([]infoServiceInterfaces.HistoricalNetworkInterface, 0, len(rows)-1)
	for _, cur := range rows {
		if prev != nil {
			ts, err := time.Parse("2006-01-02 15:04:05", cur.CreatedAtStr)
			if err != nil {
				return nil, fmt.Errorf("parsing timestamp %q: %w", cur.CreatedAtStr, err)
			}

			deltas = append(deltas, infoServiceInterfaces.HistoricalNetworkInterface{
				CreatedAt:     ts,
				ReceivedBytes: cur.ReceivedBytes - prev.ReceivedBytes,
				SentBytes:     cur.SentBytes - prev.SentBytes,
			})
		}
		prev = &cur
	}

	if len(deltas) == 0 {
		return nil, nil
	}
	return deltas, nil
}
