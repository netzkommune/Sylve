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
	"strconv"
	"strings"
	vmModels "sylve/internal/db/models/vm"
	"sylve/pkg/utils"

	"github.com/beevik/etree"
)

func updateMemory(xml string, ram int) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	memory := doc.FindElement("//memory")
	if memory == nil {
		return "", fmt.Errorf("<memory> tag not found")
	}

	memory.SetText(fmt.Sprintf("%d", ram))
	memory.RemoveAttr("unit")
	memory.CreateAttr("unit", "B")

	out, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("failed to serialize XML: %w", err)
	}

	return out, nil
}

func updateCPU(xml string, cpuSockets, cpuCores, cpuThreads int, cpuPinning []int) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	vcpu := doc.FindElement("//vcpu")
	if vcpu == nil {
		return "", fmt.Errorf("<vcpu> tag not found")
	}

	vcpu.SetText(strconv.Itoa(cpuSockets * cpuCores * cpuThreads))

	cpu := doc.FindElement("//cpu")
	if cpu == nil {
		cpu = doc.CreateElement("cpu")
	}

	topology := cpu.FindElement("topology")
	if topology == nil {
		topology = cpu.CreateElement("topology")
	}

	topology.CreateAttr("sockets", strconv.Itoa(cpuSockets))
	topology.CreateAttr("cores", strconv.Itoa(cpuCores))
	topology.CreateAttr("threads", strconv.Itoa(cpuThreads))

	if len(cpuPinning) > 0 {
		bhyveCommandline := doc.FindElement("//commandline")
		if bhyveCommandline == nil || bhyveCommandline.Space != "bhyve" {
			root := doc.Root()
			if root.SelectAttr("xmlns:bhyve") == nil {
				root.CreateAttr("xmlns:bhyve", "http://libvirt.org/schemas/domain/bhyve/1.0")
			}
			bhyveCommandline = root.CreateElement("bhyve:commandline")
		}

		for _, arg := range bhyveCommandline.ChildElements() {
			valueAttr := arg.SelectAttr("value")
			if valueAttr != nil {
				value := valueAttr.Value
				if value != "" && len(value) >= 2 && value[0:2] == "-p" {
					bhyveCommandline.RemoveChild(arg)
				}
			}
		}

		pinStr := ""

		for i, cpu := range cpuPinning {
			if i > 0 {
				pinStr += " "
			}

			pinStr += fmt.Sprintf("-p %d:%d", i, cpu)
		}

		if pinStr != "" {
			arg := bhyveCommandline.CreateElement("bhyve:arg")
			arg.CreateAttr("value", pinStr)
		}
	} else {
		bhyveCommandline := doc.FindElement("//bhyve:commandline")
		if bhyveCommandline != nil {
			for _, arg := range bhyveCommandline.SelectElements("bhyve:arg") {
				if arg.Text() != "" && arg.Text()[0:2] == "-p" {
					bhyveCommandline.RemoveChild(arg)
				}
			}
		}
	}

	out, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("failed to serialize XML: %w", err)
	}

	return out, nil
}

func updateVNC(xml string, vncPort int, vncResolution string, vncPassword string, vncWait bool) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	bhyveCommandline := doc.FindElement("//bhyve:commandline")
	if bhyveCommandline == nil || bhyveCommandline.Space != "bhyve" {
		root := doc.Root()
		if root.SelectAttr("xmlns:bhyve") == nil {
			root.CreateAttr("xmlns:bhyve", "http://libvirt.org/schemas/domain/bhyve/1.0")
		}
		bhyveCommandline = root.CreateElement("bhyve:commandline")
	}

	index := 0

	for _, arg := range bhyveCommandline.ChildElements() {
		valueAttr := arg.SelectAttr("value")
		if valueAttr != nil {
			value := valueAttr.Value
			if value != "" && strings.Contains(value, "fbuf,tcp") {
				fmt.Println("Found existing VNC argument:", value)
				start := strings.Index(value, "-s")
				end := strings.Index(value, ":")
				if start != -1 && end != -1 && end > start {
					indexStr := value[start+2 : end]
					if idx, err := strconv.Atoi(indexStr); err == nil {
						index = idx
					}
				}
				bhyveCommandline.RemoveChild(arg)
			}
		}
	}

	resolutionParts := strings.Split(vncResolution, "x")
	if len(resolutionParts) != 2 {
		return "", fmt.Errorf("invalid_vnc_resolution_format: %s", vncResolution)
	}

	width, err := strconv.Atoi(resolutionParts[0])
	if err != nil {
		return "", fmt.Errorf("invalid_vnc_resolution_width: %s", resolutionParts[0])
	}

	height, err := strconv.Atoi(resolutionParts[1])
	if err != nil {
		return "", fmt.Errorf("invalid_vnc_resolution_height: %s", resolutionParts[1])
	}

	wait := ""

	if vncWait {
		wait = ",wait"
	}

	if index == 0 {
		index, err = findLowestIndex(xml)
		if err != nil {
			return "", fmt.Errorf("failed_to_find_lowest_index: %w", err)
		}
	}

	vnc := fmt.Sprintf("-s %d:0,fbuf,tcp=0.0.0.0:%d,w=%d,h=%d,password=%s%s", index, vncPort, width, height, vncPassword, wait)

	if vnc != "" {
		arg := bhyveCommandline.CreateElement("bhyve:arg")
		arg.CreateAttr("value", vnc)
	}

	out, err := doc.WriteToString()
	if err != nil {
		return "", fmt.Errorf("failed to serialize XML: %w", err)
	}

	return out, nil
}

func (s *Service) ModifyHardware(vmId int,
	cpuSockets int,
	cpuCores int,
	cpuThreads int,
	cpuPinning []int,
	ram int,
	vncPort int,
	vncResolution string,
	vncPassword string,
	vncWait bool) error {

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

	if vm.CPUCores != cpuCores ||
		vm.CPUSockets != cpuSockets ||
		vm.CPUsThreads != cpuThreads ||
		len(vm.CPUPinning) != len(cpuPinning) {
		vm.CPUCores = cpuCores
		vm.CPUSockets = cpuSockets
		vm.CPUsThreads = cpuThreads
		vm.CPUPinning = cpuPinning

		if err := s.DB.Save(&vm).Error; err != nil {
			return fmt.Errorf("failed_to_update_vm_cpu_in_db: %w", err)
		}

		updatedXML, err = updateCPU(xml, cpuSockets, cpuCores, cpuThreads, cpuPinning)

		if err != nil {
			return fmt.Errorf("failed_to_update_cpu_in_xml: %w", err)
		}
	}

	if vm.VNCPort != vncPort ||
		vm.VNCResolution != vncResolution ||
		vm.VNCPassword != vncPassword ||
		vm.VNCWait != vncWait {
		if utils.IsValidPort(vncPort) == false {
			return fmt.Errorf("invalid_vnc_port: %d", vncPort)
		}

		if utils.IsPortInUse(vncPort) {
			return fmt.Errorf("vnc_port_in_use: %d", vncPort)
		}

		vm.VNCPort = vncPort
		vm.VNCResolution = vncResolution
		vm.VNCPassword = vncPassword
		vm.VNCWait = vncWait
		if err := s.DB.Save(&vm).Error; err != nil {
			return fmt.Errorf("failed_to_update_vm_vnc_in_db: %w", err)
		}

		updatedXML, err = updateVNC(xml, vncPort, vncResolution, vncPassword, vncWait)
		if err != nil {
			return fmt.Errorf("failed_to_update_vnc_in_xml: %w", err)
		}
	}

	if err := s.Conn.DomainUndefineFlags(domain, 0); err != nil {
		return fmt.Errorf("failed_to_undefine_domain: %w", err)
	}

	if _, err := s.Conn.DomainDefineXML(updatedXML); err != nil {
		return fmt.Errorf("failed_to_define_domain_with_modified_xml: %w", err)
	}

	return nil
}

func findLowestIndex(xml string) (int, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return -1, fmt.Errorf("failed to parse XML: %w", err)
	}
	bhyveCommandline := doc.FindElement("//commandline")
	if bhyveCommandline == nil || bhyveCommandline.Space != "bhyve" {
		return 10, nil
	}

	usedIndices := make(map[int]bool)
	for _, arg := range bhyveCommandline.ChildElements() {
		valueAttr := arg.SelectAttr("value")
		if valueAttr == nil {
			continue
		}
		value := valueAttr.Value
		if len(value) >= 2 && value[0:2] == "-s" {
			parts := strings.Fields(value)
			if len(parts) >= 2 {
				indexPart := parts[1]
				colonIndex := strings.Index(indexPart, ":")
				if colonIndex > 0 {
					indexStr := indexPart[0:colonIndex] // "10"
					if index, err := strconv.Atoi(indexStr); err == nil {
						usedIndices[index] = true
					}
				}
			}
		}
	}

	for i := 10; i < 30; i++ {
		if !usedIndices[i] {
			return i, nil
		}
	}

	return -1, fmt.Errorf("all indices 10-29 are in use")
}
