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
	"sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

type NetworkDetachRequest struct {
	VMID      int `json:"vmId" binding:"required"`
	NetworkId int `json:"networkId" binding:"required"`
}

// @Summary Detach Network from a Virtual Machine
// @Description Detach a network interface from a virtual machine
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/detach [post]
func NetworkDetach(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req NetworkDetachRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_request: " + err.Error(),
			})
			return
		}

		if err := libvirtService.NetworkDetach(req.VMID, req.NetworkId); err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "storage_detached",
			Data:    nil,
			Error:   "",
		})
	}
}
