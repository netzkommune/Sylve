// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsHandlers

import (
	"net/http"
	"sylve/internal"
	infoModels "sylve/internal/db/models/info"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/services/zfs"

	"github.com/gin-gonic/gin"
)

type AvgIODelayResponse struct {
	Delay float64 `json:"delay"`
}

// @Summary Get Average IO Delay
// @Description Get the average IO delay of all pools
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[AvgIODelayResponse] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/avg-io-delay [get]
func AvgIODelay(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := zfsSerice.GetTotalIODelay()
		c.JSON(http.StatusOK, internal.APIResponse[AvgIODelayResponse]{
			Status:  "success",
			Message: "avg_io_delay",
			Error:   "",
			Data:    AvgIODelayResponse{Delay: info},
		})
	}
}

// @Summary Get Average IO Delay Historical
// @Description Get the historical IO delays of all pools
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]infoModels.IODelay] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pool/io-delay/historical [get]
func AvgIODelayHistorical(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := zfsSerice.GetTotalIODelayHisorical()
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]infoModels.IODelay]{
			Status:  "success",
			Message: "avg_io_delay_historical",
			Error:   "",
			Data:    info,
		})
	}
}

// @Summary Get Pools
// @Description Get all ZFS pools
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]zfsServiceInterfaces.Zpool] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools [get]
func GetPools(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		pools, err := zfsSerice.GetPools()
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]zfsServiceInterfaces.Zpool]{
			Status:  "success",
			Message: "pools",
			Error:   "",
			Data:    pools,
		})
	}
}
