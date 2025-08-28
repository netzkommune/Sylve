package cluster

import (
	"errors"
	"fmt"
	"net"

	"github.com/alchemillahq/sylve/internal/config"
	clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"
	"gorm.io/gorm"
)

func (s *Service) PopulateClusterNodes() error {
	var c clusterModels.Cluster
	if err := s.DB.First(&c).Error; err != nil {
		return err
	}

	if !c.Enabled {
		return nil
	}

	if s.Raft == nil {
		return fmt.Errorf("raft_not_initialized")
	}

	fut := s.Raft.GetConfiguration()
	if err := fut.Error(); err != nil {
		return fmt.Errorf("failed_to_get_raft_configuration: %w", err)
	}
	cfg := fut.Configuration()

	current := make(map[string]string)
	for _, server := range cfg.Servers {
		addr := string(server.Address)

		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}

		current[host] = fmt.Sprintf("%s:%d", host, config.ParsedConfig.Port)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var existing []clusterModels.ClusterNode
		if err := tx.Find(&existing).Error; err != nil {
			return err
		}

		exByHost := make(map[string]clusterModels.ClusterNode, len(existing))
		for _, n := range existing {
			exByHost[n.Hostname] = n
		}

		for host, api := range current {
			if _, ok := exByHost[host]; ok {
				if err := tx.Model(&clusterModels.ClusterNode{}).
					Where("hostname = ?", host).
					Updates(map[string]any{"api": api, "status": "online"}).Error; err != nil {
					return err
				}
				delete(exByHost, host)
			} else {
				n := clusterModels.ClusterNode{Hostname: host, API: api, Status: "online"}
				if err := tx.Create(&n).Error; err != nil {
					if !errors.Is(err, gorm.ErrDuplicatedKey) {
						return err
					}
					if err := tx.Model(&clusterModels.ClusterNode{}).
						Where("hostname = ?", host).
						Updates(map[string]any{"api": api, "status": "online"}).Error; err != nil {
						return err
					}
				}
			}
		}

		if len(exByHost) > 0 {
			hosts := make([]string, 0, len(exByHost))
			for h := range exByHost {
				hosts = append(hosts, h)
			}
			if err := tx.Model(&clusterModels.ClusterNode{}).
				Where("hostname IN ?", hosts).
				Update("status", "offline").Error; err != nil {
				return err
			}
		}

		return nil
	})
}
