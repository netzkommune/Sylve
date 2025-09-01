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

	cpu      int
	cpuUsage float64

	memory   uint64  // bytes
	memUsage float64 // %

	disk      uint64  // bytes
	diskUsage float64 // %
}

// ---------- CPU ----------

type CPUInfo struct {
	Name           string   `json:"name"`
	PhysicalCores  int16    `json:"physicalCores"`
	ThreadsPerCore int16    `json:"threadsPerCore"`
	LogicalCores   int16    `json:"logicalCores"`
	Family         int16    `json:"family"`
	Model          int16    `json:"model"`
	Features       []string `json:"features"`
	CacheLine      int16    `json:"cacheLine"`
	Cache          struct {
		L1D int16 `json:"l1d"`
		L1I int16 `json:"l1i"`
		L2  int16 `json:"l2"`
		L3  int16 `json:"l3"`
	} `json:"cache"`
	Frequency int64   `json:"frequency"`
	Usage     float64 `json:"usage"`
}

func (s *Service) fetchCPUInfo(host string, port int, clusterToken, clusterKey string) (int, float64, bool) {
	url := fmt.Sprintf("https://%s:%d/api/info/cpu", host, port)

	body, _, err := utils.HTTPGetJSONRead(
		url,
		map[string]string{
			"Accept":          "application/json",
			"Content-Type":    "application/json",
			"X-Cluster-Token": fmt.Sprintf("Bearer %s", clusterToken),
		},
	)
	if err != nil {
		return 0, 0, false
	}

	var resp internal.APIResponse[CPUInfo]
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, 0, false
	}
	if resp.Status != "success" {
		return 0, 0, false
	}

	// Store physical cores; switch to logical if you prefer.
	cores := int(resp.Data.PhysicalCores)
	usage := resp.Data.Usage
	return cores, usage, true
}

// ---------- RAM ----------

type RAMInfo struct {
	Total       uint64  `json:"total"`       // bytes
	Free        uint64  `json:"free"`        // bytes
	UsedPercent float64 `json:"usedPercent"` // 0..100
}

func (s *Service) fetchRAMInfo(host string, port int, clusterToken, clusterKey string) (uint64, float64, bool) {
	url := fmt.Sprintf("https://%s:%d/api/info/ram", host, port)

	body, _, err := utils.HTTPGetJSONRead(
		url,
		map[string]string{
			"Accept":          "application/json",
			"Content-Type":    "application/json",
			"X-Cluster-Token": fmt.Sprintf("Bearer %s", clusterToken),
		},
	)
	if err != nil {
		return 0, 0, false
	}

	var resp internal.APIResponse[RAMInfo]
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, 0, false
	}
	if resp.Status != "success" {
		return 0, 0, false
	}

	// Store bytes directly; UI can format (MB/GiB).
	return resp.Data.Total, resp.Data.UsedPercent, true
}

// ---------- DISK/ZFS ----------

// Assumption: endpoint returns bytes for Total and Usage (used bytes).
// If your endpoint returns different semantics, adjust here centrally.
type PoolDisksUsageResponse struct {
	Total float64 `json:"total"` // bytes
	Usage float64 `json:"usage"` // used bytes
}

func (s *Service) fetchDiskInfo(host string, port int, clusterToken, clusterKey string) (uint64, float64, bool) {
	url := fmt.Sprintf("https://%s:%d/api/zfs/pools/disks-usage", host, port)

	body, _, err := utils.HTTPGetJSONRead(
		url,
		map[string]string{
			"Accept":          "application/json",
			"Content-Type":    "application/json",
			"X-Cluster-Token": fmt.Sprintf("Bearer %s", clusterToken),
		},
	)

	if err != nil {
		return 0, 0, false
	}

	var resp internal.APIResponse[PoolDisksUsageResponse]
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, 0, false
	}

	if resp.Status != "success" {
		return 0, 0, false
	}

	total := uint64(resp.Data.Total)
	used := uint64(resp.Data.Usage)
	pct := float64(0)
	if total > 0 {
		pct = (float64(used) / float64(total)) * 100.0
	}

	return uint64(resp.Data.Total), pct, true
}

// ---------- Health + combined fetch ----------

func (s *Service) fetchCanonicalHostnameAndCPU(host string, port int, clusterToken, clusterKey, selfHostname string) (string, bool, int, float64, bool) {
	if utils.IsLocalIP(host) {
		hostname := selfHostname
		cpuCores, usage, okCPU := s.fetchCPUInfo(host, port, clusterToken, clusterKey)
		return hostname, true, cpuCores, usage, okCPU
	}
	canon, ok := s.fetchCanonicalHostnameWithToken(host, port, clusterToken, clusterKey)
	cpuCores, usage, okCPU := s.fetchCPUInfo(host, port, clusterToken, clusterKey)
	return canon, ok, cpuCores, usage, okCPU
}

func (s *Service) getClusterToken(hostname string) (string, error) {
	return s.AuthService.CreateClusterJWT(0, hostname, "", "")
}

func (s *Service) fetchCanonicalHostnameWithToken(host string, port int, clusterToken, clusterKey string) (string, bool) {
	url := fmt.Sprintf("https://%s:%d/api/health/basic", host, port)

	body, _, err := utils.HTTPPostJSONRead(
		url,
		map[string]any{"clusterKey": clusterKey},
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

// ---------- Main ----------

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

	selfHostname, err := utils.GetSystemHostname()
	if err != nil {
		return err
	}
	clusterToken, err := s.getClusterToken(selfHostname)
	if err != nil {
		return err
	}
	clusterDetails, err := s.GetClusterDetails()
	if err != nil {
		return err
	}
	clusterKey := clusterDetails.Cluster.Key

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

		canon, okHealth, cores, cpuUsage, okCPU :=
			s.fetchCanonicalHostnameAndCPU(host, config.ParsedConfig.Port, clusterToken, clusterKey, selfHostname)

		// RAM (bytes + %)
		memBytes, memUsedPct, okRAM := s.fetchRAMInfo(host, config.ParsedConfig.Port, clusterToken, clusterKey)

		// DISK (bytes + computed %)
		diskBytes, diskUsedPct, okDisk := s.fetchDiskInfo(host, config.ParsedConfig.Port, clusterToken, clusterKey)

		ci := curInfo{
			nodeUUID:  uuid,
			api:       api,
			canonHost: canon,
			rawHost:   host,
			healthOK:  okHealth,
		}
		if okCPU {
			ci.cpu = cores
			ci.cpuUsage = cpuUsage
		}
		if okRAM {
			ci.memory = memBytes
			ci.memUsage = memUsedPct
		}
		if okDisk {
			ci.disk = diskBytes
			ci.diskUsage = diskUsedPct
		}

		current[uuid] = ci
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
			status := "offline"
			if cur.healthOK {
				status = "online"
			}

			insertRow := clusterModels.ClusterNode{
				NodeUUID: cur.nodeUUID,
				Hostname: func() string {
					if cur.canonHost != "" {
						return cur.canonHost
					}
					return cur.rawHost
				}(),
				API:         cur.api,
				Status:      status,
				CPU:         cur.cpu,       // 0 if unknown
				CPUUsage:    cur.cpuUsage,  // 0 if unknown
				Memory:      cur.memory,    // bytes; 0 if unknown
				MemoryUsage: cur.memUsage,  // %
				Disk:        cur.disk,      // bytes; 0 if unknown
				DiskUsage:   cur.diskUsage, // %
			}

			updates := map[string]any{
				"api":        cur.api,
				"status":     status,
				"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}
			if cur.canonHost != "" {
				updates["hostname"] = cur.canonHost
			}
			// CPU fields
			if cur.cpu > 0 {
				updates["cpu"] = cur.cpu
			}
			if cur.cpu > 0 || cur.cpuUsage > 0 {
				updates["cpu_usage"] = cur.cpuUsage
			}
			// RAM fields (write only if fetched)
			if cur.memory > 0 {
				updates["memory"] = cur.memory
			}
			if cur.memory > 0 || cur.memUsage > 0 {
				updates["memory_usage"] = cur.memUsage
			}
			// DISK fields
			if cur.disk > 0 {
				updates["disk"] = cur.disk
			}
			if cur.disk > 0 || cur.diskUsage > 0 {
				updates["disk_usage"] = cur.diskUsage
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "node_uuid"}},
				DoUpdates: clause.Assignments(updates),
			}).Create(&insertRow).Error; err != nil {
				return err
			}

			delete(exByUUID, cur.nodeUUID)
		}

		// Mark nodes not in RAFT config as offline
		if len(exByUUID) > 0 {
			ids := make([]string, 0, len(exByUUID))
			for uuid := range exByUUID {
				ids = append(ids, uuid)
			}
			if err := tx.Model(&clusterModels.ClusterNode{}).
				Where("node_uuid IN ?", ids).
				Updates(map[string]any{
					"status":     "offline",
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
