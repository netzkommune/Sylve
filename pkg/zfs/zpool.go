package zfs

import (
	"fmt"
	"strconv"
	"strings"
	"sylve/pkg/utils"
)

type RW struct {
	Read  uint64 `json:"read"`
	Write uint64 `json:"write"`
}

type VdevDevice struct {
	Name   string `json:"name"`
	Size   uint64 `json:"size"`
	Health string `json:"health"`
}

type ReplacingDevice struct {
	Name     string     `json:"name"`
	Health   string     `json:"health"`
	OldDrive VdevDevice `json:"oldDrive"`
	NewDrive VdevDevice `json:"newDrive"`
}

type Vdev struct {
	Name             string            `json:"name"`
	Alloc            uint64            `json:"alloc"`
	Free             uint64            `json:"free"`
	Size             uint64            `json:"size"`
	Health           string            `json:"health"`
	Operations       RW                `json:"operations"`
	Bandwidth        RW                `json:"bandwidth"`
	VdevDevices      []VdevDevice      `json:"devices"`
	ReplacingDevices []ReplacingDevice `json:"replacingDevices,omitempty"`
}

type Zpool struct {
	z             *zfs    `json:"-"`
	Name          string  `json:"name"`
	Health        string  `json:"health"`
	Allocated     uint64  `json:"allocated"`
	Size          uint64  `json:"size"`
	Free          uint64  `json:"free"`
	Fragmentation uint64  `json:"fragmentation"`
	ReadOnly      bool    `json:"readOnly"`
	Freeing       uint64  `json:"freeing"`
	Leaked        uint64  `json:"leaked"`
	DedupRatio    float64 `json:"dedupRatio"`
	Vdevs         []Vdev  `json:"vdevs"`
}

func (z *zfs) zpool(arg ...string) error {
	_, err := z.zpoolOutput(arg...)
	return err
}

func (z *zfs) zpoolOutput(arg ...string) ([][]string, error) {
	return z.run(nil, nil, "zpool", arg...)
}

func (z *Zpool) parseLine(line []string) error {
	prop := line[1]
	val := line[2]

	var err error

	switch prop {
	case "name":
		setString(&z.Name, val)
	case "health":
		setString(&z.Health, val)
	case "allocated":
		err = setUint(&z.Allocated, val)
	case "size":
		err = setUint(&z.Size, val)
	case "free":
		err = setUint(&z.Free, val)
	case "fragmentation":
		// Trim trailing "%" before parsing uint
		i := strings.Index(val, "%")
		if i < 0 {
			i = len(val)
		}
		err = setUint(&z.Fragmentation, val[:i])
	case "readonly":
		z.ReadOnly = val == "on"
	case "freeing":
		err = setUint(&z.Freeing, val)
	case "leaked":
		err = setUint(&z.Leaked, val)
	case "dedupratio":
		// Trim trailing "x" before parsing float64
		z.DedupRatio, err = strconv.ParseFloat(val[:len(val)-1], 64)
	}
	return err
}

func (z *zfs) GetZpool(name string) (*Zpool, error) {
	args := zpoolArgs
	args = append(args, name)
	out, err := z.zpoolOutput(args...)
	if err != nil {
		return nil, err
	}

	pool := &Zpool{z: z, Name: name}
	for _, line := range out {
		if err := pool.parseLine(line); err != nil {
			return nil, err
		}
	}

	vdevOut, err := z.zpoolOutput(append(zpoolVdevArgs, name)...)
	if err != nil {
		return nil, err
	}

	var vdevPtrs []*Vdev
	var currentVdev *Vdev
	var currentReplacing *ReplacingDevice

	for i, line := range vdevOut {
		if len(line) < 10 {
			continue
		}

		vdevName := line[0]

		if i == 0 && vdevName == pool.Name {
			continue
		}

		if strings.HasPrefix(vdevName, "mirror") || strings.HasPrefix(vdevName, "raidz") {
			// This is a mirror or raidz vdev
			currentVdev = &Vdev{
				Name:   vdevName,
				Alloc:  pool.Allocated,
				Free:   pool.Free,
				Size:   utils.StringToUint64(line[1]),
				Health: line[9],
				Operations: RW{
					Read:  utils.StringToUint64(line[5]),
					Write: utils.StringToUint64(line[6]),
				},
				Bandwidth: RW{
					Read:  utils.StringToUint64(line[7]),
					Write: utils.StringToUint64(line[8]),
				},
				VdevDevices:      []VdevDevice{},
				ReplacingDevices: []ReplacingDevice{},
			}
			vdevPtrs = append(vdevPtrs, currentVdev)
			currentReplacing = nil
		} else if strings.HasPrefix(vdevName, "replacing") {
			// This is a replacing vdev
			if currentVdev != nil {
				currentReplacing = &ReplacingDevice{
					Name:   vdevName,
					Health: line[9],
				}
				// We'll add it to the current vdev once we've processed its devices
			} else {
				// Standalone replacing vdev (unusual, but handle it)
				vdev := &Vdev{
					Name:   vdevName,
					Alloc:  pool.Allocated,
					Free:   pool.Free,
					Size:   pool.Size,
					Health: line[9],
					Operations: RW{
						Read:  utils.StringToUint64(line[5]),
						Write: utils.StringToUint64(line[6]),
					},
					Bandwidth: RW{
						Read:  utils.StringToUint64(line[7]),
						Write: utils.StringToUint64(line[8]),
					},
					VdevDevices:      []VdevDevice{},
					ReplacingDevices: []ReplacingDevice{},
				}
				vdevPtrs = append(vdevPtrs, vdev)
				currentVdev = vdev
				currentReplacing = nil
			}
		} else if strings.HasPrefix(vdevName, "/dev/") {
			// This is a device
			device := VdevDevice{
				Name:   vdevName,
				Size:   utils.StringToUint64(line[1]),
				Health: line[9],
			}

			if currentReplacing != nil {
				// This device is part of a replacing operation
				if currentReplacing.OldDrive.Name == "" {
					// First device is the old one
					currentReplacing.OldDrive = device
				} else {
					// Second device is the new one
					currentReplacing.NewDrive = device

					// Now that we have both old and new drives, add the replacing vdev to the parent
					if currentVdev != nil {
						currentVdev.ReplacingDevices = append(currentVdev.ReplacingDevices, *currentReplacing)
					}

					// Reset currentReplacing to handle multiple replacing operations
					currentReplacing = nil
				}
			} else if currentVdev != nil {
				// This device is part of the current vdev (mirror/raidz)
				currentVdev.VdevDevices = append(currentVdev.VdevDevices, device)
			} else {
				// This is a standalone device (not part of a mirror/raidz)
				vdev := &Vdev{
					Name:   vdevName,
					Alloc:  pool.Allocated,
					Free:   pool.Free,
					Size:   pool.Size,
					Health: line[9],
					Operations: RW{
						Read:  utils.StringToUint64(line[5]),
						Write: utils.StringToUint64(line[6]),
					},
					Bandwidth: RW{
						Read:  utils.StringToUint64(line[7]),
						Write: utils.StringToUint64(line[8]),
					},
					VdevDevices: []VdevDevice{
						device,
					},
					ReplacingDevices: []ReplacingDevice{},
				}
				vdevPtrs = append(vdevPtrs, vdev)
			}
		} else {
			// Any other line that's not a vdev or device
			// Only reset if we're not in a replacing operation
			if currentReplacing == nil {
				currentVdev = nil
			}
		}
	}

	var vdevs []Vdev
	for _, v := range vdevPtrs {
		vdevs = append(vdevs, *v)
	}

	pool.Vdevs = vdevs

	return pool, nil
}

func (z *Zpool) Datasets() ([]*Dataset, error) {
	return z.z.Datasets(z.Name)
}

func (z *Zpool) Snapshots() ([]*Dataset, error) {
	return z.z.Snapshots(z.Name)
}

func (z *zfs) CreateZpool(name string, properties map[string]string, args ...string) (*Zpool, error) {
	cli := make([]string, 1, 4)
	cli[0] = "create"

	var forceFlag bool
	var otherArgs []string
	for _, arg := range args {
		if arg == "-f" {
			forceFlag = true
		} else {
			otherArgs = append(otherArgs, arg)
		}
	}

	if forceFlag {
		cli = append(cli, "-f")
	}

	if properties != nil {
		cli = append(cli, propsSlice(properties)...)
	}
	cli = append(cli, name)
	cli = append(cli, otherArgs...)

	if err := z.zpool(cli...); err != nil {
		return nil, err
	}

	return &Zpool{z: z, Name: name}, nil
}

func (z *Zpool) Destroy() error {
	err := z.z.zpool("destroy", z.Name)
	return err
}

func (z *Zpool) Replace(oldDevice string, newDevice string) error {
	found := false

	for _, vdev := range z.Vdevs {
		for _, device := range vdev.VdevDevices {
			if device.Name == oldDevice {
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("device %s not found in pool %s", oldDevice, z.Name)
	}

	err := z.z.zpool("replace", z.Name, oldDevice, newDevice)

	return err
}

func (z *zfs) ListZpools() ([]*Zpool, error) {
	args := []string{"list", "-Ho", "name"}
	out, err := z.zpoolOutput(args...)
	if err != nil {
		return nil, err
	}

	var pools []*Zpool

	for _, line := range out {
		z, err := z.GetZpool(line[0])
		if err != nil {
			return nil, err
		}
		pools = append(pools, z)
	}
	return pools, nil
}

func (z *zfs) GetPoolIODelay(poolName string) (float64, error) {
	pool, err := z.GetZpool(poolName)
	if err != nil {
		return 0.0, err
	}

	rows, err := z.zpoolOutput("iostat", "-l", "-H", "-v", pool.Name, "1", "2")
	if err != nil {
		return 0.0, err
	}

	var sampleIndices []int
	for i, row := range rows {
		if len(row) > 0 && row[0] == poolName {
			sampleIndices = append(sampleIndices, i)
		}
	}

	if len(sampleIndices) < 2 {
		return 0.0, fmt.Errorf("not enough samples for pool %s", poolName)
	}

	secondSampleRow := rows[sampleIndices[1]]
	if len(secondSampleRow) < 9 {
		return 0.0, fmt.Errorf("not enough fields in iostat output")
	}

	readOps := utils.StringToUint64(secondSampleRow[3])
	writeOps := utils.StringToUint64(secondSampleRow[4])
	if (readOps + writeOps) == 0 {
		return 0.0, nil
	}

	readWait := ParseTimeUnit(secondSampleRow[7])
	writeWait := ParseTimeUnit(secondSampleRow[8])

	totalWait := (readOps * readWait) + (writeOps * writeWait)
	avgWait := totalWait / (readOps + writeOps)
	delayPercentage := (float64(avgWait) / 1_000_000.0) * 100

	return delayPercentage, nil
}

func (z *zfs) GetTotalIODelay() float64 {
	pools, err := z.ListZpools()
	if err != nil {
		return 0.0
	}

	var totalDelay float64
	count := 0

	for _, pool := range pools {
		delay, _ := GetPoolIODelay(pool.Name)
		if delay > 0 {
			totalDelay += delay
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalDelay / float64(count)
}
