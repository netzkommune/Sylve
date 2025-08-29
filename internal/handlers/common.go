// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package handlers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"
	"github.com/alchemillahq/sylve/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var hostname string

func ReverseProxy(c *gin.Context, backend string) {
	remote, err := url.Parse(backend)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse proxy URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		if !strings.Contains(err.Error(), "context canceled") {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func ReverseProxyInsecure(c *gin.Context, backend string) {
	remote, err := url.Parse(backend)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse proxy URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		if !strings.Contains(err.Error(), "context canceled") {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func EnsureCorrectHost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error

		if hostname == "" {
			hostname, err = utils.GetSystemHostname()
			if err != nil {
				c.Next()
				return
			}
		}

		reqHost, err := utils.GetCurrentHostnameFromHeader(c.Request.Header)
		if err != nil {
			c.Next()
			return
		}

		if reqHost != hostname {
			var node clusterModels.ClusterNode
			if err := db.Where("hostname = ?", reqHost).First(&node).Error; err != nil {
				c.Next()
				return
			}

			if node.Status == "online" {
				ReverseProxyInsecure(c, fmt.Sprintf("https://%s", node.API))
			}

			c.Next()
		}
	}
}
