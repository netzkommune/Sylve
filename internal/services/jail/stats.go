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
				JID       int      `json:"jid"`
				Hostname  string   `json:"hostname"`
				Path      string   `json:"path"`
				Name      string   `json:"name"`
				State     string   `json:"state"`
				IPv4Addrs []string `json:"ipv4_addrs"`
				IPv6Addrs []string `json:"ipv6_addrs"`
			} `json:"jail"`
		} `json:"jail-information"`
	}

	if err := json.Unmarshal([]byte(output), &jlsData); err != nil {
		return nil, fmt.Errorf("failed to parse jls JSON: %w", err)
	}

	activeMap := make(map[int]struct{})
	for _, jail := range jlsData.JailInformation.Jail {
		activeMap[jail.JID] = struct{}{}
	}

	var ctIDs []int
	err = s.DB.Model(&jailModels.Jail{}).Pluck("ct_id", &ctIDs).Error
	if err != nil {
		return nil, err
	}

	for _, ctID := range ctIDs {
		state := "INACTIVE"
		var pcpu int64
		var memory int64
		var wallclock int64

		if _, ok := activeMap[ctID]; ok {
			state = "ACTIVE"

			rctlOutput, err := utils.RunCommand("rctl", "-u", fmt.Sprintf("jail:%d", ctID))
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
