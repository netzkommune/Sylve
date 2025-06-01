package libvirt

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sylve/internal/config"
	vmModels "sylve/internal/db/models/vm"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"

	"github.com/google/uuid"
)

func (s *Service) CreateVmXML(vm vmModels.VM, vmPath string) (string, error) {
	var memoryBacking *libvirtServiceInterfaces.MemoryBacking

	if vm.PCIDevices != nil && len(vm.PCIDevices) > 0 {
		memoryBacking = &libvirtServiceInterfaces.MemoryBacking{
			Locked: struct{}{},
		}
	}

	var devices libvirtServiceInterfaces.Devices
	var isoPath string

	if vm.ISO != "" {
		var err error

		isoPath, err = s.FindISOByUUID(vm.ISO)
		if err != nil {
			return "", fmt.Errorf("failed_to_find_iso: %w", err)
		}
	}

	if isoPath != "" {
		devices.Disks = append(devices.Disks, libvirtServiceInterfaces.Disk{
			Type:   "file",
			Device: "cdrom",
			Driver: &libvirtServiceInterfaces.Driver{
				Name: "file",
				Type: "raw",
			},
			Source: libvirtServiceInterfaces.Source{
				File: isoPath,
			},
			Target: libvirtServiceInterfaces.Target{
				Dev: "hdc",
				Bus: "sata",
			},
			ReadOnly: &struct{}{},
		})
	}

	if vm.Storages != nil && len(vm.Storages) > 0 {
		for _, storage := range vm.Storages {
			if storage.Dataset != "" && storage.Type == "zvol" {
				fmt.Println("Using ZFS dataset for storage:", storage.Dataset)

				datasets, err := zfs.Datasets("")
				if err != nil {
					return "", fmt.Errorf("failed_to_get_datasets: %w", err)
				}

				var dataset *zfs.Dataset

				for _, d := range datasets {
					properties, err := d.GetAllProperties()
					if err != nil {
						return "", fmt.Errorf("failed_to_get_dataset_properties: %w", err)
					}

					if guid, exists := properties["guid"]; exists && guid == storage.Dataset {
						dataset = d
						break
					}
				}

				if dataset == nil {
					return "", fmt.Errorf("dataset_not_found: %s", storage.Dataset)
				}

				pool := strings.SplitN(dataset.Name, "/", 2)[0]
				volume := dataset.Name

				if idx := strings.LastIndex(volume, "/"); idx != -1 {
					volume = volume[idx+1:]
				}

				devices.Disks = append(devices.Disks, libvirtServiceInterfaces.Disk{
					Type:   "volume",
					Device: "disk",
					Source: libvirtServiceInterfaces.Volume{
						Pool:   pool,
						Volume: volume,
					},
					Target: libvirtServiceInterfaces.Target{
						Dev: "vdb",
						Bus: storage.Emulation,
					},
				})
			}
		}
	}

	uefi := fmt.Sprintf("%s,%s/%d_vars.fd", "/usr/local/share/uefi-firmware/BHYVE_UEFI.fd", vmPath, vm.VmID)

	domain := libvirtServiceInterfaces.Domain{
		Type:       "bhyve",
		XMLNSBhyve: "http://libvirt.org/schemas/domain/bhyve/1.0",
		Name:       strconv.Itoa(vm.VmID),
		Memory: libvirtServiceInterfaces.Memory{
			Unit: "B",
			Text: strconv.Itoa(vm.RAM),
		},
		MemoryBacking: memoryBacking,
		CPU: libvirtServiceInterfaces.CPU{
			Topology: libvirtServiceInterfaces.Topology{
				Sockets: strconv.Itoa(vm.CPUSockets),
				Cores:   strconv.Itoa(vm.CPUCores),
				Threads: strconv.Itoa(vm.CPUsThreads),
			},
		},
		VCPU: (vm.CPUSockets * vm.CPUCores * vm.CPUsThreads),
		OS: libvirtServiceInterfaces.OS{
			Type: "hvm",
			Loader: libvirtServiceInterfaces.Loader{
				ReadOnly: "yes",
				Type:     "pflash",
				Path:     uefi,
			},
		},
		Features: libvirtServiceInterfaces.Features{
			ACPI: struct{}{},
			APIC: struct{}{},
		},
		Clock: libvirtServiceInterfaces.Clock{
			Offset: "utc",
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "destroy",
		Devices:    devices,
	}

	width, height, f := strings.Cut(vm.VNCResolution, "x")
	if f != true {
		return "", fmt.Errorf("invalid_vnc_resolution")
	}

	vncWait := ""

	if vm.VNCWait {
		vncWait = ",wait"
	}

	vncArg := fmt.Sprintf("-s 20:0,fbuf,tcp=0.0.0.0:%d,w=%s,h=%s,password=%s%s",
		vm.VNCPort,
		width,
		height,
		vm.VNCPassword,
		vncWait,
	)

	domain.BhyveCommandline = &libvirtServiceInterfaces.BhyveCommandline{
		Args: []libvirtServiceInterfaces.BhyveArg{
			{Value: vncArg},
		},
	}

	out, err := xml.Marshal(domain)
	if err != nil {
		return "", fmt.Errorf("failed_to_marshal_vm_xml: %w", err)
	}

	return string(out), nil
}

func (s *Service) CreateLvVm(id int) error {
	var vm vmModels.VM
	if err := s.DB.Preload("Storages").Preload("Networks").First(&vm, id).Error; err != nil {
		return fmt.Errorf("failed_to_find_vm: %w", err)
	}

	vmDir, err := config.GetVMsPath()

	if err != nil {
		return fmt.Errorf("failed to get VMs path: %w", err)
	}

	vmPath := fmt.Sprintf("%s/%d", vmDir, vm.VmID)

	if _, err := os.Stat(vmPath); err == nil {
		if err := os.RemoveAll(vmPath); err != nil {
			return fmt.Errorf("failed to clear VM directory: %w", err)
		}
	}

	if err := os.MkdirAll(vmPath, 0755); err != nil {
		return fmt.Errorf("failed to create VM directory: %w", err)
	}

	uefiVarsBase := "/usr/local/share/uefi-firmware/BHYVE_UEFI_VARS.fd"
	uefiVarsPath := filepath.Join(vmPath, fmt.Sprintf("%d_vars.fd", vm.VmID))

	err = utils.CopyFile(uefiVarsBase, uefiVarsPath)

	if err != nil {
		return fmt.Errorf("failed to copy UEFI vars file: %w", err)
	}

	generated, err := s.CreateVmXML(vm, vmPath)
	if err != nil {
		return fmt.Errorf("failed to generate VM XML: %w", err)
	}

	fmt.Println("Generated VM XML:", generated)

	_, err = s.Conn.DomainDefineXML(generated)

	if err != nil {
		return fmt.Errorf("failed to define VM domain: %w", err)
	}

	return nil
}

func (s *Service) RemoveLvVm(vmId int) error {
	domain, err := s.Conn.DomainLookupByName(strconv.Itoa(vmId))
	if err != nil {
		return fmt.Errorf("failed_to_lookup_domain: %w", err)
	}

	if err := s.Conn.DomainDestroy(domain); err != nil {
		return fmt.Errorf("failed_to_destroy_domain: %w", err)
	}

	if err := s.Conn.DomainUndefine(domain); err != nil {
		return fmt.Errorf("failed_to_undefine_domain: %w", err)
	}

	vmDir, err := config.GetVMsPath()
	if err != nil {
		return fmt.Errorf("failed to get VMs path: %w", err)
	}

	vmPath := filepath.Join(vmDir, strconv.Itoa(vmId))
	if _, err := os.Stat(vmPath); err == nil {
		if err := os.RemoveAll(vmPath); err != nil {
			return fmt.Errorf("failed to remove VM directory: %w", err)
		}
	}

	return nil
}

func (s *Service) GetLvDomain(vmId int) (*libvirtServiceInterfaces.LvDomain, error) {
	var dom libvirtServiceInterfaces.LvDomain

	domain, err := s.Conn.DomainLookupByName(strconv.Itoa(vmId))
	if err != nil {
		return nil, fmt.Errorf("failed_to_lookup_domain: %w", err)
	}

	stateMap := map[int32]string{
		0: "No State",
		1: "Running",
		2: "Blocked",
		3: "Paused",
		4: "Shutdown",
		5: "Shutoff",
		6: "Crashed",
		7: "PMSuspended",
	}

	state, _, err := s.Conn.DomainGetState(domain, 0)
	if err != nil {
		return nil, fmt.Errorf("failed_to_get_domain_state: %w", err)
	}

	dom.ID = domain.ID
	dom.UUID = uuid.UUID(domain.UUID).String()
	dom.Name = domain.Name
	dom.Status = stateMap[state]

	return &dom, nil
}
