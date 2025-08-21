// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package db

import (
	"errors"
	"fmt"

	networkModels "github.com/alchemillahq/sylve/internal/db/models/network"
	vmModels "github.com/alchemillahq/sylve/internal/db/models/vm"

	"gorm.io/gorm"
)

func Fixups(db *gorm.DB) error {
	if err := NetworkObjectFixups(db); err != nil {
		return err
	}

	return nil
}

func NetworkObjectFixups(db *gorm.DB) error {
	/* Enforce Host objects usage in Switches */
	var switches []networkModels.StandardSwitch
	if err := db.Find(&switches).Error; err != nil {
		return fmt.Errorf("fetch switches: %w", err)
	}

	for _, sw := range switches {
		updates := map[string]interface{}{}
		if sw.Address != "" {
			objName := fmt.Sprintf("%s-auto-ipv4", sw.Name)
			var obj networkModels.Object
			err := db.Where("name = ?", objName).First(&obj).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				obj = networkModels.Object{
					Name:    objName,
					Type:    "Host",
					Comment: fmt.Sprintf("Auto-generated IPv4 for switch %s", sw.Name),
				}

				if err := db.Create(&obj).Error; err != nil {
					return fmt.Errorf("create IPv4 object: %w", err)
				}

				entry := networkModels.ObjectEntry{
					ObjectID: obj.ID,
					Value:    sw.Address,
				}
				if err := db.Create(&entry).Error; err != nil {
					return fmt.Errorf("create IPv4 entry: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("lookup IPv4 object: %w", err)
			}
			updates["address_object_id"] = obj.ID
			updates["address"] = ""
		}

		if sw.Address6 != "" {
			objName6 := fmt.Sprintf("%s-auto-ipv6", sw.Name)
			var obj6 networkModels.Object
			err := db.Where("name = ?", objName6).First(&obj6).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				obj6 = networkModels.Object{
					Name:    objName6,
					Type:    "Host",
					Comment: fmt.Sprintf("Auto-generated IPv6 for switch %s", sw.Name),
				}

				if err := db.Create(&obj6).Error; err != nil {
					return fmt.Errorf("create IPv6 object: %w", err)
				}

				entry6 := networkModels.ObjectEntry{
					ObjectID: obj6.ID,
					Value:    sw.Address6,
				}

				if err := db.Create(&entry6).Error; err != nil {
					return fmt.Errorf("create IPv6 entry: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("lookup IPv6 object: %w", err)
			}
			updates["address6_object_id"] = obj6.ID
			updates["address6"] = ""
		}

		if len(updates) > 0 {
			if err := db.
				Model(&networkModels.StandardSwitch{}).
				Where("id = ?", sw.ID).
				Updates(updates).Error; err != nil {
				return fmt.Errorf("update switch %d: %w", sw.ID, err)
			}
		}
	}

	/* Enforce MAC address objects for VM networks */
	var vmNets []vmModels.Network
	if err := db.Where("mac_id IS NULL AND mac != ''").Find(&vmNets).Error; err != nil {
		return fmt.Errorf("fetch vm networks: %w", err)
	}

	for i, net := range vmNets {
		var vm vmModels.VM
		if err := db.First(&vm, net.VMID).Error; err != nil {
			return fmt.Errorf("fetch VM %d for network %d: %w", net.VMID, net.ID, err)
		}

		objName := fmt.Sprintf("vm-%d-mac-%s", vm.VmID, net.MAC)
		var obj networkModels.Object
		err := db.Where("name = ?", objName).First(&obj).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			obj = networkModels.Object{
				Name:    objName,
				Type:    "Mac",
				Comment: fmt.Sprintf("MAC address for VM %d network %d", net.VMID, i),
			}
			if err := db.Create(&obj).Error; err != nil {
				return fmt.Errorf("create MAC object: %w", err)
			}

			entry := networkModels.ObjectEntry{
				ObjectID: obj.ID,
				Value:    net.MAC,
			}
			if err := db.Create(&entry).Error; err != nil {
				return fmt.Errorf("create MAC entry: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("lookup MAC object: %w", err)
		}

		if err := db.Model(&vmModels.Network{}).
			Where("id = ?", net.ID).
			Updates(map[string]interface{}{
				"mac_id": obj.ID,
				"mac":    "",
			}).Error; err != nil {
			return fmt.Errorf("update network %d: %w", net.ID, err)
		}
	}

	return nil
}
