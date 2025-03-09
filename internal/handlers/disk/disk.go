// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package diskHandlers

import (
	"sylve/internal/services/disk"

	"github.com/gin-gonic/gin"
)

func List(diskService *disk.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		disks, err := diskService.GetDiskDevices()

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "error_listing_devices", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success", "data": disks})
	}
}

func WipeDisk(diskService *disk.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Device string `json:"device"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
			return
		}

		err := diskService.DestroyPartitionTable(req.Device)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "error_wiping_disk", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	}
}

func InitializeGPT(diskService *disk.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Device string `json:"device"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
			return
		}

		err := diskService.InitializeGPT(req.Device)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "error_initializing_gpt", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	}
}
