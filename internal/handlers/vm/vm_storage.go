package libvirtHandlers

import (
	"sylve/internal"
	"sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

type StorageDetachRequest struct {
	VMID      int `json:"vmId" binding:"required"`
	StorageId int `json:"storageId" binding:"required"`
}

// @Summary Detach Storage from a Virtual Machine
// @Description Detach a storage volume from a virtual machine
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[any] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /storage/detach [post]
func StorageDetach(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StorageDetachRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_request: " + err.Error(),
			})
			return
		}

		if err := libvirtService.StorageDetach(req.VMID, req.StorageId); err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "storage_detached",
			Data:    nil,
			Error:   "",
		})
	}
}
