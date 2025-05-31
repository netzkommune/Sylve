// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"fmt"
	"net"
)

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

func BridgeIfName(name string) string {
	return ShortHash("syl" + name)
}

func IsPortInUse(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}
	addr := fmt.Sprintf(":%d", port)

	tcpLn, tcpErr := net.Listen("tcp", addr)
	if tcpErr != nil {
		return false
	} else {
		tcpLn.Close()
	}

	udpAddr, udpResErr := net.ResolveUDPAddr("udp", addr)
	if udpResErr != nil {
		return false
	}

	udpConn, udpErr := net.ListenUDP("udp", udpAddr)
	if udpErr != nil {
		return false
	} else {
		udpConn.Close()
	}

	return false
}
