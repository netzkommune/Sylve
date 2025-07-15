package system

import (
	"fmt"
	"sylve/pkg/utils"
)

func ServiceAction(name string, action string) error {
	args := []string{name, action}

	_, err := utils.RunCommand("service", args...)

	if err != nil {
		return fmt.Errorf("failed to %s service %s: %w", action, name, err)
	}

	return nil
}
