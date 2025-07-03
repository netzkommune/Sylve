// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirtHandlers

import (
	"strconv"
	"sylve/internal"
	vmModels "sylve/internal/db/models/vm"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	"sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

type VMEditDescRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// @Summary List all Virtual Machines
// @Description Retrieve a list of all virtual machines
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]vmModels.VM] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm [get]
func ListVMs(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		vms, err := libvirtService.ListVMs()

		for i := range vms {
			if vms[i].PCIDevices == nil {
				vms[i].PCIDevices = []int{}
			}
			if vms[i].CPUPinning == nil {
				vms[i].CPUPinning = []int{}
			}
		}

		if err != nil {
			c.JSON(500, internal.APIResponse[any]{Error: "failed_to_list_vms: " + err.Error()})
			return
		}

		c.JSON(200, internal.APIResponse[[]vmModels.VM]{
			Status:  "success",
			Message: "vm_listed",
			Data:    vms,
			Error:   "",
		})
	}
}

// @Summary Get a Virtual Machine's Domain
// @Description Retrieve the domain information of a virtual machine by its ID
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Virtual Machine ID"
// @Success 200 {object} internal.APIResponse[libvirtServiceInterfaces.LvDomain] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 404 {object} internal.APIResponse[any] "Not Found"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm/domain/{id} [get]
func GetLvDomain(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		vmID := c.Param("id")
		if vmID == "" {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_vm_id",
				Data:    nil,
				Error:   "Virtual Machine ID is required",
			})
			return
		}

		vmInt, err := strconv.Atoi(vmID)
		if err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_vm_id_format",
				Data:    nil,
				Error:   "Virtual Machine ID must be a valid integer",
			})
			return
		}

		domain, err := libvirtService.GetLvDomain(vmInt)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_get_domain",
				Data:    nil,
				Error:   "failed_to_get_domain: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[*libvirtServiceInterfaces.LvDomain]{
			Status:  "success",
			Message: "vm_domain_retrieved",
			Data:    domain,
			Error:   "",
		})
	}
}

// @Summary Create a new Virtual Machine
// @Description Create a new virtual machine with the specified parameters
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body libvirtServiceInterfaces.CreateVMRequest true "Create Virtual Machine Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm [post]
func CreateVM(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req libvirtServiceInterfaces.CreateVMRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		err := libvirtService.CreateVM(req)

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

// @Summary Remove a Virtual Machine
// @Description Remove a virtual machine by its ID
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Virtual Machine ID"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 404 {object} internal.APIResponse[any] "Not Found"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm/{id} [delete]
func RemoveVM(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		vmID := c.Param("id")
		if vmID == "" {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_vm_id",
				Data:    nil,
				Error:   "Virtual Machine ID is required",
			})
			return
		}

		vmInt, err := strconv.Atoi(vmID)
		if err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_vm_id_format",
				Data:    nil,
				Error:   "Virtual Machine ID must be a valid integer",
			})
			return
		}

		err = libvirtService.RemoveVM(uint(vmInt))

		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_remove_vm",
				Data:    nil,
				Error:   "failed_to_remove: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "vm_removed",
			Data:    nil,
			Error:   "",
		})
	}
}

// @Summary Perform an action on a Virtual Machine
// @Description Perform a specified action (start, stop, reboot) on a virtual machine by its ID
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Virtual Machine ID"
// @Param action path string true "Action to perform (start, stop, reboot)"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 404 {object} internal.APIResponse[any] "Not Found"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm/{id}/{action} [post]
func VMActionHandler(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		vmID := c.Param("id")
		action := c.Param("action")

		if vmID == "" || action == "" {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "Virtual Machine ID and action are required",
			})
			return
		}

		vmInt, err := strconv.Atoi(vmID)
		if err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_vm_id_format",
				Data:    nil,
				Error:   "Virtual Machine ID must be a valid integer",
			})
			return
		}

		err = libvirtService.PerformAction(uint(vmInt), action)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_perform_action",
				Data:    nil,
				Error:   "failed_to_perform_action: " + err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "action_performed",
			Data:    nil,
			Error:   "",
		})
	}
}

// @Summary Edit a Virtual Machine's description
// @Description Update the description of a virtual machine by its ID
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body VMEditDescRequest true "Edit Virtual Machine Description Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Router /vm/description [put]
func UpdateVMDescription(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VMEditDescRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		err := libvirtService.UpdateDescription(req.ID, req.Description)
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
			Message: "vm_description_updated",
			Data:    nil,
			Error:   "",
		})
	}
}
