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

type CreatePoolRequest struct {
	Name    string            `json:"name" binding:"required,min=3,max=128"`
	Vdevs   []string          `json:"vdevs" binding:"required"`
	Raid    string            `json:"raid"`
	Options map[string]string `json:"options" binding:"required"`
}

type DeletePoolRequest struct {
	Name string `json:"name"`
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

// @Summary Create Pool
// @Description Create a new ZFS pool
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePoolRequest true "Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools [post]
func CreatePool(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CreatePoolRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		err := zfsSerice.CreatePool(request.Name, request.Vdevs, request.Raid, request.Options)
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
// @Param request body DeletePoolRequest true "Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools/{name} [delete]
func DeletePool(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request DeletePoolRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		err := zfsSerice.DestroyPool(request.Name)
		if err != nil {
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
