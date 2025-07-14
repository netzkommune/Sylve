package authHandlers

import (
	"net/http"
	"sylve/internal"
	"sylve/internal/db/models"
	"sylve/internal/services/auth"

	"github.com/gin-gonic/gin"
)

// @Summary List Users
// @Description List all users in the system
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]models.User] "Success"
// @Failure 401 {object} internal.APIResponse[any] "Unauthorized"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /auth/users [get]
func ListUsersHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := authService.ListUsers()

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_list_users",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[[]models.User]{
			Status:  "success",
			Message: "users_listed_successfully",
			Error:   "",
			Data:    users,
		})
	}
}
