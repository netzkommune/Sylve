package network

import (
	"fmt"
	"strconv"
	networkModels "sylve/internal/db/models/network"
	utils "sylve/pkg/utils"
)

func (s *Service) GetObjects() ([]networkModels.Object, error) {
	var objects []networkModels.Object

	err := s.DB.
		Preload("Entries").
		Preload("Resolutions").
		Find(&objects).Error

	if err != nil {
		return nil, err
	}

	return objects, nil
}

func validateType(oType string) error {
	validTypes := map[string]bool{
		"Host":    true,
		"Network": true,
		"Port":    true,
		"Country": true,
		"List":    true,
		"FQDN":    true,
	}

	if !validTypes[oType] {
		return fmt.Errorf("invalid object type: %s", oType)
	}

	return nil
}

func validateValues(oType string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("values cannot be empty for type: %s", oType)
	}

	if oType == "Host" {
		isIPv4 := false
		isIPv6 := false

		for _, value := range values {
			if utils.IsValidIPv4(value) {
				isIPv4 = true
			} else if utils.IsValidIPv6(value) {
				isIPv6 = true
			} else {
				return fmt.Errorf("invalid host value: %s", value)
			}

			if isIPv4 && isIPv6 {
				return fmt.Errorf("cannot mix IPv4 and IPv6 in host values")
			}
		}
	}

	if oType == "Network" {
		isIPv4 := false
		isIPv6 := false

		for _, value := range values {
			if utils.IsValidIPv4CIDR(value) {
				isIPv4 = true
			} else if utils.IsValidIPv6CIDR(value) {
				isIPv6 = true
			} else {
				return fmt.Errorf("invalid network value: %s", value)
			}

			if isIPv4 && isIPv6 {
				return fmt.Errorf("cannot mix IPv4 and IPv6 in network values")
			}
		}
	}

	for _, value := range values {
		if value == "" {
			return fmt.Errorf("value cannot be empty for type: %s", oType)
		}

		if oType == "Port" {
			vInt, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid port value: %s", value)
			}

			if !utils.IsValidPort(vInt) {
				return fmt.Errorf("invalid port value: %s", value)
			}
		}

		if oType == "Country" {
			if !utils.IsValidCountryCode(value) {
				return fmt.Errorf("invalid country code: %s", value)
			}
		}

		if oType == "MAC" {
			if !utils.IsValidMAC(value) {
				return fmt.Errorf("invalid MAC address: %s", value)
			}
		}
	}

	return nil
}

func (s *Service) CreateObject(name string, oType string, values []string) error {
	if err := validateType(oType); err != nil {
		return err
	}

	if err := validateValues(oType, values); err != nil {
		return err
	}

	entries := make([]networkModels.ObjectEntry, len(values))
	for i, value := range values {
		entries[i] = networkModels.ObjectEntry{
			Value: value,
		}
	}

	object := networkModels.Object{
		Name:    name,
		Type:    oType,
		Entries: entries,
	}

	if err := s.DB.Create(&object).Error; err != nil {
		return err
	}

	return nil
}
