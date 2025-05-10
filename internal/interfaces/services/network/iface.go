package networkServiceInterfaces

import networkModels "sylve/internal/db/models/network"

type NetworkServiceInterface interface {
	GetStandardSwitches() ([]networkModels.StandardSwitch, error)
	NewStandardSwitch(name string, mtu int, vlan int, address string, ports []string) error
	EditStandardSwitch(id int, name string, mtu int, vlan int, address string, ports []string) error
	DeleteStandardSwitch(id int) error
}
