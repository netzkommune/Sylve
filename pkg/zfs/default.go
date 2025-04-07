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

func DestroyPool(poolName string) error {
	var pools []*Zpool
	pools, err := ListZpools()
	if err != nil {
		return err
	}

	var found *Zpool

	for _, pool := range pools {
		if pool.Name == poolName {
			found = pool
			break
		}
	}

	if found == nil {
		return fmt.Errorf("error_getting_pool: pool %s not found", poolName)
	}

	err = found.Destroy()
	if err != nil {
		return fmt.Errorf("failed to destroy pool: %w", err)
	}

	return nil
}

func ReplaceInPool(poolName string, oldDevice string, newDevice string) error {
	pool, err := GetZpool(poolName)

	if err != nil {
		return err
	}

	err = pool.Replace(oldDevice, newDevice)
	if err != nil {
		return fmt.Errorf("failed to replace device in pool: %w", err)
	}

	return nil
}
