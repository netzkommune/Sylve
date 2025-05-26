// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoModels

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sylve/pkg/zfs"
)

type ZpoolJSON zfs.Zpool

func (z ZpoolJSON) Value() (driver.Value, error) {
	return json.Marshal(z)
}

func (z *ZpoolJSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ZpoolJSON: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, z)
}

func (ZpoolJSON) GormDataType() string {
	return "text"
}

type ZPoolHistorical struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Pools     ZpoolJSON `json:"pools" gorm:"type:text"`
	CreatedAt int64     `json:"created_at" gorm:"autoCreateTime:milli"`
}
