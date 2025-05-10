// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package db

import (
	"sylve/internal"
	"sylve/internal/db/models"
	infoModels "sylve/internal/db/models/info"
	networkModels "sylve/internal/db/models/network"
	zfsModels "sylve/internal/db/models/zfs"
	"sylve/internal/logger"
	"sylve/pkg/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func SetupDatabase(cfg *internal.SylveConfig, isTest bool) *gorm.DB {
	ormConfig := &gorm.Config{
		Logger:         gormLogger.Default.LogMode(gormLogger.Silent),
		TranslateError: true,
	}

	var db *gorm.DB
	var err error

	if isTest {
		db, err = gorm.Open(sqlite.Open(":memory:"), ormConfig)
	} else {
		db, err = gorm.Open(sqlite.Open(cfg.DataPath+"/sylve.db"), ormConfig)
	}

	if err != nil {
		logger.L.Fatal().Msgf("Error connecting to database: %v", err)
	}

	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode=WAL")

	err = db.AutoMigrate(
		&models.System{},
		&models.User{},
		&models.Token{},
		&models.SystemSecrets{},

		&infoModels.CPU{},
		&infoModels.IODelay{},
		&infoModels.Note{},
		&infoModels.AuditLog{},

		&zfsModels.PeriodicSnapshot{},

		&networkModels.NetworkPort{},
		&networkModels.StandardSwitch{},
	)

	if err != nil {
		logger.L.Fatal().Msgf("Error migrating database: %v", err)
	}

	err = setupInitUsers(db, cfg)

	if err != nil {
		logger.L.Fatal().Msgf("Error setting up initial users: %v", err)
	}

	return db
}

func setupInitUsers(db *gorm.DB, cfg *internal.SylveConfig) error {
	for _, admin := range cfg.Admins {
		var user models.User
		result := db.Where("email = ?", admin.Email).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				hashed, err := utils.HashPassword(admin.Password)
				if err != nil {
					logger.L.Error().Msgf("Failed to hash password for user %s: %v", admin.Email, err)
					return err
				}

				newUser := models.User{
					Username: admin.Username,
					Email:    admin.Email,
					Password: hashed,
				}
				if err := db.Create(&newUser).Error; err != nil {
					logger.L.Error().Msgf("Failed to create admin user %s: %v", admin.Email, err)
					return err
				}
				logger.L.Info().Msgf("Admin user %s created successfully", admin.Email)
			} else {
				logger.L.Error().Msgf("Error checking user %s: %v", admin.Email, result.Error)
				return result.Error
			}
		}
	}
	return nil
}
