// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package info

import (
	infoServiceInterfaces "sylve/internal/interfaces/services/info"

	ram "github.com/shirou/gopsutil/mem"
)

func (s *Service) GetRAMInfo() (infoServiceInterfaces.RAMInfo, error) {
	ramInfo, err := ram.VirtualMemory()
	if err != nil {
		return infoServiceInterfaces.RAMInfo{}, err
	}

	return infoServiceInterfaces.RAMInfo{
		Total:       ramInfo.Total,
		Free:        ramInfo.Free,
		UsedPercent: ramInfo.UsedPercent,
	}, nil
}

func (s *Service) GetSwapInfo() (infoServiceInterfaces.SwapInfo, error) {
	swapInfo, err := ram.SwapMemory()

	if err != nil {
		return infoServiceInterfaces.SwapInfo{}, err
	}

	return infoServiceInterfaces.SwapInfo{
		Total:       swapInfo.Total,
		Free:        swapInfo.Free,
		UsedPercent: swapInfo.UsedPercent,
	}, nil
}
