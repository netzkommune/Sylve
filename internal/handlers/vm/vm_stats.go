// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirtHandlers

import (
	"sylve/internal"
	vmModels "sylve/internal/db/models/vm"
	"sylve/internal/services/libvirt"
	"sylve/pkg/utils"

	"github.com/gin-gonic/gin"
)

// type StatsRequest struct {
// 	VMID  int `json:"vmId" binding:"required"`
// 	Limit int `json:"limit" binding:"required"`
// }

// @Summary Get VM Statistics
// @Description Retrieve statistics for a virtual machine
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]vmModels.VMStats] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm/stats/:vmId/:limit [get]
func GetVMStats(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		vmId := c.Param("vmId")
		limit := c.Param("limit")
		if vmId == "" || limit == "" {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "vmid and limit are required",
			})
			return
		}

		stats, err := libvirtService.GetVMUsage(int(utils.StringToUint64(vmId)), int(utils.StringToUint64(limit)))
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[[]vmModels.VMStats]{
			Status:  "success",
			Message: "vm_stats_retrieved",
			Data:    stats,
			Error:   "",
		})
	}
}
