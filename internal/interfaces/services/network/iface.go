package networkServiceInterfaces

type NetworkServiceInterface interface {
	ParseToDB() error
	SyncToRC() error
	SetupIPv4(name string, metric int, mtu int, protocol string, address string, netmask string, aliases [][]string) error
}
