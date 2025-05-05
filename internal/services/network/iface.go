package network

import (
	"fmt"
	networkModels "sylve/internal/db/models/network"
	"sylve/internal/logger"
	iface "sylve/pkg/network/iface"
	"sylve/pkg/utils"

	"gorm.io/gorm"
)

func (s *Service) List() ([]networkModels.NetworkInterface, error) {
	var interfaces []networkModels.NetworkInterface

	err := s.DB.
		Preload("IPv4s").
		Preload("IPv6s").
		Find(&interfaces).Error
	if err != nil {
		return nil, fmt.Errorf("failed_to_get_interfaces: %w", err)
	}

	for i := range interfaces {
		iface, err := iface.Get(interfaces[i].Name)
		if err != nil {
			logger.L.Warn().Msgf("Failed to get interface %s: %v", interfaces[i].Name, err)
			continue
		}
		interfaces[i].Interface = iface
	}

	return interfaces, nil
}

func (s *Service) SetupIPv4(
	name string,
	metric int,
	mtu int,
	protocol string,
	address string,
	netmask string,
	aliases [][]string,
) error {
	_, err := iface.Get(name)

	if err != nil {
		logger.L.Warn().Msgf("Failed to get interface %s: %v", name, err)
		err = s.DB.Where("name = ?", name).Delete(&networkModels.NetworkInterface{}).Error
		if err != nil {
			logger.L.Error().Msgf("Failed to delete interface %s from database: %v", name, err)
		}

		return fmt.Errorf("failed_to_get_iface_details: %w", err)
	}

	var ifaceModel networkModels.NetworkInterface

	err = s.DB.Where("name = ?", name).First(&ifaceModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("interface %s not found", name)
		}
		return err
	}

	if !utils.IsValidMTU(mtu) {
		return fmt.Errorf("invalid_mtu_value")
	}

	if !utils.IsValidMetric(metric) {
		return fmt.Errorf("invalid_metric_value")
	}

	ifaceModel.MTU = &mtu
	ifaceModel.Metric = &metric

	if protocol != "dhcp" {
		if !utils.IsValidIP(address) && address != "" {
			return fmt.Errorf("invalid_ip_address")
		}

		if !utils.IsValidIP(netmask) && netmask != "" {
			return fmt.Errorf("invalid_netmask")
		}
	}

	ipv4 := networkModels.IPv4{
		InterfaceID: ifaceModel.ID,
		Protocol:    protocol,
		Address:     &address,
		Netmask:     &netmask,
	}

	if err = s.DB.Create(&ipv4).Error; err != nil {
		return fmt.Errorf("failed_to_create_ipv4: %w", err)
	}

	for _, alias := range aliases {
		if len(alias) != 2 {
			return fmt.Errorf("invalid_alias_format")
		}
		if !utils.IsValidIP(alias[0]) {
			return fmt.Errorf("invalid_alias_ip")
		}
		if !utils.IsValidIP(alias[1]) {
			return fmt.Errorf("invalid_alias_netmask")
		}

		ipv4Alias := networkModels.IPv4{
			InterfaceID: ifaceModel.ID,
			Protocol:    "static",
			Address:     &alias[0],
			Netmask:     &alias[1],
			IsAlias:     true,
		}

		if err = s.DB.Create(&ipv4Alias).Error; err != nil {
			return fmt.Errorf("failed_to_create_ipv4_alias: %w", err)
		}
	}

	ifaceModel.IPv4s = append(ifaceModel.IPv4s, ipv4)

	err = s.DB.Save(&ifaceModel).Error
	if err != nil {
		return fmt.Errorf("saving_interface_failed: %w", err)
	}

	return nil
}
