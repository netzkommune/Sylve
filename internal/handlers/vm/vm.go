package libvirtHandlers

import (
	"sylve/internal"
	"sylve/internal/services/libvirt"

	"github.com/gin-gonic/gin"
)

type CreateVMRequest struct {
	Name                 string   `json:"name" binding:"required"`
	VMID                 *int     `json:"vmId" binding:"required"`
	Description          string   `json:"description"`
	StorageType          string   `json:"storageType" binding:"required"`
	StorageDataset       string   `json:"storageDataset" binding:"required"`
	StorageSize          *int64   `json:"storageSize" binding:"required"`
	StorageEmulationType string   `json:"storageEmulationType"`
	SwitchID             *int     `json:"switchId" binding:"required"`
	NetworkMAC           string   `json:"networkMAC"`
	CPUSockets           int      `json:"cpuSockets" binding:"required"`
	CPUCores             int      `json:"cpuCores" binding:"required"`
	CPUThreads           int      `json:"cpuThreads" binding:"required"`
	RAM                  int      `json:"ram" binding:"required"`
	PCIDevices           []string `json:"pciDevices"`
	VNCPort              int      `json:"vncPort" binding:"required"`
	VNCPassword          string   `json:"vncPassword"`
	VNCResolution        string   `json:"vncResolution"`
	StartAtBoot          *bool    `json:"startAtBoot" binding:"required"`
	StartOrder           int      `json:"startOrder"`
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
		var req CreateVMRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request_data",
				Data:    nil,
				Error:   "Invalid request data: " + err.Error(),
			})
			return
		}

		// err := libvirtService.CreateVM(
		// 	req.Name,
		// 	req.VMID,
		// )

		// if err != nil {
		// 	c.JSON(500, internal.APIResponse[any]{Error: "Failed to create VM: " + err.Error()})
		// 	return
		// }

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "vm_created",
			Data:    nil,
			Error:   "",
		})
	}
}
