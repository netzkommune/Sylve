// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsServiceInterfaces

type RW struct {
	Read  uint64 `json:"read"`
	Write uint64 `json:"write"`
}

type Vdev struct {
	Name       string `json:"name"`
	Alloc      uint64 `json:"alloc"`
	Free       uint64 `json:"free"`
	Operations RW     `json:"operations"`
	Bandwidth  RW     `json:"bandwidth"`
}

type Zpool struct {
	Name       string  `json:"name"`
	Health     string  `json:"health"`
	Allocated  uint64  `json:"allocated"`
	Size       uint64  `json:"size"`
	Free       uint64  `json:"free"`
	ReadOnly   bool    `json:"readOnly"`
	Freeing    uint64  `json:"freeing"`
	Leaked     uint64  `json:"leaked"`
	DedupRatio float64 `json:"dedupRatio"`
	Vdevs      []Vdev  `json:"vdevs"`
}
