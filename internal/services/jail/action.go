package jail

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sylve/internal/config"
	jailModels "sylve/internal/db/models/jail"
	"sylve/pkg/utils"
	"time"
)

func (s *Service) JailAction(ctId int, action string) error {
	if action != "start" && action != "stop" && action != "restart" {
		return fmt.Errorf("invalid_action: %s", action)
	}

	var flag string
	if action == "start" {
		flag = "-c"
	} else if action == "stop" {
		flag = "-r"
	} else if action == "restart" {
		flag = "-mr"
	}

	jailsPath, err := config.GetJailsPath()
	if err != nil {
		return fmt.Errorf("failed to get jails path: %w", err)
	}

	jailConf := fmt.Sprintf("%s/%d/%d.conf", jailsPath, ctId, ctId)
	ctidHash := utils.HashIntToNLetters(ctId, 5)

	var jail jailModels.Jail
	if err := s.DB.First(&jail, "ct_id = ?", ctId).Error; err != nil {
		return fmt.Errorf("failed to find jail with ct_id %d: %w", ctId, err)
	}

	cmd := exec.Command("jail", "-f", jailConf, flag, ctidHash)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start jail command: %w", err)
	}

	streamToDB := func(r io.Reader, isStart bool) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if isStart {
				jail.StartLogs += line + "\n"
				s.DB.Model(&jail).Update("start_logs", jail.StartLogs)
			} else {
				jail.StopLogs += line + "\n"
				s.DB.Model(&jail).Update("stop_logs", jail.StopLogs)
			}
		}
	}

	isStart := (action == "start")

	if isStart {
		jail.StartLogs = ""
	} else if action == "stop" {
		jail.StopLogs = ""
	} else if action == "restart" {
		jail.StartLogs = ""
		jail.StopLogs = ""
	}

	if err := s.DB.Save(&jail).Error; err != nil {
		return fmt.Errorf("failed to reset logs: %w", err)
	}

	go streamToDB(stdout, isStart)
	go streamToDB(stderr, isStart)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to %s jail %d: %w", action, ctId, err)
	}

	now := time.Now().UTC()
	if action == "start" {
		jail.StartedAt = &now
		jail.StartLogs = ""
	} else if action == "stop" {
		jail.StoppedAt = &now
		jail.StopLogs = ""
	}

	if err := s.DB.Save(&jail).Error; err != nil {
		return fmt.Errorf("failed to update jail status: %w", err)
	}

	return nil
}
