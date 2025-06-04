package libvirt

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	vmModels "sylve/internal/db/models/vm"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"
)

func (s *Service) CreateDiskImage(vmId int, guid string, size int64) error {
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

	imagePath := filepath.Join(vmPath, fmt.Sprintf("%d.img", vmId))
	if _, err := os.Stat(imagePath); !os.IsNotExist(err) {
		return fmt.Errorf("image_already_exists: %s", imagePath)
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

	if storage.Detached {
		return fmt.Errorf("storage_already_detached: %d", storageId)
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

	var parsed libvirtServiceInterfaces.Domain

	domainXML, err := s.Conn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return fmt.Errorf("failed_to_get_domain_xml_desc: %w", err)
	}

	err = xml.Unmarshal([]byte(domainXML), &parsed)

	if err != nil {
		return fmt.Errorf("failed_to_parse_domain_xml: %w", err)
	}

	if storage.Type == "iso" {
		filePath, err := s.FindISOByUUID(storage.Dataset)
		if err != nil {
			return fmt.Errorf("failed_to_find_iso_by_uuid: %w", err)
		}

		pattern := fmt.Sprintf(`\s*-s\s+\d+:0,ahci-cd,%s\s*`, regexp.QuoteMeta(filePath))
		re := regexp.MustCompile(pattern)

		var filteredLines []string
		for _, line := range strings.Split(domainXML, "\n") {
			if re.MatchString(line) {
				continue
			}
			filteredLines = append(filteredLines, line)
		}

		newXML := strings.Join(filteredLines, "\n")

		if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
			return fmt.Errorf("failed_to_undefine_domain: %w", err)
		}

		if _, err := s.Conn.DomainDefineXML(newXML); err != nil {
			return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
		}

		storage.Detached = true

		if err := s.DB.Save(&storage).Error; err != nil {
			return fmt.Errorf("failed_to_save_storage: %w", err)
		}
	}

	return nil
}
