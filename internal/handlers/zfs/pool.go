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
	"strings"
	"sylve/internal"
	infoModels "sylve/internal/db/models/info"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/services/zfs"

	"github.com/gin-gonic/gin"

	zfsUtils "sylve/pkg/zfs"
)

type AvgIODelayResponse struct {
	Delay float64 `json:"delay"`
}

type ZpoolListResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Error   string            `json:"error"`
	Data    []*zfsUtils.Zpool `json:"data"`
}

// @Summary Get Average IO Delay
// @Description Get the average IO delay of all pools
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[AvgIODelayResponse] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pool/avg-io-delay [get]
func AvgIODelay(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := zfsUtils.GetTotalIODelay()
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
// @Success 200 {object} zfsHandlers.ZpoolListResponse "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools [get]
func GetPools(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		pools, err := zfsUtils.ListZpools()
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]*zfsUtils.Zpool]{
			Status:  "success",
			Message: "pools",
			Error:   "",
			Data:    pools,
		})
	}
}

// @Summary Create Pool
// @Description Create a new ZFS pool
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body zfsServiceInterfaces.Zpool true "Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools [post]
func CreatePool(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request zfsServiceInterfaces.Zpool
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		err := zfsSerice.CreatePool(request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "pool_create_failed",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "pool_created",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Delete Pool
// @Description Delete a ZFS pool
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Pool Name"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools/{name} [delete]
func DeletePool(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		err := zfsUtils.DestroyPool(name)
		if err != nil {
			if strings.HasPrefix(err.Error(), "error_getting_pool") {
				c.JSON(http.StatusNotFound, internal.APIResponse[any]{
					Status:  "error",
					Message: "pool_not_found",
					Error:   err.Error(),
					Data:    nil,
				})
				return
			}

			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "pool_delete_failed",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "pool_deleted",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Replace Device
// @Description Replace a device in a ZFS pool
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body zfsServiceInterfaces.ReplaceDevice true "Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools/{name}/replace-device [post]
func ReplaceDevice(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		var request zfsServiceInterfaces.ReplaceDevice

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		err := zfsUtils.ReplaceInPool(name, request.Old, request.New)
		if err != nil {
			if strings.HasPrefix(err.Error(), "error_getting_pool") {
				c.JSON(http.StatusNotFound, internal.APIResponse[any]{
					Status:  "error",
					Message: "pool_not_found",
					Error:   err.Error(),
					Data:    nil,
				})
				return
			}

			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "device_replace_failed",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "device_replaced",
			Error:   "",
			Data:    nil,
		})
	}
}
