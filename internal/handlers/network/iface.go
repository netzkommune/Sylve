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
	networkModels "sylve/internal/db/models/network"
	"sylve/internal/services/network"

	"github.com/gin-gonic/gin"
)

type IPv4ConfigRequest struct {
	Name     string     `json:"name" binding:"required,min=1,max=128"`
	Metric   int        `json:"metric" binding:"required"`
	MTU      int        `json:"mtu" binding:"required"`
	Protocol string     `json:"protocol" binding:"required,oneof=dhcp static"`
	Address  string     `json:"address" binding:"omitempty,ipv4"`
	Netmask  string     `json:"netmask" binding:"omitempty,ipv4"`
	Aliases  [][]string `json:"aliases" binding:"omitempty,aliases"`
}

// @Summary List Network Interfaces
// @Description List all network interfaces on the system
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]networkModels.NetworkInterface] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /info/basic [get]
func ListInterfaces(networkService *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		interfaces, err := networkService.List()

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})

			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]networkModels.NetworkInterface]{
			Status:  "success",
			Message: "interfaces_list",
			Error:   "",
			Data:    interfaces,
		})
	}
}

// @Summary IPv4 Configuration on Interface
// @Description Configure IPv4 on a network interface
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body IPv4ConfigRequest true "IPv4 Configuration"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/interface/ipv4 [post]
func IPv4Config(networkService *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IPv4ConfigRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_payload",
				Error:   "validation_error",
				Data:    nil,
			})
			return
		}

		err := networkService.SetupIPv4(
			req.Name,
			req.Metric,
			req.MTU,
			req.Protocol,
			req.Address,
			req.Netmask,
			req.Aliases,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_configure_ipv4",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "ipv4_configured",
			Error:   "",
			Data:    nil,
		})
	}
}
