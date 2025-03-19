package disk

import (
	"strings"
	"sylve/pkg/utils"
)

func GetDiskSize(device string) (uint64, error) {
	out, err := utils.RunCommand("diskinfo", "-v", device)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(out, "\n")

	for _, line := range lines {
		if strings.Contains(line, "mediasize in bytes") {
			size := strings.Fields(line)[0]
			return utils.StringToUint64(size), nil
		}
	}

	return 0, nil
}
