// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirt

import (
	"fmt"
	"strconv"
	"strings"
	networkModels "sylve/internal/db/models/network"
	vmModels "sylve/internal/db/models/vm"
	"sylve/internal/logger"
	"sylve/pkg/utils"

	"github.com/beevik/etree"
)

func (s *Service) NetworkDetach(vmId int, networkId int) error {
	inactive, err := s.IsDomainInactive(vmId)
	if err != nil {
		return fmt.Errorf("failed_to_check_vm_inactive: %w", err)
	}

	if !inactive {
		return fmt.Errorf("vm_is_active: cannot_detach_network")
	}

	xmlDesc, err := s.GetVMXML(vmId)
	if err != nil {
		return fmt.Errorf("failed_to_get_vm_xml: %w", err)
	}

	var network vmModels.Network
	if err := s.DB.Preload("Switch").
		First(&network, "id = ?", networkId).
		Error; err != nil {
		return fmt.Errorf("failed_to_find_network: %w", err)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlDesc); err != nil {
		return fmt.Errorf("failed_to_parse_vm_xml: %w", err)
	}

	found := false
	for _, iface := range doc.FindElements("//interface[@type='bridge']") {
		macEl := iface.FindElement("mac")
		if macEl == nil {
			continue
		}
		addrAttr := macEl.SelectAttr("address")
		if addrAttr == nil {
			continue
		}

		if strings.EqualFold(addrAttr.Value, network.MAC) {
			parent := iface.Parent()
			parent.RemoveChild(iface)
			found = true
			break
		}
	}

	if !found {
		logger.L.Debug().Msgf("Network detach: network_interface_not_found_in_xml: %s", network.MAC)
		if err := s.DB.Delete(&network).Error; err != nil {
			return fmt.Errorf("failed_to_delete_network_record: %w", err)
		}

		return nil
	}

	newXML, err := doc.WriteToString()
	if err != nil {
		return fmt.Errorf("failed_to_serialize_modified_xml: %w", err)
	}

	domain, err := s.Conn.DomainLookupByName(strconv.Itoa(vmId))
	if err != nil {
		return fmt.Errorf("failed_to_lookup_domain_by_name: %w", err)
	}

	if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
		return fmt.Errorf("failed_to_undefine_domain: %w", err)
	}

	if _, err := s.Conn.DomainDefineXML(newXML); err != nil {
		return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
	}

	if err := s.DB.Delete(&network).Error; err != nil {
		return fmt.Errorf("failed_to_delete_network_record: %w", err)
	}

	return nil
}

func (s *Service) NetworkAttach(vmId int, switchId int, emulation string, macObjId uint) error {
	inactive, err := s.IsDomainInactive(vmId)
	if err != nil {
		return fmt.Errorf("failed_to_check_vm_inactive: %w", err)
	}

	if !inactive {
		return fmt.Errorf("vm_is_active: cannot_attach_network")
	}

	if emulation == "" || (emulation != "virtio" && emulation != "e1000") {
		return fmt.Errorf("invalid_emulation_type: %s", emulation)
	}

	var stdSwitch networkModels.StandardSwitch
	if err := s.DB.First(&stdSwitch, switchId).Error; err != nil {
		return fmt.Errorf("failed_to_find_switch: %w", err)
	}

	vms, err := s.ListVMs()
	if err != nil {
		return fmt.Errorf("failed_to_list_vms: %w", err)
	}

	var vm *vmModels.VM
	for _, v := range vms {
		if v.VmID == vmId {
			vm = &v
			break
		}
	}

	var existingNetwork vmModels.Network
	if err := s.DB.First(&existingNetwork, "vm_id = ? AND switch_id = ?", vm.ID, switchId).Error; err == nil {
		return fmt.Errorf("network_already_attached_to_vm: %s", existingNetwork.MAC)
	}

	if macObjId == 0 {
		macAddress := utils.GenerateRandomMAC()

		macObj := networkModels.Object{
			Name: fmt.Sprintf("vm-%d-mac-%s", vm.VmID, macAddress),
			Type: "Mac",
		}

		if err := s.DB.Create(&macObj).Error; err != nil {
			return fmt.Errorf("failed_to_create_mac_object: %w", err)
		}

		macEntry := networkModels.ObjectEntry{
			ObjectID: macObj.ID,
			Value:    macAddress,
		}

		if err := s.DB.Create(&macEntry).Error; err != nil {
			return fmt.Errorf("failed_to_create_mac_entry: %w", err)
		}

		macObjId = macObj.ID
	} else {
		var macObj networkModels.Object
		if err := s.DB.Preload("Entries").First(&macObj, macObjId).Error; err != nil {
			return fmt.Errorf("failed_to_find_mac_object: %w", err)
		}

		if macObj.Type != "Mac" {
			return fmt.Errorf("invalid_mac_object_type: %s", macObj.Type)
		}

		if len(macObj.Entries) == 0 {
			return fmt.Errorf("mac_object_has_no_entries: %d", macObjId)
		}

		var otherNetworks []vmModels.Network
		if err := s.DB.Where("mac_id = ? AND vm_id != ?", macObjId, vm.ID).
			Find(&otherNetworks).Error; err != nil {
			return fmt.Errorf("failed_to_find_other_networks_using_mac_object: %w", err)
		}
	}

	network := vmModels.Network{
		VMID:      vm.ID,
		SwitchID:  uint(switchId),
		MacID:     &macObjId,
		Emulation: emulation,
	}

	if err := s.DB.Create(&network).Error; err != nil {
		return fmt.Errorf("failed_to_create_network_record: %w", err)
	}

	var macAddress string

	if macObjId != 0 {
		var macObj networkModels.Object
		if err := s.DB.Preload("Entries").First(&macObj, macObjId).Error; err != nil {
			return fmt.Errorf("failed_to_find_mac_object: %w", err)
		}
		if len(macObj.Entries) == 0 {
			return fmt.Errorf("mac_object_has_no_entries: %d", macObjId)
		}
		macAddress = macObj.Entries[0].Value
	}

	xmlDesc, err := s.GetVMXML(vmId)
	if err != nil {
		return fmt.Errorf("failed_to_get_vm_xml: %w", err)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlDesc); err != nil {
		return fmt.Errorf("failed_to_parse_vm_xml: %w", err)
	}

	domainEl := doc.SelectElement("domain")
	if domainEl == nil {
		return fmt.Errorf("malformed_vm_xml: missing <domain> element")
	}

	devicesEl := domainEl.FindElement("devices")
	if devicesEl == nil {
		devicesEl = etree.NewElement("devices")
		domainEl.AddChild(devicesEl)
	}

	ifaceEl := etree.NewElement("interface")
	ifaceEl.CreateAttr("type", "bridge")

	macEl := etree.NewElement("mac")
	macEl.CreateAttr("address", macAddress)
	ifaceEl.AddChild(macEl)

	sourceEl := etree.NewElement("source")
	sourceEl.CreateAttr("bridge", stdSwitch.BridgeName)
	ifaceEl.AddChild(sourceEl)

	modelEl := etree.NewElement("model")
	modelEl.CreateAttr("type", network.Emulation)
	ifaceEl.AddChild(modelEl)

	devicesEl.AddChild(ifaceEl)

	newXML, err := doc.WriteToString()
	if err != nil {
		return fmt.Errorf("failed_to_serialize_modified_xml: %w", err)
	}

	domain, err := s.Conn.DomainLookupByName(strconv.Itoa(vmId))
	if err != nil {
		return fmt.Errorf("failed_to_lookup_domain_by_name: %w", err)
	}

	if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
		return fmt.Errorf("failed_to_undefine_domain: %w", err)
	}

	if _, err := s.Conn.DomainDefineXML(newXML); err != nil {
		return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
	}

	return nil
}
