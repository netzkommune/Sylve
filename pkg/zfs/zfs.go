package zfs

import (
	"io"
	"strconv"
	"sylve/pkg/exe"
)

type InodeType int
type ChangeType int
type DestroyFlag int

const (
	DatasetFilesystem = "filesystem"
	DatasetSnapshot   = "snapshot"
	DatasetVolume     = "volume"
)

const (
	DestroyDefault         DestroyFlag = 1 << iota
	DestroyRecursive                   = 1 << iota
	DestroyRecursiveClones             = 1 << iota
	DestroyDeferDeletion               = 1 << iota
	DestroyForceUmount                 = 1 << iota
)

type zfs struct {
	exec exe.Executor
	sudo bool
}

type InodeChange struct {
	Change               ChangeType
	Type                 InodeType
	Path                 string
	NewPath              string
	ReferenceCountChange int
}

type ZFS interface {
	Datasets(filter string) ([]*Dataset, error)
	GetDataset(name string) (*Dataset, error)
	CreateFilesystem(name string, properties map[string]string) (*Dataset, error)
	Filesystems(filter string) ([]*Dataset, error)
	CreateVolume(name string, size uint64, properties map[string]string) (*Dataset, error)
	Volumes(filter string) ([]*Dataset, error)
	Snapshots(filter string) ([]*Dataset, error)
	ReceiveSnapshot(input io.Reader, name string, force ...bool) (*Dataset, error)

	ListZpools() ([]*Zpool, error)
	GetZpool(name string) (*Zpool, error)
	ScrubPool(name string) error
	CreateZpool(name string, properties map[string]string, args ...string) (*Zpool, error)
	GetPoolIODelay(poolName string) (float64, error)
	GetTotalIODelay() float64
}

func (z *zfs) do(arg ...string) error {
	_, err := z.doOutput(arg...)
	return err
}

func (z *zfs) doOutput(arg ...string) ([][]string, error) {
	return z.run(nil, nil, "zfs", arg...)
}

func (z *zfs) Datasets(filter string) ([]*Dataset, error) {
	return z.listByType("all", filter)
}

func (z *zfs) Snapshots(filter string) ([]*Dataset, error) {
	return z.listByType(DatasetSnapshot, filter)
}

func (z *zfs) Filesystems(filter string) ([]*Dataset, error) {
	return z.listByType(DatasetFilesystem, filter)
}

func (z *zfs) Volumes(filter string) ([]*Dataset, error) {
	return z.listByType(DatasetVolume, filter)
}

func (z *zfs) GetDataset(name string) (*Dataset, error) {
	out, err := z.doOutput("list", "-p", "-o", "all", name)
	if err != nil {
		return nil, err
	}

	ds := &Dataset{z: z, Name: name, props: make(map[string]string)}
	return ds, ds.parseProps(out)
}

func (z *zfs) ReceiveSnapshot(input io.Reader, name string, force ...bool) (*Dataset, error) {
	args := []string{"receive"}
	if len(force) > 0 && force[0] {
		args = append(args, "-F")
	}
	args = append(args, name)
	if _, err := z.run(input, nil, "zfs", args...); err != nil {
		return nil, err
	}
	return z.GetDataset(name)
}

func (z *zfs) CreateVolume(name string, size uint64, properties map[string]string) (*Dataset, error) {
	args := make([]string, 4, 5)
	args[0] = "create"
	args[1] = "-p"
	args[2] = "-V"
	args[3] = strconv.FormatUint(size, 10)
	if properties != nil {
		args = append(args, propsSlice(properties)...)
	}
	args = append(args, name)
	if err := z.do(args...); err != nil {
		return nil, err
	}
	return z.GetDataset(name)
}

// https://openzfs.github.io/openzfs-docs/man/7/zfsprops.7.html.
func (z *zfs) CreateFilesystem(name string, properties map[string]string) (*Dataset, error) {
	args := make([]string, 1, 4)
	args[0] = "create"

	if properties != nil {
		args = append(args, propsSlice(properties)...)
	}

	args = append(args, name)
	if err := z.do(args...); err != nil {
		return nil, err
	}
	return z.GetDataset(name)
}
