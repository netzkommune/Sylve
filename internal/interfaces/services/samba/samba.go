package sambaServiceInterfaces

type SambaServiceInterface interface {
	WriteConfig(reload bool) error
	ParseAuditLogs() error
}
