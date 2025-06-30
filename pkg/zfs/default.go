package zfs

import (
	"fmt"
	"sylve/pkg/exe"
)

var z ZFS = &zfs{exec: exe.NewLocalExecutor(), sudo: false}

func SetDefault(zfs ZFS) {
	if zfs != nil {
		z = zfs
	}
}

func Datasets(filter string) ([]*Dataset, error) {
	return z.Datasets(filter)
}

func Snapshots(filter string) ([]*Dataset, error) {
	return z.Snapshots(filter)
}

func GetZpool(name string) (*Zpool, error) {
	return z.GetZpool(name)
}

func GetZpoolByGUID(guid string) (*Zpool, error) {
	return z.GetZpoolByGUID(guid)
}

func CreateZpool(name string, properties map[string]string, args ...string) (*Zpool, error) {
	return z.CreateZpool(name, properties, args...)
}

func ListZpools() ([]*Zpool, error) {
	return z.ListZpools()
}

func GetPoolIODelay(poolName string) (float64, error) {
	return z.GetPoolIODelay(poolName)
}

func GetTotalIODelay() float64 {
	return z.GetTotalIODelay()
}

func DestroyPool(guid string) error {
	var pools []*Zpool
	pools, err := ListZpools()
	if err != nil {
		return err
	}

	var found *Zpool

	for _, pool := range pools {
		if pool.GUID == guid {
			found = pool
			break
		}
	}

	if found == nil {
		return fmt.Errorf("error_getting_pool: pool %s not found", guid)
	}

	err = found.Destroy()
	if err != nil {
		return fmt.Errorf("failed to destroy pool: %w", err)
	}

	return nil
}

func RemoveSpare(guid string, device string) error {
	pool, err := GetZpoolByGUID(guid)
	if err != nil {
		return fmt.Errorf("failed to get pool by GUID %s: %w", guid, err)
	}

	err = pool.RemoveSpare(device)
	if err != nil {
		return fmt.Errorf("failed to remove spare device %s from pool %s: %w", device, pool.Name, err)
	}

	return nil
}

func AddSpare(guid string, device string) error {
	pool, err := GetZpoolByGUID(guid)
	if err != nil {
		return fmt.Errorf("failed to get pool by GUID %s: %w", guid, err)
	}

	err = pool.AddSpare(device)
	if err != nil {
		return fmt.Errorf("failed to add spare device %s to pool %s: %w", device, pool.Name, err)
	}

	return nil
}

func ReplaceInPool(guid string, oldDevice string, newDevice string) error {
	pool, err := GetZpoolByGUID(guid)

	if err != nil {
		return err
	}

	err = pool.Replace(oldDevice, newDevice)
	if err != nil {
		return fmt.Errorf("failed to replace device in pool: %w", err)
	}

	return nil
}

func GetZpoolStatus(poolName string) (ZpoolStatus, error) {
	pool, err := GetZpool(poolName)
	if err != nil {
		return ZpoolStatus{}, fmt.Errorf("failed to get pool %s: %w", poolName, err)
	}

	status, err := pool.z.GetZpoolStatus(pool.Name)
	if err != nil {
		return ZpoolStatus{}, fmt.Errorf("failed to get status of pool %s: %w", poolName, err)
	}

	return status, nil
}

func ScrubPool(guid string) error {
	err := z.ScrubPool(guid)

	if err != nil {
		return fmt.Errorf("failed to scrub pool %s: %w", guid, err)
	}

	return nil
}

func CreateVolume(name string, size uint64, props map[string]string) (*Dataset, error) {
	return z.CreateVolume(name, size, props)
}

func CreateFilesystem(name string, props map[string]string) (*Dataset, error) {
	return z.CreateFilesystem(name, props)
}

func SetZpoolProperty(pool string, property string, value string) error {
	return z.SetZpoolProperty(pool, property, value)
}
