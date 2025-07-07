// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirt

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	vmModels "sylve/internal/db/models/vm"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"

	"github.com/beevik/etree"
)

func (s *Service) CreateDiskImage(vmId int, guid string, size int64, name string) error {
	dataset, err := zfs.Datasets("")
	if err != nil {
		return fmt.Errorf("failed_to_get_datasets: %w", err)
	}

	var targetDataset *zfs.Dataset

	for _, d := range dataset {
		guidProp, err := d.GetProperty("guid")

		if err != nil {
			return fmt.Errorf("failed_to_get_dataset_properties: %w", err)
		}

		if guidProp == guid {
			targetDataset = d
			break
		}
	}

	if targetDataset == nil {
		return fmt.Errorf("dataset_not_found: %s", guid)
	}

	if targetDataset.Type != "filesystem" {
		return fmt.Errorf("invalid_dataset_type: %s", targetDataset.Type)
	}

	mountpoint, err := targetDataset.GetProperty("mountpoint")
	if err != nil {
		return fmt.Errorf("failed_to_get_mountpoint_property: %w", err)
	}

	if mountpoint == "" {
		return fmt.Errorf("mountpoint_property_is_empty_for_dataset: %s", guid)
	}

	vmPath := filepath.Join(mountpoint, "sylve-vm-images", strconv.Itoa(vmId))
	if _, err := os.Stat(vmPath); os.IsNotExist(err) {
		if err := os.MkdirAll(vmPath, 0755); err != nil {
			return fmt.Errorf("failed_to_create_vm_images_directory: %w", err)
		}
	}

	var imagePath string

	if name == "" {
		imagePath = filepath.Join(vmPath, fmt.Sprintf("%d.img", vmId))
	} else {
		imagePath = filepath.Join(vmPath, fmt.Sprintf("%s.img", name))
	}

	if _, err := os.Stat(imagePath); !os.IsNotExist(err) {
		if err := os.Remove(imagePath); err != nil {
			return fmt.Errorf("failed_to_remove_existing_image: %w", err)
		}
	}

	if err := utils.CreateOrTruncateFile(imagePath, size); err != nil {
		return fmt.Errorf("failed_to_create_or_truncate_image_file: %w", err)
	}

	return nil
}

func (s *Service) StorageDetach(vmId int, storageId int) error {
	var storage vmModels.Storage

	err := s.DB.Find(&storage, "id = ?", storageId).Error
	if err != nil {
		return fmt.Errorf("failed_to_find_storage: %w", err)
	}

	domain, err := s.Conn.DomainLookupByName(strconv.Itoa(vmId))
	if err != nil {
		return fmt.Errorf("failed_to_lookup_domain_by_name: %w", err)
	}

	state, _, err := s.Conn.DomainGetState(domain, 0)

	if err != nil {
		return fmt.Errorf("failed_to_get_domain_state: %w", err)
	}

	if state != 5 {
		return fmt.Errorf("domain_state_not_shutoff: %d", vmId)
	}

	xml, err := s.Conn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return fmt.Errorf("failed_to_get_domain_xml_desc: %w", err)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	bhyveCommandline := doc.FindElement("//commandline")
	if bhyveCommandline == nil || bhyveCommandline.Space != "bhyve" {
		root := doc.Root()
		if root.SelectAttr("xmlns:bhyve") == nil {
			root.CreateAttr("xmlns:bhyve", "http://libvirt.org/schemas/domain/bhyve/1.0")
		}
		bhyveCommandline = root.CreateElement("bhyve:commandline")
	}

	filePath := ""

	if storage.Type == "iso" {
		filePath, err = s.FindISOByUUID(storage.Dataset, false)
		if err != nil {
			return fmt.Errorf("failed_to_find_iso_by_uuid: %w", err)
		}
	}

	for _, arg := range bhyveCommandline.ChildElements() {
		valueAttr := arg.SelectAttr("value")
		if valueAttr != nil {
			value := valueAttr.Value
			if value != "" {
				/* Takes care of CD removals */
				if strings.Contains(value, "ahci-cd") &&
					strings.Contains(value, filePath) &&
					storage.Type == "iso" {
					bhyveCommandline.RemoveChild(arg)
				}

				var dataset *zfs.Dataset

				if storage.Type == "zvol" || storage.Type == "raw" {
					datasets, err := zfs.Datasets("")
					if err != nil {
						return fmt.Errorf("failed_to_get_datasets: %w", err)
					}

					for _, d := range datasets {
						guid, err := d.GetProperty("guid")
						if err != nil {
							return fmt.Errorf("failed_to_get_dataset_property: %w", err)
						}

						if guid == storage.Dataset {
							dataset = d
							break
						}
					}

					if dataset == nil {
						return fmt.Errorf("dataset_not_found: %s", storage.Dataset)
					}
				}

				/* Takes care of ZVOL removals */
				if storage.Type == "zvol" {
					if dataset.Type != "volume" {
						return fmt.Errorf("invalid_dataset_type: %s", dataset.Type)
					}

					if strings.Contains(value, fmt.Sprintf("/dev/zvol/%s", dataset.Name)) {
						bhyveCommandline.RemoveChild(arg)
					}
				}

				/* Takes care of RAW Disk removals */
				if storage.Type == "raw" {
					if strings.Contains(value, dataset.Name) &&
						strings.Contains(value, storage.Name) {
						bhyveCommandline.RemoveChild(arg)
					}

					imagePath := filepath.Join(dataset.Mountpoint, "sylve-vm-images", strconv.Itoa(vmId), fmt.Sprintf("%s.img", storage.Name))
					if _, err := os.Stat(imagePath); !os.IsNotExist(err) {
						if err := os.Remove(imagePath); err != nil {
							return fmt.Errorf("failed_to_remove_disk_image: %w", err)
						}
					}
				}
			}
		}
	}

	out, err := doc.WriteToString()
	if err != nil {
		return fmt.Errorf("failed to serialize XML: %w", err)
	}

	if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
		return fmt.Errorf("failed_to_undefine_domain: %w", err)
	}

	if _, err := s.Conn.DomainDefineXML(out); err != nil {
		return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
	}

	if err := s.DB.Delete(&storage).Error; err != nil {
		return fmt.Errorf("failed_to_delete_storage: %w", err)
	}

	return nil
}

func (s *Service) StorageAttach(vmId int, sType string, dataset string, emulation string, size int64, name string) error {
	domain, err := s.Conn.DomainLookupByName(strconv.Itoa(vmId))
	if err != nil {
		return fmt.Errorf("failed_to_lookup_domain_by_name: %w", err)
	}

	state, _, err := s.Conn.DomainGetState(domain, 0)
	if err != nil {
		return fmt.Errorf("failed_to_get_domain_state: %w", err)
	}

	if state != 5 {
		return fmt.Errorf("domain_state_not_shutoff: %d", vmId)
	}

	if sType != "zvol" && sType != "raw" && sType != "iso" {
		return fmt.Errorf("invalid_storage_type: %s", sType)
	}

	if emulation == "" {
		return fmt.Errorf("emulation_type_required: %s", sType)
	}

	if emulation != "virtio-blk" && emulation != "ahci-cd" && emulation != "ahci-hd" && emulation != "nvme" {
		return fmt.Errorf("invalid_emulation_type: %s", emulation)
	}

	var vm vmModels.VM

	if err := s.DB.Preload("Storages").Where("vm_id = ?", vmId).First(&vm).Error; err != nil {
		return fmt.Errorf("failed_to_find_vm: %w", err)
	}

	if vm.ID == 0 {
		return fmt.Errorf("vm_not_found: %d", vmId)
	}

	xml, err := s.Conn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return fmt.Errorf("failed_to_get_domain_xml_desc: %w", err)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	bhyveCommandline := doc.FindElement("//commandline")
	if bhyveCommandline == nil || bhyveCommandline.Space != "bhyve" {
		root := doc.Root()
		if root.SelectAttr("xmlns:bhyve") == nil {
			root.CreateAttr("xmlns:bhyve", "http://libvirt.org/schemas/domain/bhyve/1.0")
		}
		bhyveCommandline = root.CreateElement("bhyve:commandline")
	}

	if sType == "iso" {
		filePath, err := s.FindISOByUUID(dataset, false)
		if err != nil {
			return fmt.Errorf("failed_to_find_iso_by_uuid: %w", err)
		}

		if filePath == "" {
			return fmt.Errorf("iso_file_not_found: %s", dataset)
		}

		for _, storage := range vm.Storages {
			if storage.Type == "iso" && storage.Dataset == dataset {
				return fmt.Errorf("iso_already_attached: %s", dataset)
			}
		}

		var existingStorage vmModels.Storage
		err = s.DB.First(&existingStorage, "dataset = ? AND type = ?", dataset, sType).Error
		if err != nil {
			if err.Error() != "record not found" {
				return fmt.Errorf("failed_to_find_existing_storage: %w", err)
			}
		}

		newStorage := vmModels.Storage{
			Type:      sType,
			Dataset:   dataset,
			Size:      0,
			Emulation: "ahci-cd",
			VMID:      uint(vm.ID),
		}

		if err := s.DB.Create(&newStorage).Error; err != nil {
			return fmt.Errorf("failed_to_create_storage: %w", err)
		}

		index, err := findLowestIndex(xml)
		if err != nil {
			return fmt.Errorf("failed_to_find_lowest_index: %w", err)
		}

		argValue := fmt.Sprintf("-s %d:0,ahci-cd,%s", index, filePath)
		bhyveCommandline.CreateElement("bhyve:arg").CreateAttr("value", argValue)
	} else if sType == "zvol" {
		datasets, err := zfs.Datasets("")
		if err != nil {
			return fmt.Errorf("failed_to_get_datasets: %w", err)
		}

		var targetDataset *zfs.Dataset
		for _, d := range datasets {
			guid, err := d.GetProperty("guid")
			if err != nil {
				return fmt.Errorf("failed_to_get_dataset_property: %w", err)
			}

			if guid == dataset {
				targetDataset = d
				break
			}
		}

		if targetDataset == nil {
			return fmt.Errorf("dataset_not_found: %s", dataset)
		}

		if targetDataset.Type != "volume" {
			return fmt.Errorf("invalid_dataset_type: %s", targetDataset.Type)
		}

		for _, storage := range vm.Storages {
			if storage.Type == "zvol" && storage.Dataset == dataset {
				return fmt.Errorf("zvol_already_attached: %s", dataset)
			}
		}

		var existingStorage vmModels.Storage
		err = s.DB.First(&existingStorage, "dataset = ? AND type = ?", dataset, sType).Error
		if err != nil {
			if err.Error() != "record not found" {
				return fmt.Errorf("failed_to_find_existing_storage: %w", err)
			}
		}

		newStorage := vmModels.Storage{
			Type:      sType,
			Dataset:   dataset,
			Size:      size,
			Emulation: emulation,
			VMID:      uint(vm.ID),
		}

		if err := s.DB.Create(&newStorage).Error; err != nil {
			return fmt.Errorf("failed_to_create_storage: %w", err)
		}

		index, err := findLowestIndex(xml)
		if err != nil {
			return fmt.Errorf("failed_to_find_lowest_index: %w", err)
		}

		argValue := fmt.Sprintf("-s %d:0,%s,%s", index, emulation, fmt.Sprintf("/dev/zvol/%s", targetDataset.Name))
		bhyveCommandline.CreateElement("bhyve:arg").CreateAttr("value", argValue)
	} else if sType == "raw" {
		var existingStorage vmModels.Storage
		err = s.DB.First(&existingStorage, "dataset = ? AND type = ? AND name = ? AND vm_id = ?", dataset, sType, name, vmId).Error
		if err != nil {
			if err.Error() != "record not found" {
				return fmt.Errorf("failed_to_find_existing_storage: %w", err)
			}
		}

		if existingStorage.ID != 0 {
			return fmt.Errorf("raw_storage_already_attached: %s", dataset)
		}

		if name == "" {
			return fmt.Errorf("name_required_for_raw_storage")
		}

		datasets, err := zfs.Datasets("")
		if err != nil {
			return fmt.Errorf("failed_to_get_datasets: %w", err)
		}

		var targetDataset *zfs.Dataset
		for _, d := range datasets {
			guid, err := d.GetProperty("guid")
			if err != nil {
				return fmt.Errorf("failed_to_get_dataset_property: %w", err)
			}

			if guid == dataset {
				targetDataset = d
				break
			}
		}

		if targetDataset == nil {
			return fmt.Errorf("dataset_not_found: %s", dataset)
		}

		newStorage := vmModels.Storage{
			Type:      sType,
			Dataset:   dataset,
			Size:      size,
			Emulation: emulation,
			Name:      name,
			VMID:      uint(vm.ID),
		}

		if err := s.DB.Create(&newStorage).Error; err != nil {
			return fmt.Errorf("failed_to_create_storage: %w", err)
		}

		if err := s.CreateDiskImage(vmId, dataset, size, name); err != nil {
			return fmt.Errorf("failed_to_create_disk_image: %w", err)
		}

		index, err := findLowestIndex(xml)
		if err != nil {
			return fmt.Errorf("failed_to_find_lowest_index: %w", err)
		}

		argValue := fmt.Sprintf("-s %d:0,%s,%s/%s.img", index, emulation, filepath.Join(targetDataset.Mountpoint, "sylve-vm-images", strconv.Itoa(vmId)), name)
		bhyveCommandline.CreateElement("bhyve:arg").CreateAttr("value", argValue)
	}

	out, err := doc.WriteToString()
	if err != nil {
		return fmt.Errorf("failed to serialize XML: %w", err)
	}

	if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
		return fmt.Errorf("failed_to_undefine_domain: %w", err)
	}

	if _, err := s.Conn.DomainDefineXML(out); err != nil {
		return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
	}

	return nil
}
