// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package jailHandlers

import (
	"sylve/internal"
	"sylve/internal/services/jail"

	"github.com/gin-gonic/gin"
)

type JailUpdateMemoryRequest struct {
	CTID   uint  `json:"ctId" binding:"required"`
	Memory int64 `json:"memory" binding:"required"`
}

type JailUpdateCPURequest struct {
	CTID  uint  `json:"ctId" binding:"required"`
	Cores int64 `json:"cores" binding:"required"`
}

// @Summary Update Jail Memory
// @Description Update the memory limit of a jail by its ID
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body jailServiceInterfaces.JailUpdateMemoryRequest true "Update Jail Memory Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Router /jail/memory [put]
func UpdateJailMemory(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req JailUpdateMemoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		err := jailService.UpdateMemory(req.CTID, req.Memory)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_update_memory",
				Data:    nil,
				Error:   "failed_to_update_memory: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "jail_memory_updated",
			Data:    nil,
			Error:   "",
		})
	}
}

// @Summary Update Jail CPU
// @Description Update the CPU limit of a jail by its ID
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body jailServiceInterfaces.JailUpdateCPURequest true "Update Jail CPU Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Router /jail/cpu [put]
func UpdateJailCPU(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req JailUpdateCPURequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		err := jailService.UpdateCPU(req.CTID, req.Cores)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_update_cpu",
				Data:    nil,
				Error:   "failed_to_update_cpu: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "jail_cpu_updated",
			Data:    nil,
			Error:   "",
		})
	}
}
