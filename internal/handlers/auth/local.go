package authHandlers

import (
	"net/http"
	"strconv"
	"sylve/internal"
	"sylve/internal/db/models"
	"sylve/internal/services/auth"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=128"`
	Password string `json:"password" binding:"required,min=3,max=128"`
	Email    string `json:"email"`
	Admin    *bool  `json:"admin" binding:"required"`
}

type EditUserRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Admin    *bool  `json:"admin" binding:"required"`
}

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

// @Summary Create User
// @Description Create a new local (sylve) user in the system
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "Create User Request"
// @Success 201 {object} internal.APIResponse[models.User] "User Created"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /auth/users [post]
func CreateUserHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Data:    nil,
				Error:   "invalid_request: " + err.Error(),
			})
			return
		}

		admin := false

		if req.Admin != nil {
			admin = *req.Admin
		}

		var model models.User

		model.Username = req.Username
		model.Password = req.Password
		model.Email = req.Email
		model.Admin = admin

		err := authService.CreateUser(&model)

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_create_user",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusCreated, internal.APIResponse[any]{
			Status:  "success",
			Message: "user_created_successfully",
			Error:   "",
			Data:    nil,
		})
	}
}

// @Summary Delete User
// @Description Delete a local (sylve) user from the system
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "User ID"
// @Success 204 {object} internal.APIResponse[any] "User Deleted"
// @Failure 400 {object} internal.APIResponse[any] "Bad Request"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /auth/users/{id} [delete]
func DeleteUserHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if id == "" {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_user_id",
				Error:   "user_id_is_required",
				Data:    nil,
			})
			return
		}

		idInt, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_user_id",
				Error:   "invalid_user_id_format",
				Data:    nil,
			})
			return
		}

		err = authService.DeleteUser(uint(idInt))

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_delete_user",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(200, internal.APIResponse[any]{
			Status:  "success",
			Message: "user_deleted_successfully",
			Error:   "",
			Data:    nil,
		})
	}
}

func EditUserHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req EditUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_request",
				Error:   "invalid_request: " + err.Error(),
				Data:    nil,
			})
			return
		}

		var admin bool
		if req.Admin != nil {
			admin = *req.Admin
		} else {
			admin = false
		}

		err := authService.EditUser(req.ID, req.Username, req.Password, req.Email, admin)

		if err != nil {
			c.JSON(http.StatusInternalServerError, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_edit_user",
				Error:   err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, internal.APIResponse[any]{
			Status:  "success",
			Message: "user_edited_successfully",
			Error:   "",
			Data:    nil,
		})
	}
}
