// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoHandlers

import (
	"net/http"
	"sylve/internal"
	infoServiceInterfaces "sylve/internal/interfaces/services/info"
	"sylve/internal/services/info"

	"github.com/gin-gonic/gin"
)

type HistoricalNetworkInterfaceResponse struct {
	Status  string                                             `json:"status"`
	Message string                                             `json:"message"`
	Error   string                                             `json:"error"`
	Data    []infoServiceInterfaces.HistoricalNetworkInterface `json:"data"`
}

// @Summary Get Historical Network information
// @Description Retrieves historical Network info
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} HistoricalNetworkInterfaceResponse
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /info/network-interfaces/historical [get]
func HistoricalNetworkInterfacesInfoHandler(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := infoService.GetNetworkInterfacesHistorical()
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]infoServiceInterfaces.HistoricalNetworkInterface]{
			Status:  "success",
			Message: "network_interfaces_info",
			Error:   "",
			Data:    info,
		})
	}
}
