// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package db

import (
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func GetHistorical[T any](db *gorm.DB, limit int) ([]T, error) {
	var records []T

	tableName := schema.NamingStrategy{}.TableName(reflect.TypeOf((*T)(nil)).Elem().Name())

	err := db.Table(tableName).
		Order("id DESC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	return records, nil
}

func StoreAndTrimRecords(db *gorm.DB, model interface{}, limit int) {
	db.Create(model)

	var count int64
	db.Model(model).Count(&count)

	if count > int64(limit) {
		var oldestIDs []uint
		db.Model(model).
			Order("id ASC").
			Limit(int(count-int64(limit))).
			Pluck("id", &oldestIDs)

		if len(oldestIDs) > 0 {
			db.Where("id IN (?)", oldestIDs).Delete(model)
		}
	}
}
