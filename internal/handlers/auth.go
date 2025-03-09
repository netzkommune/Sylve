// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package handlers

import (
	"net/http"
	"sylve/internal/services/auth"
	"sylve/internal/utils"

	"github.com/gin-gonic/gin"
)

func LoginHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		type Request struct {
			Username string `json:"username" validate:"required"`
			Password string `json:"password" validate:"required"`
			AuthType string `json:"authType"`
			Remember bool   `json:"remember"`
		}

		var r Request

		if err := c.ShouldBindJSON(&r); err != nil {
			utils.SendJSONResponse(c, http.StatusBadRequest, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
			return
		}

		if err := Validate.Struct(r); err != nil {
			utils.SendJSONResponse(c, http.StatusBadRequest, gin.H{"status": "error", "message": "invalid_request", "error": err.Error()})
			return
		}

		token, err := authService.CreateJWT(r.Username, r.Password, r.AuthType, r.Remember)

		if err != nil {
			utils.SendJSONResponse(c, http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid_credentials", "error": err.Error()})
			return
		}

		hostname, err := utils.GetSystemHostname()

		if err != nil {
			utils.SendJSONResponse(c, http.StatusInternalServerError, gin.H{"status": "error", "message": "internal_server_error", "error": err.Error()})
			return
		}

		utils.SendJSONResponse(c, http.StatusOK, gin.H{"status": "success", "token": token, "hostname": hostname})
	}
}

func LogoutHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := utils.GetTokenFromHeader(c.Request.Header)

		if err != nil {
			utils.SendJSONResponse(c, http.StatusUnauthorized, gin.H{"status": "error", "message": "no_token_provided"})
			return
		}

		if err := authService.RevokeJWT(token); err != nil {
			utils.SendJSONResponse(c, http.StatusInternalServerError, gin.H{"status": "error", "message": "internal_server_error", "error": err.Error()})
			return
		}

		utils.SendJSONResponse(c, http.StatusOK, gin.H{"status": "success"})
	}
}
