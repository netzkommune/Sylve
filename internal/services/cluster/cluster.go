package cluster

import (
	"errors"
	"fmt"

	"github.com/alchemillahq/sylve/internal/config"
	clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"
	clusterServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/cluster"
	"github.com/alchemillahq/sylve/pkg/network"
	"github.com/alchemillahq/sylve/pkg/utils"
	"github.com/hashicorp/raft"
	"gorm.io/gorm"
)

var _ clusterServiceInterfaces.ClusterServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	Raft      *raft.Raft
	Transport *raft.NetworkTransport
}

func NewClusterService(db *gorm.DB) clusterServiceInterfaces.ClusterServiceInterface {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetClusterDetails() (*clusterServiceInterfaces.ClusterDetails, error) {
	out := &clusterServiceInterfaces.ClusterDetails{
		Cluster:  nil,
		Nodes:    []clusterServiceInterfaces.RaftNode{},
		LeaderID: "",
		Partial:  false,
	}

	var c clusterModels.Cluster
	if err := s.DB.First(&c).Error; err == nil {
		out.Cluster = &c
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	detail := s.Detail()
	if detail == nil {
		return out, fmt.Errorf("failed to get cluster detail")
	}

	out.NodeID = detail.NodeID

	if s.Raft == nil {
		return out, nil
	}

	leaderAddr, leaderID := s.Raft.LeaderWithID()
	out.LeaderID = string(leaderID)
	out.LeaderAddress = string(leaderAddr)

	fut := s.Raft.GetConfiguration()
	if err := fut.Error(); err != nil {
		out.Partial = true
		return out, nil
	}
	conf := fut.Configuration()

	suffrageStr := func(sf raft.ServerSuffrage) string {
		switch sf {
		case raft.Voter:
			return "voter"
		case raft.Nonvoter:
			return "nonvoter"
		case raft.Staging:
			return "staging"
		default:
			return "unknown"
		}
	}

	for _, srv := range conf.Servers {
		id := string(srv.ID)
		addr := string(srv.Address)

		out.Nodes = append(out.Nodes, clusterServiceInterfaces.RaftNode{
			ID:       id,
			Address:  addr,
			Suffrage: suffrageStr(srv.Suffrage),
			IsLeader: id == string(leaderID) || addr == string(leaderAddr),
		})
	}

	return out, nil
}

func (s *Service) CreateCluster(ip string, port int, fsm raft.FSM) error {
	if s.Raft != nil && s.Raft.State() != raft.Shutdown {
		return errors.New("raft_already_initialized")
	}

	if err := network.TryBindToPort(ip, port, "tcp"); err != nil {
		return err
	}

	if dir, _ := config.GetRaftPath(); hasExistingRaftState(dir) {
		return errors.New("raft_state_already_exists")
	}

	var c clusterModels.Cluster
	if err := s.DB.First(&c).Error; err != nil {
		return err
	}

	if c.Enabled {
		return errors.New("cluster already exists")
	}

	bootstrap := true
	newKey := c.Key
	if newKey == "" {
		newKey = utils.GenerateRandomString(32)
	}

	if err := s.DB.Model(&c).Updates(map[string]any{
		"enabled":        true,
		"key":            newKey,
		"raft_bootstrap": &bootstrap,
		"raft_ip":        ip,
		"raft_port":      port,
	}).Error; err != nil {
		return err
	}

	if _, err := s.SetupRaft(true, fsm); err != nil {
		return err
	}

	c.Enabled = true
	c.Key = newKey
	c.RaftBootstrap = &bootstrap
	c.RaftIP = ip
	c.RaftPort = port

	return nil
}

func (s *Service) StartAsJoiner(fsm raft.FSM, ip string, port int, clusterKey string) error {
	if !utils.IsValidIP(ip) {
		return errors.New("invalid_ip_address")
	}

	if !utils.IsValidPort(port) {
		return errors.New("invalid_port_number")
	}

	if err := network.TryBindToPort(ip, port, "tcp"); err != nil {
		return fmt.Errorf("failed_to_bind_to_port: %v", err)
	}

	details, err := s.GetClusterDetails()
	if err != nil {
		return err
	}

	if details.Cluster.Enabled {
		return fmt.Errorf("clustered_already")
	}

	if s.Raft != nil && s.Raft.State() != raft.Shutdown {
		return errors.New("raft_already_initialized")
	}

	err = s.CleanRaftDir()
	if err != nil {
		return err
	}

	var c clusterModels.Cluster
	if err := s.DB.First(&c).Error; err != nil {
		return err
	}

	c.RaftIP = ip
	c.RaftPort = port
	c.Enabled = true
	c.Key = clusterKey

	if err := s.DB.Save(&c).Error; err != nil {
		return err
	}

	_, err = s.SetupRaft(false, fsm)
	if err != nil {
		c.RaftIP = ""
		c.RaftPort = 0
		c.Enabled = false
		c.Key = ""

		if err := s.DB.Save(&c).Error; err != nil {
			return err
		}

		return err
	}

	return nil
}

func (s *Service) AcceptJoin(nodeID, nodeIp string, nodePort int, providedKey string) error {
	details, err := s.GetClusterDetails()
	if err != nil {
		return err
	}

	if details.Cluster == nil {
		return errors.New("cluster_not_found")
	}

	if details.Cluster.Key != providedKey {
		return errors.New("invalid_cluster_key")
	}

	if s.Raft == nil {
		return errors.New("raft_not_initialized")
	}

	if s.Raft.State() != raft.Leader {
		addr, id := s.Raft.LeaderWithID()
		return fmt.Errorf("not_leader; leader_addr=%s; leader_id=%s", string(addr), string(id))
	}

	fut := s.Raft.GetConfiguration()
	if err := fut.Error(); err != nil {
		return fmt.Errorf("get_config_failed: %w", err)
	}

	conf := fut.Configuration()
	sid := raft.ServerID(nodeID)
	saddr := raft.ServerAddress(fmt.Sprintf("%s:%d", nodeIp, nodePort))

	for _, srv := range conf.Servers {
		if srv.ID == sid {
			if srv.Address == saddr && srv.Suffrage == raft.Voter {
				return nil
			}

			rf := s.Raft.RemoveServer(srv.ID, 0, 0)
			if err := rf.Error(); err != nil {
				return fmt.Errorf("remove_existing_failed: %w", err)
			}
			break
		}
		if srv.Address == saddr && srv.ID != sid {
			rf := s.Raft.RemoveServer(srv.ID, 0, 0)
			if err := rf.Error(); err != nil {
				return fmt.Errorf("remove_conflicting_addr_failed: %w", err)
			}
		}
	}

	af := s.Raft.AddVoter(sid, saddr, 0, 0)
	if err := af.Error(); err != nil {
		return fmt.Errorf("add_voter_failed: %w", err)
	}

	return nil
}

func (s *Service) MarkClustered() error {
	var c clusterModels.Cluster
	if err := s.DB.First(&c).Error; err != nil {
		return err
	}

	c.Enabled = true
	if err := s.DB.Save(&c).Error; err != nil {
		return err
	}

	return nil
}
