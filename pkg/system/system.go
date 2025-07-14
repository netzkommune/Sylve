package system

import (
	"strings"
	"sylve/pkg/utils"
)

func UnixUserExists(name string) (bool, error) {
	output, err := utils.RunCommand("id", name)

	if err != nil {
		lowerOutput := strings.ToLower(output)
		if strings.Contains(lowerOutput, "no such user") || strings.Contains(lowerOutput, "does not exist") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func CreateUnixUser(name string, shell string, dir string) error {
	args := []string{"useradd", name, "-m"}

	if shell != "" {
		args = append(args, "-s", shell)
	} else {
		args = append(args, "-s", "/usr/sbin/nologin")
	}

	if dir != "" {
		args = append(args, "-d", dir)
	} else {
		args = append(args, "-d", "/nonexistent")
	}

	_, err := utils.RunCommand("pw", args...)

	if err != nil {
		return err
	}

	return nil
}
