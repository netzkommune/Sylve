package jailHandlers

import (
	"strconv"
	"sylve/internal"
	"sylve/internal/services/jail"

	"github.com/gin-gonic/gin"
)

type JailInheritNetworkRequest struct {
	CTID uint  `json:"ctId" binding:"required"`
	IPv4 *bool `json:"ipv4"`
	IPv6 *bool `json:"ipv6"`
}

// @Summary Update Jail to Inherit Hosts Network
// @Description Update the network settings of a jail to inherit from the host
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body jailServiceInterfaces.JailInheritNetworkRequest true "Inherit Network Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Router /jail/network/inheritance [post]
func InheritJailNetwork(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req JailInheritNetworkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		ipv4 := false
		ipv6 := false

		if req.IPv4 != nil {
			ipv4 = *req.IPv4
		}

		if req.IPv6 != nil {
			ipv6 = *req.IPv6
		}

		if ipv4 == false && ipv6 == false {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "atleast_one_of_ipv4_or_ipv6_must_be_specified",
			})
			return
		}

		err := jailService.InheritNetwork(req.CTID, ipv4, ipv6)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_inherit_network",
				Data:    nil,
				Error:   "failed_to_inherit_network: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "jail_network_inherited",
			Data:    nil,
			Error:   "",
		})
	}
}

/*
func (s *Service) DisinheritNetwork(ctId uint) error {
	var jail jailModels.Jail

	if err := s.DB.Preload("Networks").First(&jail, ctId).Error; err != nil {
		return err
	}

	jail.InheritIPv4 = false
	jail.InheritIPv6 = false

	return s.SyncNetwork(ctId, jail)
}
*/
// @Summary Update Jail to Disinherit Hosts Network
// @Description Update the network settings of a jail to disinherit from the host
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ctId path uint true "Container ID"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Router /jail/network/disinherit/{ctId} [delete]
func DisinheritJailNetwork(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctId := c.Param("ctId")
		if ctId == "" {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_ct_id",
				Data:    nil,
				Error:   "CT ID is required",
			})
			return
		}

		ctIdUint, err := strconv.ParseUint(ctId, 10, 32)
		if err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_ct_id",
				Data:    nil,
				Error:   "Invalid CT ID: " + err.Error(),
			})
			return
		}

		err = jailService.DisinheritNetwork(uint(ctIdUint))
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_disinherit_network",
				Data:    nil,
				Error:   "failed_to_disinherit_network: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "jail_network_disinherited",
			Data:    nil,
			Error:   "",
		})
	}
}
