// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package info

import (
	infoModels "github.com/alchemillahq/sylve/internal/db/models/info"
)

func (s *Service) GetAuditRecords(limit int) ([]infoModels.AuditRecord, error) {
	var records []infoModels.AuditRecord
	err := s.DB.Order("created_at desc").Limit(limit).Find(&records).Error

	return records, err
}
