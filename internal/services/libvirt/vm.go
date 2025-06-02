package libvirt

import (
	"fmt"
	"sylve/internal/db/models"
	utilitiesModels "sylve/internal/db/models/utilities"
	vmModels "sylve/internal/db/models/vm"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"

	"gorm.io/gorm"

	sdb "sylve/internal/db"
)

func (s *Service) ListVMs() ([]vmModels.VM, error) {
	var vms []vmModels.VM
	if err := s.DB.Preload("Networks").Preload("Storages").Find(&vms).Error; err != nil {
		return nil, fmt.Errorf("failed_to_list_vms: %w", err)
	}
	return vms, nil
}

func validateCreate(data libvirtServiceInterfaces.CreateVMRequest, db *gorm.DB) error {
	if data.Name == "" || !utils.IsValidVMName(data.Name) {
		return fmt.Errorf("invalid_vm_name")
	}

	if data.VMID == nil || *data.VMID <= 0 || *data.VMID > 9999 {
		return fmt.Errorf("invalid_vm_id")
	}

	if data.Description != "" && (len(data.Description) < 1 || len(data.Description) > 1024) {
		return fmt.Errorf("invalid_description")
	}

	if data.StorageType == "raw" && (data.StorageSize == nil || *data.StorageSize < 1024*1024*128) {
		return fmt.Errorf("disk_size_must_be_greater_than_128mb")
	}

	if (data.StorageType == "raw" || data.StorageType == "zvol") && (data.StorageDataset == "" || len(data.StorageDataset) < 1) {
		noun := "filesystem"
		if data.StorageType == "zvol" {
			noun = "volume"
		}
		return fmt.Errorf("no_%s_selected", noun)
	}

	if data.StorageType != "" && data.StorageEmulationType == "" {
		return fmt.Errorf("no_emulation_type_selected")
	}

	if data.StorageDataset != "" {
		count, err := sdb.Count(db, &vmModels.Storage{}, "dataset = ?", data.StorageDataset)
		if err != nil {
			return fmt.Errorf("failed_to_check_storage_dataset_usage: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("storage_dataset_already_in_use")
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

			if guid == data.StorageDataset {
				dataset = d
			}
		}

		if dataset == nil {
			return fmt.Errorf("dataset_not_found")
		}

		if data.StorageType == "raw" && dataset.Type != zfs.DatasetFilesystem {
			return fmt.Errorf("invalid_dataset_type_for_raw_storage")
		}

		if data.StorageType == "zvol" && dataset.Type != zfs.DatasetVolume {
			return fmt.Errorf("invalid_dataset_type_for_zvol_storage")
		}

		if data.StorageType == "raw" {
			if dataset.Mountpoint == "" {
				return fmt.Errorf("raw_storage_dataset_must_have_mountpoint")
			}

			if dataset.Avail > *data.StorageSize {
				return fmt.Errorf("not_enough_space_in_raw_storage_dataset")
			}
		}
	}

	if data.StorageEmulationType != "" {
		if data.StorageEmulationType != "virtio-blk" && data.StorageEmulationType != "ahci-hd" && data.StorageEmulationType != "nvme" {
			return fmt.Errorf("invalid_storage_emulation_type")
		}
	}

	if data.SwitchID != nil && *data.SwitchID != 0 {
		if data.NetworkMAC != "" && !utils.IsValidMACAddress(data.NetworkMAC) {
			return fmt.Errorf("invalid_mac_address")
		}

		if data.NetworkMAC != "" {
			count, err := sdb.Count(db, &vmModels.Network{}, "mac = ?", data.NetworkMAC)
			if err != nil {
				return fmt.Errorf("failed_to_check_network_mac_usage: %w", err)
			}
			if count > 0 {
				return fmt.Errorf("network_mac_already_in_use")
			}
		}

		if data.SwitchEmulationType == "" {
			return fmt.Errorf("no_switch_emulation_type_selected")
		}
	}

	if data.CPUSockets < 1 {
		return fmt.Errorf("sockets_must_be_greater_than_0")
	}

	if data.CPUCores < 1 {
		return fmt.Errorf("cores_must_be_greater_than_0")
	}

	if data.CPUThreads < 1 {
		return fmt.Errorf("threads_must_be_greater_than_0")
	}

	if data.RAM < 1024*1024*128 {
		return fmt.Errorf("memory_must_be_greater_than_128mb")
	}

	if data.VNCPort < 1 || data.VNCPort > 65535 {
		return fmt.Errorf("vnc_port_must_be_between_1_and_65535")
	} else {
		count, err := sdb.Count(db, &vmModels.VM{}, "vnc_port = ?", data.VNCPort)
		if err != nil {
			return fmt.Errorf("failed_to_check_vnc_port_usage: %w", err)
		}

		if count > 0 {
			return fmt.Errorf("vnc_port_already_in_use")
		} else {
			if utils.IsPortInUse(data.VNCPort) {
				return fmt.Errorf("vnc_port_already_in_use_by_another_service")
			}
		}
	}

	if data.VNCPassword != "" && len(data.VNCPassword) < 1 {
		return fmt.Errorf("vnc_password_required")
	}

	if data.VNCResolution == "" {
		return fmt.Errorf("no_vnc_resolution_selected")
	}

	if data.StartOrder < 0 {
		return fmt.Errorf("start_order_must_be_greater_than_or_equal_to_0")
	}

	if len(data.PCIDevices) > 0 {
		for _, pciID := range data.PCIDevices {
			count, err := sdb.Count(db, &models.PassedThroughIDs{}, "id = ?", pciID)
			if err != nil {
				return fmt.Errorf("passthrough_device_does_not_exist: %w", err)
			}
			if count == 0 {
				return fmt.Errorf("pci_device_not_found: %d", pciID)
			}

			// var vm vmModels.VM
			// if err := db.Preload("PCIDevices").First(&vm, "id = ?", pciID).Error; err != nil {
			// 	if err == gorm.ErrRecordNotFound {
			// 		return fmt.Errorf("pci_device_not_found: %d", pciID)
			// 	}
			// 	return fmt.Errorf("failed_to_check_pci_device_usage: %w", err)
			// }

			// for _, device := range vm.PCIDevices {
			// 	if device == pciID {
			// 		return fmt.Errorf("pci_device_already_in_use: %d", pciID)
			// 	}
			// }
		}
	}

	if data.ISO != "" && data.ISO != "none" {
		count, err := sdb.Count(db, &utilitiesModels.Downloads{}, "uuid = ?", data.ISO)

		if err != nil {
			return fmt.Errorf("failed_to_check_iso_usage: %w", err)
		}

		if count == 0 {
			return fmt.Errorf("iso_not_found: %s", data.ISO)
		}
	}

	return nil
}

func (s *Service) CreateVM(data libvirtServiceInterfaces.CreateVMRequest) error {
	if err := validateCreate(data, s.DB); err != nil {
		return err
	}

	vncWait := false
	startAtBoot := false

	if data.VNCWait != nil {
		vncWait = *data.VNCWait
	} else {
		vncWait = true
	}

	if data.StartAtBoot == nil {
		startAtBoot = true
	} else {
		startAtBoot = *data.StartAtBoot
	}

	var networks []vmModels.Network
	if data.SwitchID != nil && *data.SwitchID != 0 {
		networks = append(networks, vmModels.Network{
			MAC:       data.NetworkMAC,
			SwitchID:  uint(*data.SwitchID),
			Emulation: data.SwitchEmulationType,
		})
	}

	var storages []vmModels.Storage
	if data.StorageType != "" {
		if data.StorageSize != nil && data.StorageType == "zvol" {
			*data.StorageSize = 0
		}

		storages = append(storages, vmModels.Storage{
			Type:      data.StorageType,
			Dataset:   data.StorageDataset,
			Size:      int64(*data.StorageSize),
			Emulation: data.StorageEmulationType,
		})
	}

	vm := &vmModels.VM{
		Name:          data.Name,
		VmID:          *data.VMID,
		Description:   data.Description,
		CPUSockets:    data.CPUSockets,
		CPUCores:      data.CPUCores,
		CPUsThreads:   data.CPUThreads,
		RAM:           data.RAM,
		VNCPort:       data.VNCPort,
		VNCPassword:   data.VNCPassword,
		VNCResolution: data.VNCResolution,
		VNCWait:       vncWait,
		StartAtBoot:   startAtBoot,
		StartOrder:    data.StartOrder,
		PCIDevices:    data.PCIDevices,
		ISO:           data.ISO,
		Storages:      storages,
		Networks:      networks,
	}

	if err := s.DB.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Create(vm).Error; err != nil {
		return fmt.Errorf("failed_to_create_vm_with_associations: %w", err)
	}

	if err := s.CreateLvVm(int(vm.ID)); err != nil {
		if err := s.DB.Delete(vm).Error; err != nil {
			return fmt.Errorf("failed_to_delete_vm_after_creation_failure: %w", err)
		}

		for _, storage := range storages {
			if err := s.DB.Delete(&storage).Error; err != nil {
				return fmt.Errorf("failed_to_delete_storage_after_vm_creation_failure: %w", err)
			}
		}

		for _, network := range networks {
			if err := s.DB.Delete(&network).Error; err != nil {
				return fmt.Errorf("failed_to_delete_network_after_vm_creation_failure: %w", err)
			}
		}

		return fmt.Errorf("failed_to_create_lv_vm: %w", err)
	}

	return nil
}

func (s *Service) RemoveVM(id uint) error {
	var vm vmModels.VM
	if err := s.DB.Preload("Networks").Preload("Storages").First(&vm, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("vm_not_found: %d", id)
		}
		return fmt.Errorf("failed_to_find_vm: %w", err)
	}

	err := s.RemoveLvVm(int(vm.VmID))
	if err != nil {
		return fmt.Errorf("failed_to_remove_lv_vm: %w", err)
	}

	for _, network := range vm.Networks {
		if err := s.DB.Delete(&network).Error; err != nil {
			return fmt.Errorf("failed_to_delete_network: %w", err)
		}
	}

	for _, storage := range vm.Storages {
		if err := s.DB.Delete(&storage).Error; err != nil {
			return fmt.Errorf("failed_to_delete_storage: %w", err)
		}
	}

	if err := s.DB.Delete(&vm).Error; err != nil {
		return fmt.Errorf("failed_to_delete_vm: %w", err)
	}

	return nil
}

func (s *Service) PerformAction(id uint, action string) error {
	var vm vmModels.VM

	if err := s.DB.First(&vm, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("vm_not_found: %d", id)
		}
		return fmt.Errorf("failed_to_find_vm: %w", err)
	}

	err := s.LvVMAction(vm, action)
	if err != nil {
		return fmt.Errorf("failed_to_perform_action: %w", err)
	}

	return nil
}
