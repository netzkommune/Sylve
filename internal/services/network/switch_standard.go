package network

import (
	"fmt"
	"strconv"
	"strings"
	networkModels "sylve/internal/db/models/network"
	"sylve/pkg/utils"

	"gorm.io/gorm"
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
	private bool,
) error {
	defer s.SyncStandardSwitches()

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

	if !utils.IsValidIPv4CIDR(address) && address != "" {
		return fmt.Errorf("invalid_ip")
	}

	sw := &networkModels.StandardSwitch{
		Name:       name,
		MTU:        mtu,
		VLAN:       vlan,
		Address:    address,
		BridgeName: utils.ShortHash("vm-" + name),
		Private:    private,
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
	private bool,
) error {
	defer s.SyncStandardSwitches()

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

	if address != "" && !utils.IsValidIPv4CIDR(address) {
		return fmt.Errorf("invalid_ip")
	}

	var sw networkModels.StandardSwitch
	if err := s.DB.First(&sw, id).Error; err != nil {
		return fmt.Errorf("switch_not_found")
	}

	oldBridge := sw.BridgeName
	newBridge := utils.ShortHash("vm-" + name)

	sw.Name = name
	sw.MTU = mtu
	sw.VLAN = vlan
	sw.Address = address
	sw.BridgeName = newBridge
	sw.Private = private

	if err := s.DB.Save(&sw).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			return fmt.Errorf("switch_name_already_exists")
		}
		return fmt.Errorf("failed_to_update_switch: %v", err)
	}

	if oldBridge != newBridge {
		if _, err := utils.RunCommand("ifconfig", oldBridge); err == nil {
			if _, err := utils.RunCommand("ifconfig", oldBridge, "name", newBridge); err != nil {
				return fmt.Errorf("failed_to_rename_bridge %q→%q: %v", oldBridge, newBridge, err)
			}
		}
	}

	if err := s.DB.Where("switch_id = ?", id).
		Delete(&networkModels.NetworkPort{}).Error; err != nil {
		return fmt.Errorf("failed_to_clear_old_ports: %v", err)
	}
	for _, p := range ports {
		port := &networkModels.NetworkPort{Name: p, SwitchID: id}
		if err := s.DB.Create(port).Error; err != nil {
			return fmt.Errorf("failed_to_create_port %s: %v", p, err)
		}
	}

	return nil
}

func (s *Service) DeleteStandardSwitch(id int) error {
	var sw networkModels.StandardSwitch
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Ports").First(&sw, id).Error; err != nil {
			return fmt.Errorf("switch_not_found: %v", err)
		}

		if err := tx.Delete(&sw).Error; err != nil {
			return fmt.Errorf("failed_to_delete_switch: %v", err)
		}

		if err := tx.Where("switch_id = ?", id).
			Delete(&networkModels.NetworkPort{}).Error; err != nil {
			return fmt.Errorf("failed_to_clear_old_ports: %v", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed_to_delete_switch: %v", err)
	}

	brName := sw.BridgeName
	if err := removeBridge(brName, sw); err != nil {
		return fmt.Errorf("failed_to_remove_bridge %q: %v", brName, err)
	}

	return nil
}

func removeBridge(br string, sw networkModels.StandardSwitch) error {
	for _, port := range sw.Ports {
		if err := unconfigurePort(br, port.Name, sw.VLAN); err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				continue
			}
			return fmt.Errorf("unconfigure_port %s: %v", port.Name, err)
		}
	}

	if _, err := utils.RunCommand("ifconfig", br, "destroy"); err != nil {
		return fmt.Errorf("bridge_destroy_failed %q: %v", br, err)
	}
	return nil
}

func (s *Service) SyncStandardSwitches() error {
	var list []networkModels.StandardSwitch
	if err := s.DB.Preload("Ports").Find(&list).Error; err != nil {
		return fmt.Errorf("db_error: %v", err)
	}
	for _, sw := range list {
		ifName := sw.BridgeName
		if err := syncBridge(ifName, sw); err != nil {
			return fmt.Errorf("sync %q: %v", sw.Name, err)
		}
	}
	return nil
}

func syncBridge(br string, sw networkModels.StandardSwitch) error {
	if _, err := utils.RunCommand("ifconfig", br); err == nil {
		return updateBridge(br, sw)
	}

	return createBridge(br, sw)
}

func createBridge(br string, sw networkModels.StandardSwitch) error {
	createdRaw, err := utils.RunCommand("ifconfig", "bridge", "create")
	if err != nil {
		return fmt.Errorf("failed_to_create_bridge: %v", err)
	}
	created := strings.TrimSpace(createdRaw)

	if _, err := utils.RunCommand("ifconfig", created, "name", br); err != nil {
		return fmt.Errorf("failed_to_rename_bridge: %v", err)
	}

	if _, err := utils.RunCommand("ifconfig", br, "descr", sw.Name); err != nil {
		return fmt.Errorf("failed_to_set_bridge_description: %v", err)
	}

	if sw.MTU != 0 {
		if _, err := utils.RunCommand("ifconfig", br, "mtu", strconv.Itoa(sw.MTU)); err != nil {
			return fmt.Errorf("failed_to_set_bridge_mtu: %v", err)
		}
	}

	if sw.Address != "" {
		if _, err := utils.RunCommand("ifconfig", br, "inet", sw.Address); err != nil {
			return fmt.Errorf("failed_to_set_bridge_address: %v", err)
		}
	}

	for _, port := range sw.Ports {
		if err := configurePort(br, port.Name, sw.VLAN, sw.MTU); err != nil {
			return fmt.Errorf("failed_to_configure_port %s: %v", port.Name, err)
		}
	}
	return nil
}

func updateBridge(br string, sw networkModels.StandardSwitch) error {
	if _, err := utils.RunCommand("ifconfig", br, "descr", sw.Name); err != nil {
		return fmt.Errorf("failed_to_set_bridge_description: %v", err)
	}

	if sw.MTU != 0 {
		if _, err := utils.RunCommand("ifconfig", br, "mtu", strconv.Itoa(sw.MTU)); err != nil {
			return fmt.Errorf("failed_to_set_bridge_mtu: %v", err)
		}
	}

	if sw.Address != "" {
		if _, err := utils.RunCommand("ifconfig", br, "inet", sw.Address); err != nil {
			return fmt.Errorf("failed_to_set_bridge_address: %v", err)
		}
	}

	for _, port := range sw.Ports {
		_, err := utils.RunCommand("ifconfig", br, "deletem", port.Name)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}

			return fmt.Errorf("failed_to_remove_port %s: %v", port.Name, err)
		}
	}

	for _, port := range sw.Ports {
		if err := configurePort(br, port.Name, sw.VLAN, sw.MTU); err != nil {
			return fmt.Errorf("failed_to_configure_port %s: %v", port.Name, err)
		}
	}

	return nil
}

func configurePort(br, port string, vlan, mtu int) error {
	if mtu > 0 {
		if _, err := utils.RunCommand("ifconfig", port, "mtu", strconv.Itoa(mtu)); err != nil {
			return fmt.Errorf("failed to set MTU on %s: %v", port, err)
		}
	}

	if vlan > 0 {
		vif := fmt.Sprintf("%s.%d", port, vlan)
		if _, err := utils.RunCommand("ifconfig", vif); err != nil {
			args := []string{
				"vlan", "create",
				"vlandev", port,
				"vlan", strconv.Itoa(vlan),
				"descr", fmt.Sprintf("svm-vlan/%s/%s", br, vif),
				"name", vif,
				"group", "svm-vlan",
				"up",
			}
			if _, err := utils.RunCommand("ifconfig", args...); err != nil {
				return fmt.Errorf("failed to create VLAN iface %s: %v", vif, err)
			}
		}

		if _, err := utils.RunCommand("ifconfig", br, "addm", vif); err != nil {
			return fmt.Errorf("failed to add %s to bridge %s: %v", vif, br, err)
		}
	} else {
		if _, err := utils.RunCommand("ifconfig", br, "addm", port); err != nil {
			return fmt.Errorf("failed to add %s to bridge %s: %v", port, br, err)
		}
	}

	return nil
}

func unconfigurePort(br, port string, vlan int) error {
	if vlan > 0 {
		vif := fmt.Sprintf("%s.%d", port, vlan)
		if _, err := utils.RunCommand("ifconfig", vif, "destroy"); err != nil {
			return fmt.Errorf("failed to destroy VLAN iface %s: %v", vif, err)
		}
	} else {
		if _, err := utils.RunCommand("ifconfig", br, "deletem", port); err != nil {
			return fmt.Errorf("failed to remove %s from bridge %s: %v", port, br, err)
		}
	}
	return nil
}
