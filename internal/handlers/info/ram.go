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

func RAMInfo(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := infoService.GetRAMInfo()

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
		}

		c.JSON(200, gin.H{"status": "success", "data": info})
	}
}

func SwapInfo(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := infoService.GetSwapInfo()

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
		}

		c.JSON(200, gin.H{"status": "success", "data": info})
	}
}
