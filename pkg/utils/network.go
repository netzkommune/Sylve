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
