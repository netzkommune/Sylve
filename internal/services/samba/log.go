package samba

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	sambaModels "sylve/internal/db/models/samba"
	"time"
)

func (s *Service) ParseAuditLogs() error {
	const logPath = "/var/log/samba4/audit.log"

	// which ops we care about
	validActions := map[string]bool{
		"connect":     true,
		"disconnect":  true,
		"mkdirat":     true,
		"unlinkat":    true,
		"renameat":    true,
		"create_file": true,
	}

	// dedupe maps
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
			// only log the real create, not every open
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

			// ——— SKIP create_file if a mkdirat on the same path happened ≤5s ago ———
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
					// skip this create_file entry
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

	// clear so next run only sees new lines
	if err := os.Truncate(logPath, 0); err != nil {
		return fmt.Errorf("failed to clear audit log: %w", err)
	}
	return nil
}
