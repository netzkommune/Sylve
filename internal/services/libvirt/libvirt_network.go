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
	vmModels "sylve/internal/db/models/vm"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
)

func (s *Service) NetworkDetach(vmId int, networkId int) error {
	def, err := s.GetVMXML(vmId)
	if err != nil {
		return fmt.Errorf("failed_to_get_vm_xml: %w", err)
	}

	var network vmModels.Network
	err = s.DB.Preload("Switch").First(&network, "id = ?", networkId).Error

	if err != nil {
		return fmt.Errorf("failed_to_find_network: %w", err)
	}

	var parsed libvirtServiceInterfaces.Domain
	err = xml.Unmarshal([]byte(def), &parsed)
	if err != nil {
		return fmt.Errorf("failed_to_parse_domain_xml: %w", err)
	}

	for i, iface := range parsed.Devices.Interfaces {
		if iface.Source.Bridge != "" {
			if iface.Source.Bridge == network.Switch.BridgeName {
				parsed.Devices.Interfaces = append(parsed.Devices.Interfaces[:i], parsed.Devices.Interfaces[i+1:]...)
				break
			}
		}
	}

	_, err = xml.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return fmt.Errorf("failed_to_marshal_updated_domain_xml: %w", err)
	}

	// save updated XML back to the domain

	return nil
}
