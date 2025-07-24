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
	sambaModels "sylve/internal/db/models/samba"
	utilitiesModels "sylve/internal/db/models/utilities"
	vmModels "sylve/internal/db/models/vm"
	zfsModels "sylve/internal/db/models/zfs"
	"sylve/internal/logger"
	"sylve/pkg/system"
	"sylve/pkg/utils"

	"gorm.io/driver/sqlite"
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
	db.Exec("PRAGMA journal_mode = WAL")
	db.Exec("PRAGMA synchronous = NORMAL")

	err = db.AutoMigrate(
		&models.System{},
		&models.User{},
		&models.Group{},
		&models.Token{},
		&models.SystemSecrets{},

		&vmModels.Storage{},
		&vmModels.Network{},
		&vmModels.VMStats{},
		&vmModels.VM{},

		&models.PassedThroughIDs{},

		&infoModels.CPU{},
		&infoModels.RAM{},
		&infoModels.Swap{},
		&infoModels.IODelay{},
		&infoModels.NetworkInterface{},
		&infoModels.Note{},
		&infoModels.AuditRecord{},

		&infoModels.ZPoolHistorical{},

		&zfsModels.PeriodicSnapshot{},

		&networkModels.StandardSwitch{},
		&networkModels.NetworkPort{},

		&utilitiesModels.DownloadedFile{},
		&utilitiesModels.Downloads{},

		&sambaModels.SambaSettings{},
		&sambaModels.SambaShare{},
	)

	if err != nil {
		logger.L.Fatal().Msgf("Error migrating database: %v", err)
	}

	err = setupInitUsers(db, cfg)

	if err != nil {
		logger.L.Fatal().Msgf("Error setting up initial users: %v", err)
	}

	if !isTest {
		if err := db.Exec("VACUUM").Error; err != nil {
			logger.L.Warn().Msgf("VACUUM failed: %v", err)
		}
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
					Admin:    true,
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
		} else {
			updates := make(map[string]interface{})

			if !user.Admin {
				updates["admin"] = true
				logger.L.Info().Msgf("User %s promoted to admin", admin.Email)
			}

			passwordMatches := utils.CheckPasswordHash(admin.Password, user.Password)
			if !passwordMatches {
				hashed, err := utils.HashPassword(admin.Password)
				if err != nil {
					logger.L.Error().Msgf("Failed to hash updated password for user %s: %v", admin.Email, err)
					return err
				}
				updates["password"] = hashed
				logger.L.Info().Msgf("Password for admin user %s updated", admin.Email)
			}

			if len(updates) > 0 {
				if err := db.Model(&user).Updates(updates).Error; err != nil {
					logger.L.Error().Msgf("Failed to update user %s: %v", admin.Email, err)
					return err
				}
			}
		}

		exists, err := system.UnixUserExists(admin.Username)
		if err != nil {
			logger.L.Error().Msgf("Error checking if Unix user %s exists: %v", admin.Username, err)
		}

		if !exists {
			err = system.CreateUnixUser(admin.Username, "/usr/sbin/nologin", "/nonexistent")
			if err != nil {
				logger.L.Error().Msgf("Failed to create Unix user %s: %v", admin.Username, err)
				return err
			}
			logger.L.Info().Msgf("Unix user %s created successfully", admin.Username)
		}
	}
	return nil
}
