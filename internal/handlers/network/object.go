package networkHandlers

import (
	"net/http"
	"sylve/internal"
	networkModels "sylve/internal/db/models/network"
	"sylve/internal/services/network"

	"github.com/gin-gonic/gin"
)

type CreateNetworkObjectRequest struct {
	Name   string   `json:"name" binding:"required"`
	Type   string   `json:"type" binding:"required"`
	Values []string `json:"values" binding:"required"`
}

// @Summary List Network Objects
// @Description List all network objects
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]networkModels.Object] "Success"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /network/object [get]
func ListNetworkObjects(svc *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		objects, err := svc.GetObjects()
		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_get_objects",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]networkModels.Object]{
			Status:  "success",
			Message: "objects_retrieved",
			Error:   "",
			Data:    objects,
		})
	}
}

// @Summary Create Network Object
// @Description Create a new network object with specified type and values
// @Tags Network
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateNetworkObject true "Create Network Object Request"
// @Success 200 {string} string "Samba share created successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Router /network/object [post]
func CreateNetworkObject(svc *network.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CreateNetworkObjectRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		if err := svc.CreateObject(request.Name, request.Type, request.Values); err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_create_object",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[string]{
			Status:  "success",
			Message: "object_created",
			Error:   "",
			Data:    "Object created successfully",
		})
	}
}
