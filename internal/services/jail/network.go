package jail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	jailModels "sylve/internal/db/models/jail"
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
	// active, err := s.IsJailActive(ctId)

	// if err != nil {
	// 	return err
	// }

	// if active {
	// 	return fmt.Errorf("jail %d is active, cannot inherit network", ctId)
	// }

	var jail jailModels.Jail

	if err := s.DB.Preload("Networks").First(&jail).Where("ct_id = ?", ctId).Error; err != nil {
		return err
	}

	jail.InheritIPv4 = ipv4
	jail.InheritIPv6 = ipv6

	return s.SyncNetwork(ctId, jail)
}

func (s *Service) DeleteNetwork(ctId uint, switchId uint, networkId uint) error {
	epair := fmt.Sprintf("%s_%d", utils.HashIntToNLetters(int(ctId), 5), switchId)
	err := s.NetworkService.DeleteEpair(epair)

	if err != nil {
		return err
	}

	var jailNetwork jailModels.Network
	err = s.DB.Where("id = ?", networkId).Delete(&jailNetwork).Error

	if err != nil {
		return err
	}

	return nil
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
				err = s.DeleteNetwork(ctId, network.SwitchID, network.ID)
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
		newCfg = cfg
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
