// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package info

import (
	"fmt"
	infoModels "sylve/internal/db/models/info"
	"sylve/pkg/utils"
	"time"
)

func (s *Service) GetAuditLogs(limit int) ([]infoModels.AuditLog, error) {
	var logs []infoModels.AuditLog
	err := s.DB.Order("created_at desc").Limit(limit).Find(&logs).Error

	return logs, err
}

func (s *Service) StartAuditLog(token string, action string, status string) uint {

	hostname, err := utils.GetSystemHostname()

	if err != nil {
		hostname = "unknown"
	}

	claimsInterface, err := utils.ParseJWT(token)

	if err != nil {
		fmt.Println("Error", err)
		return 0
	}

	claimsMap, ok := claimsInterface.(map[string]interface{})
	if !ok {
		fmt.Println("Error: invalid claims format")
		return 0
	}
	claims := claimsMap["custom_claims"].(map[string]interface{})
	userID := uint(claims["userId"].(float64))
	user := claims["username"].(string)
	authType := claims["authType"].(string)

	log := infoModels.AuditLog{
		UserID:   userID,
		User:     user,
		AuthType: authType,
		Node:     hostname,
		Action:   action,
		Status:   status,
		Started:  time.Now(),
		Ended:    time.Time{},
	}

	err = s.DB.Create(&log).Error

	if err != nil {
		return 0
	}

	return log.ID
}

func (s *Service) EndAuditLog(logID uint, status string) error {
	if logID == 0 {
		return nil
	}

	var log infoModels.AuditLog
	err := s.DB.Where("id = ?", logID).First(&log).Error

	if err != nil {
		return err
	}

	log.Status = status
	log.Ended = time.Now()

	s.DB.Save(&log)

	return nil
}
