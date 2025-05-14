package networkServiceInterfaces

import networkModels "sylve/internal/db/models/network"

type NetworkServiceInterface interface {
	SyncStandardSwitches(previous *networkModels.StandardSwitch, action string) error
	GetStandardSwitches() ([]networkModels.StandardSwitch, error)
	NewStandardSwitch(name string, mtu int, vlan int, address string, address6 string, ports []string, private bool) error
	EditStandardSwitch(id int, mtu int, vlan int, address string, address6 string, ports []string, private bool) error
	DeleteStandardSwitch(id int) error
}
