package utils

import "net"

func IsValidMetric(metric int) bool {
	return metric >= 0 && metric <= 255
}

func IsValidMTU(mtu int) bool {
	return mtu >= 68 && mtu <= 65535
}

func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func IsValidVLAN(vlan int) bool {
	return vlan >= 0 && vlan <= 4095
}

func IsValidIPv4CIDR(cidr string) bool {
	ip, _, err := net.ParseCIDR(cidr)

	if err != nil {
		return false
	}

	return ip.To4() != nil
}
