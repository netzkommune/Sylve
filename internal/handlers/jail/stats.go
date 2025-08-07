package jailHandlers

import (
	"sylve/internal"
	jailServiceInterfaces "sylve/internal/interfaces/services/jail"
	"sylve/internal/services/jail"

	"github.com/gin-gonic/gin"
)

// @Summary List all Jails States
// @Description Retrieve a list of all jails states
// @Tags Jail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[[]jailServiceInterfaces.State] "Success"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /jail/state [get]
func ListJailStates(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		states, err := jailService.GetStates()
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_list_jail_states: " + err.Error(),
				Data:    nil,
				Error:   "Internal Server Error",
			})
			return
		}
		c.JSON(200, internal.APIResponse[[]jailServiceInterfaces.State]{
			Status:  "success",
			Message: "jail_states_listed",
			Data:    states,
			Error:   "",
		})
	}
}
