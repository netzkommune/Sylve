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

func AuditLogs(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		logs, err := infoService.GetAuditLogs(64)

		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "internal_server_error", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success", "data": logs})
	}
}
