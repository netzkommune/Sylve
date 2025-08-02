package sambaServiceInterfaces

import sambaModels "sylve/internal/db/models/samba"

type SambaServiceInterface interface {
	WriteConfig(reload bool) error
	ParseAuditLogs() error
}

type AuditLogsResponse struct {
	LastPage int                         `json:"last_page"`
	Data     []sambaModels.SambaAuditLog `json:"data"`
}
