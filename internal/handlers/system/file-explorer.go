package systemHandlers

import (
	"net/http"
	"sylve/internal"
	systemServiceInterfaces "sylve/internal/interfaces/services/system"
	"sylve/internal/services/system"

	"github.com/gin-gonic/gin"
)

type AddFileOrFolderRequest struct {
	Path     string `json:"path" binding:"required"`
	Name     string `json:"name" binding:"required"`
	IsFolder *bool  `json:"isFolder" binding:"required"`
}

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

// @Summary Add File or Folder
// @Description Add a file or folder to the system
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddFileOrFolderRequest true "Request body"
// @Success 200 {object} internal.APIResponse[any]
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /system/file-explorer/add [post]
func AddFileOrFolder(systemService *system.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request AddFileOrFolderRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		var isFolder bool

		if request.IsFolder != nil {
			isFolder = *request.IsFolder
		} else {
			isFolder = false
		}

		err := systemService.AddFileOrFolder(request.Path, request.Name, isFolder)
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
			Message: "file_or_folder_added",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Delete File or Folder
// @Description Delete a file or folder from the system
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Path to the file or folder"
// @Success 200 {object} internal.APIResponse[any]
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /system/file-explorer/delete [delete]
func DeleteFileOrFolder(systemService *system.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   "id is required",
				Data:    nil,
			})
			return
		}

		err := systemService.DeleteFileOrFolder(id)
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
			Message: "file_or_folder_deleted",
			Error:   "",
			Data:    nil,
		})
	}
}
