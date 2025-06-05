// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirtServiceInterfaces

type LibvirtServiceInterface interface {
	CheckVersion() error

	ListStoragePools() ([]StoragePool, error)
	CreateStoragePool(name string) error
	DeleteStoragePool(name string) error
	RescanStoragePools() error

	StoreVMUsage() error
}

type LvDomain struct {
	ID     int32  `json:"id"`
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
