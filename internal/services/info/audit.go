// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package info

import infoModels "sylve/internal/db/models/info"

func (s *Service) GetAuditLogs(limit int) ([]infoModels.AuditLog, error) {
	var logs []infoModels.AuditLog
	err := s.DB.Order("created_at desc").Limit(limit).Find(&logs).Error

	return logs, err
}
