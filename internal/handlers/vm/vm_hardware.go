// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirtHandlers

import (
	"strconv"
	"sylve/internal"
	"sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

type ModifyHardwareRequest struct {
	CPUSockets    int    `json:"cpuSockets" binding:"required"`
	CPUCores      int    `json:"cpuCores" binding:"required"`
	CPUThreads    int    `json:"cpuThreads" binding:"required"`
	RAM           int    `json:"ram" binding:"required"`
	CPUPinning    []int  `json:"cpuPinning" binding:"required"`
	VNCPort       int    `json:"vncPort" binding:"required"`
	VNCResolution string `json:"vncResolution" binding:"required"`
	VNCPassword   string `json:"vncPassword" binding:"required"`
	VNCWait       *bool  `json:"vncWait" binding:"required"`
	PCIDevices    []int  `json:"pciDevices" binding:"required"`
}

// @Summary Modify Hardware of a Virtual Machine
// @Description Modify the hardware configuration of a virtual machine
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ModifyHardwareRequest true "Modify Hardware Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /hardware/:vmid [put]
func ModifyHardware(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ModifyHardwareRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_request: " + err.Error(),
			})
			return
		}

		vmID, exists := c.Params.Get("vmid")
		if !exists {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "vmid_not_provided",
			})
			return
		}

		vmIdInt, err := strconv.Atoi(vmID)
		if err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_vmid_format",
			})
			return
		}

		wait := false
		if req.VNCWait != nil {
			wait = *req.VNCWait
		}

		if err := libvirtService.ModifyHardware(vmIdInt,
			req.CPUSockets,
			req.CPUCores,
			req.CPUThreads,
			req.CPUPinning,
			req.RAM,
			req.VNCPort,
			req.VNCResolution,
			req.VNCPassword,
			wait,
			req.PCIDevices); err != nil {
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
			Message: "hardware_modified",
			Data:    nil,
			Error:   "",
		})
	}
}
