// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsHandlers

import (
	"sylve/internal/services/zfs"

	"github.com/gin-gonic/gin"
)

func AvgIODelay(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			info interface{}
			err  error
		)

		if c.Query("historical") == "1" {
			info, err = zfsSerice.GetTotalIODelayHisorical()
			if err != nil {
				info = []gin.H{}
			}
		} else {
			info = zfsSerice.GetTotalIODelay()
		}

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success", "data": info})
	}
}

func GetPools(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		pools, err := zfsSerice.GetPools()
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": pools})
	}
}
