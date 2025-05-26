// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsHandlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sylve/internal"
	"sylve/internal/db"
	infoModels "sylve/internal/db/models/info"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/services/info"
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

type PoolStatPointResponse struct {
	PoolStatPoint map[string][]zfsServiceInterfaces.PoolStatPoint `json:"poolStatPoint"`
	IntervalMap   []db.IntervalOption                             `json:"intervalMap"`
}

type PoolEditRequest struct {
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties"`
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
func AvgIODelay(zfsService *zfs.Service) gin.HandlerFunc {
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
func AvgIODelayHistorical(zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := zfsService.GetTotalIODelayHisorical()
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
func GetPools(zfsService *zfs.Service) gin.HandlerFunc {
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
func CreatePool(infoService *info.Service, zfsService *zfs.Service) gin.HandlerFunc {
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

		id := infoService.StartAuditLog(c.GetString("Token"), fmt.Sprintf("zfs.pool.create_pool|-|%s", request.Name), "started")
		err := zfsService.CreatePool(request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "pool_create_failed",
				Error:   err.Error(),
				Data:    nil,
			})

			infoService.EndAuditLog(id, "failed")
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "pool_created",
			Error:   "",
			Data:    nil,
		})

		infoService.EndAuditLog(id, "success")
	}
}

// @Summary Scrub Pool
// @Description Start a scrub on a ZFS pool
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Pool Name"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools/{name}/scrub [post]
func ScrubPool(infoService *info.Service, zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		id := infoService.StartAuditLog(c.GetString("Token"), fmt.Sprintf("zfs.pool.scrub_pool|-|%s", name), "started")
		err := zfsUtils.ScrubPool(name)
		if err != nil {
			if strings.HasPrefix(err.Error(), "error_getting_pool") {
				c.JSON(http.StatusNotFound, internal.APIResponse[any]{
					Status:  "error",
					Message: "pool_not_found",
					Error:   err.Error(),
					Data:    nil,
				})

				infoService.EndAuditLog(id, "failed")
				return
			}

			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "pool_scrub_failed",
				Error:   err.Error(),
				Data:    nil,
			})

			infoService.EndAuditLog(id, "failed")
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "pool_scrub_started",
			Error:   "",
			Data:    nil,
		})

		infoService.EndAuditLog(id, "success")
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
func DeletePool(infoService *info.Service, zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		id := infoService.StartAuditLog(c.GetString("Token"), fmt.Sprintf("zfs.pool.delete_pool|-|%s", name), "started")
		err := zfsService.DeletePool(name)
		if err != nil {
			if strings.HasPrefix(err.Error(), "error_getting_pool") {
				c.JSON(http.StatusNotFound, internal.APIResponse[any]{
					Status:  "error",
					Message: "pool_not_found",
					Error:   err.Error(),
					Data:    nil,
				})

				infoService.EndAuditLog(id, "failed")
				return
			}

			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "pool_delete_failed",
				Error:   err.Error(),
				Data:    nil,
			})

			infoService.EndAuditLog(id, "failed")
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "pool_deleted",
			Error:   "",
			Data:    nil,
		})
		infoService.EndAuditLog(id, "success")
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
func ReplaceDevice(infoService *info.Service, zfsService *zfs.Service) gin.HandlerFunc {
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

		id := infoService.StartAuditLog(c.GetString("Token"), fmt.Sprintf("zfs.pool.replace_device|-|%s", fmt.Sprintf("%s in %s", request.Old, name)), "started")
		err := zfsUtils.ReplaceInPool(name, request.Old, request.New)
		if err != nil {
			if strings.HasPrefix(err.Error(), "error_getting_pool") {
				c.JSON(http.StatusNotFound, internal.APIResponse[any]{
					Status:  "error",
					Message: "pool_not_found",
					Error:   err.Error(),
					Data:    nil,
				})

				infoService.EndAuditLog(id, "failed")
				return
			}

			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "device_replace_failed",
				Error:   err.Error(),
				Data:    nil,
			})

			infoService.EndAuditLog(id, "failed")
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "device_replaced",
			Error:   "",
			Data:    nil,
		})

		infoService.EndAuditLog(id, "success")
	}
}

// @Summary Get Pool Stats
// @Description Get the historical stats of a ZFS pool
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param interval path int true "Interval in minutes"
// @Param limit path int true "Limit"
// @Success 200 {object} internal.APIResponse[PoolStatPointResponse] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pool/stats/{interval}/{limit} [get]
func PoolStats(zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		interval := c.Param("interval")
		limit := c.Param("limit")

		intervalInt, err := strconv.Atoi(interval)
		if err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_interval",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_limit",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		stats, count, err := zfsService.GetZpoolHistoricalStats(intervalInt, limitInt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		response := PoolStatPointResponse{
			PoolStatPoint: stats,
			IntervalMap:   db.IntervalToMap(count),
		}

		c.JSON(http.StatusOK, internal.APIResponse[PoolStatPointResponse]{
			Status:  "success",
			Message: "pool_stats",
			Error:   "",
			Data:    response,
		})
	}
}

// @Summary Edit Pool
// @Description Edit a ZFS pool's properties
// @Tags ZFS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PoolEditRequest true "Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /zfs/pools [patch]
func EditPool(infoService *info.Service, zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request PoolEditRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		id := infoService.StartAuditLog(c.GetString("Token"), fmt.Sprintf("zfs.pool.edit_pool|-|%s", request.Name), "started")
		err := zfsService.EditZpool(request.Name, request.Properties)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "pool_edit_failed",
				Error:   err.Error(),
				Data:    nil,
			})

			infoService.EndAuditLog(id, "failed")
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "pool_edited",
			Error:   "",
			Data:    nil,
		})

		infoService.EndAuditLog(id, "success")
	}
}
