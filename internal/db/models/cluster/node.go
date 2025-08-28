package clusterModels

type ClusterNode struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Status   string `json:"status"`
	Hostname string `json:"hostname"`
	API      string `json:"api"`
}
