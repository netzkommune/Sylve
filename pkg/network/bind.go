package network

import (
	"fmt"
	"net"
)

func TryBindToPort(ip string, port int, proto string) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen(proto, addr)
	if err != nil {
		return err
	}

	defer listener.Close()
	return nil
}
