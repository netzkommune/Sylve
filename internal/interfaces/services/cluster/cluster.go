package clusterServiceInterfaces

import (
	clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"
	"github.com/hashicorp/raft"
)

type Detail struct {
	NodeID   string `json:"nodeId"`
	Hostname string `json:"hostname"`
	APIPort  int    `json:"apiPort"`
}

type RaftNode struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Suffrage string `json:"suffrage"`
	IsLeader bool   `json:"isLeader"`
}

type ClusterDetails struct {
	Cluster       *clusterModels.Cluster `json:"cluster"`
	NodeID        string                 `json:"nodeId"`
	Nodes         []RaftNode             `json:"nodes"`
	LeaderID      string                 `json:"leaderId"`
	LeaderAddress string                 `json:"leaderAddress"`
	Partial       bool                   `json:"partial"`
}

type ClusterServiceInterface interface {
	Detail() *Detail
	InitRaft(fsm raft.FSM) error
	CreateCluster(ip string, port int, fsm raft.FSM) error
	SetupRaft(bootstrap bool, fsm raft.FSM) (*raft.Raft, error)
	GetClusterDetails() (*ClusterDetails, error)
	PopulateClusterNodes() error
}
