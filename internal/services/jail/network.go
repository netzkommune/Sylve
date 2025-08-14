package jail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	jailModels "sylve/internal/db/models/jail"
	networkModels "sylve/internal/db/models/network"
	"sylve/pkg/utils"
)

func (s *Service) DisinheritNetwork(ctId uint) error {
	var jail jailModels.Jail

	if err := s.DB.Preload("Networks").First(&jail).Where("ct_id = ?", ctId).Error; err != nil {
		return err
	}

	jail.InheritIPv4 = false
	jail.InheritIPv6 = false

	return s.SyncNetwork(ctId, jail)
}

func (s *Service) InheritNetwork(ctId uint, ipv4 bool, ipv6 bool) error {
	var jail jailModels.Jail

	if err := s.DB.Preload("Networks").First(&jail).Where("ct_id = ?", ctId).Error; err != nil {
		return err
	}

	jail.InheritIPv4 = ipv4
	jail.InheritIPv6 = ipv6

	return s.SyncNetwork(ctId, jail)
}

func (s *Service) AddNetwork(ctId uint,
	switchId uint,
	macId uint,
	ip4 uint,
	ip4gw uint,
	ip6 uint,
	ip6gw uint,
	dhcp bool,
	slaac bool) error {
	var jail jailModels.Jail
	var network jailModels.Network

	if err := s.DB.Preload("Networks").First(&jail).Where("ct_id = ?", ctId).Error; err != nil {
		return err
	}

	if jail.InheritIPv4 || jail.InheritIPv6 {
		return fmt.Errorf("cannot_add_network_when_inheriting_network")
	}

	for _, network := range jail.Networks {
		if network.SwitchID == switchId {
			return fmt.Errorf("switch_id_already_used_by_jail")
		}
	}

	network.SwitchID = switchId

	if !dhcp {
		if ip4 == 0 || ip4gw == 0 {
			return fmt.Errorf("ip4_and_ip4gw_must_be_specified_when_dhcp_is_disabled")
		}

		_, err := s.NetworkService.GetObjectEntryByID(ip4)
		if err != nil {
			return fmt.Errorf("failed_to_get_ip4_object: %w", err)
		}

		_, err = s.NetworkService.GetObjectEntryByID(ip4gw)
		if err != nil {
			return fmt.Errorf("failed_to_get_ip4gw_object: %w", err)
		}

		network.IPv4ID = &ip4
		network.IPv4GwID = &ip4gw
	} else {
		network.DHCP = true
	}

	if !slaac {
		if ip6 == 0 || ip6gw == 0 {
			return fmt.Errorf("ip6_and_ip6gw_must_be_specified_when_slaac_is_disabled")
		}

		_, err := s.NetworkService.GetObjectEntryByID(ip6)
		if err != nil {
			return fmt.Errorf("failed_to_get_ip6_object: %w", err)
		}

		_, err = s.NetworkService.GetObjectEntryByID(ip6gw)
		if err != nil {
			return fmt.Errorf("failed_to_get_ip6gw_object: %w", err)
		}

		network.IPv6ID = &ip6
		network.IPv6GwID = &ip6gw
	} else {
		network.SLAAC = true
	}

	if macId == 0 {
		var sw networkModels.StandardSwitch
		if err := s.DB.First(&sw, "id = ?", switchId).Error; err != nil {
			return fmt.Errorf("failed_to_find_switch: %w", err)
		}

		macAddress := utils.GenerateRandomMAC()
		base := fmt.Sprintf("%s-%s", jail.Name, sw.Name)
		name := base

		for i := 0; ; i++ {
			if i > 0 {
				name = fmt.Sprintf("%s-%d", base, i)
			}

			var exists int64

			if err := s.DB.
				Model(&networkModels.Object{}).
				Where("name = ?", name).
				Limit(1).
				Count(&exists).Error; err != nil {
				return fmt.Errorf("failed_to_check_mac_object_exists: %w", err)
			}

			if exists == 0 {
				break
			}
		}

		macObj := networkModels.Object{
			Name: name,
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

		network.MacID = &macObj.ID
	} else {
		_, err := s.NetworkService.GetObjectEntryByID(macId)
		if err != nil {
			return fmt.Errorf("failed_to_get_mac_object: %w", err)
		}

		network.MacID = &macId
	}

	network.CTID = ctId
	err := s.DB.Create(&network).Error
	if err != nil {
		return fmt.Errorf("failed_to_create_network: %w", err)
	}

	jail.Networks = append(jail.Networks, network)

	err = s.NetworkService.SyncEpairs()

	if err != nil {
		return fmt.Errorf("failed_to_sync_epairs: %w", err)
	}

	return s.SyncNetwork(ctId, jail)
}

func (s *Service) DeleteNetwork(ctId uint, networkId uint) error {
	var network jailModels.Network
	err := s.DB.Find(&network, networkId).Error
	if err != nil {
		return fmt.Errorf("failed_to_find_network: %w", err)
	}

	epair := fmt.Sprintf("%s_%d", utils.HashIntToNLetters(int(ctId), 5), network.SwitchID)
	err = s.NetworkService.DeleteEpair(epair)
	if err != nil {
		return err
	}

	var jailNetwork jailModels.Network
	err = s.DB.Where("id = ?", networkId).Delete(&jailNetwork).Error

	if err != nil {
		return err
	}

	var jail jailModels.Jail

	err = s.DB.Preload("Networks").First(&jail).Where("ct_id = ?", ctId).Error
	if err != nil {
		return err
	}

	return s.SyncNetwork(ctId, jail)
}

func (s *Service) GetNetworkCleanedConfig(ctId uint) (string, error) {
	cfg, err := s.GetJailConfig(ctId)
	if err != nil {
		return "", err
	}

	lines := strings.Split(cfg, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "ip4=") ||
			strings.Contains(lines[i], "ip6=") ||
			strings.Contains(lines[i], "vnet;") ||
			strings.Contains(lines[i], "vnet.interface") ||
			strings.Contains(lines[i], "ifconfig") ||
			strings.Contains(lines[i], "route add default") ||
			strings.Contains(lines[i], "sysrc ipv6") ||
			strings.Contains(lines[i], "dhclient") {
			lines = append(lines[:i], lines[i+1:]...)
			i--
		}
	}

	cfg = strings.Join(lines, "\n")

	return cfg, nil
}

func (s *Service) SyncNetwork(ctId uint, jail jailModels.Jail) error {
	err := s.DB.Save(&jail).Error
	if err != nil {
		return err
	}

	cfg, err := s.GetNetworkCleanedConfig(ctId)
	if err != nil {
		return err
	}

	var newCfg string

	/* Moving from VNET to Inherited */
	if jail.InheritIPv4 || jail.InheritIPv6 {
		if jail.Networks != nil && len(jail.Networks) > 0 {
			for _, network := range jail.Networks {
				err = s.DeleteNetwork(ctId, network.ID)
				if err != nil {
					return err
				}
			}
		}

		toAppend := ""

		if jail.InheritIPv4 {
			toAppend += fmt.Sprintf("\tip4=inherit;\n")
		}

		if jail.InheritIPv6 {
			toAppend += fmt.Sprintf("\tip6=inherit;\n")
		}

		newCfg, err = s.AppendToConfig(ctId, cfg, toAppend)
		if err != nil {
			return err
		}
	} else {
		if jail.Networks != nil && len(jail.Networks) > 0 {
			ctidHash := utils.HashIntToNLetters(int(ctId), 5)

			// Ensure epairs exist
			if err := s.NetworkService.SyncEpairs(); err != nil {
				return err
			}

			var b strings.Builder

			// vnet declaration once
			b.WriteString("\tvnet;\n")

			// Add one vnet.interface line per NIC
			for _, n := range jail.Networks {
				if n.SwitchID == 0 {
					continue
				}
				b.WriteString(fmt.Sprintf("\tvnet.interface += \"%s_%db\";\n", ctidHash, n.SwitchID))
			}

			// Guard: only set default routes once
			setV4Default := false
			setV6Default := false

			for _, n := range jail.Networks {
				if n.SwitchID == 0 {
					continue
				}

				networkId := n.SwitchID

				// MAC + Bridge membership
				if n.MacID != nil && *n.MacID > 0 {
					mac, err := s.NetworkService.GetObjectEntryByID(*n.MacID)
					if err != nil {
						return fmt.Errorf("failed to get mac address: %w", err)
					}
					prevMAC, err := utils.PreviousMAC(mac)
					if err != nil {
						return fmt.Errorf("failed to get previous mac: %w", err)
					}

					b.WriteString(fmt.Sprintf("\texec.prestart += \"ifconfig %s_%da ether %s up\";\n", ctidHash, networkId, prevMAC))
					b.WriteString(fmt.Sprintf("\texec.prestart += \"ifconfig %s_%db ether %s up\";\n", ctidHash, networkId, mac))

					bridgeName, err := s.NetworkService.GetBridgeNameByID(n.SwitchID)
					if err != nil {
						return fmt.Errorf("failed to get bridge name: %w", err)
					}
					b.WriteString(fmt.Sprintf(
						"\texec.prestart += \"if ! ifconfig %s | grep -qw %s_%da; then ifconfig %s addm %s_%da; fi\";\n",
						bridgeName, ctidHash, networkId, bridgeName, ctidHash, networkId,
					))
				}

				// Addressing
				switch {
				case n.DHCP && n.SLAAC:
					b.WriteString(fmt.Sprintf("\texec.start += \"dhclient %s_%db\";\n", ctidHash, networkId))
					b.WriteString(fmt.Sprintf("\texec.start += \"sysrc ifconfig_%s_%db=\\\"DHCP\\\"\";\n", ctidHash, networkId))
					b.WriteString(fmt.Sprintf("\texec.start += \"sysrc ifconfig_%s_%db_ipv6=\\\"inet6 accept_rtadv\\\"\";\n", ctidHash, networkId))

				case n.DHCP:
					b.WriteString(fmt.Sprintf("\texec.start += \"dhclient %s_%db\";\n", ctidHash, networkId))
					b.WriteString(fmt.Sprintf("\texec.start += \"sysrc ifconfig_%s_%db=\\\"DHCP\\\"\";\n", ctidHash, networkId))

				case n.SLAAC:
					b.WriteString(fmt.Sprintf("\texec.start += \"ifconfig %s_%db inet6 accept_rtadv up\";\n", ctidHash, networkId))
					b.WriteString(fmt.Sprintf("\texec.start += \"sysrc ifconfig_%s_%db_ipv6=\\\"inet6 accept_rtadv\\\"\";\n", ctidHash, networkId))

				default:
					// Static IPv4
					if n.IPv4ID != nil && *n.IPv4ID > 0 && n.IPv4GwID != nil && *n.IPv4GwID > 0 {
						ipv4, err := s.NetworkService.GetObjectEntryByID(*n.IPv4ID)
						if err != nil {
							return fmt.Errorf("failed to get ipv4 address: %w", err)
						}
						ipv4Gw, err := s.NetworkService.GetObjectEntryByID(*n.IPv4GwID)
						if err != nil {
							return fmt.Errorf("failed to get ipv4 gateway: %w", err)
						}
						ip, mask, err := utils.SplitIPv4AndMask(ipv4)
						if err != nil {
							return fmt.Errorf("failed to split ipv4 address and mask: %w", err)
						}

						b.WriteString(fmt.Sprintf("\texec.start += \"ifconfig %s_%db inet %s netmask %s\";\n", ctidHash, networkId, ip, mask))
						if !setV4Default {
							b.WriteString(fmt.Sprintf("\texec.start += \"route add default %s\";\n", ipv4Gw))
							setV4Default = true
						}
						b.WriteString(fmt.Sprintf("\texec.start += \"sysrc ifconfig_%s_%db=\\\"inet %s netmask %s\\\"\";\n", ctidHash, networkId, ip, mask))
					}

					// Static IPv6
					if n.IPv6ID != nil && *n.IPv6ID > 0 && n.IPv6GwID != nil && *n.IPv6GwID > 0 {
						ipv6, err := s.NetworkService.GetObjectEntryByID(*n.IPv6ID)
						if err != nil {
							return fmt.Errorf("failed to get ipv6 address: %w", err)
						}
						ipv6Gw, err := s.NetworkService.GetObjectEntryByID(*n.IPv6GwID)
						if err != nil {
							return fmt.Errorf("failed to get ipv6 gateway: %w", err)
						}

						b.WriteString(fmt.Sprintf("\texec.start += \"ifconfig %s_%db inet6 %s\";\n", ctidHash, networkId, ipv6))
						if !setV6Default {
							b.WriteString(fmt.Sprintf("\texec.start += \"sysrc ipv6_defaultrouter=\\\"%s\\\"\";\n", ipv6Gw))
							setV6Default = true
						}
						b.WriteString(fmt.Sprintf("\texec.start += \"sysrc ifconfig_%s_%db_ipv6=\\\"inet6 %s\\\"\";\n", ctidHash, networkId, ipv6))
					}
				}
			}

			newCfg, err = s.AppendToConfig(ctId, cfg, b.String())
			if err != nil {
				return err
			}
		} else {
			toAppend := "\tip4=disable;\n\tip6=disable;\n"
			newCfg, err = s.AppendToConfig(ctId, cfg, toAppend)
			if err != nil {
				return err
			}
		}
	}

	err = s.SaveJailConfig(ctId, newCfg)

	if err != nil {
		return err
	}

	mountPoint, err := s.GetJailMountPoint(ctId)
	if err != nil {
		return err
	}

	rcConfPath := filepath.Join(mountPoint, "etc", "rc.conf")

	var exists bool
	if _, err := os.Stat(rcConfPath); err == nil {
		exists = true
	}

	if exists {
		if jail.InheritIPv4 || jail.InheritIPv6 {
			rcConf, err := os.ReadFile(rcConfPath)
			if err != nil {
				return err
			}

			lines := strings.Split(string(rcConf), "\n")
			for i := 0; i < len(lines); i++ {
				if strings.HasPrefix(lines[i], "ifconfig") ||
					strings.HasPrefix(lines[i], "ipv6") {
					lines = append(lines[:i], lines[i+1:]...)
					i--
				}
			}

			err = os.WriteFile(rcConfPath, []byte(strings.Join(lines, "\n")), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
