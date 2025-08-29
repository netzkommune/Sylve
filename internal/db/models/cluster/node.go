package clusterModels

import "time"

type ClusterNode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeUUID  string    `json:"nodeUUID" gorm:"column:node_uuid;uniqueIndex;not null"`
	Status    string    `json:"status"`
	Hostname  string    `json:"hostname"`
	API       string    `json:"api"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
