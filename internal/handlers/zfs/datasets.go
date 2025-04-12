package zfsHandlers

import (
	"net/http"
	"sylve/internal"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/services/zfs"

	"github.com/gin-gonic/gin"
)

type CreateSnapshotRequest struct {
	GUID      string `json:"guid" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Recursive bool   `json:"recursive"`
}

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

func DeleteSnapshot(zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		guid := c.Param("guid")
		err := zfsService.DeleteSnapshot(guid)

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
			Message: "deleted_snapshot",
			Error:   "",
			Data:    nil,
		})
	}
}

func CreateSnapshot(zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CreateSnapshotRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		err := zfsService.CreateSnapshot(request.GUID, request.Name, request.Recursive)

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
			Message: "created_snapshot",
			Error:   "",
			Data:    nil,
		})
	}
}

func CreateFilesystem(zfsService *zfs.Service) gin.HandlerFunc {
	return func(c *gin.Context) {}
}