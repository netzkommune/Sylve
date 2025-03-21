// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"strings"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
)

func ParseZpoolListOutput(pools string, vdevs string) (*zfsServiceInterfaces.Zpool, error) {
	poolSlice := strings.Split(strings.TrimSpace(pools), "\n")
	vdevSlice := strings.Split(strings.TrimSpace(vdevs), "\n")

	zpool := &zfsServiceInterfaces.Zpool{}

	for _, pool := range poolSlice {
		if pool == "" {
			continue
		}

		parts := strings.Fields(pool)
		if len(parts) < 9 {
			continue
		}

		zpool.Name = parts[0]
		zpool.Health = parts[1]
		zpool.Allocated = StringToUint64(parts[2])
		zpool.Size = StringToUint64(parts[3])
		zpool.Free = StringToUint64(parts[4])
		zpool.ReadOnly = parts[5] == "on"
		zpool.Freeing = StringToUint64(parts[6])
		zpool.Leaked = StringToUint64(parts[7])
		zpool.DedupRatio = StringToFloat64(parts[8])

		for _, vdev := range vdevSlice {
			if strings.HasPrefix(vdev, zpool.Name) {
				continue
			}

			vdevParts := strings.Fields(vdev)

			if len(vdevParts) < 7 {
				continue
			}

			vdev := zfsServiceInterfaces.Vdev{}

			vdev.Name = vdevParts[0]
			vdev.Alloc = StringToUint64(vdevParts[1])
			vdev.Free = StringToUint64(vdevParts[2])
			vdev.Operations.Read = StringToUint64(vdevParts[3])
			vdev.Operations.Write = StringToUint64(vdevParts[4])
			vdev.Bandwidth.Read = StringToUint64(vdevParts[5])
			vdev.Bandwidth.Write = StringToUint64(vdevParts[6])

			zpool.Vdevs = append(zpool.Vdevs, vdev)
		}
	}

	return zpool, nil
}
