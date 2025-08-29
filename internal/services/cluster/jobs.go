package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/alchemillahq/sylve/internal"
	"github.com/alchemillahq/sylve/internal/config"
	clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"
	"github.com/alchemillahq/sylve/pkg/utils"
	"gorm.io/gorm"
)

type basicHealthData struct {
	Hostname string `json:"hostname"`
}

func (s *Service) fetchCanonicalHostname(host string, port int) (string, error) {
	cluster, err := s.GetClusterDetails()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster details: %w", err)
	}

	hostname, err := utils.GetSystemHostname()
	if err != nil {
		return "", fmt.Errorf("failed to get system hostname: %w", err)
	}

	url := fmt.Sprintf("https://%s:%d/api/health/basic", host, port)
	clusterToken, err := s.AuthService.CreateClusterJWT(0, hostname, "", "cluster-token")
	if err != nil {
		return "", fmt.Errorf("failed to create cluster JWT: %w", err)
	}

	b, _, err := utils.HTTPPostJSONRead(url, map[string]any{
		"clusterKey": cluster.Cluster.Key,
	}, map[string]string{
		"Accept":          "application/json",
		"X-Cluster-Token": fmt.Sprintf("Bearer %s", clusterToken),
	})

	if err != nil {
		return "", fmt.Errorf("https request failed: %w", err)
	}

	var resp internal.APIResponse[basicHealthData]
	if err := json.Unmarshal(b, &resp); err != nil {
		return "", fmt.Errorf("unmarshal failed: %w", err)
	}
	if resp.Status == "success" && resp.Data.Hostname != "" {
		return resp.Data.Hostname, nil
	}

	return "", fmt.Errorf("hostname not returned by %s", url)
}

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

	current := make(map[string]struct {
		api      string
		hostname string
	})

	for _, server := range cfg.Servers {
		addr := string(server.Address)

		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}

		api := fmt.Sprintf("%s:%d", host, config.ParsedConfig.Port)
		canon, err := s.fetchCanonicalHostname(host, config.ParsedConfig.Port)

		if err != nil {
			canon = host
		}

		current[host] = struct {
			api      string
			hostname string
		}{
			api:      api,
			hostname: canon,
		}
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

		for raftHost, cur := range current {
			hostKey := cur.hostname
			if hostKey == "" {
				hostKey = raftHost
			}

			if _, ok := exByHost[hostKey]; ok {
				if err := tx.Model(&clusterModels.ClusterNode{}).
					Where("hostname = ?", hostKey).
					Updates(map[string]any{
						"api":    cur.api,
						"status": "online",
					}).Error; err != nil {
					return err
				}
				delete(exByHost, hostKey)
			} else {
				n := clusterModels.ClusterNode{
					Hostname: hostKey,
					API:      cur.api,
					Status:   "online",
				}
				if err := tx.Create(&n).Error; err != nil {
					if !errors.Is(err, gorm.ErrDuplicatedKey) {
						return err
					}
					if err := tx.Model(&clusterModels.ClusterNode{}).
						Where("hostname = ?", hostKey).
						Updates(map[string]any{
							"api":    cur.api,
							"status": "online",
						}).Error; err != nil {
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
