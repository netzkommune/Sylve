package network

import (
	"fmt"
	"strconv"
	"strings"
	networkModels "sylve/internal/db/models/network"
	iface "sylve/pkg/network/iface"
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
	address6 string,
	ports []string,
	private bool,
) error {
	var existingPorts []networkModels.NetworkPort
	if err := s.DB.Where("name IN ?", ports).Find(&existingPorts).Error; err != nil {
		return fmt.Errorf("db_error_checking_ports: %v", err)
	}

	if len(existingPorts) > 0 {
		return fmt.Errorf("port_overlap")
	}

	if !utils.IsValidMTU(mtu) && mtu != 0 {
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
		Address6:   address6,
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

	var new networkModels.StandardSwitch
	if err := s.DB.Preload("Ports").First(&new, sw.ID).Error; err != nil {
		return fmt.Errorf("reload switch: %v", err)
	}

	return s.SyncStandardSwitches(&new, "create")
}

func (s *Service) DeleteStandardSwitch(id int) error {
	var oldSw networkModels.StandardSwitch

	var sw networkModels.StandardSwitch
	if err := s.DB.Preload("Ports").First(&sw, id).Error; err != nil {
		return fmt.Errorf("switch_not_found")
	}

	oldSw = sw

	i, err := iface.Get(sw.BridgeName)
	ifaceMissing := false

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ifaceMissing = true
		} else {
			return fmt.Errorf("delete_standard_bridge: failed to get interface %s: %v", sw.BridgeName, err)
		}
	}

	if !ifaceMissing {
		missing := []string{}

		for _, member := range i.BridgeMembers {
			found := false
			for _, port := range sw.Ports {
				if port.Name == member.Name {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, member.Name)
			}
		}

		if sw.VLAN > 0 {
			for i, port := range missing {
				if strings.Contains(port, fmt.Sprintf(".%d", sw.VLAN)) {
					missing = append(missing[:i], missing[i+1:]...)
					break
				}
			}
		}

		if len(missing) > 0 {
			return fmt.Errorf("delete_standard_bridge: missing_ports_in_db: %v", strings.Join(missing, ", "))
		}
	}

	if err := s.DB.Delete(&sw).Error; err != nil {
		return fmt.Errorf("failed_to_delete_switch: %v", err)
	}

	if err := s.DB.Where("switch_id = ?", id).
		Delete(&networkModels.NetworkPort{}).Error; err != nil {
		return fmt.Errorf("failed_to_delete_ports: %v", err)
	}

	return s.SyncStandardSwitches(&oldSw, "delete")
}

func (s *Service) EditStandardSwitch(
	id int,
	mtu int,
	vlan int,
	address string,
	address6 string,
	ports []string,
	private bool,
) error {
	var oldSw networkModels.StandardSwitch
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
	if err := s.DB.Preload("Ports").First(&sw, id).Error; err != nil {
		return fmt.Errorf("switch_not_found")
	}

	oldSw = sw

	sw.MTU = mtu
	sw.VLAN = vlan
	sw.Address = address
	sw.Address6 = address6
	sw.Private = private

	if err := s.DB.Save(&sw).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			return fmt.Errorf("switch_name_already_exists")
		}
		return fmt.Errorf("failed_to_update_switch: %v", err)
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

	return s.SyncStandardSwitches(&oldSw, "edit")
}

func (s *Service) SyncStandardSwitches(sw *networkModels.StandardSwitch, action string) error {
	if action == "create" {
		if err := createStandardBridge(*sw); err != nil {
			return err
		}
	}

	if action == "delete" {
		if err := deleteStandardBridge(*sw); err != nil {
			return err
		}
	}

	if action == "edit" {
		var newSw networkModels.StandardSwitch
		if err := s.DB.Preload("Ports").First(&newSw, sw.ID).Error; err != nil {
			return fmt.Errorf("switch_not_found")
		}
		if err := editStandardBridge(*sw, newSw); err != nil {
			return err
		}
	}

	return nil
}

func createStandardBridge(sw networkModels.StandardSwitch) error {
	// Create
	raw, err := utils.RunCommand("ifconfig", "bridge", "create")
	if err != nil {
		return fmt.Errorf("create_standard_bridge: failed_to_create: %v", err)
	}

	raw = strings.TrimSpace(raw)

	// Rename
	if _, err := utils.RunCommand("ifconfig", raw, "name", sw.BridgeName); err != nil {
		return fmt.Errorf("create_standard_bridge: failed_to_rename: %v", err)
	}

	// Set description
	if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "descr", sw.Name); err != nil {
		return fmt.Errorf("create_standard_bridge: failed_to_set_descr: %v", err)
	}

	// Set MTU
	if sw.MTU != 0 {
		if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "mtu", strconv.Itoa(sw.MTU)); err != nil {
			return fmt.Errorf("create_standard_bridge: failed_to_set_bridge_mtu: %v", err)
		}
	}

	// Set Address4
	if sw.Address != "" {
		if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "inet", sw.Address); err != nil {
			return fmt.Errorf("create_standard_bridge: failed_to_set_bridge_address: %v", err)
		}
	}

	// Set Address6
	if sw.Address6 != "" {
		if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "inet6", sw.Address6); err != nil {
			return fmt.Errorf("create_standard_bridge: failed_to_set_bridge_address6: %v", err)
		}
	}

	// Configure Each Port
	for _, port := range sw.Ports {
		// Set MTU on port if bridge has MTU
		if sw.MTU > 0 {
			if _, err := utils.RunCommand("ifconfig", port.Name, "mtu", strconv.Itoa(sw.MTU)); err != nil {
				return fmt.Errorf("create_standard_bridge: failed_to_set_port_mtu: %v", err)
			}
		}

		if sw.VLAN > 0 {
			vif := fmt.Sprintf("%s.%d", port.Name, sw.VLAN)

			// Create VLAN interface if it doesn't exist
			if _, err := utils.RunCommand("ifconfig", vif); err != nil {
				args := []string{
					"vlan", "create",
					"vlandev", port.Name,
					"vlan", strconv.Itoa(sw.VLAN),
					"descr", fmt.Sprintf("svm-vlan/%s/%s", sw.BridgeName, vif),
					"name", vif,
					"group", "svm-vlan",
					"up",
				}

				if _, err := utils.RunCommand("ifconfig", args...); err != nil {
					return fmt.Errorf("create_standard_bridge: failed to create VLAN iface %s: %v", vif, err)
				}
			}

			if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "addm", vif, "up"); err != nil {
				return fmt.Errorf("create_standard_bridge: failed to add %s to bridge %s: %v", vif, sw.BridgeName, err)
			}
		} else {
			// Port without VLAN, just add it to bridge
			if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "addm", port.Name, "up"); err != nil {
				return fmt.Errorf("create_standard_bridge: failed to add %s to bridge %s: %v", port.Name, sw.BridgeName, err)
			}
		}
	}

	return nil
}

func editStandardBridge(oldSw, newSw networkModels.StandardSwitch) error {
	ifaceObj, err := iface.Get(oldSw.BridgeName)

	if err != nil {
		return fmt.Errorf("edit_standard_bridge: get %s: %v", oldSw.BridgeName, err)
	}

	var originalBM []string
	for _, m := range ifaceObj.BridgeMembers {
		originalBM = append(originalBM, m.Name)
	}

	if _, err := utils.RunCommand("ifconfig", oldSw.BridgeName, "destroy"); err != nil {
		return fmt.Errorf("edit_standard_bridge: failed to destroy %s: %v", oldSw.BridgeName, err)
	}

	var oldPorts []string
	for _, port := range oldSw.Ports {
		oldPorts = append(oldPorts, port.Name)
	}

	var taps []string

	for _, m := range originalBM {
		if !utils.Contains(oldPorts, m) {
			oIfaceObj, err := iface.Get(m)
			if err != nil {
				return fmt.Errorf("edit_standard_bridge: get %s: %v", m, err)
			}

			if strings.Contains(oIfaceObj.Driver, "tap") || utils.Contains(oIfaceObj.Groups, "tap") {
				taps = append(taps, m)
				continue
			}

			if _, err := utils.RunCommand("ifconfig", m, "destroy"); err != nil {
				return fmt.Errorf("edit_standard_bridge: failed to destroy %s: %v", m, err)
			}
		}
	}

	err = createStandardBridge(newSw)

	if err != nil {
		return fmt.Errorf("edit_standard_bridge: create %s: %v", newSw.BridgeName, err)
	}

	for _, tap := range taps {
		if _, err := utils.RunCommand("ifconfig", newSw.BridgeName, "addm", tap); err != nil {
			return fmt.Errorf("edit_standard_bridge: failed to add %s to bridge %s: %v", tap, newSw.BridgeName, err)
		}
	}

	return nil
}

func deleteStandardBridge(sw networkModels.StandardSwitch) error {
	if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "destroy"); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("delete_standard_bridge: failed_to_destroy: %v", err)
		}
	}

	for _, port := range sw.Ports {
		vif := fmt.Sprintf("%s.%d", port.Name, sw.VLAN)
		if _, err := utils.RunCommand("ifconfig", vif); err == nil {
			if _, err := utils.RunCommand("ifconfig", vif, "destroy"); err != nil {
				return fmt.Errorf("delete_standard_bridge: failed to destroy VLAN iface %s: %v", vif, err)
			}
		}
	}

	return nil
}
