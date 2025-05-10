// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package networkHandlers

import (
	"net/http"
	"sylve/internal"
	"sylve/internal/services/network"
	iface "sylve/pkg/network/iface"

	"github.com/gin-gonic/gin"
)

// @Summary List Network Interfaces
// @Description List all network interfaces on the system
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]*iface.Interface] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/interface [get]
func ListInterfaces(networkService *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		interfaces, err := iface.List()

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})

			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]*iface.Interface]{
			Status:  "success",
			Message: "interfaces_list",
			Error:   "",
			Data:    interfaces,
		})
	}
}
