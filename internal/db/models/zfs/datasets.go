package zfsModels

type Dataset struct {
	ID            int64  `json:"id" gorm:"primaryKey"`
	Name          string `json:"name"`
	Origin        string `json:"origin"`
	Used          uint64 `json:"used"`
	Avail         uint64 `json:"avail"`
	Mountpoint    string `json:"mountpoint"`
	Compression   string `json:"compression"`
	Type          string `json:"type"`
	Written       uint64 `json:"written"`
	Volsize       uint64 `json:"volsize"`
	Logicalused   uint64 `json:"logicalused"`
	Usedbydataset uint64 `json:"usedbydataset"`
	Quota         uint64 `json:"quota"`
	Referenced    uint64 `json:"referenced"`

	SnapshotJobs []SnapshotJob `json:"snapshotJobs" gorm:"foreignKey:DatasetID;constraint:OnDelete:CASCADE;"`
}

type SnapshotJob struct {
	ID        int64   `json:"id" gorm:"primaryKey"`
	DatasetID int64   `json:"datasetId"`
	Dataset   Dataset `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Interval  uint64  `json:"interval"`
}
