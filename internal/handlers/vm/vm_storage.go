// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirtHandlers

import (
	"github.com/alchemillahq/sylve/internal"
	"github.com/alchemillahq/sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

type StorageDetachRequest struct {
	VMID      int `json:"vmId" binding:"required"`
	StorageId int `json:"storageId" binding:"required"`
}
type StorageAttachRequest struct {
	VMID        int    `json:"vmId" binding:"required"`
	StorageType string `json:"storageType" binding:"required"`
	Dataset     string `json:"dataset" binding:"required"`
	Emulation   string `json:"emulation" binding:"required"`
	Size        *int64 `json:"size" binding:"required"`
	Name        string `json:"name"`
}

// @Summary Detach Storage from a Virtual Machine
// @Description Detach a storage volume from a virtual machine
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /storage/detach [post]
func StorageDetach(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StorageDetachRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_request: " + err.Error(),
			})
			return
		}

		if err := libvirtService.StorageDetach(req.VMID, req.StorageId); err != nil {
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

// @Summary Attach Storage to a Virtual Machine
// @Description Attach a storage volume to a virtual machine
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /storage/attach [post]
func StorageAttach(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StorageAttachRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_request: " + err.Error(),
			})
			return
		}

		var size int64

		if req.Size == nil {
			size = 0
		} else {
			size = *req.Size
		}

		var name string

		if req.Name == "" {
			name = ""
		} else {
			name = req.Name
		}

		if err := libvirtService.StorageAttach(req.VMID, req.StorageType, req.Dataset, req.Emulation, size, name); err != nil {
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
			Message: "storage_attached",
			Data:    nil,
			Error:   "",
		})
	}
}
