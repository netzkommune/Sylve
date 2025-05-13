// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utilitiesHandlers

import (
	"fmt"
	"net/http"
	"sylve/internal"
	utilitiesModels "sylve/internal/db/models/utilities"
	"sylve/internal/services/utilities"

	"github.com/gin-gonic/gin"
)

type DownloadFileRequest struct {
	URL string `json:"url" binding:"required"`
}

// @Summary List Downloads
// @Description List all downloads
// @Tags Utilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]utilitiesModels.Downloads] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /utilities/downloads [get]
func ListDownloads(utilitiesService *utilities.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		downloads, err := utilitiesService.ListDownloads()
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_list_downloads",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		// c.JSON(200, downloads)
		c.JSON(http.StatusOK, internal.APIResponse[[]utilitiesModels.Downloads]{
			Status:  "success",
			Message: "downloads_listed",
			Error:   "",
			Data:    downloads,
		})
	}
}

// @Summary Download File
// @Description Download a file from a Magnet or HTTP(s) URL
// @Tags Utilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url query string true "URL"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /utilities/downloads [post]
func DownloadFile(utilitiesService *utilities.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request DownloadFileRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		fmt.Println("Request URL:", request.URL)

		if err := utilitiesService.DownloadFile(request.URL); err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_download_file",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "file_download_started",
			Error:   "",
			Data:    nil,
		})
	}
}
