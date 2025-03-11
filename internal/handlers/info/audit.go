// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoHandlers

import (
	"net/http"
	"sylve/internal"
	"sylve/internal/services/info"

	_ "sylve/internal/db/models/info"

	"github.com/gin-gonic/gin"
)

// @Summary Get Audit Logs
// @Description Get the latest audit logs
// @Tags Info
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]infoModels.AuditLog] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /info/audit-logs [get]
func AuditLogs(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		logs, err := infoService.GetAuditLogs(64)

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "audit_logs",
			Error:   "",
			Data:    logs,
		})
	}
}
