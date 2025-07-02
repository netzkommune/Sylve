package libvirt

import (
	"fmt"
	"regexp"
	"strconv"
	vmModels "sylve/internal/db/models/vm"
)

func updateMemory(xml string, ram int) (string, error) {
	re := regexp.MustCompile(`<memory unit='[^']*'>[^<]*</memory>`)
	replacement := fmt.Sprintf("<memory unit='B'>%d</memory>", ram)
	updatedXML := re.ReplaceAllString(xml, replacement)
	return updatedXML, nil
}

func (s *Service) ModifyHardware(vmId int,
	cpuSockets int,
	cpuCores int,
	cpuThreads int,
	cpuPinning []int,
	ram int) error {

	vms, err := s.ListVMs()
	if err != nil {
		return fmt.Errorf("failed_to_get_vm_by_id: %w", err)
	}

	var vm vmModels.VM

	for _, v := range vms {
		if v.VmID == vmId {
			vm = v
			break
		}
	}

	if vm.VmID == 0 {
		return fmt.Errorf("vm_not_found: %d", vmId)
	}

	if vm.CPUCores == cpuCores &&
		vm.CPUSockets == cpuSockets &&
		vm.CPUsThreads == cpuThreads &&
		vm.RAM == ram &&
		len(vm.CPUPinning) == len(cpuPinning) {
		for i, cpu := range vm.CPUPinning {
			if i >= len(cpuPinning) || cpu != cpuPinning[i] {
				return fmt.Errorf("no_changes_detected: %d", vmId)
			}
		}
	}

	vCPUs := cpuSockets * cpuCores * cpuThreads
	if vCPUs <= 0 {
		return fmt.Errorf("invalid_cpu_configuration: sockets=%d, cores=%d, threads=%d", cpuSockets, cpuCores, cpuThreads)
	}

	if len(cpuPinning) > 0 {
		for _, v := range vms {
			if v.VmID != vmId && len(v.CPUPinning) > 0 {
				for _, pinnedCPU := range v.CPUPinning {
					for _, cpu := range cpuPinning {
						if pinnedCPU == cpu {
							return fmt.Errorf("cpu_pinning_conflict: %d", cpu)
						}
					}
				}
			}
		}

		if len(cpuPinning) > vCPUs {
			return fmt.Errorf("cpu_pinning_exceeds_vcpus: %d", vCPUs)
		}
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

	domainXML, err := s.Conn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return fmt.Errorf("failed_to_get_domain_xml_desc: %w", err)
	}

	xml := string(domainXML)
	updatedXML := xml

	if vm.RAM != ram {
		if err := s.DB.Model(&vm).Update("ram", ram).Error; err != nil {
			return fmt.Errorf("failed_to_update_vm_ram_in_db: %w", err)
		}

		updatedXML, err = updateMemory(xml, ram)
		if err != nil {
			return fmt.Errorf("failed_to_update_memory_in_xml: %w", err)
		}
	}

	if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
		return fmt.Errorf("failed_to_undefine_domain: %w", err)
	}

	if _, err := s.Conn.DomainDefineXML(updatedXML); err != nil {
		return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
	}

	// var parsed libvirtServiceInterfaces.Domain

	// domainXML, err := s.Conn.DomainGetXMLDesc(domain, 0)
	// if err != nil {
	// 	return fmt.Errorf("failed_to_get_domain_xml_desc: %w", err)
	// }

	// err = xml.Unmarshal([]byte(domainXML), &parsed)

	// if err != nil {
	// 	return fmt.Errorf("failed_to_parse_domain_xml: %w", err)
	// }

	// if vm.RAM != ram {
	// 	if err := s.DB.Model(&vm).Update("ram", ram).Error; err != nil {
	// 		return fmt.Errorf("failed_to_update_vm_ram_in_db: %w", err)
	// 	}

	// 	parsed.Memory = libvirtServiceInterfaces.Memory{
	// 		Unit: "B",
	// 		Text: strconv.Itoa(ram),
	// 	}
	// }

	// if vm.CPUCores != cpuCores {
	// 	if err := s.DB.Model(&vm).Update("cpu_cores", cpuCores).Error; err != nil {
	// 		return fmt.Errorf("failed_to_update_vm_cpu_cores_in_db: %w", err)
	// 	}

	// 	parsed.CPU.Topology.Cores = strconv.Itoa(cpuCores)
	// }

	// if vm.CPUSockets != cpuSockets {
	// 	if err := s.DB.Model(&vm).Update("cpu_sockets", cpuSockets).Error; err != nil {
	// 		return fmt.Errorf("failed_to_update_vm_cpu_sockets_in_db: %w", err)
	// 	}

	// 	parsed.CPU.Topology.Sockets = strconv.Itoa(cpuSockets)
	// }

	// if vm.CPUsThreads != cpuThreads {
	// 	if err := s.DB.Model(&vm).Update("cpu_threads", cpuThreads).Error; err != nil {
	// 		return fmt.Errorf("failed_to_update_vm_cpu_threads_in_db: %w", err)
	// 	}

	// 	parsed.CPU.Topology.Threads = strconv.Itoa(cpuThreads)
	// }

	// if parsed.VCPU != vCPUs {
	// 	parsed.VCPU = vCPUs
	// }

	// if len(cpuPinning) > 0 {
	// 	vm.CPUPinning = cpuPinning

	// 	if err := s.DB.Save(&vm).Error; err != nil {
	// 		return fmt.Errorf("failed_to_update_vm_cpu_pinning_in_db: %w", err)
	// 	}

	// 	if parsed.BhyveCommandline == nil {
	// 		parsed.BhyveCommandline = &libvirtServiceInterfaces.BhyveCommandline{}
	// 	}

	// 	cleanedArgs := make([]libvirtServiceInterfaces.BhyveArg, 0, len(parsed.BhyveCommandline.Args))
	// 	for _, arg := range parsed.BhyveCommandline.Args {
	// 		if !strings.HasPrefix(arg.Value, "-p") {
	// 			cleanedArgs = append(cleanedArgs, arg)
	// 		}
	// 	}
	// 	parsed.BhyveCommandline.Args = cleanedArgs

	// 	fmt.Println("1.2")

	// 	for i, cpu := range cpuPinning {
	// 		parsed.BhyveCommandline.Args = append(parsed.BhyveCommandline.Args, libvirtServiceInterfaces.BhyveArg{
	// 			Value: fmt.Sprintf("-p %d:%d", i, cpu),
	// 		})
	// 	}
	// }

	// newXML, err := xml.MarshalIndent(parsed, "", "  ")
	// if err != nil {
	// 	return fmt.Errorf("failed_to_marshal_updated_domain_xml: %w", err)
	// }

	// newXMLStr := string(newXML)

	// fmt.Println("3", newXMLStr)

	// // if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
	// // 	return fmt.Errorf("failed_to_undefine_domain: %w", err)
	// // }

	// // fmt.Println(newXMLStr)

	// // if _, err := s.Conn.DomainDefineXMLFlags(newXMLStr, libvirt.DomainDefineValidate); err != nil {
	// // 	return fmt.Errorf("xml_validation_failed: %w", err)
	// // }

	// if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
	// 	return fmt.Errorf("failed_to_undefine_domain: %w", err)
	// }

	// if _, err := s.Conn.DomainDefineXML(newXMLStr); err != nil {
	// 	return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
	// }

	return nil
}
