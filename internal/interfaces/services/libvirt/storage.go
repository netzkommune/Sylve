package libvirtServiceInterfaces

import (
	"encoding/xml"
)

type StoragePoolXML struct {
	XMLName xml.Name `xml:"pool"`
	Text    string   `xml:",chardata"`
	Type    string   `xml:"type,attr"`
	Name    string   `xml:"name"`
	UUID    string   `xml:"uuid"`
	Source  struct {
		Text string `xml:",chardata"`
		Name string `xml:"name"`
	} `xml:"source"`
}

type StoragePool struct {
	Name   string
	Source string
	UUID   string
}
