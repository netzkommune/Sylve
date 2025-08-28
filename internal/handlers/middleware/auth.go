// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authService "github.com/alchemillahq/sylve/internal/services/auth"
	"github.com/alchemillahq/sylve/pkg/utils"
)

func EnsureAuthenticated(authService *authService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api/utilities/downloads/") &&
			len(path) > len("/api/utilities/downloads/") {
			c.Next()
			return
		}

		if path == "/api/auth/login" {
			c.Next()
			return
		}

		if clusterJWT, err := utils.GetClusterTokenFromHeader(c.Request.Header); err == nil && clusterJWT != "" {
			clusterClaims, err := authService.VerifyClusterJWT(clusterJWT)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "invalid_cluster_token"})
				return
			}

			c.Set("Token", clusterJWT)
			c.Set("AuthScope", "cluster")
			c.Set("UserID", clusterClaims.UserID)
			c.Set("Username", clusterClaims.Username)
			c.Set("AuthType", clusterClaims.AuthType)
			c.Next()
			return
		}

		var localJWT string
		if hash := c.Query("hash"); hash != "" {
			tok, err := authService.GetTokenBySHA256(hash)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "invalid_hash"})
				return
			}
			localJWT = tok
		}

		if localJWT == "" {
			var err error
			localJWT, err = utils.GetTokenFromHeader(c.Request.Header)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no_token_provided"})
				return
			}
		}

		claims, err := authService.ValidateToken(localJWT)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.Set("Token", localJWT)
		c.Set("AuthScope", "local")
		c.Set("UserID", claims.UserID)
		c.Set("Username", claims.Username)
		c.Set("AuthType", claims.AuthType)
		authService.UpdateLastUsageTime(claims.UserID)
		c.Next()
	}
}
