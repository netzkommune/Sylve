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
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	switch action {
	case "sync":
		var switches []networkModels.StandardSwitch
		if err := s.DB.Preload("Ports").Find(&switches).Error; err != nil {
			return fmt.Errorf("db_error_checking_switches: %v", err)
		}
		for _, switchObj := range switches {
			if err := deleteStandardBridge(switchObj); err != nil {
				return fmt.Errorf("sync_standard_switches: failed_to_delete: %v", err)
			}
		}
		for _, switchObj := range switches {
			if err := createStandardBridge(switchObj); err != nil {
				return fmt.Errorf("sync_standard_switches: failed_to_create: %v", err)
			}
		}

	case "create":
		if err := createStandardBridge(*sw); err != nil {
			return err
		}

	case "delete":
		if err := deleteStandardBridge(*sw); err != nil {
			return err
		}

	case "edit":
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

	for _, port := range sw.Ports {
		if err := addBridgeMember(sw.BridgeName, port.Name, sw.MTU, sw.VLAN); err != nil {
			return fmt.Errorf("create_standard_bridge: %v", err)
		}
	}

	return nil
}

func editStandardBridge(oldSw, newSw networkModels.StandardSwitch) error {
	br := oldSw.BridgeName

	// 1) snapshot existing members
	ifaceObj, err := iface.Get(br)
	if err != nil {
		return fmt.Errorf("edit_standard_bridge: get %s: %v", br, err)
	}
	var original []string
	for _, m := range ifaceObj.BridgeMembers {
		original = append(original, m.Name)
	}

	// 2) build sets of old & new DB ports (incl. VLAN ifaces)
	oldSet := make(map[string]bool, len(oldSw.Ports)*2)
	for _, p := range oldSw.Ports {
		oldSet[p.Name] = true
		if oldSw.VLAN > 0 {
			oldSet[fmt.Sprintf("%s.%d", p.Name, oldSw.VLAN)] = true
		}
	}
	newSet := make(map[string]bool, len(newSw.Ports)*2)
	for _, p := range newSw.Ports {
		newSet[p.Name] = true
		if newSw.VLAN > 0 {
			newSet[fmt.Sprintf("%s.%d", p.Name, newSw.VLAN)] = true
		}
	}

	// 3) remove only the *old* DB ports (and their VLAN sub-ifs)
	for _, p := range oldSw.Ports {
		if err := removeBridgeMember(br, p.Name, oldSw.VLAN); err != nil {
			return fmt.Errorf("edit_standard_bridge: remove old port %s: %v", p.Name, err)
		}
	}

	// 4) reconfigure bridge in place
	if _, err := utils.RunCommand("ifconfig", br, "descr", newSw.Name); err != nil {
		return fmt.Errorf("edit_standard_bridge: set descr: %v", err)
	}
	if oldSw.MTU != newSw.MTU && newSw.MTU > 0 {
		if _, err := utils.RunCommand("ifconfig", br, "mtu", strconv.Itoa(newSw.MTU)); err != nil {
			return fmt.Errorf("edit_standard_bridge: set mtu: %v", err)
		}
	}
	// IPv4
	if oldSw.Address != newSw.Address {
		if oldSw.Address != "" {
			if _, err := utils.RunCommand("ifconfig", br, "inet", oldSw.Address, "delete"); err != nil {
				return fmt.Errorf("edit_standard_bridge: del old inet: %v", err)
			}
		}
		if newSw.Address != "" {
			if _, err := utils.RunCommand("ifconfig", br, "inet", newSw.Address); err != nil {
				return fmt.Errorf("edit_standard_bridge: set inet: %v", err)
			}
		}
	}
	// IPv6
	if oldSw.Address6 != newSw.Address6 {
		if oldSw.Address6 != "" {
			if _, err := utils.RunCommand("ifconfig", br, "inet6", oldSw.Address6, "delete"); err != nil {
				return fmt.Errorf("edit_standard_bridge: del old inet6: %v", err)
			}
		}
		if newSw.Address6 != "" {
			if _, err := utils.RunCommand("ifconfig", br, "inet6", newSw.Address6); err != nil {
				return fmt.Errorf("edit_standard_bridge: set inet6: %v", err)
			}
		}
	}

	// 5) add the *new* DB ports (and VLAN sub-ifs)
	for _, p := range newSw.Ports {
		if err := addBridgeMember(br, p.Name, newSw.MTU, newSw.VLAN); err != nil {
			return fmt.Errorf("edit_standard_bridge: add port %s: %v", p.Name, err)
		}
	}

	// 6) re-attach only non-DB members (e.g. taps), skip old/new DB ports
	for _, m := range original {
		if oldSet[m] || newSet[m] {
			continue
		}

		oif, err := iface.Get(m)
		if err != nil {
			continue
		}
		if strings.Contains(oif.Driver, "tap") || utils.Contains(oif.Groups, "tap") {
			if _, err := utils.RunCommand("ifconfig", br, "addm", m, "up"); err != nil {
				return fmt.Errorf("edit_standard_bridge: re-add tap %s: %v", m, err)
			}
		}
	}

	return nil
}

func deleteStandardBridge(sw networkModels.StandardSwitch) error {
	if _, err := utils.RunCommand("ifconfig", sw.BridgeName, "destroy"); err != nil {
		if !strings.Contains(err.Error(), "does not exist") {
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

func addBridgeMember(br, portName string, mtu, vlan int) error {
	// set port MTU
	if mtu > 0 {
		if _, err := utils.RunCommand("ifconfig", portName, "mtu", strconv.Itoa(mtu)); err != nil {
			return fmt.Errorf("set mtu for %s: %v", portName, err)
		}
	}
	if vlan > 0 {
		vif := fmt.Sprintf("%s.%d", portName, vlan)
		// create VLAN iface if missing
		if _, err := utils.RunCommand("ifconfig", vif); err != nil {
			args := []string{
				"vlan", "create",
				"vlandev", portName,
				"vlan", strconv.Itoa(vlan),
				"descr", fmt.Sprintf("svm-vlan/%s/%s", br, vif),
				"name", vif,
				"group", "svm-vlan",
				"up",
			}
			if _, err := utils.RunCommand("ifconfig", args...); err != nil {
				return fmt.Errorf("create vlan %s: %v", vif, err)
			}
		}
		// add VLAN iface to bridge
		if _, err := utils.RunCommand("ifconfig", br, "addm", vif, "up"); err != nil {
			return fmt.Errorf("add vlan %s: %v", vif, err)
		}
	} else {
		// add plain port to bridge
		if _, err := utils.RunCommand("ifconfig", br, "addm", portName, "up"); err != nil {
			return fmt.Errorf("add port %s: %v", portName, err)
		}
	}
	return nil
}

func removeBridgeMember(br, portName string, vlan int) error {
	if vlan > 0 {
		vif := fmt.Sprintf("%s.%d", portName, vlan)
		// remove from bridge
		if _, err := utils.RunCommand("ifconfig", br, "deletem", vif); err != nil {
			return fmt.Errorf("remove vlan member %s: %v", vif, err)
		}
		// destroy VLAN iface
		if _, err := utils.RunCommand("ifconfig", vif, "destroy"); err != nil {
			return fmt.Errorf("destroy vlan iface %s: %v", vif, err)
		}
	} else {
		// remove plain port
		if _, err := utils.RunCommand("ifconfig", br, "deletem", portName); err != nil {
			return fmt.Errorf("remove port member %s: %v", portName, err)
		}
	}
	return nil
}
