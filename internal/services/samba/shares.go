package samba

import (
	"fmt"
	"sylve/internal/db/models"
	sambaModels "sylve/internal/db/models/samba"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"
)

func (s *Service) GetShares() ([]sambaModels.SambaShare, error) {
	var shares []sambaModels.SambaShare
	if err := s.DB.Preload("ReadOnlyGroups").Preload("WriteableGroups").Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("failed_to_get_shares: %w", err)
	}
	return shares, nil
}

func (s *Service) CreateShare(
	name string,
	dataset string,
	readOnlyGroups []string,
	writeableGroups []string,
	createMask string,
	directoryMask string,
	guestOk bool,
	readOnly bool) error {
	if err := s.DB.Where("name = ?", name).First(&sambaModels.SambaShare{}).Error; err == nil {
		return fmt.Errorf("share_with_name_exists")
	}

	if len(readOnlyGroups) > 0 && readOnly {
		return fmt.Errorf("cannot_create_read_only_share_with_read_only_groups")
	}

	datasets, err := zfs.Datasets("")
	if err != nil {
		return fmt.Errorf("failed_to_fetch_datasets: %v", err)
	}

	var fDataset *zfs.Dataset

	for _, ds := range datasets {
		properties, err := ds.GetAllProperties()
		if err != nil {
			return fmt.Errorf("failed_to_get_properties_for_dataset: %v", err)
		}

		if properties["guid"] == dataset {
			fDataset = ds
			break
		}
	}

	if fDataset == nil {
		return fmt.Errorf("dataset_not_found")
	}

	if fDataset.Mountpoint == "" {
		return fmt.Errorf("dataset_not_mounted")
	}

	allGroups := utils.JoinStringSlices(readOnlyGroups, writeableGroups)

	if len(allGroups) == 0 && !guestOk {
		return fmt.Errorf("no_groups_selected_and_guests_not_allowed")
	}

	for _, group := range allGroups {
		if err := s.DB.Where("name = ?", group).First(&models.Group{}).Error; err != nil {
			return fmt.Errorf("group_not_found: %s", group)
		}
	}

	var roGroups []models.Group
	var wrGroups []models.Group

	for _, group := range readOnlyGroups {
		var g models.Group
		if err := s.DB.Where("name = ?", group).First(&g).Error; err != nil {
			return fmt.Errorf("read_only_group_not_found: %s", group)
		}
		roGroups = append(roGroups, g)
	}

	for _, group := range writeableGroups {
		var g models.Group
		if err := s.DB.Where("name = ?", group).First(&g).Error; err != nil {
			return fmt.Errorf("writeable_group_not_found: %s", group)
		}
		wrGroups = append(wrGroups, g)
	}

	share := sambaModels.SambaShare{
		Name:            name,
		Dataset:         dataset,
		ReadOnlyGroups:  roGroups,
		WriteableGroups: wrGroups,
		CreateMask:      createMask,
		DirectoryMask:   directoryMask,
		GuestOk:         guestOk,
		ReadOnly:        readOnly,
	}

	if err := s.DB.Create(&share).Error; err != nil {
		return fmt.Errorf("failed_to_create_share: %w", err)
	}

	return s.WriteConfig(true)
}

func (s *Service) DeleteShare(id uint) error {
	var share sambaModels.SambaShare
	if err := s.DB.Where("id = ?", id).First(&share).Error; err != nil {
		return fmt.Errorf("share_not_found: %w", err)
	}

	if err := s.DB.Delete(&share).Error; err != nil {
		return fmt.Errorf("failed_to_delete_share: %w", err)
	}

	return s.WriteConfig(true)
}
