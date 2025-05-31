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

	fmt.Println("ISO UUID:", vm.ISO)
	fmt.Println("ISO Path:", isoPath)

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
	if err := s.DB.First(&vm, id).Error; err != nil {
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
