package clusterModels

type Cluster struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Enabled       bool   `json:"enabled"`
	Key           string `json:"key"`
	RaftBootstrap *bool  `json:"raftBootstrap"`
	RaftIP        string `json:"raftIP"`
	RaftPort      int    `json:"raftPort"`
}
