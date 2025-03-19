package disk

import (
	"fmt"
	"os"
	"sylve/pkg/utils"
)

func CheckDevice(device string) error {
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

	return nil
}

func DestroyDisk(device string) error {
	err := CheckDevice(device)

	if err != nil {
		return err
	}

	output, err := utils.RunCommand("gpart", "destroy", "-F", device)
	if err != nil {
		return fmt.Errorf("error destroying disk %s: %v, output: %s", device, err, output)
	}

	return nil
}

func CreatePartition(device string, size uint64) error {
	err := CheckDevice(device)

	if err != nil {
		return err
	}

	fmt.Println("Creating partition on disk", device, "with size", size)

	mbytes := uint64(utils.BytesToSize("MB", float64(size)))
	if mbytes < 1 {
		return fmt.Errorf("size must be at least 1MB")
	}

	output, err := utils.RunCommand("gpart", "add", "-t", "freebsd-zfs", "-s", fmt.Sprintf("%dMB", mbytes), device)
	if err != nil {
		return fmt.Errorf("error creating partition on disk %s: %v, output: %s", device, err, output)
	}

	return nil
}

func CreatePartitions(device string, sizes []uint64) error {
	err := CheckDevice(device)

	if err != nil {
		return err
	}

	totalRequiredSize := uint64(0)

	for _, size := range sizes {
		totalRequiredSize += size
	}

	diskSize, err := GetDiskSize(device)

	if err != nil {
		return fmt.Errorf("failed to get disk size: %v", err)
	}

	if diskSize < totalRequiredSize {
		return fmt.Errorf("disk size is too small for partitions")
	}

	for _, size := range sizes {
		err = CreatePartition(device, size)

		if err != nil {
			return err
		}
	}

	return nil
}

func DeletePartition(device string, partition int) error {
	err := CheckDevice(device)

	if err != nil {
		return err
	}

	output, err := utils.RunCommand("gpart", "delete", "-i", fmt.Sprintf("%d", partition), device)
	if err != nil {
		return fmt.Errorf("error deleting partition %d from disk %s: %v, output: %s", partition, device, err, output)
	}

	return nil
}
