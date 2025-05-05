package network

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	networkModels "sylve/internal/db/models/network"
	networkServiceInterfaces "sylve/internal/interfaces/services/network"
	iface "sylve/pkg/network/iface"
	"sylve/pkg/rcconf"

	"gorm.io/gorm"
)

var _ networkServiceInterfaces.NetworkServiceInterface = (*Service)(nil)

type Service struct {
	DB *gorm.DB
}

func extractOptions(parts []string) (enabled []string, disabled []string) {
	for _, p := range parts {
		if strings.HasPrefix(p, "-") {
			disabled = append(disabled, p)
		} else if !strings.HasPrefix(p, "inet") && !strings.HasPrefix(p, "netmask") &&
			!strings.HasPrefix(p, "prefixlen") && !strings.Contains(p, ":") &&
			!strings.Contains(p, ".") {
			enabled = append(enabled, p)
		}
	}
	return
}

func (s *Service) ParseToDB() error {
	parsed, err := rcconf.Parse("/etc/rc.conf")
	if err != nil {
		return err
	}

	err = s.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&networkModels.NetworkInterface{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete interfaces: %w", err)
	}

	createdIfaces := make(map[string]*networkModels.NetworkInterface)

	knownKeywordList := []string{
		"dhcp", "inet", "inet6", "netmask", "prefixlen",
		"autoconf", "accept_rtadv", "alias",
	}

	knownKeywords := make(map[string]struct{}, len(knownKeywordList))
	for _, kw := range knownKeywordList {
		knownKeywords[kw] = struct{}{}
	}

	var isIPorNumber = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$|^[0-9]+$|^[a-f0-9:]+$`)

	extractOptions := func(parts []string) (enabled []string, disabled []string) {
		for _, p := range parts {
			pl := strings.ToLower(p)
			if _, skip := knownKeywords[pl]; skip || isIPorNumber.MatchString(pl) {
				continue
			}
			if strings.HasPrefix(p, "-") {
				disabled = append(disabled, p)
			} else {
				enabled = append(enabled, p)
			}
		}
		return
	}

	for key, val := range parsed {
		if !strings.HasPrefix(key, "ifconfig_") {
			continue
		}

		ifaceKey := strings.TrimPrefix(key, "ifconfig_")
		isIPv6 := strings.Contains(ifaceKey, "_ipv6")
		isAlias := strings.Contains(ifaceKey, "_alias")

		suffix := ifaceKey
		suffix = strings.ReplaceAll(suffix, "_ipv6", "")
		suffix = strings.SplitN(suffix, "_alias", 2)[0]
		ifaceName := suffix

		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}

		var ifaceModel *networkModels.NetworkInterface
		if cached, ok := createdIfaces[ifaceName]; ok {
			ifaceModel = cached
		} else {
			ifaceDetails, err := iface.Get(ifaceName)
			if err != nil {
				return err
			}

			model := &networkModels.NetworkInterface{
				Name:   ifaceDetails.Name,
				MAC:    ifaceDetails.Ether,
				MTU:    &ifaceDetails.MTU,
				Metric: &ifaceDetails.Metric,
			}

			if err := s.DB.Create(model).Error; err != nil {
				return fmt.Errorf("failed to create interface %s: %w", ifaceName, err)
			}

			createdIfaces[ifaceName] = model
			ifaceModel = model
		}

		valLower := strings.ToLower(val)
		parts := strings.Fields(val)
		enabled, disabled := extractOptions(parts)
		options := strings.Join(append(enabled, disabled...), " ")

		if isIPv6 {
			if strings.Contains(valLower, "accept_rtadv") {
				ipv6 := networkModels.IPv6{
					InterfaceID: ifaceModel.ID,
					Protocol:    "SLAAC",
					IsAlias:     isAlias,
					Options:     &options,
				}
				if err := s.DB.Create(&ipv6).Error; err != nil {
					return fmt.Errorf("failed to create SLAAC IPv6 for %s: %w", ifaceName, err)
				}
			} else if strings.Contains(valLower, "inet6") && strings.Contains(valLower, "prefixlen") {
				if len(parts) < 4 {
					return fmt.Errorf("invalid inet6 configuration for %s: %s", key, val)
				}
				addr := parts[1]
				prefixStr := parts[3]
				prefixLen, err := strconv.Atoi(prefixStr)
				if err != nil {
					return fmt.Errorf("invalid prefix length in %s: %v", key, err)
				}
				ipv6 := networkModels.IPv6{
					InterfaceID:  ifaceModel.ID,
					Protocol:     "Static",
					Address:      &addr,
					PrefixLength: &prefixLen,
					IsAlias:      isAlias,
					Options:      &options,
				}
				if err := s.DB.Create(&ipv6).Error; err != nil {
					return fmt.Errorf("failed to create static IPv6 for %s: %w", key, err)
				}
			}
			continue
		}

		if strings.Contains(valLower, "dhcp") {
			if isAlias {
				continue
			}
			ipv4 := networkModels.IPv4{
				InterfaceID: ifaceModel.ID,
				Protocol:    "DHCP",
				Options:     &options,
			}
			if err := s.DB.Create(&ipv4).Error; err != nil {
				return fmt.Errorf("failed to create DHCP IPv4 for %s: %w", ifaceName, err)
			}
			continue
		}

		if strings.Contains(valLower, "inet") && strings.Contains(valLower, "netmask") {
			if len(parts) < 4 {
				return fmt.Errorf("invalid inet configuration for %s: %s", key, val)
			}
			ipv4 := networkModels.IPv4{
				InterfaceID: ifaceModel.ID,
				Protocol:    "Static",
				Address:     &parts[1],
				Netmask:     &parts[3],
				IsAlias:     isAlias,
				Options:     &options,
			}
			if err := s.DB.Create(&ipv4).Error; err != nil {
				return fmt.Errorf("failed to create static IPv4 for %s: %w", key, err)
			}
		}
	}

	return nil
}

func (s *Service) SyncToRC() error {
	return nil
}

func NewNetworkService(db *gorm.DB) networkServiceInterfaces.NetworkServiceInterface {
	return &Service{
		DB: db,
	}
}
