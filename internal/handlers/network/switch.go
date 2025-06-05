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
	"strconv"
	"sylve/internal"
	networkModels "sylve/internal/db/models/network"
	"sylve/internal/services/network"

	"github.com/gin-gonic/gin"
)

type ListSwitchResponse struct {
	Standard []networkModels.StandardSwitch `json:"standard"`
}

type CreateStandardSwitchRequest struct {
	Name     string   `json:"name" binding:"required"`
	MTU      *int     `json:"mtu"`
	VLAN     *int     `json:"vlan"`
	Address  string   `json:"address"`
	Address6 string   `json:"address6"`
	Private  *bool    `json:"private" binding:"required"`
	DHCP     *bool    `json:"dhcp"`
	Ports    []string `json:"ports" binding:"required"`
}

type UpdateStandardSwitchRequest struct {
	ID       int      `json:"id" binding:"required"`
	MTU      *int     `json:"mtu"`
	VLAN     *int     `json:"vlan"`
	Address  string   `json:"address"`
	Address6 string   `json:"address6"`
	Private  *bool    `json:"private" binding:"required"`
	Ports    []string `json:"ports" binding:"required"`
}

// @Summary List Network Switches
// @Description List all network switches on the system
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[ListSwitchResponse] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/switch [get]
func ListSwitches(networkService *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var response ListSwitchResponse
		switches, err := networkService.GetStandardSwitches()

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_get_switches",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		response.Standard = switches
		c.JSON(http.StatusOK, internal.APIResponse[ListSwitchResponse]{
			Status:  "success",
			Message: "switches_list",
			Error:   "",
			Data:    response,
		})
	}
}

// @Summary Create a new Standard Switch
// @Description Create a new standard switch
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateStandardSwitchRequest true "Create Standard Switch Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/switch [post]
func CreateStandardSwitch(networkService *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CreateStandardSwitchRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		mtu := 0
		vlan := 0

		if request.VLAN != nil {
			if *request.VLAN < 0 {
				c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
					Status:  "error",
					Message: "invalid_vlan",
					Error:   "vlan_must_be_positive_or_zero",
					Data:    nil,
				})
				return
			}
			vlan = *request.VLAN
		}

		if request.MTU != nil {
			if *request.MTU < 0 {
				c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
					Status:  "error",
					Message: "invalid_mtu",
					Error:   "mtu_must_be_positive_or_zero",
					Data:    nil,
				})
				return
			}
			mtu = *request.MTU
		}

		if request.Private == nil {
			request.Private = new(bool)
			*request.Private = false
		}

		if request.DHCP == nil {
			request.DHCP = new(bool)
			*request.DHCP = false
		}

		if *request.DHCP == true {
			request.Address = ""
		}

		err := networkService.NewStandardSwitch(request.Name, mtu, vlan, request.Address, request.Address6, request.Ports, *request.Private, *request.DHCP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_create_switch",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "switch_created",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Delete a Standard Switch
// @Description Delete a standard switch by ID
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Switch ID"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/switch/{id} [delete]
func DeleteStandardSwitch(networkService *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_switch_id",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		err = networkService.DeleteStandardSwitch(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_delete_switch",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "switch_deleted",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Update a Standard Switch
// @Description Update a standard switch by ID
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Switch ID"
// @Param request body UpdateStandardSwitchRequest true "Update Standard Switch Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/switch [put]
func UpdateStandardSwitch(networkService *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request UpdateStandardSwitchRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		mtu := 0
		vlan := 0

		if request.MTU != nil {
			if *request.MTU < 0 {
				c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
					Status:  "error",
					Message: "invalid_mtu",
					Error:   "mtu_must_be_positive_or_zero",
					Data:    nil,
				})
				return
			}
			mtu = *request.MTU
		}

		if request.VLAN != nil {
			if *request.VLAN < 0 {
				c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
					Status:  "error",
					Message: "invalid_vlan",
					Error:   "vlan_must_be_positive_or_zero",
					Data:    nil,
				})
				return
			}
			vlan = *request.VLAN
		}

		err := networkService.EditStandardSwitch(request.ID, mtu, vlan, request.Address, request.Address6, request.Ports, *request.Private)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_update_switch",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "switch_updated",
			Error:   "",
			Data:    nil,
		})
	}
}
