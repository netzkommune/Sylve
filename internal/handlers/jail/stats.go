package jailHandlers

import (
	"strconv"
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

// @Summary Get Jail Logs
// @Description Retrieve start/stop logs for a specific jail
// @Tags Jail
// @Accept json
// @Produce json
// @Param id path int true "Jail ID"
// @Param start query bool false "Get start logs (default: false for stop logs)"
// @Security BearerAuth
// @Success 200 {object} internal.APIResponse[string] "Success"
// @Failure 404 {object} internal.APIResponse[any] "Jail Not Found"
// @Failure 500 {object} internal.APIResponse[any] "Internal Server Error"
// @Router /jail/:id/logs [get]
func GetJailLogs(jailService *jail.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_jail_id",
				Data:    nil,
				Error:   "Bad Request",
			})
			return
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(400, internal.APIResponse[any]{
				Status:  "error",
				Message: "invalid_jail_id_format: " + err.Error(),
				Data:    nil,
				Error:   "Bad Request",
			})
			return
		}

		start := c.Query("start") == "true"

		logs, err := jailService.GetJailLogs(uint(idInt), start)
		if err != nil {
			c.JSON(500, internal.APIResponse[any]{
				Status:  "error",
				Message: "failed_to_get_jail_logs: " + err.Error(),
				Data:    nil,
				Error:   "Internal Server Error",
			})
			return
		}

		type LogsResponse struct {
			Logs string `json:"logs"`
		}

		c.JSON(200, internal.APIResponse[LogsResponse]{
			Status:  "success",
			Message: "jail_logs_retrieved",
			Data:    LogsResponse{Logs: logs},
			Error:   "",
		})
	}
}
