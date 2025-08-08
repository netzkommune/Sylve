package jailHandlers

import (
	"fmt"
	"strconv"
	"sylve/internal"
	jailModels "sylve/internal/db/models/jail"
	jailServiceInterfaces "sylve/internal/interfaces/services/jail"
	"sylve/internal/services/jail"

	"github.com/gin-gonic/gin"
)

type JailEditDescRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Description string `json:"description"`
}

// @Summary List all Jails
// @Description Retrieve a list of all jails
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]jailModels.Jail] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /jail [get]
func ListJails(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		jails, err := jailService.GetJails()

		if err != nil {
			c.JSON(500, internal.APIResponse[any]{Error: "failed_to_list_jails: " + err.Error()})
			return
		}

		c.JSON(200, internal.APIResponse[[]jailModels.Jail]{
			Status:  "success",
			Message: "jail_listed",
			Data:    jails,
			Error:   "",
		})
	}
}

// @Summary List all Jails (Simple)
// @Description Retrieve a simple list of all jails
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]jailServiceInterfaces.SimpleList] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /jail/simple [get]
func ListJailsSimple(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		jails, err := jailService.GetJailsSimple()

		if err != nil {
			c.JSON(500, internal.APIResponse[any]{Error: "failed_to_list_jails_simple: " + err.Error()})
			return
		}

		c.JSON(200, internal.APIResponse[[]jailServiceInterfaces.SimpleList]{
			Status:  "success",
			Message: "jail_listed_simple",
			Data:    jails,
			Error:   "",
		})
	}
}

// @Summary Create a new Jail
// @Description Create a new jail with the provided configuration
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body jailServiceInterfaces.CreateJailRequest true "Create Jail Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /jail [post]
func CreateJail(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req jailServiceInterfaces.CreateJailRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		err := jailService.CreateJail(req)

		if err != nil {
			c.JSON(500, internal.APIResponse[any]{Error: "failed_to_create: " + err.Error()})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "vm_created",
			Data:    nil,
			Error:   "",
		})
	}
}

// @Summary Delete a Jail
// @Description Delete a jail by its CTID
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ctid path int true "CTID of the Jail"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 404 {object} internal.APIResponse[any] "Not Found"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /jail/{ctid} [delete]
func DeleteJail(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctid, berr := c.Params.Get("ctid")
		if !berr || ctid == "" {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_ctid",
				Data:    nil,
				Error:   "invalid_ctid: ",
			})
			return
		}

		var ctidInt int
		ctidInt, err := strconv.Atoi(ctid)
		if err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_ctid_format",
				Data:    nil,
				Error:   fmt.Sprintf("invalid_ctid_format: %s", err.Error()),
			})
			return
		}

		err = jailService.DeleteJail(uint(ctidInt))

		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_delete_jail",
				Data:    nil,
				Error:   "failed_to_delete_jail: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "jail_deleted",
			Data:    nil,
			Error:   "",
		})
	}
}

// @Summary Edit a Jail's description
// @Description Update the description of a jail by its ID
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body jailServiceInterfaces.JailEditDescRequest true "Edit Jail Description Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Router /jail/description [put]
func UpdateJailDescription(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req JailEditDescRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		err := jailService.UpdateDescription(req.ID, req.Description)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_update_description",
				Data:    nil,
				Error:   "failed_to_update_description: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "jail_description_updated",
			Data:    nil,
			Error:   "",
		})
	}
}
