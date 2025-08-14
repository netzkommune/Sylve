// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package sambaHandlers

import (
	"net/http"
	"strconv"
	"sylve/internal"
	sambaServiceInterfaces "sylve/internal/interfaces/services/samba"
	"sylve/internal/services/samba"

	"github.com/gin-gonic/gin"
)

// @Summary Get Samba Audit Logs
// @Description Retrieve Samba audit logs
// @Tags Samba
// @Accept json
// @Produce json
// @Param hash query string true "Auth hash"
// @Param page query int false "Page number (default 1)"
// @Param size query int false "Page size  (default 100)"
// @Param sort[0][field] query string false "Field to sort by (e.g. id, action, share, created_at)"
// @Param sort[0][dir]   query string false "Sort direction (asc or desc)"
// @Success 200 {object} internal.APIResponse[sambaServiceInterfaces.AuditLogsResponse] "Samba audit logs"
// @Failure 500 {string} string "Internal server error"
// @Router /samba/audit-logs [get]
func GetAuditLogs(smbService *samba.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse page & size with sensible defaults
		pageStr := c.DefaultQuery("page", "1")
		sizeStr := c.DefaultQuery("size", "100")
		page, _ := strconv.Atoi(pageStr)
		size, _ := strconv.Atoi(sizeStr)

		sortField := c.Query("sort[0][field]")
		sortDir := c.Query("sort[0][dir]")
		logs, err := smbService.GetAuditLogs(page, size, sortField, sortDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_get_samba_audit_logs",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[*sambaServiceInterfaces.AuditLogsResponse]{
			Status:  "success",
			Message: "samba_audit_logs_retrieved",
			Error:   "",
			Data:    logs,
		})
	}
}
