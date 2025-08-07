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

	// Make a map of running jail names
	activeMap := make(map[string]struct{})
	for _, jail := range jlsData.JailInformation.Jail {
		activeMap[jail.Name] = struct{}{}
	}

	// Fetch all ctids from DB
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
