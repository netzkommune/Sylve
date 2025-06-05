package libvirtHandlers

import (
	"sylve/internal"
	vmModels "sylve/internal/db/models/vm"
	"sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

type StatsRequest struct {
	VMID  int `json:"vmId" binding:"required"`
	Limit int `json:"limit" binding:"required"`
}

// @Summary Get VM Statistics
// @Description Retrieve statistics for a virtual machine
// @Tags VM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]vmModels.VMStats] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /vm/stats [post]
func GetVMStats(libvirtService *libvirt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StatsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_request: " + err.Error(),
			})
			return
		}

		stats, err := libvirtService.GetVMUsage(req.VMID, req.Limit)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(200, internal.APIResponse[[]vmModels.VMStats]{
			Status:  "success",
			Message: "vm_stats_retrieved",
			Data:    stats,
			Error:   "",
		})
	}
}
