package libvirtServiceInterfaces

import "encoding/xml"

type CreateVMRequest struct {
	Name                 string  `json:"name" binding:"required"`
	VMID                 *int    `json:"vmId" binding:"required"`
	Description          string  `json:"description"`
	ISO                  string  `json:"iso"`
	StorageType          string  `json:"storageType" binding:"required"`
	StorageDataset       string  `json:"storageDataset" binding:"required"`
	StorageSize          *uint64 `json:"storageSize" binding:"required"`
	StorageEmulationType string  `json:"storageEmulationType"`
	SwitchID             *int    `json:"switchId" binding:"required"`
	SwitchEmulationType  string  `json:"switchEmulationType"`
	NetworkMAC           string  `json:"macAddress"`
	CPUSockets           int     `json:"cpuSockets" binding:"required"`
	CPUCores             int     `json:"cpuCores" binding:"required"`
	CPUThreads           int     `json:"cpuThreads" binding:"required"`
	RAM                  int     `json:"ram" binding:"required"`
	PCIDevices           []int   `json:"pciDevices"`
	VNCPort              int     `json:"vncPort" binding:"required"`
	VNCPassword          string  `json:"vncPassword"`
	VNCResolution        string  `json:"vncResolution"`
	VNCWait              *bool   `json:"vncWait"`
	StartAtBoot          *bool   `json:"startAtBoot" binding:"required"`
	StartOrder           int     `json:"startOrder"`
}

type Memory struct {
	Unit string `xml:"unit,attr"`
	Text string `xml:",chardata"`
}

type MemoryBacking struct {
	Locked struct{} `xml:"locked"`
}

type Topology struct {
	Sockets string `xml:"sockets,attr"`
	Cores   string `xml:"cores,attr"`
	Threads string `xml:"threads,attr"`
}

type CPU struct {
	Topology Topology `xml:"topology"`
}

type Loader struct {
	ReadOnly string `xml:"readonly,attr"`
	Type     string `xml:"type,attr"`
	Path     string `xml:",chardata"`
}

type OS struct {
	Type   string `xml:"type"`
	Loader Loader `xml:"loader"`
}

type Features struct {
	APIC struct{} `xml:"apic"`
	ACPI struct{} `xml:"acpi"`
}

type Clock struct {
	Offset string `xml:"offset,attr"`
}

type Driver struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Target struct {
	Dev string `xml:"dev,attr"`
	Bus string `xml:"bus,attr"`
}

type Source struct {
	File string `xml:"file,attr"`
}

type Volume struct {
	Pool   string `xml:"pool,attr"`
	Volume string `xml:"volume,attr"`
}

type Disk struct {
	Type     string    `xml:"type,attr"`
	Device   string    `xml:"device,attr"`
	Driver   *Driver   `xml:"driver,omitempty"`
	Source   any       `xml:"source"`
	Target   Target    `xml:"target"`
	ReadOnly *struct{} `xml:"readonly,omitempty"`
}

type MACAddress struct {
	Address string `xml:"address,attr"`
}

type BridgeSource struct {
	Bridge string `xml:"bridge,attr"`
}

type Model struct {
	Type string `xml:"type,attr"`
}

type Interface struct {
	Type   string       `xml:"type,attr"`
	MAC    *MACAddress  `xml:"mac,omitempty"`
	Source BridgeSource `xml:"source"`
	Model  Model        `xml:"model"`
}

type Devices struct {
	Disks      []Disk      `xml:"disk,omitempty"`
	Interfaces []Interface `xml:"interface,omitempty"`
}

type BhyveArg struct {
	Value string `xml:"value,attr"`
}

type BhyveCommandline struct {
	Args []BhyveArg `xml:"bhyve:arg"`
}

type Domain struct {
	XMLName       xml.Name       `xml:"domain"`
	Type          string         `xml:"type,attr"`
	XMLNSBhyve    string         `xml:"xmlns:bhyve,attr"`
	Name          string         `xml:"name"`
	Memory        Memory         `xml:"memory"`
	MemoryBacking *MemoryBacking `xml:"memoryBacking,omitempty"`
	CPU           CPU            `xml:"cpu"`
	VCPU          int            `xml:"vcpu"`
	OS            OS             `xml:"os"`
	Features      Features       `xml:"features"`
	Clock         Clock          `xml:"clock"`

	OnPoweroff string `xml:"on_poweroff,omitempty"`
	OnReboot   string `xml:"on_reboot,omitempty"`
	OnCrash    string `xml:"on_crash,omitempty"`

	Devices Devices `xml:"devices"`

	BhyveCommandline *BhyveCommandline `xml:"bhyve:commandline,omitempty"`
}
