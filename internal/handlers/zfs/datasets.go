package zfsHandlers

import (
	"net/http"
	"sylve/internal"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/services/zfs"

	"github.com/gin-gonic/gin"
)

func GetDatasets(zfsSerice *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		datasets, err := zfsSerice.GetDatasets()

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]zfsServiceInterfaces.Dataset]{
			Status:  "success",
			Message: "datasets",
			Error:   "",
			Data:    datasets,
		})
	}
}
