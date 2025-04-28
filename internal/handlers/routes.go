// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
	static "github.com/soulteary/gin-static"

	"sylve/internal/assets"
	diskHandlers "sylve/internal/handlers/disk"
	infoHandlers "sylve/internal/handlers/info"
	"sylve/internal/handlers/middleware"
	zfsHandlers "sylve/internal/handlers/zfs"
	authService "sylve/internal/services/auth"
	diskService "sylve/internal/services/disk"
	infoService "sylve/internal/services/info"
	zfsService "sylve/internal/services/zfs"
)

// @title           Sylve API
// @version         0.0.1
// @description     Sylve is a lightweight GUI for managing Bhyve, Jails, ZFS, networking, and more on FreeBSD.
// @termsOfService  https://github.com/AlchemillaHQ/Sylve/blob/master/LICENSE

// @contact.name   Alchemilla Ventures Pvt. Ltd.
// @contact.url    https://alchemilla.io
// @contact.email  hello@alchemilla.io

// @license.name  BSD-2-Clause
// @license.url   https://github.com/AlchemillaHQ/Sylve/blob/master/LICENSE

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @host      sylve.lan:8181
// @BasePath  /api
func RegisterRoutes(r *gin.Engine,
	environment string,
	proxyToVite bool,
	authService *authService.Service,
	infoService *infoService.Service,
	zfsService *zfsService.Service,
	diskService *diskService.Service,
) {
	api := r.Group("/api")

	health := api.Group("/health")
	health.Use(middleware.EnsureAuthenticated(authService))
	{
		health.GET("/basic", BasicHealthCheckHandler)
		health.GET("/http", HTTPHealthCheckHandler)
	}

	info := api.Group("/info")
	info.Use(middleware.EnsureAuthenticated(authService))
	{
		info.GET("/basic", infoHandlers.BasicInfo(infoService))

		info.GET("/cpu", infoHandlers.RealTimeCPUInfoHandler(infoService))
		info.GET("/cpu/historical", infoHandlers.HistoricalCPUInfoHandler(infoService))

		info.GET("/ram", infoHandlers.RAMInfo(infoService))
		info.GET("/swap", infoHandlers.SwapInfo(infoService))

		notes := info.Group("/notes")
		{
			notes.GET("", infoHandlers.NotesHandler(infoService))
			notes.POST("", infoHandlers.NotesHandler(infoService))
			notes.DELETE("/:id", infoHandlers.NotesHandler(infoService))
			notes.PUT("/:id", infoHandlers.NotesHandler(infoService))
		}

		info.GET("/audit-logs", infoHandlers.AuditLogs(infoService))
		info.GET("/terminal", infoHandlers.HandleTerminalWebsocket)
	}

	zfs := api.Group("/zfs")
	zfs.Use(middleware.EnsureAuthenticated(authService))
	{
		zfs.GET("/pool/io-delay", zfsHandlers.AvgIODelay(zfsService))
		zfs.GET("/pool/io-delay/historical", zfsHandlers.AvgIODelayHistorical(zfsService))

		pools := zfs.Group("/pools")
		{
			pools.GET("", zfsHandlers.GetPools(zfsService))
			pools.POST("", zfsHandlers.CreatePool(infoService, zfsService))
			pools.POST("/:name/scrub", zfsHandlers.ScrubPool(infoService, zfsService))
			pools.DELETE("/:name", zfsHandlers.DeletePool(infoService, zfsService))
			pools.POST("/:name/replace-device", zfsHandlers.ReplaceDevice(infoService, zfsService))
		}

		datasets := zfs.Group("/datasets")
		{
			datasets.GET("", zfsHandlers.GetDatasets(zfsService))
			datasets.POST("/snapshot", zfsHandlers.CreateSnapshot(zfsService))
			datasets.POST("/snapshot/rollback", zfsHandlers.RollbackSnapshot(zfsService))
			datasets.DELETE("/snapshot/:guid", zfsHandlers.DeleteSnapshot(zfsService))

			datasets.GET("/snapshot/periodic", zfsHandlers.GetPeriodicSnapshots(zfsService))
			datasets.POST("/snapshot/periodic", zfsHandlers.CreatePeriodicSnapshot(zfsService))
			datasets.DELETE("/snapshot/periodic/:guid", zfsHandlers.DeletePeriodicSnapshot(zfsService))

			datasets.POST("/filesystem", zfsHandlers.CreateFilesystem(zfsService))
			datasets.DELETE("/filesystem/:guid", zfsHandlers.DeleteFilesystem(zfsService))

			datasets.POST("/volume", zfsHandlers.CreateVolume(zfsService))
			datasets.DELETE("/volume/:guid", zfsHandlers.DeleteVolume(zfsService))
		}
	}

	disk := api.Group("/disk")
	disk.Use(middleware.EnsureAuthenticated(authService))
	{
		disk.GET("/list", diskHandlers.List(diskService))
		disk.POST("/wipe", diskHandlers.WipeDisk(diskService, infoService))
		disk.POST("/initialize-gpt", diskHandlers.InitializeGPT(diskService, infoService))
		disk.POST("/create-partitions", diskHandlers.CreatePartition(infoService))
		disk.POST("/delete-partition", diskHandlers.DeletePartition(infoService))
	}

	auth := api.Group("/auth")
	{
		auth.POST("/login", LoginHandler(authService))
		auth.Any("/logout", LogoutHandler(authService))
	}

	if proxyToVite {
		r.NoRoute(func(c *gin.Context) {
			ReverseProxy(c, "http://127.0.0.1:5173")
		})
	} else {
		staticFiles, err := static.EmbedFolder(assets.SvelteKitFiles, "web-files")
		if err != nil {
			log.Fatalln("Initialization of embed folder failed:", err)
		} else {
			r.Use(static.Serve("/", staticFiles))
			r.NoRoute(func(c *gin.Context) {
				c.FileFromFS("200.html", staticFiles)
			})
		}
	}
}
