package zfsServiceInterfaces

import "sylve/pkg/zfs"

type Dataset struct {
	zfs.Dataset
	Properties map[string]string `json:"properties"`
}
