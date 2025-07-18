package systemHandlers

import (
	"net/http"
	"path/filepath"
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

type RenameFileOrFolderRequest struct {
	ID      string `json:"id" binding:"required"`
	NewName string `json:"newName" binding:"required"`
}

type CopyOrMoveFileOrFolderRequest struct {
	ID      string `json:"id" binding:"required"`
	NewPath string `json:"newPath" binding:"required"`
	Cut     *bool  `json:"cut" binding:"required"`
}

type DeleteFilesOrFoldersRequest struct {
	Paths []string `json:"paths" binding:"required"`
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
// @Router /system/file-explorer [get]
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
// @Router /system/file-explorer [post]
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
// @Router /system/file-explorer [delete]
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

// @Summary Delete Files or Folders
// @Description Delete multiple files or folders from the system
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Request body DeleteFilesOrFoldersRequest true "Delete Files or Folders Request"
// @Success 200 {object} internal.APIResponse[any]
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /system/file-explorer/delete [post]
func DeleteFilesOrFolders(systemService *system.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request DeleteFilesOrFoldersRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		if len(request.Paths) == 0 {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   "no paths provided",
				Data:    nil,
			})
			return
		}

		err := systemService.DeleteFilesOrFolders(request.Paths)
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
			Message: "files_or_folders_deleted",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Rename File or Folder
// @Description Rename a file or folder in the system
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Path to the file or folder"
// @Request body RenameFileOrFolderRequest true "Rename File or Folder Request"
// @Success 200 {object} internal.APIResponse[any]
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /system/file-explorer/rename [post]
func RenameFileOrFolder(systemService *system.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request RenameFileOrFolderRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		if request.ID == "" || request.NewName == "" {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   "id and new name are required",
				Data:    nil,
			})
			return
		}

		err := systemService.RenameFileOrFolder(request.ID, request.NewName)
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
			Message: "file_or_folder_renamed",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Download File
// @Description Download a file from the system
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Path to the file"
// @Success 200 {file} file "File content"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /system/file-explorer/download [get]
func DownloadFile(systemService *system.Service) gin.HandlerFunc {
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

		file, err := systemService.DownloadFile(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "internal_server_error",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+filepath.Base(id))
		c.Header("Content-Type", "application/octet-stream")
		c.File(file)
	}
}

// @Summary Copy or Move File or Folder
// @Description Copy or move a file or folder to a new path
// @Tags System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CopyOrMoveFileOrFolderRequest true "Request body"
// @Success 200 {object} internal.APIResponse[any]
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /system/file-explorer/copy-or-move [post]
func CopyOrMoveFileOrFolder(systemService *system.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CopyOrMoveFileOrFolderRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		if request.ID == "" || request.NewPath == "" {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "bad_request",
				Error:   "id and new path are required",
				Data:    nil,
			})
			return
		}

		move := false
		if request.Cut != nil {
			move = *request.Cut
		}

		if err := systemService.CopyOrMoveFileOrFolder(request.ID, request.NewPath, move); err != nil {
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
			Message: "file_or_folder_copied_or_moved",
			Error:   "",
			Data:    nil,
		})
	}
}
