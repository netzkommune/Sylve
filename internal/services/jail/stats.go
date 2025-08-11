package jail

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	jailModels "sylve/internal/db/models/jail"
	jailServiceInterfaces "sylve/internal/interfaces/services/jail"
	"sylve/pkg/utils"
)

func (s *Service) GetStates() ([]jailServiceInterfaces.State, error) {
	var states []jailServiceInterfaces.State

	output, err := utils.RunCommand("jls", "-v", "--libxo", "json")
	if err != nil {
		return states, err
	}

	var jlsData struct {
		JailInformation struct {
			Jail []struct {
				JID   int    `json:"jid"`
				Name  string `json:"name"`
				State string `json:"state"`
			} `json:"jail"`
		} `json:"jail-information"`
	}

	if err := json.Unmarshal([]byte(output), &jlsData); err != nil {
		return nil, fmt.Errorf("failed to parse jls JSON: %w", err)
	}

	activeMap := make(map[string]struct{})
	for _, jail := range jlsData.JailInformation.Jail {
		activeMap[jail.Name] = struct{}{}
	}

	var ctIDs []int
	err = s.DB.Model(&jailModels.Jail{}).Pluck("ct_id", &ctIDs).Error
	if err != nil {
		return nil, err
	}

	for _, ctID := range ctIDs {
		name := utils.HashIntToNLetters(ctID, 5)
		state := "INACTIVE"
		var pcpu, memory, wallclock int64

		if _, ok := activeMap[name]; ok {
			state = "ACTIVE"
			rctlOutput, err := utils.RunCommand("rctl", "-u", fmt.Sprintf("jail:%s", name))
			if err == nil {
				scanner := bufio.NewScanner(strings.NewReader(rctlOutput))
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, "pcpu=") {
						val := strings.TrimPrefix(line, "pcpu=")
						if f, err := strconv.ParseFloat(val, 64); err == nil {
							pcpu = int64(math.Round(f))
						}
					} else if strings.HasPrefix(line, "memoryuse=") {
						val := strings.TrimPrefix(line, "memoryuse=")
						if b, err := strconv.ParseInt(val, 10, 64); err == nil {
							memory = b
						}
					} else if strings.HasPrefix(line, "wallclock=") {
						val := strings.TrimPrefix(line, "wallclock=")
						if sec, err := strconv.ParseInt(val, 10, 64); err == nil {
							wallclock = sec
						}
					}
				}
			}
		}

		states = append(states, jailServiceInterfaces.State{
			CTID:      ctID,
			State:     state,
			PCPU:      pcpu,
			Memory:    memory,
			WallClock: wallclock,
		})
	}

	return states, nil
}

func (s *Service) StoreJailUsage() error {
	if !s.crudMutex.TryLock() {
		return nil
	}
	defer s.crudMutex.Unlock()

	var jails []jailModels.Jail

	if err := s.DB.Select("id, ct_id, memory").Find(&jails).Error; err != nil {
		return fmt.Errorf("failed_to_load_jails: %w", err)
	}

	if len(jails) == 0 {
		return nil
	}

	states, err := s.GetStates()
	if err != nil {
		return fmt.Errorf("failed_to_get_jail_states: %w", err)
	}

	type sInfo struct {
		CPUPercent   int64
		MemBytesUsed int64
		Active       bool
	}

	stateByCTID := make(map[int]sInfo, len(states))
	for _, st := range states {
		stateByCTID[st.CTID] = sInfo{
			CPUPercent:   st.PCPU,
			MemBytesUsed: st.Memory,
			Active:       st.State == "ACTIVE",
		}
	}

	for _, j := range jails {
		live, ok := stateByCTID[j.CTID]
		if !ok || !live.Active {
			continue
		}

		cpuPct := live.CPUPercent

		var memPct int64
		if j.Memory > 0 {
			memPct = int64(math.Round((float64(live.MemBytesUsed) / float64(j.Memory)) * 100.0))
			if memPct < 0 {
				memPct = 0
			} else if memPct > 100 {
				memPct = 100
			}
		}

		stat := &jailModels.JailStats{
			CTID:        int(j.ID),
			CPUUsage:    cpuPct,
			MemoryUsage: memPct,
		}
		if err := s.DB.Save(stat).Error; err != nil {
			continue
		}
	}

	jDBIDs := make([]uint, 0, len(jails))
	for _, j := range jails {
		jDBIDs = append(jDBIDs, j.ID)
	}

	for _, dbID := range jDBIDs {
		var stats []jailModels.JailStats
		if err := s.DB.
			Where("ct_id = ?", dbID).
			Order("id DESC").
			Limit(256).
			Find(&stats).Error; err != nil {
			return fmt.Errorf("failed_to_get_jail_stats: %w", err)
		}
		if len(stats) < 256 {
			continue
		}
		cutoff := stats[len(stats)-1].ID
		if err := s.DB.
			Where("ct_id = ? AND id < ?", dbID, cutoff).
			Delete(&jailModels.JailStats{}).Error; err != nil {
			return fmt.Errorf("failed_to_delete_old_jail_stats: %w", err)
		}
	}

	if err := s.pruneOrphanedJailStats(jDBIDs); err != nil {
		return err
	}

	return nil
}

func (s *Service) pruneOrphanedJailStats(validJailDBIDs []uint) error {
	if len(validJailDBIDs) == 0 {
		return s.DB.Where("1 = 1").Delete(&jailModels.JailStats{}).Error
	}

	valid := make([]int, len(validJailDBIDs))
	for i, id := range validJailDBIDs {
		valid[i] = int(id)
	}

	if err := s.DB.
		Where("ct_id NOT IN ?", valid).
		Delete(&jailModels.JailStats{}).Error; err != nil {
		return fmt.Errorf("failed_to_prune_orphaned_jail_stats: %w", err)
	}
	return nil
}

func (s *Service) GetJailUsage(ctId int, limit int) ([]jailModels.JailStats, error) {
	var jailDbId uint
	if err := s.DB.Model(&jailModels.Jail{}).
		Where("ct_id = ?", ctId).
		Select("id").
		First(&jailDbId).Error; err != nil {
		return nil, fmt.Errorf("failed_to_get_actual_jail_id: %w", err)
	}

	if jailDbId == 0 {
		return nil, fmt.Errorf("jail_not_found")
	}

	var jailStats []jailModels.JailStats
	sub := s.DB.
		Model(&jailModels.JailStats{}).
		Where("ct_id = ?", jailDbId).
		Order("id DESC").
		Limit(limit)

	if err := s.DB.Table("(?) as sub", sub).
		Order("id ASC").
		Find(&jailStats).Error; err != nil {
		return nil, fmt.Errorf("failed_to_get_jail_usage: %w", err)
	}

	return jailStats, nil
}
