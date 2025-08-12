package jail

import (
	"fmt"
	jailModels "sylve/internal/db/models/jail"
)

func (s *Service) GetJailLogs(id uint, start bool) (string, error) {
	var jail jailModels.Jail

	if err := s.DB.First(&jail, "id = ?", id).Error; err != nil {
		return "", fmt.Errorf("failed to find jail with id %d: %w", id, err)
	}

	if start {
		return jail.StartLogs, nil
	}

	return jail.StopLogs, nil
}
