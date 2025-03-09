// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoHandlers

import (
	"sylve/internal/services/info"

	"github.com/gin-gonic/gin"
)

func CPUInfo(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			info interface{}
			err  error
		)

		if c.Query("historical") == "1" {
			info, err = infoService.GetCPUUsageHistorical()
		} else {
			info, err = infoService.GetCPUInfo(false)
		}

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success", "data": info})
	}
}
