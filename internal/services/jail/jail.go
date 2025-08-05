// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package jail

import (
	"fmt"
	"os"
	"path/filepath"
	"sylve/internal/config"
	jailModels "sylve/internal/db/models/jail"
	utilitiesModels "sylve/internal/db/models/utilities"
	jailServiceInterfaces "sylve/internal/interfaces/services/jail"
	networkServiceInterfaces "sylve/internal/interfaces/services/network"
	"sylve/internal/logger"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"

	"gorm.io/gorm"

	sdb "sylve/internal/db"
)

var _ jailServiceInterfaces.JailServiceInterface = (*Service)(nil)

type Service struct {
	DB             *gorm.DB
	NetworkService networkServiceInterfaces.NetworkServiceInterface
}

func NewJailService(db *gorm.DB, networkService networkServiceInterfaces.NetworkServiceInterface) jailServiceInterfaces.JailServiceInterface {
	return &Service{
		DB:             db,
		NetworkService: networkService,
	}
}

func (s *Service) GetJails() ([]jailModels.Jail, error) {
	var jails []jailModels.Jail
	if err := s.DB.Preload("Networks").Find(&jails).Error; err != nil {
		logger.L.Error().Err(err).Msg("get_jails: failed to fetch jails")
		return nil, fmt.Errorf("failed_to_fetch_jails: %w", err)
	}
	return jails, nil
}

func (s *Service) GetJailsSimple() ([]jailServiceInterfaces.SimpleList, error) {
	var jails []jailModels.Jail

	if err := s.DB.Model(&jailModels.Jail{}).Select("id, name, ct_id").Find(&jails).Error; err != nil {
		logger.L.Error().Err(err).Msg("get_jails_simple: failed to fetch jails")
		return nil, fmt.Errorf("failed_to_fetch_jails_simple: %w", err)
	}

	var list []jailServiceInterfaces.SimpleList

	for _, jail := range jails {
		list = append(list, jailServiceInterfaces.SimpleList{
			ID:   jail.ID,
			Name: jail.Name,
			CTID: jail.CTID,
		})
	}

	return list, nil
}

func (s *Service) ValidateCreate(data jailServiceInterfaces.CreateJailRequest) error {
	if data.Name == "" || !utils.IsValidVMName(data.Name) {
		return fmt.Errorf("invalid_vm_name")
	}

	if data.CTID == nil || *data.CTID <= 0 || *data.CTID > 9999 {
		return fmt.Errorf("invalid_ct_id")
	}

	if data.Description != "" && (len(data.Description) < 1 || len(data.Description) > 1024) {
		return fmt.Errorf("invalid_description")
	}

	datasets, err := zfs.Datasets("")
	if err != nil {
		return fmt.Errorf("failed_to_get_dataset: %w", err)
	}

	if datasets == nil {
		return fmt.Errorf("dataset_not_found")
	}

	var dataset *zfs.Dataset

	for _, d := range datasets {
		guid, err := d.GetProperty("guid")
		if err != nil {
			return fmt.Errorf("failed_to_get_dataset_properties: %w", err)
		}

		if guid == data.Dataset {
			dataset = d
		}
	}

	if dataset == nil {
		return fmt.Errorf("dataset_not_found")
	}

	mountPoint, err := dataset.GetProperty("mountpoint")
	if err != nil {
		return fmt.Errorf("failed_to_get_dataset_mountpoint: %w", err)
	}

	if mountPoint == "" {
		return fmt.Errorf("dataset_mountpoint_not_found")
	}

	if emptyDir := utils.IsEmptyDir(mountPoint); !emptyDir {
		return fmt.Errorf("dataset_mountpoint_not_empty: %w", err)
	}

	if data.Base == "" {
		return fmt.Errorf("base_download_uuid_required")
	}

	dCount, err := sdb.Count(s.DB, &utilitiesModels.Downloads{}, "uuid = ?", data.Base)
	if err != nil {
		return fmt.Errorf("failed_to_count_downloads: %w", err)
	}

	if dCount == 0 {
		return fmt.Errorf("iso_not_found")
	}

	_, err = s.FindBaseByUUID(data.Base)

	if err != nil {
		return fmt.Errorf("failed_to_find_base_by_uuid: %w", err)
	}

	switchId := uint(0)
	mac := uint(0)
	dhcp := false
	slaac := false

	if data.SwitchId != nil {
		switchId = uint(*data.SwitchId)
	}

	if data.MAC != nil {
		mac = uint(*data.MAC)
		if mac == 0 && switchId != 0 {
			return fmt.Errorf("mac_required_if_switch_id_provided")
		} else {
			if mac != 0 {
				used, err := s.NetworkService.IsObjectUsed(mac)
				if err != nil {
					return fmt.Errorf("failed_to_check_mac_usage: %w", err)
				}

				if used {
					return fmt.Errorf("mac_already_used")
				}
			}
		}
	}

	if data.DHCP != nil {
		dhcp = *data.DHCP
	}

	if data.SLAAC != nil {
		slaac = *data.SLAAC
	}

	if switchId != 0 {
		if !dhcp {
			if data.IPv4 != nil {
				ipv4Id := uint(*data.IPv4)
				if ipv4Id != 0 && data.IPv4Gw != nil {
					ipv4GwId := uint(*data.IPv4Gw)
					if ipv4GwId == 0 {
						return fmt.Errorf("invalid_ipv4_gateway")
					}

					isUsed, err := s.NetworkService.IsObjectUsed(ipv4Id)
					if err != nil {
						return fmt.Errorf("failed_to_check_ipv4_usage: %w", err)
					}

					if isUsed {
						return fmt.Errorf("ipv4_already_used")
					}
				}
			}
		}

		if !slaac {
			if data.IPv6 != nil {
				ipv6Id := uint(*data.IPv6)

				if ipv6Id != 0 && data.IPv6Gw != nil {
					ipv6GwId := uint(*data.IPv6Gw)
					if ipv6GwId == 0 {
						return fmt.Errorf("invalid_ipv6_gateway")
					}

					isUsed, err := s.NetworkService.IsObjectUsed(ipv6Id)
					if err != nil {
						return fmt.Errorf("failed_to_check_ipv6_usage: %w", err)
					}

					if isUsed {
						return fmt.Errorf("ipv6_already_used")
					}
				}
			}
		}
	}

	if data.StartOrder < 0 {
		return fmt.Errorf("start_order_must_be_greater_than_or_equal_to_0")
	}

	return nil
}

func (s *Service) CreateJailConfig(data jailServiceInterfaces.CreateJailRequest) (string, error) {
	var dataset *zfs.Dataset
	datasets, err := zfs.Datasets("")
	if err != nil {
		return "", fmt.Errorf("failed_to_get_datasets: %w", err)
	}

	for _, d := range datasets {
		guid, err := d.GetProperty("guid")
		if err != nil {
			return "", fmt.Errorf("failed_to_get_dataset_properties: %w", err)
		}

		if guid == data.Dataset {
			dataset = d
			break
		}
	}

	if dataset == nil {
		return "", fmt.Errorf("dataset_not_found")
	}

	var mountPoint string
	mountPoint, err = dataset.GetProperty("mountpoint")
	if err != nil {
		return "", fmt.Errorf("failed_to_get_dataset_mountpoint: %w", err)
	}

	var config string
	ctid := *data.CTID
	config += fmt.Sprintf("%d {\n", ctid)
	config += fmt.Sprintf("\t$ctid = \"%d\";\n", ctid)
	config += fmt.Sprintf("\tpath = \"%s\";\n", mountPoint)
	config += fmt.Sprintf("\thost.hostname = \"%s\";\n", utils.MakeValidHostname(data.Name))
	config += fmt.Sprintf("\tpersist;\n")
	config += fmt.Sprintf("\texec.clean;\n\n")

	config += fmt.Sprintf("\tmount.devfs;\n")
	config += fmt.Sprintf("\tdevfs_ruleset=\"8181\";\n\n")

	config += fmt.Sprintf("\tallow.raw_sockets;\n")
	config += fmt.Sprintf("\tallow.socket_af;\n\n")

	config += fmt.Sprintf("\texec.start = \"/bin/sh /etc/rc\";\n")
	config += fmt.Sprintf("\texec.stop = \"/bin/sh /etc/rc.shutdown\";\n")

	config += fmt.Sprintf("}\n")

	baseTxz, err := s.FindBaseByUUID(data.Base)
	if err != nil {
		logger.L.Error().Err(err).Msg("create_jail: failed to find base")
		return "", fmt.Errorf("failed_to_find_base: %w", err)
	}

	output, err := s.ExtractBase(mountPoint, baseTxz)
	if err != nil {
		logger.L.Error().Err(err).Msg("create_jail: failed to extract base")
		return "", fmt.Errorf("failed_to_extract_base: %w", err)
	}

	logger.L.Info().Msgf("Base extracted successfully: %s", output)

	return config, nil
}

func (s *Service) CreateJail(data jailServiceInterfaces.CreateJailRequest) error {
	if err := s.ValidateCreate(data); err != nil {
		logger.L.Debug().Err(err).Msg("create_jail: validation failed")
		return err
	}

	jCfg, err := s.CreateJailConfig(data)
	if err != nil {
		logger.L.Error().Err(err).Msg("create_jail: failed to create jail config")
		return fmt.Errorf("failed_to_create_jail_config: %w", err)
	}

	jailsPath, err := config.GetJailsPath()
	if jailsPath == "" {
		logger.L.Error().Msg("create_jail: jails path is empty")
		return fmt.Errorf("jails_path_not_found")
	}

	if err != nil {
		logger.L.Error().Err(err).Msg("create_jail: failed to get jails path")
		return fmt.Errorf("failed_to_get_jails_path: %w", err)
	}

	jailDir := filepath.Join(jailsPath, fmt.Sprintf("%d", *data.CTID))
	if err := os.MkdirAll(jailDir, 0755); err != nil {
		logger.L.Error().Err(err).Msg("create_jail: failed to create jail directory")
		return fmt.Errorf("failed_to_create_jail_directory: %w", err)
	}

	jailConfigPath := filepath.Join(jailDir, fmt.Sprintf("%d.conf", *data.CTID))
	if err := os.WriteFile(jailConfigPath, []byte(jCfg), 0644); err != nil {
		logger.L.Error().Err(err).Msg("create_jail: failed to write jail config file")
		return fmt.Errorf("failed_to_write_jail_config_file: %w", err)
	}

	var jail jailModels.Jail

	jail.Name = data.Name
	jail.CTID = *data.CTID
	jail.Description = data.Description
	jail.Dataset = data.Dataset
	jail.Base = data.Base
	jail.StartAtBoot = *data.StartAtBoot
	jail.StartOrder = data.StartOrder

	if data.SwitchId != nil && *data.SwitchId > 0 {
		var mac uint
		if data.MAC != nil {
			mac = uint(*data.MAC)
		}

		var ipv4Id, ipv4GwId, ipv6Id, ipv6GwId *uint
		if data.IPv4 != nil {
			ipv4Id = new(uint)
			*ipv4Id = uint(*data.IPv4)
		}

		if data.IPv4Gw != nil {
			ipv4GwId = new(uint)
			*ipv4GwId = uint(*data.IPv4Gw)
		}

		if data.IPv6 != nil {
			ipv6Id = new(uint)
			*ipv6Id = uint(*data.IPv6)
		}

		if data.IPv6Gw != nil {
			ipv6GwId = new(uint)
			*ipv6GwId = uint(*data.IPv6Gw)
		}

		jail.Networks = append(jail.Networks, jailModels.Network{
			SwitchID: uint(*data.SwitchId),
			MacID:    &mac,
			IPv4ID:   ipv4Id,
			IPv4GwID: ipv4GwId,
			IPv6ID:   ipv6Id,
			IPv6GwID: ipv6GwId,
		})
	}

	if err := s.DB.Create(&jail).Error; err != nil {
		logger.L.Error().Err(err).Msg("create_jail: failed to create jail")
		return fmt.Errorf("failed_to_create_jail: %w", err)
	}

	return nil
}

func (s *Service) DeleteJail(ctId uint) error {
	if ctId == 0 {
		return fmt.Errorf("invalid_ct_id")
	}

	var jail jailModels.Jail
	if err := s.DB.Where("ct_id = ?", ctId).Preload("Networks").First(&jail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("jail_not_found")
		}
		logger.L.Error().Err(err).Msg("delete_jail: failed to find jail")
		return fmt.Errorf("failed_to_find_jail: %w", err)
	}

	if len(jail.Networks) > 0 {
		for _, network := range jail.Networks {
			if err := s.DB.Delete(&network).Error; err != nil {
				logger.L.Error().Err(err).Msg("delete_jail: failed to delete network")
				return fmt.Errorf("failed_to_delete_network: %w", err)
			}
		}
	}

	if err := s.DB.Delete(&jail).Error; err != nil {
		logger.L.Error().Err(err).Msg("delete_jail: failed to delete jail")
		return fmt.Errorf("failed_to_delete_jail: %w", err)
	}

	jailsPath, err := config.GetJailsPath()
	if err != nil {
		logger.L.Error().Err(err).Msg("delete_jail: failed to get jails path")
		return fmt.Errorf("failed_to_get_jails_path: %w", err)
	}

	jailDir := filepath.Join(jailsPath, fmt.Sprintf("%d", ctId))
	if err := os.RemoveAll(jailDir); err != nil {
		logger.L.Error().Err(err).Msg("delete_jail: failed to remove jail directory")
		return fmt.Errorf("failed_to_remove_jail_directory: %w", err)
	}

	return nil
}
