package infoHandlers

import (
	"net/http"
	"sylve/internal"
	"sylve/internal/services/info"

	"github.com/gin-gonic/gin"
)

// @Summary Get Historical Network information
// @Description Retrieves historical Network info
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} internal.APIResponse[[]infoServiceInterfaces.HistoricalNetworkInterface]
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /info/network-interfaces/historical [get]
func HistoricalNetworkInterfacesInfoHandler(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := infoService.GetNetworkInterfacesHistorical()
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "network_interfaces_info",
			Error:   "",
			Data:    info,
		})
	}
}
