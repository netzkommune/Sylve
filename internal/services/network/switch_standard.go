package network

import (
	"fmt"
	"strings"
	networkModels "sylve/internal/db/models/network"
	"sylve/pkg/utils"
)

func (s *Service) GetStandardSwitches() ([]networkModels.StandardSwitch, error) {
	var switches []networkModels.StandardSwitch
	if err := s.DB.
		Preload("Ports").
		Find(&switches).Error; err != nil {
		return nil, err
	}
	return switches, nil
}

func (s *Service) NewStandardSwitch(
	name string,
	mtu int,
	vlan int,
	address string,
	ports []string,
) error {
	var existingPorts []networkModels.NetworkPort
	if err := s.DB.Where("name IN ?", ports).Find(&existingPorts).Error; err != nil {
		return fmt.Errorf("db_error_checking_ports: %v", err)
	}
	if len(existingPorts) > 0 {
		return fmt.Errorf("port_overlap")
	}

	if !utils.IsValidMTU(mtu) {
		return fmt.Errorf("invalid_mtu")
	}

	if !utils.IsValidVLAN(vlan) && vlan != 0 {
		return fmt.Errorf("invalid_vlan")
	}

	if !utils.IsValidIP(address) && address != "" {
		return fmt.Errorf("invalid_ip")
	}

	sw := &networkModels.StandardSwitch{
		Name:    name,
		MTU:     mtu,
		VLAN:    vlan,
		Address: address,
	}

	if err := s.DB.Create(sw).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			return fmt.Errorf("switch_name_already_exists")
		}
		return fmt.Errorf("failed_to_create_switch: %v", err)
	}

	for _, p := range ports {
		port := &networkModels.NetworkPort{
			Name:     p,
			SwitchID: sw.ID,
		}
		if err := s.DB.Create(port).Error; err != nil {
			return fmt.Errorf("failed_to_create_port %s: %v", p, err)
		}
	}

	return nil
}

func (s *Service) EditStandardSwitch(
	id int,
	name string,
	mtu int,
	vlan int,
	address string,
	ports []string,
) error {
	var conflictingPorts []networkModels.NetworkPort
	if err := s.DB.
		Where("name IN ?", ports).
		Where("switch_id <> ?", id).
		Find(&conflictingPorts).Error; err != nil {
		return fmt.Errorf("db_error_checking_ports: %v", err)
	}
	if len(conflictingPorts) > 0 {
		return fmt.Errorf("port_overlap")
	}

	if !utils.IsValidMTU(mtu) {
		return fmt.Errorf("invalid_mtu")
	}
	if !utils.IsValidVLAN(vlan) && vlan != 0 {
		return fmt.Errorf("invalid_vlan")
	}
	if address != "" && !utils.IsValidIP(address) {
		return fmt.Errorf("invalid_ip")
	}

	var sw networkModels.StandardSwitch
	if err := s.DB.First(&sw, id).Error; err != nil {
		return fmt.Errorf("switch_not_found")
	}

	sw.Name = name
	sw.MTU = mtu
	sw.VLAN = vlan
	sw.Address = address

	if err := s.DB.Save(&sw).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			return fmt.Errorf("switch_name_already_exists")
		}
		return fmt.Errorf("failed_to_update_switch: %v", err)
	}

	if err := s.DB.Where("switch_id = ?", id).Delete(&networkModels.NetworkPort{}).Error; err != nil {
		return fmt.Errorf("failed_to_clear_old_ports: %v", err)
	}

	for _, p := range ports {
		port := &networkModels.NetworkPort{
			Name:     p,
			SwitchID: id,
		}
		if err := s.DB.Create(port).Error; err != nil {
			return fmt.Errorf("failed_to_create_port %s: %v", p, err)
		}
	}

	return nil
}

func (s *Service) DeleteStandardSwitch(id int) error {
	var sw networkModels.StandardSwitch

	if err := s.DB.Where("id = ?", id).Delete(&sw).Error; err != nil {
		return fmt.Errorf("failed_to_delete_switch %v", err)
	}

	return nil
}
