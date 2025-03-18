package disk

import (
	"fmt"
	"os"
	"sylve/pkg/utils"
)

func DestroyDisk(device string) error {
	info, err := os.Stat(device)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("device %s does not exist", device)
		}
		return fmt.Errorf("error accessing device %s: %v", device, err)
	}

	if info.Mode()&os.ModeDevice == 0 {
		return fmt.Errorf("%s exists but is not a block device", device)
	}

	output, err := utils.RunCommand("gpart", "destroy", "-F", device)
	if err != nil {
		return fmt.Errorf("error destroying disk %s: %v, output: %s", device, err, output)
	}

	return nil
}
