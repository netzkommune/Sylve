package jail

import (
	"fmt"
	"strings"
	"sylve/internal/config"
	jailModels "sylve/internal/db/models/jail"
	"sylve/pkg/utils"
	"time"
)

func (s *Service) JailAction(ctId int, action string) error {
	if action != "start" && action != "stop" {
		return fmt.Errorf("invalid_action: %s", action)
	}

	var flag string
	if action == "start" {
		flag = "-c"
	} else {
		flag = "-r"
	}

	jailsPath, err := config.GetJailsPath()
	if err != nil {
		return fmt.Errorf("failed to get jails path: %w", err)
	}

	jailConf := fmt.Sprintf("%s/%d/%d.conf", jailsPath, ctId, ctId)
	output, err := utils.RunCommand("jail", "-f", jailConf, flag, fmt.Sprintf("%d", ctId))

	if err != nil {
		return fmt.Errorf("failed to %s jail %d: %w", action, ctId, err)
	}

	if action == "start" && !strings.Contains(output, ": created") {
		return fmt.Errorf("unexpected output from jail command: %s", output)
	}

	if action == "stop" && !strings.Contains(output, ": removed") {
		return fmt.Errorf("unexpected output from jail command: %s", output)
	}

	var jail jailModels.Jail
	err = s.DB.First(&jail, "ct_id = ?", ctId).Error
	if err != nil {
		return fmt.Errorf("failed to find jail with ct_id %d: %w", ctId, err)
	}

	now := time.Now().UTC()

	if action == "start" {
		jail.StartedAt = &now
	} else {
		jail.StoppedAt = &now
	}

	err = s.DB.Save(&jail).Error

	if err != nil {
		return fmt.Errorf("failed to update jail status: %w", err)
	}

	return nil
}
