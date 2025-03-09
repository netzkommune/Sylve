// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package handlers

import (
	"net/http"

	"sylve/internal/utils"

	"github.com/gin-gonic/gin"
)

func BasicHealthCheckHandler(c *gin.Context) {
	h, err := utils.GetSystemHostname()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError, gin.H{"status": "error", "message": "internal_server_error"})
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, gin.H{"hostanme": h, "message": "Basic health is OK"})
}

func HTTPHealthCheckHandler(c *gin.Context) {
	utils.SendJSONResponse(c, http.StatusOK, gin.H{"message": "HTTP health is OK"})
}
