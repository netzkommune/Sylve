package systemHandlers

import (
	"net/http"
	"sylve/internal"
	systemServiceInterfaces "sylve/internal/interfaces/services/system"
	"sylve/internal/services/system"

	"github.com/gin-gonic/gin"
)

// /api/files?id="
// @Summary Find Files on System
// @Description Find files on the system based on a search term
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Query id string "Search term"
// @Success 200 {object} internal.APIResponse[[]systemServiceInterfaces.FileNode]
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /system/file-explorer/files [get]
func Files(systemService *system.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		nodes, err := systemService.Traverse(id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]systemServiceInterfaces.FileNode]{
			Status:  "success",
			Message: "files_listed",
			Error:   "",
			Data:    nodes,
		})
	}
}
