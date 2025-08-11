package jail

import (
	"fmt"
	"strings"
	jailModels "sylve/internal/db/models/jail"
	"sylve/pkg/utils"
)

func (s *Service) UpdateMemory(ctId uint, memoryBytes int64) error {
	if memoryBytes < 0 {
		return fmt.Errorf("invalid memory value: %d", memoryBytes)
	}

	const MiB = int64(1024 * 1024)
	mb := (memoryBytes + MiB - 1) / MiB
	if mb < 1 {
		return fmt.Errorf("memory must be at least 1MB, got: %dMB", mb)
	}

	cfg, err := s.GetJailConfig(ctId)
	if err != nil {
		return err
	}

	if strings.TrimSpace(cfg) == "" {
		return fmt.Errorf("jail config not found for CTID: %d", ctId)
	}

	ctIdHash := utils.HashIntToNLetters(int(ctId), 5)
	prefix := fmt.Sprintf(`exec.poststart += "rctl -a jail:%s:memoryuse:deny=`, ctIdHash)

	lines := strings.Split(cfg, "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			lines[i] = fmt.Sprintf(`	exec.poststart += "rctl -a jail:%s:memoryuse:deny=%dM";`, ctIdHash, mb)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf(`exec.poststart += "rctl -a jail:%s:memoryuse:deny=%dM";`, ctIdHash, mb))
	}

	newCfg := strings.Join(lines, "\n")
	if err := s.SaveJailConfig(ctId, newCfg); err != nil {
		return fmt.Errorf("failed to save jail config: %w", err)
	}

	var jail jailModels.Jail
	if err := s.DB.Find(&jail, "ct_id = ?", ctId).Error; err != nil {
		return fmt.Errorf("failed to find jail with CTID %d: %w", ctId, err)
	}

	jail.Memory = int(memoryBytes)

	if err := s.DB.Save(&jail).Error; err != nil {
		return fmt.Errorf("failed to update jail memory in database: %w", err)
	}

	_, err = utils.RunCommand("rctl", "-a", fmt.Sprintf("jail:%s:memoryuse:deny=%dM", ctIdHash, mb))

	if err != nil {
		return fmt.Errorf("failed to apply memory limit with rctl: %w", err)
	}

	return nil
}
