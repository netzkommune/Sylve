package libvirtHandlers

import (
	"strconv"
	"sylve/internal"
	vmModels "sylve/internal/db/models/vm"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	"sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

// @Summary List all Virtual Machines
// @Description Retrieve a list of all virtual machines
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]vmModels.vm] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm/list [get]
func ListVMs(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		vms, err := libvirtService.ListVMs()
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

// @Summary Create a new Virtual Machine
// @Description Create a new virtual machine with the specified parameters
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateVMRequest true "Create Virtual Machine Request"
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm/create [post]
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
// @Router /vm/remove/{id} [delete]
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
			c.JSON(500, internal.APIResponse[any]{Error: "failed_to_remove: " + err.Error()})
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
