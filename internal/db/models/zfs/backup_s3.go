package zfsModels

import "time"

type BackupRunStatus string

const (
	RunPending  BackupRunStatus = "pending"
	RunRunning  BackupRunStatus = "running"
	RunSuccess  BackupRunStatus = "success"
	RunFailed   BackupRunStatus = "failed"
	RunCanceled BackupRunStatus = "canceled"
)

func (ZFSS3BackupRun) TableName() string { return "zfs_s3_backup_runs" }
func (ZFSS3Backup) TableName() string    { return "zfs_s3_backups" }

type ZFSS3Backup struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	ConfigID       uint   `json:"configId"`
	InitialDataset string `json:"initialDataset"`
	Dataset        string `gorm:"not null;index" json:"dataset"`
	Name           string `gorm:"not null;unique" json:"name"`

	// === incremental toggle ===
	Incremental  bool    `gorm:"default:true" json:"incremental"` // if false => always full
	LastSnapshot *string `json:"lastSnapshot"`                    // last successfully sent snapshot (e.g. "pool/ds@2025-09-03T00:00:00Z")

	Recursive bool   `json:"recursive"`
	Prune     bool   `json:"prune"`
	KeepN     int    `json:"keepN"`
	CronExpr  string `json:"cronExpr"`
	Enabled   bool   `gorm:"default:true" json:"enabled"`

	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	LastRunAt     *time.Time `json:"lastRunAt"`
	LastSuccessAt *time.Time `json:"lastSuccessAt"`
}

type ZFSS3BackupRun struct {
	ID uint `gorm:"primaryKey" json:"id"`

	JobID uint        `gorm:"index;not null" json:"jobId"`
	Job   ZFSS3Backup `gorm:"foreignKey:JobID" json:"job"`

	StartedAt  time.Time  `gorm:"index" json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	DurationMs int64      `json:"durationMs"`

	// === full vs incremental for this run ===
	Full         bool    `json:"full"`                   // true => full send
	SnapshotName string  `json:"snapshotName"`           // snapshot we actually sent (e.g. "pool/ds@2025-09-03T12:00:00Z")
	FromSnapshot *string `json:"fromSnapshot,omitempty"` // base snapshot if incremental

	ObjectKey string `json:"objectKey"`
	Size      int64  `json:"size"`
	ETag      string `json:"etag"`

	Status BackupRunStatus `gorm:"type:text;index" json:"status"`
}
