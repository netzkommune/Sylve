// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirt

import (
	"encoding/xml"
	"fmt"

	libvirtServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/libvirt"

	"github.com/digitalocean/go-libvirt"
)

func (s *Service) ListStoragePools() ([]libvirtServiceInterfaces.StoragePool, error) {
	pools, _, err := s.Conn.ConnectListAllStoragePools(-1, libvirt.ConnectListStoragePoolsZfs)

	if err != nil {
		return nil, err
	}

	var result []libvirtServiceInterfaces.StoragePool

	for _, pool := range pools {
		xmlDesc, err := s.Conn.StoragePoolGetXMLDesc(pool, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get XML for pool %v: %w", pool.Name, err)
		}

		var poolXML libvirtServiceInterfaces.StoragePoolXML

		if err := xml.Unmarshal([]byte(xmlDesc), &poolXML); err != nil {
			return nil, fmt.Errorf("failed to unmarshal XML for pool %v: %w", pool.Name, err)
		}

		result = append(result, libvirtServiceInterfaces.StoragePool{
			Name:   poolXML.Name,
			UUID:   poolXML.UUID,
			Source: poolXML.Source.Name,
		})
	}

	return result, nil
}

func (s *Service) CreateStoragePool(name string) error {
	xml := fmt.Sprintf(`
		<pool type='zfs'>
		<name>%s</name>
		<source>
			<name>%s</name>
		</source>
		</pool>`, name, name)

	pool, err := s.Conn.StoragePoolDefineXML(xml, 0)
	if err != nil {
		return fmt.Errorf("failed to define storage pool: %w", err)
	}

	if err := s.Conn.StoragePoolCreate(pool, 0); err != nil {
		return fmt.Errorf("failed to create/start storage pool: %w", err)
	}

	if err := s.Conn.StoragePoolSetAutostart(pool, 1); err != nil {
		return fmt.Errorf("failed to set autostart: %w", err)
	}

	return nil
}

func (s *Service) DeleteStoragePool(name string) error {
	pool, err := s.Conn.StoragePoolLookupByName(name)
	if err != nil {
		return fmt.Errorf("failed to lookup storage pool: %w", err)
	}

	if err := s.Conn.StoragePoolDestroy(pool); err != nil {
		return fmt.Errorf("failed to destroy storage pool: %w", err)
	}

	if err := s.Conn.StoragePoolUndefine(pool); err != nil {
		return fmt.Errorf("failed to undefine storage pool: %w", err)
	}

	return nil
}

func (s *Service) RescanStoragePools() error {
	pools, _, err := s.Conn.ConnectListAllStoragePools(-1, libvirt.ConnectListStoragePoolsZfs)
	if err != nil {
		return fmt.Errorf("failed to list storage pools: %w", err)
	}

	for _, pool := range pools {
		if err := s.Conn.StoragePoolRefresh(pool, 0); err != nil {
			return fmt.Errorf("failed to refresh storage pool %s: %w", pool.Name, err)
		}
	}

	return nil
}
