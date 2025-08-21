// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package samba

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	sambaModels "github.com/alchemillahq/sylve/internal/db/models/samba"
	sambaServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/samba"
)

func (s *Service) ParseAuditLogs() error {
	const logPath = "/var/log/samba4/audit.log"

	validActions := map[string]bool{
		"connect":     true,
		"disconnect":  true,
		"mkdirat":     true,
		"unlinkat":    true,
		"renameat":    true,
		"create_file": true,
	}

	seenCreates := make(map[string]bool)

	f, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, ": ")
		if idx < 0 {
			continue
		}
		payload := line[idx+2:]
		if !strings.HasPrefix(payload, "sylve-smb-al|") {
			continue
		}

		parts := strings.Split(payload, "|")
		if len(parts) < 8 {
			continue
		}

		user := parts[1]
		ip := parts[2]
		share := parts[4]
		action := parts[6]
		result := parts[7]
		args := parts[8:]

		if !validActions[action] {
			continue
		}

		entry := sambaModels.SambaAuditLog{
			Share:  share,
			User:   user,
			IP:     ip,
			Action: action,
			Result: result,
		}

		switch action {
		case "mkdirat", "unlinkat":
			entry.Path = args[len(args)-1]

		case "renameat":
			if len(args) >= 2 {
				entry.Path = args[0]
				entry.Target = args[1]
			}

		case "create_file":
			if len(args) >= 2 && args[len(args)-2] == "create" {
				p := args[len(args)-1]
				if !seenCreates[p] {
					seenCreates[p] = true
					entry.Path = p
				}
			}
		}

		if entry.Path != "" {
			entry.Folder = filepath.Base(entry.Path)

			if action == "create_file" {
				var cnt int64
				cutoff := time.Now().Add(-5 * time.Second)
				if err := s.DB.
					Model(&sambaModels.SambaAuditLog{}).
					Where("action = ? AND path = ? AND created_at >= ?", "mkdirat", entry.Path, cutoff).
					Count(&cnt).Error; err != nil {
					return fmt.Errorf("failed to check recent mkdirat: %w", err)
				}
				if cnt > 0 {
					continue
				}
			}

			if err := s.DB.Create(&entry).Error; err != nil {
				return fmt.Errorf("failed to insert audit log entry: %w", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning audit log: %w", err)
	}

	if err := os.Truncate(logPath, 0); err != nil {
		return fmt.Errorf("failed to clear audit log: %w", err)
	}

	return nil
}

func (s *Service) GetAuditLogs(
	page int,
	size int,
	sortField, sortDir string,
) (*sambaServiceInterfaces.AuditLogsResponse, error) {
	if size <= 0 {
		size = 100
	}
	if page <= 0 {
		page = 1
	}

	var total int64
	if err := s.DB.
		Model(&sambaModels.SambaAuditLog{}).
		Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	lastPage := int(math.Ceil(float64(total) / float64(size)))
	allowed := map[string]bool{
		"id":         true,
		"action":     true,
		"share":      true,
		"path":       true,
		"created_at": true,
	}

	field := "id"
	direction := "DESC"
	normalized := strings.ToLower(sortField)

	if normalized == "createdat" {
		normalized = "created_at"
	}

	if allowed[normalized] {
		field = normalized
		dir := strings.ToUpper(sortDir)
		if dir == "ASC" || dir == "DESC" {
			direction = dir
		} else {
			direction = "ASC"
		}
	}

	orderExpr := fmt.Sprintf("%s %s", field, direction)
	offset := (page - 1) * size

	var logs []sambaModels.SambaAuditLog
	if err := s.DB.
		Order(orderExpr).
		Offset(offset).
		Limit(size).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	return &sambaServiceInterfaces.AuditLogsResponse{
		LastPage: lastPage,
		Data:     logs,
	}, nil
}
