package cluster

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/alchemillahq/sylve/internal"
	"github.com/alchemillahq/sylve/internal/config"
	clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"
	"github.com/alchemillahq/sylve/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type basicHealthData struct {
	Hostname string `json:"hostname"`
}

type curInfo struct {
	nodeUUID  string
	api       string
	canonHost string
	rawHost   string
	healthOK  bool
}

func (s *Service) fetchCanonicalHostname(host string, port int) (string, bool) {
	cluster, err := s.GetClusterDetails()
	if err != nil {
		return "", false
	}

	selfHostname, err := utils.GetSystemHostname()
	if err != nil {
		return "", false
	}

	clusterToken, err := s.AuthService.CreateClusterJWT(0, selfHostname, "", "")
	if err != nil {
		return "", false
	}

	url := fmt.Sprintf("https://%s:%d/api/health/basic", host, port)

	body, _, err := utils.HTTPPostJSONRead(
		url,
		map[string]any{"clusterKey": cluster.Cluster.Key},
		map[string]string{
			"Accept":          "application/json",
			"Content-Type":    "application/json",
			"X-Cluster-Token": fmt.Sprintf("Bearer %s", clusterToken),
		},
	)
	if err != nil {
		return "", false
	}

	var resp internal.APIResponse[basicHealthData]
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", false
	}

	if resp.Status == "success" && resp.Data.Hostname != "" {
		return resp.Data.Hostname, true
	}
	return "", false
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

	current := make(map[string]curInfo, len(cfg.Servers))

	for _, server := range cfg.Servers {
		uuid := string(server.ID)
		addr := string(server.Address)

		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}
		api := fmt.Sprintf("%s:%d", host, config.ParsedConfig.Port)
		canon, ok := s.fetchCanonicalHostname(host, config.ParsedConfig.Port)

		current[uuid] = curInfo{
			nodeUUID:  uuid,
			api:       api,
			canonHost: canon,
			rawHost:   host,
			healthOK:  ok,
		}
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var existing []clusterModels.ClusterNode
		if err := tx.Find(&existing).Error; err != nil {
			return err
		}
		exByUUID := make(map[string]clusterModels.ClusterNode, len(existing))
		for _, n := range existing {
			exByUUID[n.NodeUUID] = n
		}

		for _, cur := range current {
			// Decide desired status from health check
			status := "offline"
			if cur.healthOK {
				status = "online"
			}

			// Build insert (for new rows)
			insertRow := clusterModels.ClusterNode{
				NodeUUID: cur.nodeUUID,
				// For new nodes: prefer canonical hostname if known, else use raft host
				Hostname: func() string {
					if cur.canonHost != "" {
						return cur.canonHost
					}
					return cur.rawHost
				}(),
				API:    cur.api,
				Status: status,
			}

			// Build selective updates:
			updates := map[string]any{
				"api":        cur.api,
				"status":     status,
				"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}
			// Only overwrite hostname if we actually resolved it
			if cur.canonHost != "" {
				updates["hostname"] = cur.canonHost
			}

			// Upsert by node_uuid with selective updates
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "node_uuid"}},
				DoUpdates: clause.Assignments(updates),
			}).Create(&insertRow).Error; err != nil {
				return err
			}

			delete(exByUUID, cur.nodeUUID)
		}

		// Anything not in Raft config -> mark offline (removed/evicted)
		if len(exByUUID) > 0 {
			ids := make([]string, 0, len(exByUUID))
			for uuid := range exByUUID {
				ids = append(ids, uuid)
			}
			if err := tx.Model(&clusterModels.ClusterNode{}).
				Where("node_uuid IN ?", ids).
				Updates(map[string]any{
					"status":     "offline",
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"), // <-- add this
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
